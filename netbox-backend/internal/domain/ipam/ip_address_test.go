package ipam_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaindcim "netbox-go/internal/domain/dcim"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var ipAddressDomainTime = shared.NewTimestamp(
	time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
)

func TestIPAddressPreservesHostMaskAndNormalizesDNS(t *testing.T) {
	t.Parallel()

	address, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: " 192.0.2.17/24 ", VRF: domainipam.NullVRFReference(),
		Status:  domainipam.IPAddressStatusActive.String(),
		DNSName: " EDGE_01.Example. ",
	}, ipAddressDomainTime)
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.17/24", address.Address().String())
	assert.Equal(t, uint32(4), address.Family())
	assert.Equal(t, "edge_01.example.", address.DNSName())
	assert.True(t, address.Role().IsNull())

	ipv6, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "2001:db8::17/64", VRF: domainipam.NullVRFReference(),
		Status: domainipam.IPAddressStatusSLAAC.String(),
	}, ipAddressDomainTime)
	require.NoError(t, err)
	assert.Equal(t, "2001:db8::17/64", ipv6.Display())
	assert.Equal(t, uint32(6), ipv6.Family())
}

func TestIPAddressRejectsSlashZeroSLAACIPv4AndInvalidDNS(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"0.0.0.1/0", "2001:db8::1/0"} {
		_, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
			Address: value, VRF: domainipam.NullVRFReference(),
			Status: domainipam.IPAddressStatusActive.String(),
		}, ipAddressDomainTime)
		require.Error(t, err)
		assert.Equal(
			t, "Cannot create IP address with /0 mask.",
			shared.ViolationsOf(err)[0].Description,
		)
	}

	_, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "192.0.2.1/24", VRF: domainipam.NullVRFReference(),
		Status: domainipam.IPAddressStatusSLAAC.String(), DNSName: "bad name",
	}, ipAddressDomainTime)
	require.Error(t, err)
	assert.Len(t, shared.ViolationsOf(err), 2)
}

func TestIPAddressAssignmentRejectsOrdinaryNetworkAndBroadcastEdges(t *testing.T) {
	t.Parallel()

	assignment := domainipam.NonNullInterfaceAssignment(
		requiredIPAddressAssignment(t),
	)
	tests := []struct {
		address string
		valid   bool
	}{
		{address: "192.0.2.0/24"},
		{address: "192.0.2.255/24"},
		{address: "192.0.2.0/31", valid: true},
		{address: "192.0.2.1/32", valid: true},
		{address: "2001:db8::/64"},
		{address: "2001:db8::/127", valid: true},
		{address: "2001:db8::1/128", valid: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.address, func(t *testing.T) {
			t.Parallel()
			_, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
				Address: test.address, VRF: domainipam.NullVRFReference(),
				Status:     domainipam.IPAddressStatusActive.String(),
				Assignment: assignment,
			}, ipAddressDomainTime)
			if test.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			violations := shared.ViolationsOf(err)
			require.Len(t, violations, 1)
			assert.Equal(t, "non_field_errors", violations[0].Field)
		})
	}
}

func TestIPAddressExceptionalDuplicateRolesAreExplicit(t *testing.T) {
	t.Parallel()

	for _, role := range []domainipam.IPAddressRole{
		domainipam.IPAddressRoleAnycast,
		domainipam.IPAddressRoleVIP,
		domainipam.IPAddressRoleVRRP,
		domainipam.IPAddressRoleHSRP,
		domainipam.IPAddressRoleGLBP,
		domainipam.IPAddressRoleCARP,
	} {
		assert.True(t, role.AllowsDuplicateHost(), role.String())
	}
	assert.False(t, domainipam.IPAddressRoleLoopback.AllowsDuplicateHost())
	assert.False(t, domainipam.IPAddressRoleSecondary.AllowsDuplicateHost())
	assert.False(t, domainipam.IPAddressRole("").AllowsDuplicateHost())
}

func requiredIPAddressAssignment(
	t *testing.T,
) domainipam.InterfaceAssignment {
	t.Helper()
	name := domaindcim.NonNullDeviceValue("edge-01")
	device, err := domaindcim.NewDeviceReference(7, name, "edge-01")
	require.NoError(t, err)
	networkInterface, err := domaindcim.RestoreInterface(
		domaindcim.InterfaceState{
			ID: 11, Device: device, Name: "Ethernet1", Type: "1000base-t",
			Enabled: true, Created: ipAddressDomainTime,
			LastUpdated: ipAddressDomainTime,
		},
	)
	require.NoError(t, err)
	assignment, err := domainipam.NewInterfaceAssignment(networkInterface)
	require.NoError(t, err)
	return assignment
}
