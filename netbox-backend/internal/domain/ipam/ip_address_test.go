package ipam_test

import (
	"strings"
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
		Address: "192.0.2.17/24", VRF: domainipam.NullVRFReference(),
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

func TestIPAddressScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		input  string
		want   string
		family uint32
	}{
		{name: "maskless IPv4", input: "192.0.2.17", want: "192.0.2.17/32", family: 4},
		{name: "maskless IPv6", input: "2001:db8::17", want: "2001:db8::17/128", family: 6},
		{name: "IPv4 dotted netmask", input: "192.0.2.17/255.255.255.0", want: "192.0.2.17/24", family: 4},
		{name: "IPv4 inverse hostmask", input: "192.0.2.17/0.0.0.255", want: "192.0.2.17/24", family: 4},
		{name: "zero-padded prefix", input: "192.0.2.17/024", want: "192.0.2.17/24", family: 4},
		{name: "signed spaced prefix", input: "192.0.2.17/ +024 ", want: "192.0.2.17/24", family: 4},
		{name: "Unicode decimal prefix", input: "192.0.2.17/ ٢_٤ ", want: "192.0.2.17/24", family: 4},
		{name: "astral Unicode decimal prefix", input: "192.0.2.17/\U0001D7F8_\U0001D7FA", want: "192.0.2.17/24", family: 4},
		{name: "IPv6 netmask", input: "2001:db8::17/ffff:ffff:ffff:ffff::", want: "2001:db8::17/64", family: 6},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			parsed, err := domainipam.ParseHostAddress(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.want, parsed.String())
			assert.Equal(t, test.family, parsed.Family())
		})
	}

	_, err := domainipam.ParseHostAddress("")
	require.Error(t, err)
	require.Len(t, shared.ViolationsOf(err), 1)
	assert.Equal(t, shared.FieldViolation{
		Field: "address", Reason: "blank",
		Description: "This field may not be blank.",
	}, shared.ViolationsOf(err)[0])

	_, err = domainipam.ParseHostAddress("fe80::1%eth0")
	require.Error(t, err)
	require.Len(t, shared.ViolationsOf(err), 1)
	assert.Equal(t, "address", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "invalid", shared.ViolationsOf(err)[0].Reason)

	for _, value := range []string{" 192.0.2.17/24", "192.0.2.17 ", "192.0.2.17 /24"} {
		_, err = domainipam.ParseHostAddress(value)
		require.Error(t, err)
		require.Len(t, shared.ViolationsOf(err), 1)
		assert.Equal(t, "Invalid IP address format: "+value, shared.ViolationsOf(err)[0].Description)
	}

	status, valid := domainipam.ParseIPAddressStatus(" active ")
	require.True(t, valid, "non-write parsing must retain its existing normalization")
	assert.Equal(t, domainipam.IPAddressStatusActive, status)
	role, valid := domainipam.ParseIPAddressRole(" loopback ")
	require.True(t, valid, "non-write parsing must retain its existing normalization")
	assert.Equal(t, domainipam.IPAddressRoleLoopback, role)

	_, err = domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
		Status: " active ",
	}, ipAddressDomainTime)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	assert.Equal(t, "status", violations[0].Field)
	assert.Equal(t, "invalid_choice", violations[0].Reason)
	assert.Equal(t, " active  is not a valid choice.", violations[0].Description)

	_, err = domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
		Status: domainipam.IPAddressStatusActive.String(),
		Role:   domainipam.NonNullIPAddressRole(" loopback "),
	}, ipAddressDomainTime)
	require.Error(t, err)
	violations = shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	assert.Equal(t, "role", violations[0].Field)
	assert.Equal(t, "invalid_choice", violations[0].Reason)
	assert.Equal(t, " loopback  is not a valid choice.", violations[0].Description)

	for _, test := range []struct {
		name        string
		status      string
		role        domainipam.NullableIPAddressRole
		field       string
		description string
	}{
		{name: "boolean-like status", status: "true", field: "status", description: "True is not a valid choice."},
		{name: "integer-like role", status: "active", role: domainipam.NonNullIPAddressRole("001"), field: "role", description: "1 is not a valid choice."},
		{name: "Unicode integer-like role", status: "active", role: domainipam.NonNullIPAddressRole("٠٠١"), field: "role", description: "1 is not a valid choice."},
		{name: "signed mixed Unicode integer-like role", status: "active", role: domainipam.NonNullIPAddressRole("+२_４"), field: "role", description: "24 is not a valid choice."},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
				Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
				Status: test.status, Role: test.role,
			}, ipAddressDomainTime)
			require.Error(t, err)
			violations := shared.ViolationsOf(err)
			require.Len(t, violations, 1)
			assert.Equal(t, test.field, violations[0].Field)
			assert.Equal(t, test.description, violations[0].Description)
		})
	}

	for _, test := range []struct {
		field string
		apply func(*domainipam.IPAddressValues)
	}{
		{field: "dns_name", apply: func(values *domainipam.IPAddressValues) { values.DNSName = "contains\x00null" }},
		{field: "description", apply: func(values *domainipam.IPAddressValues) { values.Description = "contains\x00null" }},
		{field: "comments", apply: func(values *domainipam.IPAddressValues) { values.Comments = "contains\x00null" }},
	} {
		values := domainipam.IPAddressValues{
			Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
			Status: domainipam.IPAddressStatusActive.String(),
		}
		test.apply(&values)
		_, err := domainipam.NewIPAddress(values, ipAddressDomainTime)
		require.Error(t, err)
		violations := shared.ViolationsOf(err)
		if test.field == "dns_name" {
			require.Len(t, violations, 2)
			assert.Equal(t, "Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names", violations[0].Description)
		} else {
			require.Len(t, violations, 1)
		}
		assert.Equal(t, test.field, violations[len(violations)-1].Field)
		assert.Equal(t, "Null characters are not allowed.", violations[len(violations)-1].Description)
	}

	_, err = domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
		Status:  domainipam.IPAddressStatusActive.String(),
		DNSName: strings.Repeat("a", 254) + "!\x00",
	}, ipAddressDomainTime)
	require.Error(t, err)
	violations = shared.ViolationsOf(err)
	require.Len(t, violations, 3)
	assert.Equal(t, []string{
		"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
		"Ensure this field has no more than 255 characters.",
		"Null characters are not allowed.",
	}, []string{
		violations[0].Description,
		violations[1].Description,
		violations[2].Description,
	})

	address, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: "192.0.2.18", VRF: domainipam.NullVRFReference(),
		Status:  domainipam.IPAddressStatusActive.String(),
		Role:    domainipam.NonNullIPAddressRole(domainipam.IPAddressRoleLoopback),
		DNSName: " EDGE_02.Example. ", Description: " description ",
		Comments: " comments ",
	}, ipAddressDomainTime)
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.18/32", address.Display())
	assert.Equal(t, "edge_02.example.", address.DNSName())
	assert.Equal(t, "description", address.Description())
	assert.Equal(t, "comments", address.Comments())
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
	violations := shared.ViolationsOf(err)
	assert.Len(t, violations, 2)
	assert.Equal(
		t,
		"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
		violations[1].Description,
	)
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
