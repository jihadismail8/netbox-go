package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestTypedIPAddressListAndUpdatePreservePresence(t *testing.T) {
	zero := uint32(0)
	family := uint32(4)
	assigned := false
	interfaceID, deviceID := int64(-7), int64(-8)
	parent := "192.0.2.99/24"
	query := typedIPAddressListQuery(&ipamv1.ListIPAddressesRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Id: []int64{-1, 0, 17},
			Ordering: []string{"vrf", "-address"},
		},
		Family: &family, Assigned: &assigned, InterfaceId: &interfaceID,
		DeviceId: &deviceID, Parent: &parent,
	})
	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumIPAddressPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 0, 17}, query.IDs)
	assert.Equal(t, int64(4), *query.Family)
	assert.False(t, *query.Assigned)
	assert.Equal(t, []int64{-7}, query.InterfaceIDs)
	assert.Equal(t, []int64{-8}, query.DeviceIDs)
	assert.Equal(t, parent, *query.Parent)

	command, err := typedIPAddressUpdateCommand(
		17,
		&ipamv1.IPAddressInput{},
		&fieldmaskpb.FieldMask{Paths: []string{
			"role", "assigned_object_type", "assigned_object_id",
		}},
	)
	require.NoError(t, err)
	assert.Equal(t, applicationipam.FieldNull, command.Role.State())
	assert.Equal(t, applicationipam.FieldNull, command.AssignedObjectType.State())
	assert.Equal(t, applicationipam.FieldNull, command.AssignedObjectID.State())
	assert.Equal(t, applicationipam.FieldOmitted, command.Address.State())

	command, err = typedIPAddressUpdateCommand(
		17,
		&ipamv1.IPAddressInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"address"}},
	)
	require.NoError(t, err)
	assert.Equal(t, applicationipam.FieldNull, command.Address.State())
}

func TestIPAddressGRPCScalarPresenceMatrix(t *testing.T) {
	type scalarCase struct {
		name   string
		input  *ipamv1.IPAddressInput
		mask   *fieldmaskpb.FieldMask
		state  applicationipam.FieldState
		values map[string]string
	}
	withoutNull := []scalarCase{
		{
			name:  "omitted",
			input: &ipamv1.IPAddressInput{},
			state: applicationipam.FieldOmitted,
		},
		{
			name:  "blank",
			input: grpcIPAddressScalarInput("", "", "", "", "", ""),
			state: applicationipam.FieldPresent,
			values: map[string]string{
				"address": "", "status": "", "role": "", "dns_name": "",
				"description": "", "comments": "",
			},
		},
		{
			name: "concrete raw values",
			input: grpcIPAddressScalarInput(
				"192.0.2.17", " reserved ", " loopback ", " EDGE.EXAMPLE ",
				" description ", " comments ",
			),
			state: applicationipam.FieldPresent,
			values: map[string]string{
				"address": "192.0.2.17", "status": " reserved ",
				"role": " loopback ", "dns_name": " EDGE.EXAMPLE ",
				"description": " description ", "comments": " comments ",
			},
		},
		{
			name: "baseline-sensitive raw values",
			input: grpcIPAddressScalarInput(
				"192.0.2.17/255.255.255.0", "true", "001", "123",
				"contains\x00null", "contains\x00null",
			),
			state: applicationipam.FieldPresent,
			values: map[string]string{
				"address": "192.0.2.17/255.255.255.0", "status": "true",
				"role": "001", "dns_name": "123",
				"description": "contains\x00null", "comments": "contains\x00null",
			},
		},
	}

	for _, test := range withoutNull {
		t.Run("create/"+test.name, func(t *testing.T) {
			command := typedIPAddressCreateCommand(test.input)
			assertIPAddressGRPCScalarFields(
				t, grpcIPAddressCreateScalarFields(command), test.state, test.values,
			)
		})
		t.Run("replace/"+test.name, func(t *testing.T) {
			command := typedIPAddressReplaceCommand(17, test.input)
			assertIPAddressGRPCScalarFields(
				t,
				grpcIPAddressCreateScalarFields(command.CreateIPAddressCommand),
				test.state,
				test.values,
			)
		})
	}

	updateCases := append(withoutNull, scalarCase{
		name:  "explicit null",
		input: &ipamv1.IPAddressInput{},
		mask: &fieldmaskpb.FieldMask{Paths: []string{
			"address", "status", "role", "dns_name", "description", "comments",
		}},
		state: applicationipam.FieldNull,
	})
	for _, test := range updateCases {
		t.Run("update/"+test.name, func(t *testing.T) {
			command, err := typedIPAddressUpdateCommand(17, test.input, test.mask)
			require.NoError(t, err)
			assertIPAddressGRPCScalarFields(
				t, grpcIPAddressUpdateScalarFields(command), test.state, test.values,
			)
		})
	}

	t.Run("response role nullability", func(t *testing.T) {
		assert.Nil(t, typedIPAddressRole(domainipam.NullIPAddressRole()))
		assert.Nil(t, typedIPAddressRole(domainipam.NonNullIPAddressRole("")))
		role := typedIPAddressRole(
			domainipam.NonNullIPAddressRole(domainipam.IPAddressRoleLoopback),
		)
		require.NotNil(t, role)
		assert.Equal(t, domainipam.IPAddressRoleLoopback.String(), role.Value)
	})
}

func grpcIPAddressCreateScalarFields(
	command applicationipam.CreateIPAddressCommand,
) map[string]applicationipam.Field[string] {
	return map[string]applicationipam.Field[string]{
		"address": command.Address, "status": command.Status,
		"role": command.Role, "dns_name": command.DNSName,
		"description": command.Description, "comments": command.Comments,
	}
}

