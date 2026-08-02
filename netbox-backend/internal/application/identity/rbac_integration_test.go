package identity_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	postgres "netbox-go/internal/adapters/postgres/identity"
	"netbox-go/internal/application/authz"
	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestEffectivePermissionsCombineDirectGroupAndObjectGrants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_rbac?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	admin, err := service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	operator, err := service.CreateLocalUser(t.Context(), admin.Principal(), application.CreateUserInput{
		Username: "operator", Password: "Operator-Horse-2026!",
	})
	require.NoError(t, err)
	group, err := service.CreateGroup(t.Context(), admin.Principal(), "DCIM operators")
	require.NoError(t, err)
	require.NoError(t, service.AddGroupMember(t.Context(), admin.Principal(), operator.ID, group.ID))

	firstSite, secondSite := int64(41), int64(42)
	_, err = service.GrantPermissionToGroup(t.Context(), admin.Principal(), group.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site", ObjectID: &firstSite,
	})
	require.NoError(t, err)
	_, err = service.GrantPermissionToUser(t.Context(), admin.Principal(), operator.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site", ObjectID: &secondSite,
	})
	require.NoError(t, err)
	_, err = service.GrantPermissionToGroup(t.Context(), admin.Principal(), group.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "add", Model: "device_type",
	})
	require.NoError(t, err)
	_, err = service.GrantPermissionToUser(t.Context(), admin.Principal(), operator.ID, application.PermissionGrantInput{
		AppLabel: "ipam", Action: "change", Model: "ip_address",
	})
	require.NoError(t, err)

	loaded, err := service.AuthenticatePassword(t.Context(), "operator", "Operator-Horse-2026!")
	require.NoError(t, err)
	require.Equal(t, []string{"dcim.add_devicetype", "dcim.view_site", "ipam.change_ipaddress"}, loaded.Permissions)
	principal := loaded.Principal()
	require.True(t, principal.Has("dcim.view_site"))
	require.Contains(t, principal.ObjectVisibility["dcim.view_site"], firstSite)
	require.Contains(t, principal.ObjectVisibility["dcim.view_site"], secondSite)
	require.NotContains(t, principal.ObjectVisibility, "dcim.add_devicetype")
	require.NotContains(t, principal.ObjectVisibility, "ipam.change_ipaddress")

	// A global grant dominates narrower grants for the same codename.
	_, err = service.GrantPermissionToUserByUsername(t.Context(), admin.Principal(), "operator", application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site",
	})
	require.NoError(t, err)
	loaded, err = service.AuthenticatePassword(t.Context(), "operator", "Operator-Horse-2026!")
	require.NoError(t, err)
	require.NotContains(t, loaded.Principal().ObjectVisibility, "dcim.view_site")
}

func TestRBACAdministrationRequiresSuperuserAndNeverConstrainsSuperuser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_rbac_admin?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	service := application.NewService(postgres.NewStore(db), application.RealClock{})
	admin, err := service.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	operator, err := service.CreateLocalUser(t.Context(), admin.Principal(), application.CreateUserInput{Username: "operator", Password: "Operator-Horse-2026!"})
	require.NoError(t, err)

	_, err = service.CreateGroup(t.Context(), operator.Principal(), "forbidden")
	assertSharedReason(t, err, shared.ErrorReasonForbidden)
	_, err = service.CreateGroup(t.Context(), domain.Principal{}, "anonymous")
	assertSharedReason(t, err, shared.ErrorReasonUnauthenticated)
	objectID := int64(9)
	_, err = service.GrantPermissionToUser(t.Context(), admin.Principal(), operator.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "add", Model: "site", ObjectID: &objectID,
	})
	assertSharedReason(t, err, shared.ErrorReasonValidation)

	_, err = service.GrantPermissionToUser(t.Context(), admin.Principal(), admin.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site", ObjectID: &objectID,
	})
	require.NoError(t, err)
	reloaded, err := service.AuthenticatePassword(t.Context(), "admin", "Correct-Horse-2026!")
	require.NoError(t, err)
	principal := reloaded.Principal()
	require.True(t, principal.Has("ipam.delete_prefix"), "superusers retain implicit permission to every model action")
	require.Empty(t, principal.ObjectVisibility, "object grants must not narrow a superuser")
}

func TestPersistedObjectGrantFiltersTypedResourceAuthorization(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:identity_rbac_workflow?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(postgres.Models()...))
	identityService := application.NewService(postgres.NewStore(db), application.RealClock{})
	admin, err := identityService.BootstrapAdministrator(t.Context(), "admin", "", "Correct-Horse-2026!")
	require.NoError(t, err)
	firstID, secondID := int64(41), int64(42)

	operator, err := identityService.CreateLocalUser(t.Context(), admin.Principal(), application.CreateUserInput{Username: "reader", Password: "Reader-Horse-2026!"})
	require.NoError(t, err)
	group, err := identityService.CreateGroup(t.Context(), admin.Principal(), "site readers")
	require.NoError(t, err)
	require.NoError(t, identityService.AddGroupMember(t.Context(), admin.Principal(), operator.ID, group.ID))
	_, err = identityService.GrantPermissionToGroup(t.Context(), admin.Principal(), group.ID, application.PermissionGrantInput{
		AppLabel: "dcim", Action: "view", Model: "site", ObjectID: &secondID,
	})
	require.NoError(t, err)
	operator, err = identityService.AuthenticatePassword(t.Context(), "reader", "Reader-Horse-2026!")
	require.NoError(t, err)

	authorizer := authz.PermissionAuthorizer{}
	scope := authorizer.ResourceListScope(
		t.Context(), operator.Principal(), authz.View, authz.ResourceSite,
	)
	require.True(t, scope.Constrained)
	require.Equal(t, []int64{secondID}, scope.ObjectIDs)

	err = authorizer.AuthorizeResource(
		t.Context(), operator.Principal(), authz.View, authz.ResourceSite, authz.NewObject(firstID),
	)
	assertSharedReason(t, err, shared.ErrorReasonForbidden)
	err = authorizer.AuthorizeResource(
		t.Context(), operator.Principal(), authz.View, authz.ResourceSite, authz.NewObject(secondID),
	)
	require.NoError(t, err)
}

func assertSharedReason(t *testing.T, err error, reason shared.ErrorReason) {
	t.Helper()
	var applicationError *shared.Error
	require.ErrorAs(t, err, &applicationError)
	require.Equal(t, reason, applicationError.Reason)
}
