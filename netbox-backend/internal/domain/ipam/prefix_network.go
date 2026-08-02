package ipam

import (
	"fmt"
	"net/netip"
	"strings"

	"netbox-go/internal/domain/shared"
)

// PrefixNetwork is a canonical IPv4 or IPv6 network. Unlike filter values,
// persisted Prefix values reject host bits and /0 networks.
type PrefixNetwork struct {
	value netip.Prefix
}

func ParsePrefixNetwork(value string) (PrefixNetwork, error) {
	value = strings.TrimSpace(value)
	parsed, err := parseNetworkOrHost(value)
	if err != nil {
		return PrefixNetwork{}, shared.NewValidationError(shared.FieldViolation{
			Field: "prefix", Reason: "invalid",
			Description: "Enter a valid IPv4 or IPv6 prefix.",
		})
	}
	if parsed.Bits() == 0 {
		return PrefixNetwork{}, shared.NewValidationError(shared.FieldViolation{
			Field: "prefix", Reason: "invalid",
			Description: "Cannot create prefix with /0 mask.",
		})
	}
	canonical := parsed.Masked()
	if parsed.Addr() != canonical.Addr() {
		return PrefixNetwork{}, shared.NewValidationError(shared.FieldViolation{
			Field: "prefix", Reason: "invalid",
			Description: fmt.Sprintf(
				"%s is not a valid prefix. Did you mean %s?", parsed, canonical,
			),
		})
	}
	return PrefixNetwork{value: canonical}, nil
}

// ParsePrefixFilter mirrors netaddr.IPNetwork for list filters: bare hosts are
// accepted at maximum mask length and masked inputs are canonicalized.
func ParsePrefixFilter(value string) (PrefixNetwork, bool, error) {
	value = strings.TrimSpace(value)
	explicitMask := strings.Contains(value, "/")
	parsed, err := parseNetworkOrHost(value)
	if err != nil {
		return PrefixNetwork{}, explicitMask, err
	}
	return PrefixNetwork{value: parsed.Masked()}, explicitMask, nil
}

func parseNetworkOrHost(value string) (netip.Prefix, error) {
	if strings.Contains(value, "/") {
		return netip.ParsePrefix(value)
	}
	address, err := netip.ParseAddr(value)
	if err != nil {
		return netip.Prefix{}, err
	}
	bits := 128
	if address.Is4() {
		bits = 32
	}
	return netip.PrefixFrom(address, bits), nil
}

func (network PrefixNetwork) String() string { return network.value.String() }

func (network PrefixNetwork) Family() uint32 {
	if network.value.Addr().Is4() {
		return 4
	}
	if network.value.IsValid() {
		return 6
	}
	return 0
}

func (network PrefixNetwork) Bits() int { return network.value.Bits() }

func (network PrefixNetwork) Valid() bool {
	return network.value.IsValid() && network.value.Bits() > 0 && network.value == network.value.Masked()
}

func (network PrefixNetwork) Contains(other PrefixNetwork, includeEqual bool) bool {
	if !network.Valid() || !other.Valid() || network.Family() != other.Family() {
		return false
	}
	if !network.value.Contains(other.value.Addr()) || network.Bits() > other.Bits() {
		return false
	}
	return includeEqual || network.Bits() < other.Bits()
}

func (network PrefixNetwork) Compare(other PrefixNetwork) int {
	if network.Family() < other.Family() {
		return -1
	}
	if network.Family() > other.Family() {
		return 1
	}
	if compared := network.value.Addr().Compare(other.value.Addr()); compared != 0 {
		return compared
	}
	if network.Bits() < other.Bits() {
		return -1
	}
	if network.Bits() > other.Bits() {
		return 1
	}
	return 0
}
