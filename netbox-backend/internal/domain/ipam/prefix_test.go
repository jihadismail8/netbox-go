package ipam_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var prefixDomainTime = shared.NewTimestamp(time.Date(2026, time.July, 22, 14, 0, 0, 0, time.UTC))

func TestPrefixRejectsHostBitsWithCanonicalSuggestionAndSlashZero(t *testing.T) {
	t.Parallel()

	_, err := domainipam.NewPrefix(validPrefixValues("192.0.2.7/24"), prefixDomainTime)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	assert.Equal(t, "prefix", violations[0].Field)
	assert.Equal(t, "192.0.2.7/24 is not a valid prefix. Did you mean 192.0.2.0/24?", violations[0].Description)

	for _, value := range []string{"0.0.0.0/0", "::/0"} {
		_, err = domainipam.NewPrefix(validPrefixValues(value), prefixDomainTime)
		require.Error(t, err)
		assert.Equal(t, "Cannot create prefix with /0 mask.", shared.ViolationsOf(err)[0].Description)
	}
}

func TestPrefixAcceptsBareHostsCanonicalNetworksAndDefaults(t *testing.T) {
	t.Parallel()

	prefix, err := domainipam.NewPrefix(validPrefixValues(" 2001:db8::1 "), prefixDomainTime)
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::1/128", prefix.Network().String())
	assert.Equal(t, uint32(6), prefix.Family())
	assert.Equal(t, domainipam.PrefixStatusActive, prefix.Status())
	assert.False(t, prefix.IsPool())
	assert.False(t, prefix.MarkUtilized())
	assert.True(t, prefix.VRF().IsNull())
}

func TestPrefixValidatesChoiceReferenceAndDescription(t *testing.T) {
	t.Parallel()

	values := validPrefixValues("192.0.2.0/24")
	values.Status = "unknown"
	values.Description = strings.Repeat("x", domainipam.PrefixDescriptionMaxLength+1)
	values.VRF = domainipam.NonNullVRFReference(domainipam.VRFReference{})
	_, err := domainipam.NewPrefix(values, prefixDomainTime)
	require.Error(t, err)
	assert.Len(t, shared.ViolationsOf(err), 3)
}

func TestPrefixPatchPreservesNullableVRFAndComputedProjection(t *testing.T) {
	t.Parallel()

	reference := prefixVRFReference(t, 9, "Tenant", "65000:9", true)
	prefix, err := domainipam.RestorePrefix(domainipam.PrefixState{
		ID: 3, Prefix: "10.0.0.0/8", VRF: domainipam.NonNullVRFReference(reference),
		Status:  domainipam.PrefixStatusContainer.String(),
		Created: prefixDomainTime, LastUpdated: prefixDomainTime, Children: 7, Depth: 2,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), prefix.Children())
	assert.Equal(t, uint32(2), prefix.Depth())

	nullVRF := domainipam.NullVRFReference()
	updated := shared.NewTimestamp(prefixDomainTime.Add(time.Hour))
	require.NoError(t, prefix.ApplyPatch(domainipam.PrefixPatch{
		VRF: &nullVRF, Description: stringPointer(" moved "),
	}, updated))
	assert.True(t, prefix.VRF().IsNull())
	assert.Equal(t, "moved", prefix.Description())
	assert.Equal(t, updated, prefix.LastUpdated())
	assert.Equal(t, shared.ID(9), reference.ID())
}

func TestPrefixNetworkContainmentAndOrdering(t *testing.T) {
	t.Parallel()

	outer := requiredPrefixNetwork(t, "10.0.0.0/8")
	inner := requiredPrefixNetwork(t, "10.1.0.0/16")
	equal := requiredPrefixNetwork(t, "10.0.0.0/8")
	v6 := requiredPrefixNetwork(t, "2001:db8::/32")
	assert.True(t, outer.Contains(inner, false))
	assert.False(t, outer.Contains(equal, false))
	assert.True(t, outer.Contains(equal, true))
	assert.False(t, outer.Contains(v6, true))
	assert.Less(t, outer.Compare(inner), 0)
	assert.Less(t, inner.Compare(v6), 0)

	filtered, explicit, err := domainipam.ParsePrefixFilter("10.1.2.3/16")
	require.NoError(t, err)
	assert.True(t, explicit)
	assert.Equal(t, "10.1.0.0/16", filtered.String())
}

func validPrefixValues(value string) domainipam.PrefixValues {
	return domainipam.PrefixValues{
		Prefix: value, VRF: domainipam.NullVRFReference(), Status: domainipam.PrefixStatusActive.String(),
	}
}

func prefixVRFReference(
	t *testing.T,
	id shared.ID,
	name string,
	rdValue string,
	enforceUnique bool,
) domainipam.VRFReference {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	reference, err := domainipam.NewVRFReference(
		id, name, domainipam.NonNullRouteDistinguisher(rd), enforceUnique,
	)
	require.NoError(t, err)
	return reference
}

func requiredPrefixNetwork(t *testing.T, value string) domainipam.PrefixNetwork {
	t.Helper()
	network, err := domainipam.ParsePrefixNetwork(value)
	require.NoError(t, err)
	return network
}

func stringPointer(value string) *string { return &value }
