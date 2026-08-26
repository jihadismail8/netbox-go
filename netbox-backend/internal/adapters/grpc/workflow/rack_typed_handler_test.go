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

func TestTypedRackGRPCListPreservesFiltersAndProjectsCanonicalShape(t *testing.T) {
	service := &rackGRPCServiceSpy{rack: grpcRackFixture(t)}
	handler := NewRackRPCHandler(service)
	zero, offset := uint32(0), uint32(7)
	siteID, roleID, rackTypeID := int64(-1), int64(9), int64(8)
	siteSlug, name, rackStatus := "moscow", "A01", "active"
	roleSlug, rackTypeSlug := "production", "r24"

	response, err := handler.ListRacks(rackGRPCContext(t), &dcimv1.ListRacksRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Offset: &offset, Id: []int64{-7, 0, 41},
			Ordering: []string{"site", "-name"},
		},
		SiteId: &siteID, SiteSlug: &siteSlug, Name: &name, Status: &rackStatus,
		RoleId: &roleID, RoleSlug: &roleSlug,
		RackTypeId: &rackTypeID, RackTypeSlug: &rackTypeSlug,
	})
	require.NoError(t, err)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumRackPageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, uint32(7), service.listQuery.Offset)
	assert.Equal(t, []int64{-7, 0, 41}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1}, service.listQuery.SiteIDs)
	assert.Equal(t, []string{"moscow"}, service.listQuery.SiteSlugs)
	assert.Equal(t, []string{"A01"}, service.listQuery.Names)
	assert.Equal(t, []string{"active"}, service.listQuery.Statuses)
	assert.Equal(t, []int64{9}, service.listQuery.RoleIDs)
	assert.Equal(t, []string{"production"}, service.listQuery.RoleSlugs)
	assert.Equal(t, []int64{8}, service.listQuery.RackTypeIDs)
	assert.Equal(t, []string{"r24"}, service.listQuery.RackTypeSlugs)
	require.Len(t, response.Results, 1)
	assert.Equal(t, uint64(1), response.Page.Count)
	assert.Equal(t, int64(41), response.Results[0].Id)
	assert.Equal(t, "/api/dcim/racks/41/", response.Results[0].Url)
	require.NotNil(t, response.Results[0].Site)
	assert.Equal(t, int64(3), response.Results[0].Site.Id)
	assert.Equal(t, wrapperspb.Int64(8), response.Results[0].RackTypeId)
	assert.Equal(t, wrapperspb.Int64(9), response.Results[0].RoleId)
	assert.Equal(t, uint64(2), response.Results[0].DeviceCount)
}

