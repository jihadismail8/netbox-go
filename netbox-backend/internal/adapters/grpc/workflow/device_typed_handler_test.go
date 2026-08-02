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

func TestTypedDeviceGRPCListPreservesFiltersAndProjectsCanonicalShape(t *testing.T) {
	service := &deviceGRPCServiceSpy{device: grpcDeviceFixture(t)}
	handler := NewDeviceRPCHandler(service)
	zero, offset := uint32(0), uint32(7)
	siteID, rackID, deviceTypeID, roleID := int64(-1), int64(6), int64(8), int64(9)
	siteSlug, deviceTypeSlug, roleSlug := "moscow", "router-9000", "core-router"
	name, deviceStatus, query := "edge-01", "active", "edge"

	response, err := handler.ListDevices(deviceGRPCContext(t), &dcimv1.ListDevicesRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Offset: &offset, Query: &query,
			Id: []int64{-7, 0, 41}, Ordering: []string{"site", "-name"},
		},
		SiteId: &siteID, SiteSlug: &siteSlug, RackId: &rackID,
		DeviceTypeId: &deviceTypeID, DeviceTypeSlug: &deviceTypeSlug,
		RoleId: &roleID, RoleSlug: &roleSlug, Name: &name, Status: &deviceStatus,
	})
	require.NoError(t, err)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDevicePageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, uint32(7), service.listQuery.Offset)
	assert.Equal(t, "edge", service.listQuery.Query)
	assert.Equal(t, []int64{-7, 0, 41}, service.listQuery.IDs)
	assert.Equal(t, []string{"site", "-name"}, service.listQuery.Ordering)
	assert.Equal(t, []int64{-1}, service.listQuery.SiteIDs)
	assert.Equal(t, []string{"moscow"}, service.listQuery.SiteSlugs)
	assert.Equal(t, []int64{6}, service.listQuery.RackIDs)
	assert.Equal(t, []int64{8}, service.listQuery.DeviceTypeIDs)
	assert.Equal(t, []string{"router-9000"}, service.listQuery.DeviceTypeSlugs)
	assert.Equal(t, []int64{9}, service.listQuery.RoleIDs)
	assert.Equal(t, []string{"core-router"}, service.listQuery.RoleSlugs)
	assert.Equal(t, []string{"edge-01"}, service.listQuery.Names)
	assert.Equal(t, []string{"active"}, service.listQuery.Statuses)

	require.NotNil(t, response.Page)
	assert.Equal(t, uint64(1), response.Page.Count)
	require.Len(t, response.Results, 1)
	device := response.Results[0]
	assert.Equal(t, int64(41), device.Id)
	assert.Equal(t, "/api/dcim/devices/41/", device.Url)
	assert.Equal(t, "edge-01 (ASSET-1)", device.Display)
	require.NotNil(t, device.DeviceType)
	assert.Equal(t, int64(8), device.DeviceType.Id)
	assert.Equal(t, "/api/dcim/device-types/8/", device.DeviceType.Url)
	require.NotNil(t, device.Role)
	assert.Equal(t, int64(9), device.Role.Id)
	require.NotNil(t, device.Site)
	assert.Equal(t, int64(3), device.Site.Id)
	assert.Equal(t, wrapperspb.Int64(6), device.RackId)
	assert.Equal(t, "10.5", device.Position)
	assert.Equal(t, "rear", device.Face)
	assert.Equal(t, "active", device.Status)
	assert.Equal(t, "ASSET-1", device.AssetTag)
	assert.Equal(t, "rear-to-front", device.Airflow)
	assert.Equal(t, uint64(4), device.InterfaceCount)
}

