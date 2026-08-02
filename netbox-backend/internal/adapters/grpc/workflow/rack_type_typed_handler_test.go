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

func TestTypedRackTypeGRPCListPreservesPresenceFiltersAndProjection(t *testing.T) {
	fixture := grpcRackTypeFixture(t)
	service := &rackTypeGRPCServiceSpy{rackType: fixture}
	handler := NewRackTypeRPCHandler(service)
	zero := uint32(0)
	offset := uint32(7)
	manufacturerID := int64(-1)
	manufacturerSlug := "acme"
	model := "R42"
	slug := "r42"

	response, err := handler.ListRackTypes(rackTypeGRPCContext(t), &dcimv1.ListRackTypesRequest{
		Page: &typesv1.PageRequest{
			Limit: &zero, Offset: &offset, Id: []int64{-7, 0, 41},
			Ordering: []string{"manufacturer", "-model"},
		},
		ManufacturerId: &manufacturerID, ManufacturerSlug: &manufacturerSlug,
		Model: &model, Slug: &slug,
	})
	require.NoError(t, err)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumRackTypePageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, uint32(7), service.listQuery.Offset)
	assert.Equal(t, []int64{-7, 0, 41}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1}, service.listQuery.ManufacturerIDs)
	assert.Equal(t, []string{"acme"}, service.listQuery.ManufacturerSlugs)
	assert.Equal(t, []string{"R42"}, service.listQuery.Models)
	assert.Equal(t, []string{"r42"}, service.listQuery.Slugs)
	require.Len(t, response.Results, 1)
	assert.Equal(t, uint64(1), response.Page.Count)
	assert.Equal(t, int64(41), response.Results[0].Id)
	assert.Equal(t, "R42", response.Results[0].Display)
	require.NotNil(t, response.Results[0].Manufacturer)
	assert.Equal(t, int64(9), response.Results[0].Manufacturer.Id)
	assert.Equal(t, "/api/dcim/manufacturers/9/", response.Results[0].Manufacturer.Url)
	assert.Equal(t, "Acme", response.Results[0].Manufacturer.Display)
	assert.Equal(t, uint32(19), response.Results[0].Width)
}