func TestTypedRackGRPCCreatePreservesOmissionAndExplicitBlank(t *testing.T) {
	service := &rackGRPCServiceSpy{rack: grpcRackFixture(t)}
	handler := NewRackRPCHandler(service)
	siteID := int64(3)
	name, blank := "A01", ""
	descUnits := false

	response, err := handler.CreateRack(rackGRPCContext(t), &dcimv1.CreateRackRequest{
		Rack: &dcimv1.RackInput{
			Site: &siteID, Name: &name, AssetTag: &blank, Airflow: &blank,
			DescUnits: &descUnits,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Width.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.RackType.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.Airflow.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.DescUnits.State())
	value, present := service.createCommand.DescUnits.Get()
	assert.True(t, present)
	assert.False(t, value)
	assert.Equal(t, int64(41), response.Rack.Id)
}

func TestTypedRackGRPCFieldMaskClearsNullableFieldsAndRequiresNonNullableValues(t *testing.T) {
	service := &rackGRPCServiceSpy{rack: grpcRackFixture(t)}
	handler := NewRackRPCHandler(service)

	response, err := handler.UpdateRack(rackGRPCContext(t), &dcimv1.UpdateRackRequest{
		Id: 41, Rack: &dcimv1.RackInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
			"facility_id", "rack_type", "role", "asset_tag", "form_factor",
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.FacilityID.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.RackType.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Role.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.FormFactor.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Airflow.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Name.State())
	assert.Equal(t, int64(41), response.Rack.Id)

	for _, field := range []string{
		"site", "name", "status", "serial", "width", "u_height", "starting_unit",
		"desc_units", "airflow", "description", "comments",
	} {
		field := field
		t.Run("supported absent field reaches shared service validation/"+field, func(t *testing.T) {
			service.updateErr = shared.NewValidationError(shared.FieldViolation{
				Field: field, Reason: "null", Description: "This field may not be null.",
			})
			beforeCalls := service.updateCalls
			_, updateErr := handler.UpdateRack(rackGRPCContext(t), &dcimv1.UpdateRackRequest{
				Id: 41, Rack: &dcimv1.RackInput{},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
			})
			require.Error(t, updateErr)
			assert.Equal(t, codes.InvalidArgument, status.Code(updateErr))
			assert.Equal(t, beforeCalls+1, service.updateCalls)
			assert.Equal(t, applicationdcim.FieldNull, rackUpdateGRPCStates(service.updateCommand)[field])
			assert.Equal(t, []shared.FieldViolation{{
				Field: field, Reason: "null", Description: "This field may not be null.",
			}}, shared.ViolationsOf(service.updateErr))
		})
	}

	service.updateErr = nil
	beforeCalls := service.updateCalls
	_, err = handler.UpdateRack(rackGRPCContext(t), &dcimv1.UpdateRackRequest{
		Id: 41, Rack: &dcimv1.RackInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Equal(t, beforeCalls, service.updateCalls, "unknown masks fail before service invocation")
}

func TestRackRPCDispatchesAllMethodsAndRequiresAuthentication(t *testing.T) {
	service := &rackGRPCServiceSpy{rack: grpcRackFixture(t)}
	handler := NewRackRPCHandler(service)
	ctx := rackGRPCContext(t)

	_, err := handler.GetRack(ctx, &dcimv1.GetRackRequest{Id: 41})
	require.NoError(t, err)
	siteID := int64(3)
	name := "A01"
	_, err = handler.ReplaceRack(ctx, &dcimv1.ReplaceRackRequest{
		Id: 41, Rack: &dcimv1.RackInput{Site: &siteID, Name: &name},
	})
	require.NoError(t, err)
	_, err = handler.DeleteRack(ctx, &dcimv1.DeleteRackRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)

	_, err = handler.ListRacks(context.Background(), &dcimv1.ListRacksRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	services := completeTypedDCIMTestServices()
	services.racks = service
	server := services.server()
	_, err = server.ListRacks(ctx, &dcimv1.ListRacksRequest{})
	require.NoError(t, err)
	assert.Equal(t, 1, service.listCalls)
}

func rackGRPCContext(t *testing.T) context.Context {
	t.Helper()
	return identity.WithPrincipal(t.Context(), identity.Principal{
		ID: 17, Username: "rack-grpc", IsSuperuser: true,
	})
}

func grpcRackFixture(t *testing.T) *domaindcim.Rack {
	t.Helper()
	site, err := domaindcim.NewSiteReference(3, "Moscow", "moscow")
	require.NoError(t, err)
	rackType, err := domaindcim.NewRackTypeReference(
		8, "R24", "r24", domaindcim.RackPhysicalAttributes{
			FormFactor: domaindcim.RackFormFactorWallFrame,
			Width:      domaindcim.RackWidth23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		},
	)
	require.NoError(t, err)
	role, err := domaindcim.NewRackRoleReference(9, "Production", "production")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	rack, err := domaindcim.RestoreRack(domaindcim.RackState{
		ID: 41, Site: site, Name: "A01",
		FacilityID: domaindcim.NonNullRackValue("F01"),
		RackType:   domaindcim.NonNullRackValue(rackType),
		Status:     "active", Role: domaindcim.NonNullRackValue(role),
		Serial: "serial", AssetTag: domaindcim.NonNullRackValue(""),
		FormFactor: domaindcim.NonNullRackValue("wall-frame"),
		Width:      23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		Airflow:     domaindcim.NonNullRackValue("front-to-rear"),
		Description: "Rack", Comments: "Notes",
		Created: now, LastUpdated: now, DeviceCount: 2,
	})
	require.NoError(t, err)
	return rack
}

type rackGRPCServiceSpy struct {
	rack           *domaindcim.Rack
	listQuery      applicationdcim.ListRacksQuery
	getQuery       applicationdcim.GetRackQuery
	createCommand  applicationdcim.CreateRackCommand
	replaceCommand applicationdcim.ReplaceRackCommand
	updateCommand  applicationdcim.UpdateRackCommand
	deleteCommand  applicationdcim.DeleteRackCommand
	listCalls      int
	updateCalls    int
	updateErr      error
}

func (service *rackGRPCServiceSpy) ListRacks(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListRacksQuery,
) (applicationdcim.RackPage, error) {
	service.listCalls++
	service.listQuery = query
	if service.rack == nil {
		return applicationdcim.RackPage{}, nil
	}
	return applicationdcim.RackPage{
		Count: 1, Results: []*domaindcim.Rack{service.rack},
	}, nil
}

func (service *rackGRPCServiceSpy) GetRack(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetRackQuery,
) (*domaindcim.Rack, error) {
	service.getQuery = query
	return service.rack, nil
}

func (service *rackGRPCServiceSpy) CreateRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateRackCommand,
) (*domaindcim.Rack, error) {
	service.createCommand = command
	return service.rack, nil
}

func (service *rackGRPCServiceSpy) ReplaceRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceRackCommand,
) (*domaindcim.Rack, error) {
	service.replaceCommand = command
	return service.rack, nil
}

func (service *rackGRPCServiceSpy) UpdateRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateRackCommand,
) (*domaindcim.Rack, error) {
	service.updateCalls++
	service.updateCommand = command
	if service.updateErr != nil {
		return nil, service.updateErr
	}
	return service.rack, nil
}

func (service *rackGRPCServiceSpy) DeleteRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteRackCommand,
) error {
	service.deleteCommand = command
	return nil
}
