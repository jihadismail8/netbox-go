package dcim

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRoleRepositoryProjectsTreeOrderDepthParentAndCumulativeDeviceCount(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewDeviceRoleRepository(db)
	alpha := newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Alpha", "alpha")
	beta := newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Beta", "beta")
	require.NoError(t, repository.Create(t.Context(), beta))
	require.NoError(t, repository.Create(t.Context(), alpha))
	zulu := newDeviceRoleFixture(t, domaindcim.NonRootDeviceRoleParent(alpha.ID()), "Zulu", "zulu")
	bravo := newDeviceRoleFixture(t, domaindcim.NonRootDeviceRoleParent(alpha.ID()), "Bravo", "bravo")
	require.NoError(t, repository.Create(t.Context(), zulu))
	require.NoError(t, repository.Create(t.Context(), bravo))
	leaf := newDeviceRoleFixture(t, domaindcim.NonRootDeviceRoleParent(bravo.ID()), "Leaf", "leaf")
	require.NoError(t, repository.Create(t.Context(), leaf))
	seedDeviceRoleDevice(t, db, zulu.ID(), "zulu-device", "ZT-1")
	seedDeviceRoleDevice(t, db, leaf.ID(), "leaf-device", "LF-1")

	page, err := repository.List(t.Context(), applicationdcim.DeviceRoleListCriteria{
		Limit: 50, DefaultTreeOrder: true,
	})
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, uint64(5), page.Count)
	require.Len(t, page.Results, 5)
	assert.Equal(t, []string{"Alpha", "Bravo", "Leaf", "Zulu", "Beta"}, deviceRoleNames(page.Results))
	assert.Equal(t, uint64(2), page.Results[0].DeviceCount())
	assert.Equal(t, uint64(1), page.Results[1].DeviceCount())
	assert.Equal(t, uint32(2), page.Results[2].Depth())
	reference, present := page.Results[2].ParentReference()
	assert.True(t, present)
	assert.Equal(t, "Bravo", reference.Display)
	assert.Equal(t, uint64(1), page.Results[3].DeviceCount())
	assert.Zero(t, page.Results[4].DeviceCount())
}

func TestDeviceRoleRepositoryRepeatedFiltersSignedIDsVisibilityAndPagination(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewDeviceRoleRepository(db)
	roles := []*domaindcim.DeviceRole{
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Alpha", "alpha"),
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Beta", "beta"),
		newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Gamma", "gamma"),
	}
	for _, role := range roles {
		require.NoError(t, repository.Create(t.Context(), role))
	}
	page, err := repository.List(t.Context(), applicationdcim.DeviceRoleListCriteria{
		Limit: 1, IDs: []int64{-1, roles[0].ID().Int64(), roles[1].ID().Int64()},
		Names: []string{"Alpha", "Beta"}, Slugs: []string{"alpha", "beta"},
		VisibleObjectIDs: []shared.ID{roles[1].ID(), roles[2].ID()}, VisibilityConstrained: true,
		Ordering: []applicationdcim.DeviceRoleSort{{Field: applicationdcim.DeviceRoleSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, roles[1].ID(), page.Results[0].ID())

	empty, err := repository.List(t.Context(), applicationdcim.DeviceRoleListCriteria{
		Limit: 50, IDs: []int64{-1, 0}, DefaultTreeOrder: true,
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)
}

func TestDeviceRoleRepositoryEnforcesSiblingAndRootUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewDeviceRoleRepository(db)
	root := newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Root", "root")
	require.NoError(t, repository.Create(t.Context(), root))
	child := newDeviceRoleFixture(t, domaindcim.NonRootDeviceRoleParent(root.ID()), "Child", "child")
	require.NoError(t, repository.Create(t.Context(), child))

	for _, test := range []struct {
		role    *domaindcim.DeviceRole
		message string
	}{
		{
			role:    newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Root", "other"),
			message: "A top-level device role with this name already exists.",
		},
		{
			role:    newDeviceRoleFixture(t, domaindcim.NonRootDeviceRoleParent(root.ID()), "Other", "child"),
			message: "Device role with this Parent and Slug already exists.",
		},
	} {
		err := repository.Create(t.Context(), test.role)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, test.message, err.Error())
		assert.False(t, test.role.ID().IsValid())
	}
}

func TestDeviceRoleRepositoryFindsDeviceProtection(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewDeviceRoleRepository(db)
	role := newDeviceRoleFixture(t, domaindcim.RootDeviceRoleParent(), "Protected", "protected")
	require.NoError(t, repository.Create(t.Context(), role))
	seedDeviceRoleDevice(t, db, role.ID(), "edge", "EDGE-1")

	dependent, err := repository.FindDeviceUsingRoles(t.Context(), []shared.ID{role.ID()})
	require.NoError(t, err)
	require.NotNil(t, dependent)
	assert.Equal(t, "edge (EDGE-1)", dependent.Display)

	missing, err := repository.FindDeviceUsingRoles(t.Context(), []shared.ID{999})
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func newDeviceRoleFixture(
	t *testing.T,
	parent domaindcim.DeviceRoleParent,
	name string,
	slug string,
) *domaindcim.DeviceRole {
	t.Helper()
	role, err := domaindcim.NewDeviceRole(domaindcim.DeviceRoleValues{
		Parent: parent, Name: name, Slug: slug, Color: domaindcim.DeviceRoleDefaultColor,
		VMRole: true, Description: name + " description", Comments: name + " comments",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return role
}

func deviceRoleNames(roles []*domaindcim.DeviceRole) []string {
	names := make([]string, len(roles))
	for index, role := range roles {
		names[index] = role.Name()
	}
	return names
}

func seedDeviceRoleDevice(
	t *testing.T,
	db interface{ Create(any) *gorm.DB },
	roleID shared.ID,
	name string,
	assetTag string,
) {
	t.Helper()
	site := dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "DeviceRole Site " + roleID.String(), Slug: "device-role-site-" + roleID.String(), Status: "active",
	}
	require.NoError(t, db.Create(&site).Error)
	manufacturer := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "DeviceRole Vendor " + roleID.String(), Slug: "device-role-vendor-" + roleID.String(),
	}
	require.NoError(t, db.Create(&manufacturer).Error)
	deviceType := dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		ManufacturerID: manufacturer.ID, Model: "Model " + roleID.String(), Slug: "model-" + roleID.String(),
		UHeight: 1, IsFullDepth: true,
	}
	require.NoError(t, db.Create(&deviceType).Error)
	deviceName := name
	deviceAssetTag := assetTag
	require.NoError(t, db.Create(&dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: deviceType.ID, RoleID: roleID.Int64(), Name: &deviceName,
		SiteID: site.ID, Status: "active", AssetTag: &deviceAssetTag,
	}).Error)
}
