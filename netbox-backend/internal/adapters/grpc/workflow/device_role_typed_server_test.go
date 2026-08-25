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

func TestTypedDeviceRoleGRPCListPreservesPresenceAndHierarchyProjection(t *testing.T) {
	role := restoredGRPCDeviceRole(t, true)
	service := &deviceRoleGRPCServiceSpy{page: applicationdcim.DeviceRolePage{
		Count: 1, Results: []*domaindcim.DeviceRole{role},
	}}
	server := NewDCIMDeviceRoleServer(service)
	zero := uint32(0)
	name := "Leaf"

	response, err := server.ListDeviceRoles(deviceRoleGRPCContext(t), &dcimv1.ListDeviceRolesRequest{
		Page: &typesv1.PageRequest{Limit: &zero, Id: []int64{-1, 8}}, Name: &name,
	})

	require.NoError(t, err)
	assert.True(t, service.query.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDeviceRolePageLimit, service.query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 8}, service.query.IDs)
	assert.Equal(t, []string{"Leaf"}, service.query.Names)
	require.Len(t, response.Results, 1)
	require.NotNil(t, response.Results[0].ParentId)
	assert.Equal(t, int64(7), response.Results[0].ParentId.Value)
	assert.Equal(t, uint64(4), response.Results[0].DeviceCount)
	assert.Equal(t, uint32(1), response.Results[0].Depth)
	assert.Equal(t, uint64(1), response.Page.Count)
}

func TestTypedDeviceRoleGRPCUpdateMaskCarriesExplicitNullIntent(t *testing.T) {
	service := &deviceRoleGRPCServiceSpy{role: restoredGRPCDeviceRole(t, false)}
	server := NewDCIMDeviceRoleServer(service)
	ctx := deviceRoleGRPCContext(t)

	_, err := server.UpdateDeviceRole(ctx, &dcimv1.UpdateDeviceRoleRequest{
		Id: 8, DeviceRole: &dcimv1.DeviceRoleInput{},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"color"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, service.updateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.update.Color.State())

	vmRole := false
	response, err := server.UpdateDeviceRole(ctx, &dcimv1.UpdateDeviceRoleRequest{
		Id: 8, DeviceRole: &dcimv1.DeviceRoleInput{VmRole: &vmRole},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"parent", "vm_role"}},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, service.updateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.update.Parent.State())
	value, present := service.update.VMRole.Get()
	require.True(t, present)
	assert.False(t, value)
	assert.Nil(t, response.DeviceRole.ParentId)
}

func deviceRoleGRPCContext(t *testing.T) context.Context {
	t.Helper()
	return identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-device-role"})
}

func restoredGRPCDeviceRole(t *testing.T, child bool) *domaindcim.DeviceRole {
	t.Helper()
	now := shared.NewTimestamp(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	state := domaindcim.DeviceRoleState{
		ID: 8, Parent: domaindcim.RootDeviceRoleParent(), Name: "Leaf", Slug: "leaf",
		Color: "9e9e9e", VMRole: false, Created: now, LastUpdated: now, DeviceCount: 4,
	}
	if child {
		state.Parent = domaindcim.NonRootDeviceRoleParent(7)
		state.ParentDisplay = "Root"
		state.Depth = 1
	}
	role, err := domaindcim.RestoreDeviceRole(state)
	require.NoError(t, err)
	return role
}

type deviceRoleGRPCServiceSpy struct {
	page        applicationdcim.DeviceRolePage
	query       applicationdcim.ListDeviceRolesQuery
	role        *domaindcim.DeviceRole
	update      applicationdcim.UpdateDeviceRoleCommand
	updateCalls int
}

func (service *deviceRoleGRPCServiceSpy) ListDeviceRoles(
	_ context.Context, _ identity.Principal, query applicationdcim.ListDeviceRolesQuery,
) (applicationdcim.DeviceRolePage, error) {
	service.query = query
	return service.page, nil
}

func (*deviceRoleGRPCServiceSpy) GetDeviceRole(
	context.Context, identity.Principal, applicationdcim.GetDeviceRoleQuery,
) (*domaindcim.DeviceRole, error) {
	return nil, nil
}

func (*deviceRoleGRPCServiceSpy) CreateDeviceRole(
	context.Context, identity.Principal, applicationdcim.CreateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	return nil, nil
}

func (*deviceRoleGRPCServiceSpy) ReplaceDeviceRole(
	context.Context, identity.Principal, applicationdcim.ReplaceDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	return nil, nil
}

func (service *deviceRoleGRPCServiceSpy) UpdateDeviceRole(
	_ context.Context, _ identity.Principal, command applicationdcim.UpdateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.updateCalls++
	service.update = command
	return service.role, nil
}

func (*deviceRoleGRPCServiceSpy) DeleteDeviceRole(
	context.Context, identity.Principal, applicationdcim.DeleteDeviceRoleCommand,
) error {
	return nil
}
