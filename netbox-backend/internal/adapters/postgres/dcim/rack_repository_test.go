package dcim

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestRackRepositoryProjectsRelationshipsCountersFiltersAndNullableBlanks(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	site, rackType, role := seedRackDependencies(t, db, "Moscow", "moscow")
	repository := NewRackRepository(db)

	rack := newRackFixture(
		t, site, rackType, role, "A01",
		domaindcim.NullRackValue[string](),
		domaindcim.NonNullRackValue(""),
	)
	require.NoError(t, repository.Create(t.Context(), rack))
	device := seedRackDevice(t, db, site.ID(), rack.ID(), rackType.Manufacturer().ID(), "edge-1", 10.5)

	loaded, err := repository.Get(t.Context(), rack.ID())
	require.NoError(t, err)
	assert.Equal(t, site.ID(), loaded.Site().ID())
	assert.Equal(t, "Moscow", loaded.Site().Name())
	loadedType, present := loaded.RackType().Get()
	require.True(t, present)
	assert.Equal(t, rackType.ID(), loadedType.ID())
	assert.Equal(t, "R42", loadedType.Model())
	loadedRole, present := loaded.Role().Get()
	require.True(t, present)
	assert.Equal(t, role.ID(), loadedRole.ID())
	assert.True(t, loaded.FacilityID().IsNull())
	assert.Equal(t, "", rackRepositoryNullableValue(t, loaded.AssetTag()))
	assert.Equal(t, uint64(1), loaded.DeviceCount())
	assert.Positive(t, device.ID)

	page, err := repository.List(t.Context(), applicationdcim.RackListCriteria{
		Limit:   50,
		IDs:     []int64{-1, rack.ID().Int64()},
		SiteIDs: []int64{-1, site.ID().Int64()}, SiteSlugs: []string{"missing", "moscow"},
		Names:    []string{"missing", "A01"},
		Statuses: []domaindcim.RackStatus{domaindcim.RackStatusActive},
		RoleIDs:  []int64{-1, role.ID().Int64()}, RoleSlugs: []string{"missing", "production"},
		RackTypeIDs: []int64{-1, rackType.ID().Int64()}, RackTypeSlugs: []string{"missing", "r42"},
		Ordering: []applicationdcim.RackSort{{Field: applicationdcim.RackSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, rack.ID(), page.Results[0].ID())
	assert.Equal(t, uint64(1), page.Results[0].DeviceCount())
}

func TestRackRepositoryEnforcesBlankAssetTagUniquenessAndProtectedDelete(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	site, rackType, role := seedRackDependencies(t, db, "Moscow", "moscow")
	repository := NewRackRepository(db)

	protected := newRackFixture(
		t, site, rackType, role, "A01",
		domaindcim.NullRackValue[string](),
		domaindcim.NonNullRackValue(""),
	)
	require.NoError(t, repository.Create(t.Context(), protected))
	duplicateBlank := newRackFixture(
		t, site, rackType, role, "A02",
		domaindcim.NullRackValue[string](),
		domaindcim.NonNullRackValue(""),
	)
	err := repository.Create(t.Context(), duplicateBlank)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, []shared.FieldViolation{{
		Field: "asset_tag", Reason: "unique",
		Description: "rack with this asset tag already exists.",
	}}, shared.ViolationsOf(err))

	for _, name := range []string{"A03", "A04"} {
		candidate := newRackFixture(
			t, site, rackType, role, name,
			domaindcim.NullRackValue[string](),
			domaindcim.NullRackValue[string](),
		)
		require.NoError(t, repository.Create(t.Context(), candidate), "multiple NULL asset tags are allowed")
	}

	seedRackDevice(t, db, site.ID(), protected.ID(), rackType.Manufacturer().ID(), "edge-1", 10)
	err = repository.Delete(t.Context(), protected)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	_, getErr := repository.Get(t.Context(), protected.ID())
	require.NoError(t, getErr)
}

func TestRackRepositoryPropagatesDeviceSitesAtomicallyAndTranslatesNameConflict(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	oldSite, rackType, role := seedRackDependencies(t, db, "Old", "old")
	newSite := newSiteFixture(t, "New", "new", "")
	require.NoError(t, NewSiteRepository(db).Create(t.Context(), newSite))
	repository := NewRackRepository(db)
	rack := newRackFixture(
		t, oldSite, rackType, role, "A01",
		domaindcim.NullRackValue[string](),
		domaindcim.NullRackValue[string](),
	)
	require.NoError(t, repository.Create(t.Context(), rack))
	device := seedRackDevice(t, db, oldSite.ID(), rack.ID(), rackType.Manufacturer().ID(), "edge-1", 10)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)

	var changes []applicationdcim.RackSitePropagationChange
	err := unitOfWork.WithinTransaction(t.Context(), func(ctx context.Context) error {
		var propagationErr error
		changes, propagationErr = repository.PropagateSiteToDevices(
			ctx, rack.ID(), newSite.ID(), repositoryUpdatedAt,
		)
		return propagationErr
	})
	require.NoError(t, err)
	require.Len(t, changes, 1)
	before, ok := changes[0].Before.(domaindcim.DeviceSnapshot)
	require.True(t, ok)
	after, ok := changes[0].After.(domaindcim.DeviceSnapshot)
	require.True(t, ok)
	assert.Equal(t, oldSite.ID(), before.SiteID)
	assert.Equal(t, newSite.ID(), after.SiteID)

	// Return the Device to its old site, create a case-insensitive collision at
	// the target, then prove the failed propagation leaves the original site.
	require.NoError(t, db.Model(&dcimrow.DeviceRow{}).
		Where("id = ?", device.ID).
		Updates(map[string]any{"site_id": oldSite.ID().Int64()}).Error)
	seedRackDevice(t, db, newSite.ID(), 0, rackType.Manufacturer().ID(), "EDGE-1", 0)

	err = unitOfWork.WithinTransaction(t.Context(), func(ctx context.Context) error {
		_, propagationErr := repository.PropagateSiteToDevices(
			ctx, rack.ID(), newSite.ID(), repositoryUpdatedAt,
		)
		return propagationErr
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, "A device with this name already exists at this site.", err.Error())

	var persisted dcimrow.DeviceRow
	require.NoError(t, db.First(&persisted, device.ID).Error)
	assert.Equal(t, oldSite.ID().Int64(), persisted.SiteID)
}

func TestRackServiceRollsBackRackDeviceAndAuditOnSitePropagationConflict(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	oldSite, rackType, role := seedRackDependencies(t, db, "Old", "old")
	newSite := newSiteFixture(t, "New", "new", "")
	require.NoError(t, NewSiteRepository(db).Create(t.Context(), newSite))
	repository := NewRackRepository(db)
	rack := newRackFixture(
		t, oldSite, rackType, role, "A01",
		domaindcim.NullRackValue[string](),
		domaindcim.NullRackValue[string](),
	)
	require.NoError(t, repository.Create(t.Context(), rack))
	device := seedRackDevice(t, db, oldSite.ID(), rack.ID(), rackType.Manufacturer().ID(), "edge-1", 10)
	collision := seedRackDevice(t, db, newSite.ID(), 0, rackType.Manufacturer().ID(), "EDGE-1", 0)

	service, err := applicationdcim.NewRackService(
		repository,
		NewSiteRepository(db),
		NewRackTypeRepository(db),
		NewRackRoleRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		rackTypeClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)

	_, err = service.UpdateRack(
		t.Context(),
		rackTypePrincipal(),
		applicationdcim.UpdateRackCommand{
			ID: rack.ID(), Site: applicationdcim.FieldValue(newSite.ID()),
		},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))

	rolledBackRack, getErr := repository.Get(t.Context(), rack.ID())
	require.NoError(t, getErr)
	assert.Equal(t, oldSite.ID(), rolledBackRack.Site().ID())
	var rolledBackDevice dcimrow.DeviceRow
	require.NoError(t, db.First(&rolledBackDevice, device.ID).Error)
	assert.Equal(t, oldSite.ID().Int64(), rolledBackDevice.SiteID)
	assertTableCount(t, db, &postgreschangelog.ChangeRow{}, 0)

	require.NoError(t, db.Delete(&collision).Error)
	updated, err := service.UpdateRack(
		t.Context(),
		rackTypePrincipal(),
		applicationdcim.UpdateRackCommand{
			ID: rack.ID(), Site: applicationdcim.FieldValue(newSite.ID()),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, newSite.ID(), updated.Site().ID())
	require.NoError(t, db.First(&rolledBackDevice, device.ID).Error)
	assert.Equal(t, newSite.ID().Int64(), rolledBackDevice.SiteID)

	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&changes).Error)
	require.Len(t, changes, 2)
	assert.Equal(t, domaindcim.RackObjectType, changes[0].Kind)
	assert.Equal(t, domaindcim.DeviceObjectType, changes[1].Kind)
}

