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

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestTypedDeviceTypeGRPCListPreservesFiltersCountersAndProjection(t *testing.T) {
	service := &deviceTypeGRPCServiceSpy{deviceType: grpcDeviceTypeFixture(t)}
	handler := NewDeviceTypeRPCHandler(service)
	zero := uint32(0)
	offset := uint32(7)
	manufacturerID := int64(-1)
	manufacturerSlug := "acme"
	model := "Router 9000"
	slug := "router-9000"

	response, err := handler.ListDeviceTypes(
		deviceTypeGRPCContext(t),
		&dcimv1.ListDeviceTypesRequest{
			Page: &typesv1.PageRequest{
				Limit: &zero, Offset: &offset, Id: []int64{-7, 0, 41},
				Ordering: []string{"manufacturer", "-model"},
			},
			ManufacturerId: &manufacturerID, ManufacturerSlug: &manufacturerSlug,
			Model: &model, Slug: &slug,
		},
	)
	require.NoError(t, err)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDeviceTypePageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, uint32(7), service.listQuery.Offset)
	assert.Equal(t, []int64{-7, 0, 41}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1}, service.listQuery.ManufacturerIDs)
	assert.Equal(t, []string{"acme"}, service.listQuery.ManufacturerSlugs)
	assert.Equal(t, []string{"Router 9000"}, service.listQuery.Models)
	assert.Equal(t, []string{"router-9000"}, service.listQuery.Slugs)
	require.Len(t, response.Results, 1)
	assert.Equal(t, uint64(1), response.Page.Count)
	result := response.Results[0]
	assert.Equal(t, int64(41), result.Id)
	assert.Equal(t, "Router 9000", result.Display)
	require.NotNil(t, result.Manufacturer)
	assert.Equal(t, int64(9), result.Manufacturer.Id)
	assert.Equal(t, "/api/dcim/manufacturers/9/", result.Manufacturer.Url)
	assert.Equal(t, "1.5", result.UHeight)
	assert.Equal(t, "front-to-rear", result.Airflow)
	assert.Equal(t, uint64(4), result.DeviceCount)
	assert.Equal(t, uint64(6), result.InterfaceTemplateCount)
}

func TestTypedDeviceTypeGRPCCreatePreservesDefaultsHeightAndExplicitFalse(t *testing.T) {
	service := &deviceTypeGRPCServiceSpy{deviceType: grpcDeviceTypeFixture(t)}
	handler := NewDeviceTypeRPCHandler(service)
	manufacturerID := int64(9)
	model := "Router 9000"
	slug := "router-9000"
	height := "1.5"
	excluded := false

	response, err := handler.CreateDeviceType(
		deviceTypeGRPCContext(t),
		&dcimv1.CreateDeviceTypeRequest{
			DeviceType: &dcimv1.DeviceTypeInput{
				Manufacturer: &manufacturerID, Model: &model, Slug: &slug,
				UHeight: &height, ExcludeFromUtilization: &excluded,
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.UHeight.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.ExcludeFromUtilization.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.IsFullDepth.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Airflow.State())
	heightValue, present := service.createCommand.UHeight.Get()
	assert.True(t, present)
	assert.Equal(t, "1.5", heightValue)
	excludedValue, present := service.createCommand.ExcludeFromUtilization.Get()
	assert.True(t, present)
	assert.False(t, excludedValue)
	assert.Equal(t, int64(41), response.DeviceType.Id)
}

func TestTypedDeviceTypeGRPCUpdateMaskCarriesNullAndAllowsBlankAirflow(t *testing.T) {
	service := &deviceTypeGRPCServiceSpy{deviceType: grpcDeviceTypeFixture(t)}
	handler := NewDeviceTypeRPCHandler(service)

	_, err := handler.UpdateDeviceType(
		deviceTypeGRPCContext(t),
		&dcimv1.UpdateDeviceTypeRequest{
			Id: 41, DeviceType: &dcimv1.DeviceTypeInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"airflow"}},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 1, service.updateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Airflow.State())

	blankAirflow := ""
	height := "2.5"
	response, err := handler.UpdateDeviceType(
		deviceTypeGRPCContext(t),
		&dcimv1.UpdateDeviceTypeRequest{
			Id: 41, DeviceType: &dcimv1.DeviceTypeInput{
				Airflow: &blankAirflow, UHeight: &height,
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"airflow", "u_height"}},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, 2, service.updateCalls)
	assert.Equal(t, shared.ID(41), service.updateCommand.ID)
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Airflow.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.UHeight.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Model.State())
	airflow, present := service.updateCommand.Airflow.Get()
	assert.True(t, present)
	assert.Empty(t, airflow)
	assert.Equal(t, int64(41), response.DeviceType.Id)
}

