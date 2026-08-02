package ipam

import (
	"net/netip"
	"strings"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

const (
	DefaultIPAddressPageLimit uint32 = 50
	MaximumIPAddressPageLimit uint32 = 1000
)

type ListIPAddressesQuery struct {
	Limit        uint32
	LimitPresent bool
	Offset       uint32
	Query        string
	IDs          []int64
	Ordering     []string
	VRFIDs       []int64
	VRFRDs       []string
	Addresses    []string
	Family       *int64
	Parent       *string
	Statuses     []string
	Assigned     *bool
	InterfaceIDs []int64
	DeviceIDs    []int64
}

func (query ListIPAddressesQuery) EffectiveLimit() uint32 {
	if query.Limit > 0 {
		return query.Limit
	}
	if query.LimitPresent {
		return MaximumIPAddressPageLimit
	}
	return DefaultIPAddressPageLimit
}

type GetIPAddressQuery struct{ ID shared.ID }

type IPAddressSortField string

const (
	IPAddressSortID          IPAddressSortField = "id"
	IPAddressSortVRF         IPAddressSortField = "vrf"
	IPAddressSortAddress     IPAddressSortField = "address"
	IPAddressSortStatus      IPAddressSortField = "status"
	IPAddressSortDNSName     IPAddressSortField = "dns_name"
	IPAddressSortCreated     IPAddressSortField = "created"
	IPAddressSortLastUpdated IPAddressSortField = "last_updated"
)

type IPAddressSort struct {
	Field      IPAddressSortField
	Descending bool
}

type IPAddressFilter struct {
	Address      domainipam.HostAddress
	ExplicitMask bool
	Valid        bool
}

type IPAddressParentFilter struct {
	Network netip.Prefix
	Valid   bool
}

func validateListIPAddressesQuery(
	query ListIPAddressesQuery,
) (IPAddressListCriteria, error) {
	var violations []shared.FieldViolation
	limit := query.EffectiveLimit()
	if limit > MaximumIPAddressPageLimit {
		violations = append(violations, shared.FieldViolation{
			Field: "limit", Reason: "max_value",
			Description: "Ensure this value is less than or equal to 1000.",
		})
	}
	ordering, orderingViolations := parseIPAddressOrdering(query.Ordering)
	violations = append(violations, orderingViolations...)
	statuses := make([]domainipam.IPAddressStatus, 0, len(query.Statuses))
	for _, raw := range query.Statuses {
		status, valid := domainipam.ParseIPAddressStatus(raw)
		if !valid {
			violations = append(violations, shared.FieldViolation{
				Field: "status", Reason: "invalid_choice", Description: "Select a valid choice.",
			})
			continue
		}
		statuses = append(statuses, status)
	}
	addresses := make([]IPAddressFilter, 0, len(query.Addresses))
	for _, raw := range query.Addresses {
		addresses = append(addresses, parseIPAddressFilter(raw))
	}
	if len(violations) > 0 {
		return IPAddressListCriteria{}, shared.NewValidationError(violations...)
	}
	return IPAddressListCriteria{
		Limit: limit, Offset: query.Offset, Query: strings.TrimSpace(query.Query),
		IDs: append([]int64(nil), query.IDs...), Ordering: ordering,
		VRFIDs: append([]int64(nil), query.VRFIDs...), VRFRDs: trimmedStrings(query.VRFRDs),
		Addresses: addresses, AddressesPresent: len(query.Addresses) > 0,
		Family: cloneInt64(query.Family), Parent: parseIPAddressParentFilter(query.Parent),
		Statuses: statuses, Assigned: cloneBool(query.Assigned),
		InterfaceIDs: append([]int64(nil), query.InterfaceIDs...),
		DeviceIDs:    append([]int64(nil), query.DeviceIDs...),
	}, nil
}

func parseIPAddressFilter(raw string) IPAddressFilter {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return IPAddressFilter{}
	}
	if host, err := netip.ParseAddr(raw); err == nil {
		bits := 128
		if host.Is4() {
			bits = 32
		}
		address, parseErr := domainipam.ParseHostAddress(
			netip.PrefixFrom(host, bits).String(),
		)
		return IPAddressFilter{
			Address: address, ExplicitMask: false, Valid: parseErr == nil,
		}
	}
	address, err := domainipam.ParseHostAddress(raw)
	return IPAddressFilter{Address: address, ExplicitMask: true, Valid: err == nil}
}

func parseIPAddressParentFilter(value *string) *IPAddressParentFilter {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	raw := strings.TrimSpace(*value)
	if host, err := netip.ParseAddr(raw); err == nil {
		bits := 128
		if host.Is4() {
			bits = 32
		}
		return &IPAddressParentFilter{
			Network: netip.PrefixFrom(host, bits), Valid: true,
		}
	}
	network, err := netip.ParsePrefix(raw)
	if err != nil || network.Bits() == 0 {
		return &IPAddressParentFilter{}
	}
	return &IPAddressParentFilter{Network: network.Masked(), Valid: true}
}

func parseIPAddressOrdering(
	values []string,
) ([]IPAddressSort, []shared.FieldViolation) {
	if len(values) == 0 {
		return []IPAddressSort{
			{Field: IPAddressSortVRF},
			{Field: IPAddressSortAddress},
		}, nil
	}
	var ordering []IPAddressSort
	var violations []shared.FieldViolation
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			descending := strings.HasPrefix(item, "-")
			field, valid := parseIPAddressSortField(strings.TrimPrefix(item, "-"))
			if !valid {
				violations = append(violations, shared.FieldViolation{
					Field: "ordering", Reason: "invalid_choice",
					Description: "Select a valid ordering field.",
				})
				continue
			}
			ordering = append(ordering, IPAddressSort{
				Field: field, Descending: descending,
			})
		}
	}
	return ordering, violations
}

func parseIPAddressSortField(value string) (IPAddressSortField, bool) {
	field := IPAddressSortField(value)
	switch field {
	case IPAddressSortID, IPAddressSortVRF, IPAddressSortAddress,
		IPAddressSortStatus, IPAddressSortDNSName, IPAddressSortCreated,
		IPAddressSortLastUpdated:
		return field, true
	default:
		return "", false
	}
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