func seedRackDependencies(
	t *testing.T,
	db *gorm.DB,
	siteName string,
	siteSlug string,
) (*domaindcim.Site, *domaindcim.RackType, *domaindcim.RackRole) {
	t.Helper()
	site := newSiteFixture(t, siteName, siteSlug, "")
	require.NoError(t, NewSiteRepository(db).Create(t.Context(), site))
	manufacturer := newManufacturerFixture(t, siteName+" Vendor", siteSlug+"-vendor")
	require.NoError(t, NewManufacturerRepository(db).Create(t.Context(), manufacturer))
	rackType := newRackTypeFixture(t, manufacturer, "R42", "r42", 42)
	require.NoError(t, NewRackTypeRepository(db).Create(t.Context(), rackType))
	role := newRackRoleFixture(t, "Production", "production", "00ff00")
	require.NoError(t, NewRackRoleRepository(db).Create(t.Context(), role))
	return site, rackType, role
}

func newRackFixture(
	t *testing.T,
	site *domaindcim.Site,
	rackType *domaindcim.RackType,
	role *domaindcim.RackRole,
	name string,
	facilityID domaindcim.RackNullable[string],
	assetTag domaindcim.RackNullable[string],
) *domaindcim.Rack {
	t.Helper()
	siteReference, err := domaindcim.NewSiteReference(site.ID(), site.Name(), site.Slug().String())
	require.NoError(t, err)
	rackTypeReference, err := domaindcim.NewRackTypeReference(
		rackType.ID(), rackType.Model(), rackType.Slug().String(), rackType.PhysicalAttributes(),
	)
	require.NoError(t, err)
	roleReference, err := domaindcim.NewRackRoleReference(role.ID(), role.Name(), role.Slug().String())
	require.NoError(t, err)
	rack, err := domaindcim.NewRack(domaindcim.RackValues{
		Site: siteReference, Name: name, FacilityID: facilityID,
		RackType: domaindcim.NonNullRackValue(rackTypeReference),
		Status:   "active", Role: domaindcim.NonNullRackValue(roleReference),
		Serial: "rack serial", AssetTag: assetTag,
		FormFactor: domaindcim.NonNullRackValue(rackType.FormFactor().String()),
		Width:      rackType.Width().Uint32(), UHeight: rackType.UHeight(),
		StartingUnit: rackType.StartingUnit(), DescUnits: rackType.DescUnits(),
		Airflow:     domaindcim.NonNullRackValue(""),
		Description: "rack description", Comments: "rack comments",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return rack
}

func seedRackDevice(
	t *testing.T,
	db *gorm.DB,
	siteID shared.ID,
	rackID shared.ID,
	manufacturerID shared.ID,
	name string,
	position float64,
) dcimrow.DeviceRow {
	t.Helper()
	role := dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "Role " + name, Slug: "role-" + name, Color: "9e9e9e",
	}
	require.NoError(t, db.Create(&role).Error)
	deviceType := dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		ManufacturerID: manufacturerID.Int64(),
		Model:          "Type " + name, Slug: "type-" + name,
		UHeight: 2, IsFullDepth: true,
	}
	require.NoError(t, db.Create(&deviceType).Error)
	var persistedRackID *int64
	var persistedPosition *float64
	if rackID.IsValid() {
		value := rackID.Int64()
		persistedRackID = &value
		persistedPosition = &position
	}
	device := dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: deviceType.ID, RoleID: role.ID, Name: &name,
		SiteID: siteID.Int64(), RackID: persistedRackID, Position: persistedPosition,
		Face: "front", Status: "active",
	}
	require.NoError(t, db.Create(&device).Error)
	return device
}

func rackRepositoryNullableValue[T any](
	t *testing.T,
	value domaindcim.RackNullable[T],
) T {
	t.Helper()
	out, present := value.Get()
	require.True(t, present)
	return out
}