func TestDeviceTypeRPCHandlerDispatchesGetReplaceDeleteAndRequiresAuthentication(t *testing.T) {
	service := &deviceTypeGRPCServiceSpy{deviceType: grpcDeviceTypeFixture(t)}
	handler := NewDeviceTypeRPCHandler(service)
	ctx := deviceTypeGRPCContext(t)

	get, err := handler.GetDeviceType(ctx, &dcimv1.GetDeviceTypeRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, int64(41), get.DeviceType.Id)
	manufacturerID := int64(9)
	model, slug := "Router 9000", "router-9000"
	replaced, err := handler.ReplaceDeviceType(
		ctx,
		&dcimv1.ReplaceDeviceTypeRequest{
			Id: 41, DeviceType: &dcimv1.DeviceTypeInput{
				Manufacturer: &manufacturerID, Model: &model, Slug: &slug,
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(41), replaced.DeviceType.Id)
	_, err = handler.DeleteDeviceType(ctx, &dcimv1.DeleteDeviceTypeRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)

	_, err = handler.ListDeviceTypes(context.Background(), &dcimv1.ListDeviceTypesRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestDeviceTypeRPCDispatchUsesTypedService(t *testing.T) {
	typed := &deviceTypeGRPCServiceSpy{}
	services := completeTypedDCIMTestServices()
	services.deviceTypes = typed
	server := services.server()

	response, err := server.ListDeviceTypes(
		deviceTypeGRPCContext(t), &dcimv1.ListDeviceTypesRequest{},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(0), response.Page.Count)
	require.Equal(t, 1, typed.listCalls)
}

func deviceTypeGRPCContext(t *testing.T) context.Context {
	t.Helper()
	return identity.WithPrincipal(t.Context(), identity.Principal{
		ID: 17, Username: "device-type-grpc", IsSuperuser: true,
	})
}

func grpcDeviceTypeFixture(t *testing.T) *domaindcim.DeviceType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	deviceType, err := domaindcim.RestoreDeviceType(domaindcim.DeviceTypeState{
		ID: 41, Manufacturer: reference,
		Model: "Router 9000", Slug: "router-9000", PartNumber: "PN-9",
		UHeight: "1.5", IsFullDepth: true,
		Airflow:     domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
		Description: "Core router", Comments: "Notes",
		Created: now, LastUpdated: now,
		DeviceCount: 4, InterfaceTemplateCount: 6,
	})
	require.NoError(t, err)
	return deviceType
}

type deviceTypeGRPCServiceSpy struct {
	deviceType     *domaindcim.DeviceType
	listQuery      applicationdcim.ListDeviceTypesQuery
	getQuery       applicationdcim.GetDeviceTypeQuery
	createCommand  applicationdcim.CreateDeviceTypeCommand
	replaceCommand applicationdcim.ReplaceDeviceTypeCommand
	updateCommand  applicationdcim.UpdateDeviceTypeCommand
	deleteCommand  applicationdcim.DeleteDeviceTypeCommand
	listCalls      int
	updateCalls    int
}

func (service *deviceTypeGRPCServiceSpy) ListDeviceTypes(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListDeviceTypesQuery,
) (applicationdcim.DeviceTypePage, error) {
	service.listCalls++
	service.listQuery = query
	if service.deviceType == nil {
		return applicationdcim.DeviceTypePage{}, nil
	}
	return applicationdcim.DeviceTypePage{
		Count: 1, Results: []*domaindcim.DeviceType{service.deviceType},
	}, nil
}

func (service *deviceTypeGRPCServiceSpy) GetDeviceType(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetDeviceTypeQuery,
) (*domaindcim.DeviceType, error) {
	service.getQuery = query
	return service.deviceType, nil
}

func (service *deviceTypeGRPCServiceSpy) CreateDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	service.createCommand = command
	return service.deviceType, nil
}

func (service *deviceTypeGRPCServiceSpy) ReplaceDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	service.replaceCommand = command
	return service.deviceType, nil
}

func (service *deviceTypeGRPCServiceSpy) UpdateDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	service.updateCalls++
	service.updateCommand = command
	return service.deviceType, nil
}

func (service *deviceTypeGRPCServiceSpy) DeleteDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteDeviceTypeCommand,
) error {
	service.deleteCommand = command
	return nil
}