func TestTypedDeviceGRPCCreatePreservesOmissionAndExplicitBlank(t *testing.T) {
	service := &deviceGRPCServiceSpy{device: grpcDeviceFixture(t)}
	handler := NewDeviceRPCHandler(service)
	deviceTypeID, roleID, siteID := int64(8), int64(9), int64(3)
	blank := ""

	response, err := handler.CreateDevice(deviceGRPCContext(t), &dcimv1.CreateDeviceRequest{
		Device: &dcimv1.DeviceInput{
			DeviceType: &deviceTypeID,
			Role:       &roleID,
			Name:       wrapperspb.String(""),
			Site:       &siteID,
			Face:       wrapperspb.String(""),
			AssetTag:   wrapperspb.String(""),
			Airflow:    &blank,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.DeviceType.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.Name.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Rack.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Position.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.Face.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.Airflow.State())
	name, present := service.createCommand.Name.Get()
	require.True(t, present)
	assert.Empty(t, name)
	require.NotNil(t, response.Device)
	assert.Equal(t, int64(41), response.Device.Id)
}

func TestTypedDeviceGRPCFieldMaskClearsNullableFieldsAndRequiresNonNullableValues(t *testing.T) {
	service := &deviceGRPCServiceSpy{device: grpcDeviceFixture(t)}
	handler := NewDeviceRPCHandler(service)

	response, err := handler.UpdateDevice(deviceGRPCContext(t), &dcimv1.UpdateDeviceRequest{
		Id: 41, Device: &dcimv1.DeviceInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"name", "rack", "position", "face", "asset_tag", "airflow",
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Name.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Rack.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Position.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Face.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Airflow.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Status.State())
	require.NotNil(t, response.Device)

	_, err = handler.UpdateDevice(deviceGRPCContext(t), &dcimv1.UpdateDeviceRequest{
		Id: 41, Device: &dcimv1.DeviceInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"status"}},
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, 1, service.updateCalls)
}

func TestDeviceRPCDispatchesAllMethodsAndRequiresAuthentication(t *testing.T) {
	service := &deviceGRPCServiceSpy{device: grpcDeviceFixture(t)}
	handler := NewDeviceRPCHandler(service)
	ctx := deviceGRPCContext(t)

	_, err := handler.GetDevice(ctx, &dcimv1.GetDeviceRequest{Id: 41})
	require.NoError(t, err)
	deviceTypeID, roleID, siteID := int64(8), int64(9), int64(3)
	_, err = handler.ReplaceDevice(ctx, &dcimv1.ReplaceDeviceRequest{
		Id: 41,
		Device: &dcimv1.DeviceInput{
			DeviceType: &deviceTypeID, Role: &roleID, Site: &siteID,
		},
	})
	require.NoError(t, err)
	_, err = handler.DeleteDevice(ctx, &dcimv1.DeleteDeviceRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)

	_, err = handler.ListDevices(context.Background(), &dcimv1.ListDevicesRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	services := completeTypedDCIMTestServices()
	services.devices = service
	server := services.server()
	_, err = server.ListDevices(ctx, &dcimv1.ListDevicesRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, service.listCalls)
}

func deviceGRPCContext(t *testing.T) context.Context {
	t.Helper()
	return identity.WithPrincipal(t.Context(), identity.Principal{
		ID: 17, Username: "device-grpc", IsSuperuser: true,
	})
}

func grpcDeviceFixture(t *testing.T) *domaindcim.Device {
	t.Helper()
	height, err := domaindcim.ParseDeviceHeight("1.5")
	require.NoError(t, err)
	deviceType, err := domaindcim.NewDeviceTypeInstanceReference(
		8, "Router 9000", "router-9000", "Acme", height, true,
		domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
	)
	require.NoError(t, err)
	site, err := domaindcim.NewSiteReference(3, "Moscow", "moscow")
	require.NoError(t, err)
	rack, err := domaindcim.NewRackReference(6, "Rack A", 3, 1, 42)
	require.NoError(t, err)
	position, err := domaindcim.ParseRackPosition("10.5")
	require.NoError(t, err)
	created := shared.NewTimestamp(time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC))
	updated := shared.NewTimestamp(time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC))
	device, err := domaindcim.RestoreDevice(domaindcim.DeviceState{
		ID: 41, DeviceType: deviceType,
		Role: domaindcim.DeviceRoleReference{ID: 9, Display: "Core Router"},
		Name: domaindcim.NonNullDeviceValue("edge-01"), Site: site,
		Rack:     domaindcim.NonNullDeviceValue(rack),
		Position: domaindcim.NonNullDeviceValue(position),
		Face:     "rear", Status: "active", Serial: "SN-1",
		AssetTag: domaindcim.NonNullDeviceValue("ASSET-1"),
		Airflow: domaindcim.NonNullDeviceAirflow(
			domaindcim.DeviceAirflowRearToFront,
		),
		Description: "Edge router", Comments: "Notes",
		Created: created, LastUpdated: updated, InterfaceCount: 4,
	})
	require.NoError(t, err)
	return device
}

type deviceGRPCServiceSpy struct {
	device         *domaindcim.Device
	listQuery      applicationdcim.ListDevicesQuery
	getQuery       applicationdcim.GetDeviceQuery
	createCommand  applicationdcim.CreateDeviceCommand
	replaceCommand applicationdcim.ReplaceDeviceCommand
	updateCommand  applicationdcim.UpdateDeviceCommand
	deleteCommand  applicationdcim.DeleteDeviceCommand
	listCalls      int
	updateCalls    int
}

func (service *deviceGRPCServiceSpy) ListDevices(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListDevicesQuery,
) (applicationdcim.DevicePage, error) {
	service.listCalls++
	service.listQuery = query
	if service.device == nil {
		return applicationdcim.DevicePage{}, nil
	}
	return applicationdcim.DevicePage{
		Count: 1, Results: []*domaindcim.Device{service.device},
	}, nil
}

func (service *deviceGRPCServiceSpy) GetDevice(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetDeviceQuery,
) (*domaindcim.Device, error) {
	service.getQuery = query
	return service.device, nil
}

func (service *deviceGRPCServiceSpy) CreateDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateDeviceCommand,
) (*domaindcim.Device, error) {
	service.createCommand = command
	return service.device, nil
}

func (service *deviceGRPCServiceSpy) ReplaceDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceDeviceCommand,
) (*domaindcim.Device, error) {
	service.replaceCommand = command
	return service.device, nil
}

func (service *deviceGRPCServiceSpy) UpdateDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateDeviceCommand,
) (*domaindcim.Device, error) {
	service.updateCalls++
	service.updateCommand = command
	return service.device, nil
}

func (service *deviceGRPCServiceSpy) DeleteDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteDeviceCommand,
) error {
	service.deleteCommand = command
	return nil
}
