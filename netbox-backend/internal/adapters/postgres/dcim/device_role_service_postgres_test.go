package dcim

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRoleServiceValidatesHierarchyUniquenessDefaultsAndRecursiveDelete(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	service := newPostgresDeviceRoleService(t, db, postgreschangelog.NewRecorder(db), allowDeviceRoleAuthorizer{})
	principal := identity.Principal{ID: 17, Username: "device-role-auditor"}

	root := createDeviceRoleThroughService(t, service, principal, applicationdcim.CreateDeviceRoleCommand{
		Name: applicationdcim.FieldValue("Root"), Slug: applicationdcim.FieldValue("root"),
	})
	assert.Equal(t, domaindcim.DeviceRoleDefaultColor, root.Color().String())
	assert.True(t, root.VMRole())
	child := createDeviceRoleThroughService(t, service, principal, applicationdcim.CreateDeviceRoleCommand{
		Parent: applicationdcim.FieldValue(root.ID()), Name: applicationdcim.FieldValue("Child"),
		Slug: applicationdcim.FieldValue("child"), VMRole: applicationdcim.FieldValue(false),
	})
	grandchild := createDeviceRoleThroughService(t, service, principal, applicationdcim.CreateDeviceRoleCommand{
		Parent: applicationdcim.FieldValue(child.ID()), Name: applicationdcim.FieldValue("Grandchild"),
		Slug: applicationdcim.FieldValue("grandchild"),
	})
	assert.False(t, child.VMRole())
	assert.Equal(t, uint32(2), grandchild.Depth())

	_, err := service.UpdateDeviceRole(t.Context(), principal, applicationdcim.UpdateDeviceRoleCommand{
		ID: root.ID(), Parent: applicationdcim.FieldValue(grandchild.ID()),
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	assert.Equal(t, "parent", shared.ViolationsOf(err)[0].Field)

	_, err = service.CreateDeviceRole(t.Context(), principal, applicationdcim.CreateDeviceRoleCommand{
		Parent: applicationdcim.FieldValue(root.ID()), Name: applicationdcim.FieldValue("Child"),
		Slug: applicationdcim.FieldValue("child-two"),
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, "non_field_errors", shared.ViolationsOf(err)[0].Field)

	seedDeviceRoleDevice(t, db, grandchild.ID(), "protected-edge", "PROTECTED-1")
	var changesBefore int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	err = service.DeleteDeviceRole(t.Context(), principal, applicationdcim.DeleteDeviceRoleCommand{ID: root.ID()})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	for _, id := range []shared.ID{root.ID(), child.ID(), grandchild.ID()} {
		_, getErr := service.GetDeviceRole(t.Context(), principal, applicationdcim.GetDeviceRoleQuery{ID: id})
		require.NoError(t, getErr)
	}
	var changesAfter int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	assert.Equal(t, changesBefore, changesAfter)

	require.NoError(t, db.Where("role_id = ?", grandchild.ID().Int64()).Delete(&dcimrow.DeviceRow{}).Error)
	require.NoError(t, service.DeleteDeviceRole(
		t.Context(), principal, applicationdcim.DeleteDeviceRoleCommand{ID: root.ID()},
	))
	var remainingRoles int64
	require.NoError(t, db.Model(&dcimrow.DeviceRoleRow{}).Count(&remainingRoles).Error)
	assert.Zero(t, remainingRoles)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	assert.Equal(t, changesBefore+3, changesAfter)
}

func TestDeviceRoleServiceRollsBackMutationWhenChangeRecordingFails(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository := NewDeviceRoleRepository(db)
	role := newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Original", "original")
	require.NoError(t, repository.Create(t.Context(), role))
	sentinel := errors.New("forced DeviceRole change failure")
	service := newPostgresDeviceRoleService(
		t, db, failingDeviceRoleRecorder{err: sentinel}, allowDeviceRoleAuthorizer{},
	)
	principal := identity.Principal{ID: 17, Username: "device-role-auditor"}

	_, err := service.UpdateDeviceRole(t.Context(), principal, applicationdcim.UpdateDeviceRoleCommand{
		ID: role.ID(), Description: applicationdcim.FieldValue("changed"),
	})
	require.ErrorIs(t, err, sentinel)
	loaded, getErr := repository.Get(t.Context(), role.ID())
	require.NoError(t, getErr)
	assert.Equal(t, "Original description", loaded.Description())
	var changes int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changes).Error)
	assert.Zero(t, changes)
}

func TestDeviceRoleServiceAppliesCompleteVisibilityBeforeCountAndPage(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository := NewDeviceRoleRepository(db)
	roles := []*domaindcim.DeviceRole{
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "First", "first"),
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Second", "second"),
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Third", "third"),
	}
	for _, role := range roles {
		require.NoError(t, repository.Create(t.Context(), role))
	}
	authorizer := scopedDeviceRoleAuthorizer{visible: map[int64]struct{}{
		roles[1].ID().Int64(): {}, roles[2].ID().Int64(): {},
	}}
	service := newPostgresDeviceRoleService(t, db, postgreschangelog.NewRecorder(db), authorizer)
	page, err := service.ListDeviceRoles(
		t.Context(), identity.Principal{ID: 17, Username: "device-role-viewer"},
		applicationdcim.ListDeviceRolesQuery{Limit: 1},
	)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, roles[1].ID(), page.Results[0].ID())
}

