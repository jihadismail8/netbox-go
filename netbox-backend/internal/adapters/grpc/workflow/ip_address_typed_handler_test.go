package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

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

	_, err = typedIPAddressUpdateCommand(
		17,
		&ipamv1.IPAddressInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"address"}},
	)
	require.Error(t, err)
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
