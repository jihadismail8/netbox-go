package authz_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestTypedResourcePermissionsUsePinnedNetBoxCodenames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		resource   authz.ResourceType
		action     authz.Action
		permission string
	}{
		{authz.ResourceSite, authz.View, "dcim.view_site"},
		{authz.ResourceRackRole, authz.Add, "dcim.add_rackrole"},
		{authz.ResourceDeviceType, authz.Change, "dcim.change_devicetype"},
		{
			authz.ResourceInterfaceTemplate,
			authz.Delete,
			"dcim.delete_interfacetemplate",
		},
		{authz.ResourceVRF, authz.View, "ipam.view_vrf"},
		{authz.ResourceIPAddress, authz.Change, "ipam.change_ipaddress"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.permission, func(t *testing.T) {
			t.Parallel()
			require.True(t, test.resource.Valid())
			assert.Equal(t, test.permission, test.resource.Permission(test.action))
		})
	}
}

func TestTypedAuthorizationObjectRejectsInvalidIdentity(t *testing.T) {
	t.Parallel()

	assert.Nil(t, authz.NewObject(0))
	assert.Nil(t, authz.NewObject(-1))
	assert.Equal(t, int64(42), authz.NewObject(42).ID)
}

func TestTypedPermissionAuthorizerAppliesGlobalAndObjectGrants(t *testing.T) {
	t.Parallel()

	const permission = "dcim.view_device"
	authorizer := authz.PermissionAuthorizer{}
	principal := identity.Principal{
		ID:          7,
		Permissions: map[string]struct{}{permission: {}},
		ObjectVisibility: map[string]map[int64]struct{}{
			permission: {42: {}},
		},
	}

	require.NoError(t, authorizer.AuthorizeResource(
		t.Context(), principal, authz.View, authz.ResourceDevice, authz.NewObject(42),
	))
	err := authorizer.AuthorizeResource(
		t.Context(), principal, authz.View, authz.ResourceDevice, authz.NewObject(43),
	)
	require.Error(t, err)
	assert.Equal(t, shared.ErrorReasonForbidden, shared.ReasonOf(err))

	scope := authorizer.ResourceListScope(
		t.Context(), principal, authz.View, authz.ResourceDevice,
	)
	assert.True(t, scope.Constrained)
	assert.Equal(t, []int64{42}, scope.ObjectIDs)
}

func TestTypedPermissionAuthorizerFailsClosed(t *testing.T) {
	t.Parallel()

	authorizer := authz.PermissionAuthorizer{}
	err := authorizer.AuthorizeResource(
		t.Context(),
		identity.Principal{},
		authz.View,
		authz.ResourceSite,
		nil,
	)
	require.Error(t, err)
	assert.Equal(t, shared.ErrorReasonUnauthenticated, shared.ReasonOf(err))

	err = authorizer.AuthorizeResource(
		t.Context(),
		identity.Principal{ID: 1},
		authz.View,
		authz.ResourceSite,
		nil,
	)
	require.Error(t, err)
	assert.Equal(t, shared.ErrorReasonForbidden, shared.ReasonOf(err))
}
