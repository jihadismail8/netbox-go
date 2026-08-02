package authz_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
)

func TestPermissionAuthorizerUsesNetBoxCodenames(t *testing.T) {
	tests := []struct {
		resource   authz.ResourceType
		action     authz.Action
		permission string
	}{
		{authz.ResourceSite, authz.View, "dcim.view_site"},
		{authz.ResourceRackRole, authz.Add, "dcim.add_rackrole"},
		{authz.ResourceDeviceType, authz.Change, "dcim.change_devicetype"},
		{authz.ResourceInterfaceTemplate, authz.Delete, "dcim.delete_interfacetemplate"},
		{authz.ResourceVRF, authz.View, "ipam.view_vrf"},
		{authz.ResourceIPAddress, authz.Change, "ipam.change_ipaddress"},
	}
	authorizer := authz.PermissionAuthorizer{}
	for _, test := range tests {
		t.Run(test.permission, func(t *testing.T) {
			principal := identity.Principal{ID: 7, Permissions: map[string]struct{}{test.permission: {}}}
			require.NoError(t, authorizer.AuthorizeResource(
				t.Context(), principal, test.action, test.resource, nil,
			))
		})
	}
}

func TestPermissionAuthorizerAppliesObjectVisibilityToNetBoxPermission(t *testing.T) {
	permission := "dcim.view_site"
	principal := identity.Principal{
		ID:          7,
		Permissions: map[string]struct{}{permission: {}},
		ObjectVisibility: map[string]map[int64]struct{}{
			permission: {42: {}},
		},
	}
	authorizer := authz.PermissionAuthorizer{}
	require.NoError(t, authorizer.AuthorizeResource(
		t.Context(), principal, authz.View, authz.ResourceSite, authz.NewObject(42),
	))
	require.Error(t, authorizer.AuthorizeResource(
		t.Context(), principal, authz.View, authz.ResourceSite, authz.NewObject(43),
	))
}

func TestPermissionAuthorizerExposesCompleteListScope(t *testing.T) {
	permission := "dcim.view_site"
	authorizer := authz.PermissionAuthorizer{}
	restricted := identity.Principal{
		ID:          7,
		Permissions: map[string]struct{}{permission: {}},
		ObjectVisibility: map[string]map[int64]struct{}{
			permission: {42: {}, 84: {}},
		},
	}
	scope := authorizer.ResourceListScope(
		t.Context(), restricted, authz.View, authz.ResourceSite,
	)
	require.True(t, scope.Constrained)
	require.ElementsMatch(t, []int64{42, 84}, scope.ObjectIDs)

	unrestricted := restricted
	unrestricted.ObjectVisibility = nil
	require.False(t, authorizer.ResourceListScope(
		t.Context(), unrestricted, authz.View, authz.ResourceSite,
	).Constrained)

	superuser := restricted
	superuser.IsSuperuser = true
	require.False(t, authorizer.ResourceListScope(
		t.Context(), superuser, authz.View, authz.ResourceSite,
	).Constrained)
}
