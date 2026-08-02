package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestTypedInterfaceRPCPreservesFiltersNullableMaskAndProjection(t *testing.T) {
	spy := &grpcInterfaceServiceSpy{networkInterface: grpcInterfaceFixture(t)}
	handler := NewInterfaceRPCHandler(spy)
	ctx := identity.WithPrincipal(
		t.Context(),
		identity.Principal{ID: 1, Username: "interface-grpc", IsSuperuser: true},
	)
	disabled, management := false, true
	limit, offset := uint32(0), uint32(3)
	deviceID := int64(7)
	deviceName, name, interfaceType, query := "edge-01", "Ethernet1", "10gbase-sr", "uplink"

	_, err := handler.ListInterfaces(ctx, &dcimv1.ListInterfacesRequest{
		Page: &typesv1.PageRequest{
			Limit: &limit, Offset: &offset, Query: &query,
			Ordering: []string{"-type", "name"}, Id: []int64{-1, 41},
		},
		DeviceId: &deviceID, DeviceName: &deviceName, Name: &name,
		Type: &interfaceType, Enabled: &disabled, MgmtOnly: &management,
	})
	require.NoError(t, err)
	assert.True(t, spy.listQuery.LimitPresent)
	assert.Equal(t, []int64{7}, spy.listQuery.DeviceIDs)
	assert.Equal(t, []string{"edge-01"}, spy.listQuery.DeviceNames)
	assert.Equal(t, []string{"Ethernet1"}, spy.listQuery.Names)
	assert.Equal(t, []string{"10gbase-sr"}, spy.listQuery.Types)

	create, err := handler.CreateInterface(ctx, &dcimv1.CreateInterfaceRequest{
		Interface: &dcimv1.InterfaceInput{
			Device: &deviceID, Name: &name, Type: &interfaceType,
			Enabled: &disabled, Mtu: wrapperspb.Int32(1500),
			Speed: wrapperspb.Int64(1_000_000), Duplex: wrapperspb.String(""),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, create.Interface)
	assert.Equal(t, int64(41), create.Interface.Id)
	assert.Equal(t, uint64(3), create.Interface.CountIpaddresses)
	require.NotNil(t, create.Interface.Duplex)
	assert.Empty(t, create.Interface.Duplex.Value, "gRPC retains present blank duplex")
	assert.Equal(t, "edge-01", create.Interface.Device.Display)

	_, err = handler.UpdateInterface(ctx, &dcimv1.UpdateInterfaceRequest{
		Id: 41, Interface: &dcimv1.InterfaceInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"mtu", "speed", "duplex"}},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.MTU.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.Speed.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.Duplex.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Name.State())

	_, err = handler.UpdateInterface(ctx, &dcimv1.UpdateInterfaceRequest{
		Id: 41, Interface: &dcimv1.InterfaceInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"enabled"}},
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, 1, spy.updateCalls)
}

type grpcInterfaceServiceSpy struct {
	listQuery        applicationdcim.ListInterfacesQuery
	createCommand    applicationdcim.CreateInterfaceCommand
	updateCommand    applicationdcim.UpdateInterfaceCommand
	networkInterface *domaindcim.Interface
	updateCalls      int
}

func (spy *grpcInterfaceServiceSpy) ListInterfaces(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListInterfacesQuery,
) (applicationdcim.InterfacePage, error) {
	spy.listQuery = query
	return applicationdcim.InterfacePage{}, nil
}

func (spy *grpcInterfaceServiceSpy) GetInterface(
	context.Context,
	identity.Principal,
	applicationdcim.GetInterfaceQuery,
) (*domaindcim.Interface, error) {
	return spy.networkInterface, nil
}

func (spy *grpcInterfaceServiceSpy) CreateInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateInterfaceCommand,
) (*domaindcim.Interface, error) {
	spy.createCommand = command
	return spy.networkInterface, nil
}

func (spy *grpcInterfaceServiceSpy) ReplaceInterface(
	context.Context,
	identity.Principal,
	applicationdcim.ReplaceInterfaceCommand,
) (*domaindcim.Interface, error) {
	return spy.networkInterface, nil
}

func (spy *grpcInterfaceServiceSpy) UpdateInterface(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateInterfaceCommand,
) (*domaindcim.Interface, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.networkInterface, nil
}

func (spy *grpcInterfaceServiceSpy) DeleteInterface(
	context.Context,
	identity.Principal,
	applicationdcim.DeleteInterfaceCommand,
) error {
	return nil
}

func grpcInterfaceFixture(t *testing.T) *domaindcim.Interface {
	t.Helper()
	reference, err := domaindcim.NewDeviceReference(
		7, domaindcim.NonNullDeviceValue("edge-01"), "edge-01",
	)
	require.NoError(t, err)
	networkInterface, err := domaindcim.RestoreInterface(domaindcim.InterfaceState{
		ID: 41, Device: reference, Name: "Ethernet1", Label: "WAN",
		Type: "10gbase-sr", Enabled: false, MgmtOnly: true,
		MTU:         domaindcim.NullDeviceValue[uint32](),
		Speed:       domaindcim.NonNullDeviceValue(uint64(2147483647)),
		Duplex:      domaindcim.NonNullDeviceValue(""),
		Description: "Interface description",
		Created: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
		),
		LastUpdated: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC),
		),
		IPAddressCount: 3,
	})
	require.NoError(t, err)
	return networkInterface
}