func newPostgresDeviceRoleService(
	t *testing.T,
	db *gorm.DB,
	recorder applicationchangelog.Recorder,
	authorizer authz.ResourceAuthorizer,
) *applicationdcim.DeviceRoleService {
	t.Helper()
	service, err := applicationdcim.NewDeviceRoleService(
		NewDeviceRoleRepository(db), postgresTransaction.NewUnitOfWork(db), recorder,
		authorizer, fixedDeviceRoleClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	return service
}

func createDeviceRoleThroughService(
	t *testing.T,
	service *applicationdcim.DeviceRoleService,
	principal identity.Principal,
	command applicationdcim.CreateDeviceRoleCommand,
) *domaindcim.DeviceRole {
	t.Helper()
	role, err := service.CreateDeviceRole(t.Context(), principal, command)
	require.NoError(t, err)
	return role
}

type fixedDeviceRoleClock struct{ now shared.Timestamp }

func (clock fixedDeviceRoleClock) Now() shared.Timestamp { return clock.now }

type allowDeviceRoleAuthorizer struct{}

func (allowDeviceRoleAuthorizer) AuthorizeResource(
	context.Context,
	identity.Principal,
	authz.Action,
	authz.ResourceType,
	*authz.Object,
) error {
	return nil
}

type scopedDeviceRoleAuthorizer struct{ visible map[int64]struct{} }

func (authorizer scopedDeviceRoleAuthorizer) ResourceListScope(
	context.Context,
	identity.Principal,
	authz.Action,
	authz.ResourceType,
) authz.ListScope {
	ids := make([]int64, 0, len(authorizer.visible))
	for id := range authorizer.visible {
		ids = append(ids, id)
	}
	return authz.ListScope{ObjectIDs: ids, Constrained: true}
}

func (authorizer scopedDeviceRoleAuthorizer) AuthorizeResource(
	_ context.Context,
	_ identity.Principal,
	_ authz.Action,
	_ authz.ResourceType,
	resource *authz.Object,
) error {
	if resource == nil {
		return nil
	}
	if _, visible := authorizer.visible[resource.ID]; visible {
		return nil
	}
	return shared.NewError(
		shared.ErrorReasonForbidden,
		"You do not have permission to perform this action.",
	)
}

type failingDeviceRoleRecorder struct{ err error }

func (recorder failingDeviceRoleRecorder) Record(context.Context, applicationchangelog.Change) error {
	return recorder.err
}
