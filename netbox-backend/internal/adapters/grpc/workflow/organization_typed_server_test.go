package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestTypedOrganizationGRPCListPreservesPagePresenceAndMapsDomainResult(t *testing.T) {
	manufacturer := restoredGRPCManufacturer(t)
	service := &organizationGRPCServiceSpy{
		manufacturerPage: applicationdcim.ManufacturerPage{
			Count: 1, Results: []*domaindcim.Manufacturer{manufacturer},
		},
	}
	server := NewDCIMOrganizationServer(service, service)
	zero := uint32(0)
	name := "Acme"
	response, err := server.ListManufacturers(
		identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-organization"}),
		&dcimv1.ListManufacturersRequest{
			Page: &typesv1.PageRequest{Limit: &zero, Id: []int64{-1, 7}}, Name: &name,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 1, service.manufacturerListCalls)
	assert.True(t, service.manufacturerQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumManufacturerPageLimit, service.manufacturerQuery.EffectiveLimit())
	assert.Equal(t, []int64{-1, 7}, service.manufacturerQuery.IDs)
	assert.Equal(t, []string{"Acme"}, service.manufacturerQuery.Names)
	require.Len(t, response.Results, 1)
	assert.Equal(t, int64(7), response.Results[0].Id)
	assert.Equal(t, "Acme", response.Results[0].Display)
	assert.Equal(t, uint64(3), response.Results[0].DevicetypeCount)
	assert.Equal(t, uint64(1), response.Page.Count)
}

func TestTypedOrganizationGRPCUpdateMaskPreservesExplicitNullIntent(t *testing.T) {
	service := &organizationGRPCServiceSpy{rackRole: restoredGRPCRackRole(t)}
	server := NewDCIMOrganizationServer(service, service)
	ctx := identity.WithPrincipal(
		t.Context(), identity.Principal{ID: 1, Username: "typed-organization"},
	)

	response, err := server.UpdateRackRole(ctx, &dcimv1.UpdateRackRoleRequest{
		Id: 8, RackRole: &dcimv1.RackRoleInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"color"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, service.rackRoleUpdateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.rackRoleUpdate.Color.State())
	assert.Equal(t, "00ff00", response.RackRole.Color)

	color := "00ff00"
	response, err = server.UpdateRackRole(ctx, &dcimv1.UpdateRackRoleRequest{
		Id: 8, RackRole: &dcimv1.RackRoleInput{Color: &color},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"color"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, service.rackRoleUpdateCalls)
	assert.Equal(t, shared.ID(8), service.rackRoleUpdate.ID)
	assert.Equal(t, applicationdcim.FieldOmitted, service.rackRoleUpdate.Name.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.rackRoleUpdate.Color.State())
	assert.Equal(t, "00ff00", response.RackRole.Color)
}

func restoredGRPCManufacturer(t *testing.T) *domaindcim.Manufacturer {
	t.Helper()
	now := shared.NewTimestamp(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	manufacturer, err := domaindcim.RestoreManufacturer(domaindcim.ManufacturerState{
		ID: 7, Name: "Acme", Slug: "acme", Description: "Routers",
		Created: now, LastUpdated: now, DeviceTypeCount: 3,
	})
	require.NoError(t, err)
	return manufacturer
}

func restoredGRPCRackRole(t *testing.T) *domaindcim.RackRole {
	t.Helper()
	now := shared.NewTimestamp(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	role, err := domaindcim.RestoreRackRole(domaindcim.RackRoleState{
		ID: 8, Name: "Core", Slug: "core", Color: "00ff00", Description: "Core racks",
		Created: now, LastUpdated: now, RackCount: 2,
	})
	require.NoError(t, err)
	return role
}

type organizationGRPCServiceSpy struct {
	manufacturerPage      applicationdcim.ManufacturerPage
	manufacturerQuery     applicationdcim.ListManufacturersQuery
	manufacturerListCalls int
	rackRole              *domaindcim.RackRole
	rackRoleUpdate        applicationdcim.UpdateRackRoleCommand
	rackRoleUpdateCalls   int
}

func (service *organizationGRPCServiceSpy) ListManufacturers(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListManufacturersQuery,
) (applicationdcim.ManufacturerPage, error) {
	service.manufacturerListCalls++
	service.manufacturerQuery = query
	return service.manufacturerPage, nil
}

func (*organizationGRPCServiceSpy) GetManufacturer(
	context.Context, identity.Principal, applicationdcim.GetManufacturerQuery,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) CreateManufacturer(
	context.Context, identity.Principal, applicationdcim.CreateManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) ReplaceManufacturer(
	context.Context, identity.Principal, applicationdcim.ReplaceManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) UpdateManufacturer(
	context.Context, identity.Principal, applicationdcim.UpdateManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) DeleteManufacturer(
	context.Context, identity.Principal, applicationdcim.DeleteManufacturerCommand,
) error {
	return nil
}

func (*organizationGRPCServiceSpy) ListRackRoles(
	context.Context, identity.Principal, applicationdcim.ListRackRolesQuery,
) (applicationdcim.RackRolePage, error) {
	return applicationdcim.RackRolePage{}, nil
}

func (*organizationGRPCServiceSpy) GetRackRole(
	context.Context, identity.Principal, applicationdcim.GetRackRoleQuery,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) CreateRackRole(
	context.Context, identity.Principal, applicationdcim.CreateRackRoleCommand,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (*organizationGRPCServiceSpy) ReplaceRackRole(
	context.Context, identity.Principal, applicationdcim.ReplaceRackRoleCommand,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (service *organizationGRPCServiceSpy) UpdateRackRole(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateRackRoleCommand,
) (*domaindcim.RackRole, error) {
	service.rackRoleUpdateCalls++
	service.rackRoleUpdate = command
	return service.rackRole, nil
}

func (*organizationGRPCServiceSpy) DeleteRackRole(
	context.Context, identity.Principal, applicationdcim.DeleteRackRoleCommand,
) error {
	return nil
}
