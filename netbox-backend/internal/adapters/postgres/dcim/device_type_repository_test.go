package dcim

import (
	"encoding/json"
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

func TestDeviceTypeRepositoryMapsCountersFiltersNullableAirflowAndPlacements(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewDeviceTypeRepository(db)
	alpha := newManufacturerFixture(t, "Alpha", "alpha")
	beta := newManufacturerFixture(t, "Beta", "beta")
	require.NoError(t, manufacturerRepository.Create(t.Context(), alpha))
	require.NoError(t, manufacturerRepository.Create(t.Context(), beta))

	alphaRouter := newDeviceTypeFixture(
		t, alpha, "Router", "router", "1.5",
		domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
	)
	alphaSwitch := newDeviceTypeFixture(
		t, alpha, "Switch", "switch", "1", domaindcim.NonNullDeviceAirflow(""),
	)
	betaRouter := newDeviceTypeFixture(
		t, beta, "Router", "router", "2", domaindcim.NullDeviceAirflow(),
	)
	for _, deviceType := range []*domaindcim.DeviceType{alphaRouter, alphaSwitch, betaRouter} {
		require.NoError(t, repository.Create(t.Context(), deviceType))
	}

	require.NoError(t, db.Create(&dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: alphaRouter.ID().Int64(), Name: "eth0", Type: "1000base-t",
		Enabled: true,
	}).Error)
	siteID, roleID := seedDeviceTypeDeviceDependencies(t, db)
	rack := dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		SiteID: siteID, Name: "Rack 1", Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}
	require.NoError(t, db.Create(&rack).Error)
	deviceName := "edge-1"
	position := 10.5
	require.NoError(t, db.Create(&dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: alphaRouter.ID().Int64(), RoleID: roleID,
		Name: &deviceName, SiteID: siteID, RackID: &rack.ID, Position: &position,
		Face: "front", Status: "active",
	}).Error)

	loaded, err := repository.Get(t.Context(), alphaRouter.ID())
	require.NoError(t, err)
	assert.Equal(t, alpha.ID(), loaded.Manufacturer().ID())
	assert.Equal(t, uint64(1), loaded.DeviceCount())
	assert.Equal(t, uint64(1), loaded.InterfaceTemplateCount())
	assert.Equal(t, "1.5", loaded.UHeight().String())
	airflow, present := loaded.Airflow().Get()
	assert.True(t, present)
	assert.Equal(t, domaindcim.DeviceAirflowFrontToRear, airflow)

	blank, err := repository.Get(t.Context(), alphaSwitch.ID())
	require.NoError(t, err)
	blankAirflow, present := blank.Airflow().Get()
	assert.True(t, present)
	assert.Empty(t, blankAirflow)
	nullValue, err := repository.Get(t.Context(), betaRouter.ID())
	require.NoError(t, err)
	assert.True(t, nullValue.Airflow().IsNull())

	page, err := repository.List(t.Context(), applicationdcim.DeviceTypeListCriteria{
		Limit: 50,
		IDs: []int64{
			-1, alphaRouter.ID().Int64(), alphaSwitch.ID().Int64(),
		},
		ManufacturerIDs:   []int64{-1, alpha.ID().Int64()},
		ManufacturerSlugs: []string{"missing", "alpha"},
		Models:            []string{"Router", "Switch"},
		Slugs:             []string{"router", "switch"},
		Ordering: []applicationdcim.DeviceTypeSort{
			{Field: applicationdcim.DeviceTypeSortModel},
			{Field: applicationdcim.DeviceTypeSortID},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, []string{"Router", "Switch"}, []string{
		page.Results[0].Model(), page.Results[1].Model(),
	})

	placements, err := repository.ListPositionedDevicesForUpdate(t.Context())
	require.NoError(t, err)
	require.Len(t, placements, 1)
	assert.Equal(t, alphaRouter.ID(), placements[0].DeviceTypeID)
	assert.Equal(t, uint32(21), placements[0].PositionHalfUnits)
	assert.Equal(t, uint32(3), placements[0].StoredHeightHalfUnits)
	assert.True(t, placements[0].StoredFullDepth)
	assert.Equal(t, uint32(1), placements[0].RackStartingUnit)
	assert.Equal(t, uint32(42), placements[0].RackUHeight)
}

func TestDeviceTypeRepositoryEnforcesManufacturerScopedModelAndSlugUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewDeviceTypeRepository(db)
	alpha := newManufacturerFixture(t, "Alpha", "alpha")
	beta := newManufacturerFixture(t, "Beta", "beta")
	require.NoError(t, manufacturerRepository.Create(t.Context(), alpha))
	require.NoError(t, manufacturerRepository.Create(t.Context(), beta))
	original := newDeviceTypeFixture(
		t, alpha, "Router", "router", "1", domaindcim.NullDeviceAirflow(),
	)
	require.NoError(t, repository.Create(t.Context(), original))

	for _, test := range []struct {
		candidate *domaindcim.DeviceType
		want      shared.FieldViolation
	}{
		{
			candidate: newDeviceTypeFixture(
				t, alpha, "Router", "different", "1", domaindcim.NullDeviceAirflow(),
			),
			want: shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: "Device type with this Manufacturer and Model already exists.",
			},
		},
		{
			candidate: newDeviceTypeFixture(
				t, alpha, "Different", "router", "1", domaindcim.NullDeviceAirflow(),
			),
			want: shared.FieldViolation{
				Field: "non_field_errors", Reason: "unique_together",
				Description: "Device type with this Manufacturer and Slug already exists.",
			},
		},
	} {
		err := repository.Create(t.Context(), test.candidate)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, []shared.FieldViolation{test.want}, shared.ViolationsOf(err))
		assert.False(t, test.candidate.ID().IsValid())
	}

	sameValuesDifferentManufacturer := newDeviceTypeFixture(
		t, beta, "Router", "router", "1", domaindcim.NullDeviceAirflow(),
	)
	require.NoError(t, repository.Create(t.Context(), sameValuesDifferentManufacturer))
	assert.True(t, sameValuesDifferentManufacturer.ID().IsValid())
}