func TestTypedRackTypeGRPCCreatePreservesOmittedDefaultsAndExplicitFalse(t *testing.T) {
	service := &rackTypeGRPCServiceSpy{rackType: grpcRackTypeFixture(t)}
	handler := NewRackTypeRPCHandler(service)
	manufacturerID := int64(9)
	model := "R42"
	slug := "r42"
	formFactor := "4-post-cabinet"
	descUnits := false

	response, err := handler.CreateRackType(rackTypeGRPCContext(t), &dcimv1.CreateRackTypeRequest{
		RackType: &dcimv1.RackTypeInput{
			Manufacturer: &manufacturerID, Model: &model, Slug: &slug,
			FormFactor: &formFactor, DescUnits: &descUnits,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Width.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.UHeight.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.StartingUnit.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.DescUnits.State())
	value, present := service.createCommand.DescUnits.Get()
	assert.True(t, present)
	assert.False(t, value)
	assert.Equal(t, int64(41), response.RackType.Id)
}

func TestTypedRackTypeGRPCUpdateMaskRequiresExplicitSupportedFields(t *testing.T) {
	service := &rackTypeGRPCServiceSpy{rackType: grpcRackTypeFixture(t)}
	handler := NewRackTypeRPCHandler(service)

	_, err := handler.UpdateRackType(rackTypeGRPCContext(t), &dcimv1.UpdateRackTypeRequest{
		Id: 41, RackType: &dcimv1.RackTypeInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"width"}},
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
	assert.Zero(t, service.updateCalls)

	width := uint32(23)
	description := ""
	response, err := handler.UpdateRackType(rackTypeGRPCContext(t), &dcimv1.UpdateRackTypeRequest{
		Id: 41, RackType: &dcimv1.RackTypeInput{Width: &width, Description: &description},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"width", "description"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, service.updateCalls)
	assert.Equal(t, shared.ID(41), service.updateCommand.ID)
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Manufacturer.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Width.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Description.State())
	assert.Equal(t, int64(41), response.RackType.Id)
}

func TestRackTypeRPCHandlerDispatchesGetReplaceDeleteAndRequiresAuthentication(t *testing.T) {
	service := &rackTypeGRPCServiceSpy{rackType: grpcRackTypeFixture(t)}
	handler := NewRackTypeRPCHandler(service)
	ctx := rackTypeGRPCContext(t)

	get, err := handler.GetRackType(ctx, &dcimv1.GetRackTypeRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, int64(41), get.RackType.Id)
	manufacturerID := int64(9)
	model, slug, formFactor := "R42", "r42", "4-post-cabinet"
	replaced, err := handler.ReplaceRackType(ctx, &dcimv1.ReplaceRackTypeRequest{
		Id: 41, RackType: &dcimv1.RackTypeInput{
			Manufacturer: &manufacturerID, Model: &model, Slug: &slug, FormFactor: &formFactor,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(41), replaced.RackType.Id)
	_, err = handler.DeleteRackType(ctx, &dcimv1.DeleteRackTypeRequest{Id: 41})
	require.NoError(t, err)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)

	_, err = handler.ListRackTypes(context.Background(), &dcimv1.ListRackTypesRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func rackTypeGRPCContext(t *testing.T) context.Context {
	t.Helper()
	return identity.WithPrincipal(t.Context(), identity.Principal{
		ID: 17, Username: "rack-type-grpc", IsSuperuser: true,
	})
}

func grpcRackTypeFixture(t *testing.T) *domaindcim.RackType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	rackType, err := domaindcim.RestoreRackType(domaindcim.RackTypeState{
		ID: 41, Manufacturer: reference, Model: "R42", Slug: "r42", FormFactor: "4-post-cabinet",
		Width: 19, UHeight: 42, StartingUnit: 1, Description: "Rack type", Comments: "Notes",
		Created: now, LastUpdated: now,
	})
	require.NoError(t, err)
	return rackType
}

type rackTypeGRPCServiceSpy struct {
	rackType       *domaindcim.RackType
	listQuery      applicationdcim.ListRackTypesQuery
	getQuery       applicationdcim.GetRackTypeQuery
	createCommand  applicationdcim.CreateRackTypeCommand
	replaceCommand applicationdcim.ReplaceRackTypeCommand
	updateCommand  applicationdcim.UpdateRackTypeCommand
	deleteCommand  applicationdcim.DeleteRackTypeCommand
	updateCalls    int
}

func (service *rackTypeGRPCServiceSpy) ListRackTypes(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListRackTypesQuery,
) (applicationdcim.RackTypePage, error) {
	service.listQuery = query
	if service.rackType == nil {
		return applicationdcim.RackTypePage{}, nil
	}
	return applicationdcim.RackTypePage{
		Count: 1, Results: []*domaindcim.RackType{service.rackType},
	}, nil
}

func (service *rackTypeGRPCServiceSpy) GetRackType(_ context.Context, _ identity.Principal, query applicationdcim.GetRackTypeQuery) (*domaindcim.RackType, error) {
	service.getQuery = query
	return service.rackType, nil
}

func (service *rackTypeGRPCServiceSpy) CreateRackType(_ context.Context, _ identity.Principal, command applicationdcim.CreateRackTypeCommand) (*domaindcim.RackType, error) {
	service.createCommand = command
	return service.rackType, nil
}

func (service *rackTypeGRPCServiceSpy) ReplaceRackType(_ context.Context, _ identity.Principal, command applicationdcim.ReplaceRackTypeCommand) (*domaindcim.RackType, error) {
	service.replaceCommand = command
	return service.rackType, nil
}

func (service *rackTypeGRPCServiceSpy) UpdateRackType(_ context.Context, _ identity.Principal, command applicationdcim.UpdateRackTypeCommand) (*domaindcim.RackType, error) {
	service.updateCalls++
	service.updateCommand = command
	return service.rackType, nil
}

func (service *rackTypeGRPCServiceSpy) DeleteRackType(_ context.Context, _ identity.Principal, command applicationdcim.DeleteRackTypeCommand) error {
	service.deleteCommand = command
	return nil
}