func grpcIPAddressUpdateScalarFields(
	command applicationipam.UpdateIPAddressCommand,
) map[string]applicationipam.Field[string] {
	return map[string]applicationipam.Field[string]{
		"address": command.Address, "status": command.Status,
		"role": command.Role, "dns_name": command.DNSName,
		"description": command.Description, "comments": command.Comments,
	}
}

func grpcIPAddressScalarInput(
	address string,
	status string,
	role string,
	dnsName string,
	description string,
	comments string,
) *ipamv1.IPAddressInput {
	return &ipamv1.IPAddressInput{
		Address: &address, Status: &status, Role: wrapperspb.String(role),
		DnsName: &dnsName, Description: &description, Comments: &comments,
	}
}

func assertIPAddressGRPCScalarFields(
	t *testing.T,
	fields map[string]applicationipam.Field[string],
	wantState applicationipam.FieldState,
	wantValues map[string]string,
) {
	t.Helper()
	for name, field := range fields {
		assert.Equal(t, wantState, field.State(), name)
		if wantState != applicationipam.FieldPresent {
			continue
		}
		value, present := field.Get()
		require.True(t, present, name)
		assert.Equal(t, wantValues[name], value, name)
	}
}

func TestIPAddressRPCAndIPAMServerUseOnlyTypedService(t *testing.T) {
	address := grpcIPAddressFixture(t)
	typed := &grpcIPAddressServiceSpy{address: address}
	vrfs := &grpcVRFServiceSpy{}
	prefixes := &grpcPrefixServiceSpy{}
	server := NewIPAMServer(vrfs, prefixes, typed)
	ctx := identity.WithPrincipal(
		t.Context(),
		identity.Principal{ID: 1, Username: "typed-address"},
	)

	list, err := server.ListIPAddresses(
		ctx, &ipamv1.ListIPAddressesRequest{},
	)
	require.NoError(t, err)
	require.Len(t, list.Results, 1)
	assert.Equal(t, int64(17), list.Results[0].Id)
	require.NotNil(t, list.Results[0].AssignedObject)
	assert.Equal(t, int64(11), list.Results[0].AssignedObject.Id)

	assigned, err := server.AssignIPAddress(
		ctx, &ipamv1.AssignIPAddressRequest{Id: 17, InterfaceId: 11},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(17), assigned.IpAddress.Id)
	assert.Equal(t, shared.ID(11), typed.assignCommand.InterfaceID)

	unassigned, err := server.UnassignIPAddress(
		ctx, &ipamv1.UnassignIPAddressRequest{Id: 17},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(17), unassigned.IpAddress.Id)
	assert.Equal(t, shared.ID(17), typed.unassignCommand.ID)
	assert.Equal(t, 3, typed.calls)
	assert.Zero(t, vrfs.listCalls)
	assert.Zero(t, prefixes.listCalls)
}

func grpcIPAddressFixture(t *testing.T) *domainipam.IPAddress {
	t.Helper()
	stamp := shared.NewTimestamp(
		time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
	)
	device, err := domaindcim.NewDeviceReference(
		7, domaindcim.NonNullDeviceValue("edge-01"), "edge-01",
	)
	require.NoError(t, err)
	networkInterface, err := domaindcim.RestoreInterface(
		domaindcim.InterfaceState{
			ID: 11, Device: device, Name: "Ethernet1", Label: "uplink",
			Type: "1000base-t", Enabled: true, Created: stamp,
			LastUpdated: stamp,
		},
	)
	require.NoError(t, err)
	value, err := domainipam.NewInterfaceAssignment(networkInterface)
	require.NoError(t, err)
	address, err := domainipam.RestoreIPAddress(domainipam.IPAddressState{
		ID: 17, Address: "192.0.2.1/24",
		VRF:    domainipam.NullVRFReference(),
		Status: domainipam.IPAddressStatusActive.String(),
		Role: domainipam.NonNullIPAddressRole(
			domainipam.IPAddressRoleLoopback,
		),
		Assignment: domainipam.NonNullInterfaceAssignment(value),
		Created:    stamp, LastUpdated: stamp,
	})
	require.NoError(t, err)
	return address
}

type grpcIPAddressServiceSpy struct {
	calls           int
	address         *domainipam.IPAddress
	assignCommand   applicationipam.AssignIPAddressCommand
	unassignCommand applicationipam.UnassignIPAddressCommand
}

func (spy *grpcIPAddressServiceSpy) ListIPAddresses(
	context.Context,
	identity.Principal,
	applicationipam.ListIPAddressesQuery,
) (applicationipam.IPAddressPage, error) {
	spy.calls++
	if spy.address == nil {
		return applicationipam.IPAddressPage{}, nil
	}
	return applicationipam.IPAddressPage{
		Count: 1, Results: []*domainipam.IPAddress{spy.address},
	}, nil
}

func (spy *grpcIPAddressServiceSpy) GetIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.GetIPAddressQuery,
) (*domainipam.IPAddress, error) {
	spy.calls++
	return spy.address, nil
}

func (spy *grpcIPAddressServiceSpy) CreateIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.CreateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.calls++
	return spy.address, nil
}

func (spy *grpcIPAddressServiceSpy) ReplaceIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.ReplaceIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.calls++
	return spy.address, nil
}

func (spy *grpcIPAddressServiceSpy) UpdateIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.UpdateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.calls++
	return spy.address, nil
}

func (spy *grpcIPAddressServiceSpy) DeleteIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.DeleteIPAddressCommand,
) error {
	spy.calls++
	return nil
}

func (spy *grpcIPAddressServiceSpy) AssignIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.AssignIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.calls++
	spy.assignCommand = command
	return spy.address, nil
}

func (spy *grpcIPAddressServiceSpy) UnassignIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UnassignIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.calls++
	spy.unassignCommand = command
	return spy.address, nil
}