func TestDeviceTypeServiceProtectsDeviceThenCascadesTemplatesWithTypedAudit(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewDeviceTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newDeviceTypeFixture(
		t, manufacturer, "Router", "router", "1", domaindcim.NullDeviceAirflow(),
	)
	require.NoError(t, repository.Create(t.Context(), deviceType))
	templateIDs := seedDeviceTypeTemplates(t, db, deviceType.ID())
	siteID, roleID := seedDeviceTypeDeviceDependencies(t, db)
	name, assetTag := "edge-1", "EDGE-1"
	device := dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: deviceType.ID().Int64(), RoleID: roleID, Name: &name,
		SiteID: siteID, Status: "active", AssetTag: &assetTag,
	}
	require.NoError(t, db.Create(&device).Error)
	service := newPostgresDeviceTypeService(
		t, db, repository, manufacturerRepository, postgreschangelog.NewRecorder(db),
	)

	err := service.DeleteDeviceType(
		t.Context(), deviceTypePrincipal(),
		applicationdcim.DeleteDeviceTypeCommand{ID: deviceType.ID()},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	assert.Contains(
		t, err.Error(),
		"Unable to delete object. 1 dependent objects were found: edge-1 (EDGE-1)",
	)
	assertDeviceTypeAndTemplatesRemain(t, db, repository, deviceType.ID(), len(templateIDs))
	var changeCount int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changeCount).Error)
	assert.Zero(t, changeCount)

	require.NoError(t, db.Delete(&device).Error)
	require.NoError(t, service.DeleteDeviceType(
		t.Context(), deviceTypePrincipal(),
		applicationdcim.DeleteDeviceTypeCommand{ID: deviceType.ID()},
	))
	_, getErr := repository.Get(t.Context(), deviceType.ID())
	require.Error(t, getErr)
	assert.True(t, shared.HasReason(getErr, shared.ErrorReasonNotFound))
	var templateCount int64
	require.NoError(t, db.Model(&dcimrow.InterfaceTemplateRow{}).
		Where("device_type_id = ?", deviceType.ID().Int64()).Count(&templateCount).Error)
	assert.Zero(t, templateCount)

	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&changes).Error)
	require.Len(t, changes, 3)
	assert.Equal(t, []string{
		domaindcim.InterfaceTemplateObjectType,
		domaindcim.InterfaceTemplateObjectType,
		domaindcim.DeviceTypeObjectType,
	}, []string{changes[0].Kind, changes[1].Kind, changes[2].Kind})
	assert.Equal(t, []int64{
		templateIDs[0].Int64(), templateIDs[1].Int64(), deviceType.ID().Int64(),
	}, []int64{changes[0].ObjectID, changes[1].ObjectID, changes[2].ObjectID})
	var firstTemplate map[string]any
	require.NoError(t, json.Unmarshal(changes[0].BeforeData, &firstTemplate))
	assert.Equal(t, float64(deviceType.ID().Int64()), firstTemplate["device_type"])
	assert.Equal(t, "eth0", firstTemplate["name"])
	var parent map[string]any
	require.NoError(t, json.Unmarshal(changes[2].BeforeData, &parent))
	assert.Equal(t, float64(1), parent["u_height"])
	assert.Equal(t, "Router", parent["model"])
}

func TestDeviceTypeTemplateCascadeAndAuditRollbackAtomically(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewDeviceTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newDeviceTypeFixture(
		t, manufacturer, "Router", "router", "1", domaindcim.NullDeviceAirflow(),
	)
	require.NoError(t, repository.Create(t.Context(), deviceType))
	seedDeviceTypeTemplates(t, db, deviceType.ID())
	sentinel := errors.New("forced InterfaceTemplate audit failure")
	service := newPostgresDeviceTypeService(
		t, db, repository, manufacturerRepository,
		&failOnRecorder{
			delegate: postgreschangelog.NewRecorder(db), failAt: 2, err: sentinel,
		},
	)

	err := service.DeleteDeviceType(
		t.Context(), deviceTypePrincipal(),
		applicationdcim.DeleteDeviceTypeCommand{ID: deviceType.ID()},
	)
	require.ErrorIs(t, err, sentinel)
	assertDeviceTypeAndTemplatesRemain(t, db, repository, deviceType.ID(), 2)
	var changeCount int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changeCount).Error)
	assert.Zero(t, changeCount)
}

func newDeviceTypeFixture(
	t *testing.T,
	manufacturer *domaindcim.Manufacturer,
	model string,
	slug string,
	height string,
	airflow domaindcim.NullableDeviceAirflow,
) *domaindcim.DeviceType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(
		manufacturer.ID(), manufacturer.Name(), manufacturer.Slug().String(),
	)
	require.NoError(t, err)
	deviceType, err := domaindcim.NewDeviceType(domaindcim.DeviceTypeValues{
		Manufacturer: reference, Model: model, Slug: slug, PartNumber: "PN-" + slug,
		UHeight: height, IsFullDepth: true, Airflow: airflow,
		Description: "Type description", Comments: "Type comments",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return deviceType
}

func seedDeviceTypeDeviceDependencies(t *testing.T, db *gorm.DB) (int64, int64) {
	t.Helper()
	site := dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "DeviceType Site", Slug: "device-type-site", Status: "active",
	}
	require.NoError(t, db.Create(&site).Error)
	role := dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "DeviceType Role", Slug: "device-type-role", Color: "9e9e9e",
		VMRole: true,
	}
	require.NoError(t, db.Create(&role).Error)
	return site.ID, role.ID
}

func seedDeviceTypeTemplates(
	t *testing.T,
	db *gorm.DB,
	deviceTypeID shared.ID,
) []shared.ID {
	t.Helper()
	ids := make([]shared.ID, 0, 2)
	for index, name := range []string{"eth0", "eth1"} {
		template := dcimrow.InterfaceTemplateRow{
			RowMetadata: dcimrow.RowMetadata{
				Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
			},
			DeviceTypeID: deviceTypeID.Int64(), Name: name,
			Label: "port-" + name, Type: "1000base-t", Enabled: true,
			MgmtOnly: index == 0, Description: name + " description",
		}
		require.NoError(t, db.Create(&template).Error)
		ids = append(ids, shared.ID(template.ID))
	}
	return ids
}

func newPostgresDeviceTypeService(
	t *testing.T,
	db *gorm.DB,
	repository *DeviceTypeRepository,
	manufacturers *ManufacturerRepository,
	recorder applicationchangelog.Recorder,
) *applicationdcim.DeviceTypeService {
	t.Helper()
	service, err := applicationdcim.NewDeviceTypeService(
		repository, manufacturers, postgresTransaction.NewUnitOfWork(db),
		recorder, authz.AllowAll{}, deviceTypeRepositoryClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	return service
}

func assertDeviceTypeAndTemplatesRemain(
	t *testing.T,
	db *gorm.DB,
	repository *DeviceTypeRepository,
	id shared.ID,
	wantTemplates int,
) {
	t.Helper()
	_, err := repository.Get(t.Context(), id)
	require.NoError(t, err)
	var count int64
	require.NoError(t, db.Model(&dcimrow.InterfaceTemplateRow{}).
		Where("device_type_id = ?", id.Int64()).Count(&count).Error)
	assert.Equal(t, int64(wantTemplates), count)
}

type deviceTypeRepositoryClock struct{ now shared.Timestamp }

func (clock deviceTypeRepositoryClock) Now() shared.Timestamp { return clock.now }

func deviceTypePrincipal() identity.Principal {
	return identity.Principal{
		ID: 17, Username: "device-type-auditor", IsSuperuser: true,
	}
}
