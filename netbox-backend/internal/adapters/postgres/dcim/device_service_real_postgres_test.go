package dcim

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestPostgresDeviceScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.October, 12, 9, 0, 0, 0, time.UTC)
	dependencies := seedPostgresDevicePresenceDependencies(t, db, createdAt)
	repository := NewDeviceRepository(db)
	service := newPostgresDevicePresenceService(
		t, db, createdAt, repository, postgreschangelog.NewRecorder(db),
	)

	device, err := service.CreateDevice(
		t.Context(), principal, applicationdcim.CreateDeviceCommand{
			DeviceType: applicationdcim.FieldValue(dependencies.deviceTypeA),
			Role:       applicationdcim.FieldValue(dependencies.role),
			Site:       applicationdcim.FieldValue(dependencies.site),
		},
	)
	require.NoError(t, err)
	frontToRear := domaindcim.DeviceAirflowFrontToRear.String()
	requirePostgresDevicePresenceAggregate(t, device, postgresDevicePresenceExpected{
		deviceTypeID: dependencies.deviceTypeA, roleID: dependencies.role,
		siteID: dependencies.site, status: "active", airflow: &frontToRear,
		created: createdAt, lastUpdated: createdAt,
	})
	created := loadPostgresDevicePresenceState(t, db, device.ID())
	requirePostgresDevicePresenceRow(t, created.row, dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: device.ID().Int64(), Created: createdAt, LastUpdated: createdAt,
		},
		DeviceTypeID: dependencies.deviceTypeA.Int64(), RoleID: dependencies.role.Int64(),
		SiteID: dependencies.site.Int64(), Status: "active", Airflow: &frontToRear,
	})
	require.Equal(t, int64(1), created.manufacturerCount)
	require.Equal(t, int64(1), created.siteCount)
	require.Equal(t, int64(1), created.rackCount)
	require.Equal(t, int64(1), created.deviceRoleCount)
	require.Equal(t, int64(2), created.deviceTypeCount)
	require.Equal(t, int64(1), created.deviceCount)
	require.Zero(t, created.interfaceTemplateCount)
	require.Zero(t, created.interfaceCount)
	require.Equal(t, int64(1), created.targetDeviceChangeCount)
	require.Equal(t, int64(1), created.allDeviceChangeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	requirePostgresDevicePresenceReloadEquals(t, db, device)

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresDevicePresenceService(
		t, db, patchedAt, repository, postgreschangelog.NewRecorder(db),
	)
	patched, err := service.UpdateDevice(
		t.Context(), principal, applicationdcim.UpdateDeviceCommand{
			ID: device.ID(), DeviceType: applicationdcim.FieldValue(dependencies.deviceTypeA),
			Role:     applicationdcim.FieldValue(dependencies.role),
			Name:     applicationdcim.FieldValue("  Durable Device  "),
			Site:     applicationdcim.FieldValue(dependencies.site),
			Rack:     applicationdcim.FieldValue(dependencies.rack),
			Position: applicationdcim.FieldValue("10"), Face: applicationdcim.FieldValue("front"),
			Status:      applicationdcim.FieldValue("planned"),
			Serial:      applicationdcim.FieldValue("  SERIAL-01  "),
			AssetTag:    applicationdcim.FieldValue("  ASSET-I10-01  "),
			Airflow:     applicationdcim.FieldValue("rear-to-front"),
			Description: applicationdcim.FieldValue("  durable description  "),
			Comments:    applicationdcim.FieldValue("  durable comments  "),
		},
	)
	require.NoError(t, err)
	name := "Durable Device"
	position := "10"
	serial := "SERIAL-01"
	assetTag := "ASSET-I10-01"
	rearToFront := domaindcim.DeviceAirflowRearToFront.String()
	description := "durable description"
	comments := "durable comments"
	requirePostgresDevicePresenceAggregate(t, patched, postgresDevicePresenceExpected{
		deviceTypeID: dependencies.deviceTypeA, roleID: dependencies.role,
		name: &name, siteID: dependencies.site, rackID: &dependencies.rack,
		position: &position, face: "front", status: "planned", serial: serial,
		assetTag: &assetTag, airflow: &rearToFront,
		description: description, comments: comments,
		created: createdAt, lastUpdated: patchedAt,
	})
	positionNumber := float64(10)
	patchedState := loadPostgresDevicePresenceState(t, db, device.ID())
	requirePostgresDevicePresenceRow(t, patchedState.row, dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: device.ID().Int64(), Created: createdAt, LastUpdated: patchedAt,
		},
		DeviceTypeID: dependencies.deviceTypeA.Int64(), RoleID: dependencies.role.Int64(),
		Name: &name, SiteID: dependencies.site.Int64(),
		RackID:   postgresDevicePresenceInt64Pointer(dependencies.rack.Int64()),
		Position: &positionNumber, Face: "front", Status: "planned", Serial: serial,
		AssetTag: &assetTag, Airflow: &rearToFront,
		Description: description, Comments: comments,
	})
	requirePostgresDevicePresenceMutation(t, created, patchedState, patchedAt)
	requirePostgresDevicePresenceReloadEquals(t, db, patched)

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresDevicePresenceService(
		t, db, replacedAt, repository, postgreschangelog.NewRecorder(db),
	)
	replaced, err := service.ReplaceDevice(
		t.Context(), principal, applicationdcim.ReplaceDeviceCommand{
			ID: device.ID(), CreateDeviceCommand: applicationdcim.CreateDeviceCommand{
				DeviceType: applicationdcim.FieldValue(dependencies.deviceTypeB),
				Role:       applicationdcim.FieldValue(dependencies.role),
				Site:       applicationdcim.FieldValue(dependencies.site),
			},
		},
	)
	require.NoError(t, err)
	requirePostgresDevicePresenceAggregate(t, replaced, postgresDevicePresenceExpected{
		deviceTypeID: dependencies.deviceTypeB, roleID: dependencies.role,
		name: &name, siteID: dependencies.site, status: "planned", serial: serial,
		assetTag: &assetTag, airflow: &rearToFront,
		description: description, comments: comments,
		created: createdAt, lastUpdated: replacedAt,
	})
	replacedState := loadPostgresDevicePresenceState(t, db, device.ID())
	requirePostgresDevicePresenceRow(t, replacedState.row, dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: device.ID().Int64(), Created: createdAt, LastUpdated: replacedAt,
		},
		DeviceTypeID: dependencies.deviceTypeB.Int64(), RoleID: dependencies.role.Int64(),
		Name: &name, SiteID: dependencies.site.Int64(), Status: "planned", Serial: serial,
		AssetTag: &assetTag, Airflow: &rearToFront,
		Description: description, Comments: comments,
	})
	requirePostgresDevicePresenceMutation(t, patchedState, replacedState, replacedAt)
	requirePostgresDevicePresenceReloadEquals(t, db, replaced)

	clearedAt := replacedAt.Add(time.Minute)
	service = newPostgresDevicePresenceService(
		t, db, clearedAt, repository, postgreschangelog.NewRecorder(db),
	)
	cleared, err := service.UpdateDevice(
		t.Context(), principal, applicationdcim.UpdateDeviceCommand{
			ID: device.ID(), Name: applicationdcim.NullField[string](),
			Face: applicationdcim.NullField[string](), Serial: applicationdcim.FieldValue(""),
			AssetTag: applicationdcim.FieldValue(""), Airflow: applicationdcim.NullField[string](),
			Description: applicationdcim.FieldValue(""), Comments: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	blank := ""
	requirePostgresDevicePresenceAggregate(t, cleared, postgresDevicePresenceExpected{
		deviceTypeID: dependencies.deviceTypeB, roleID: dependencies.role,
		siteID: dependencies.site, status: "planned", assetTag: &blank, airflow: &blank,
		created: createdAt, lastUpdated: clearedAt,
	})
	clearedState := loadPostgresDevicePresenceState(t, db, device.ID())
	requirePostgresDevicePresenceRow(t, clearedState.row, dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: device.ID().Int64(), Created: createdAt, LastUpdated: clearedAt,
		},
		DeviceTypeID: dependencies.deviceTypeB.Int64(), RoleID: dependencies.role.Int64(),
		SiteID: dependencies.site.Int64(), Status: "planned", AssetTag: &blank, Airflow: &blank,
	})
	requirePostgresDevicePresenceMutation(t, replacedState, clearedState, clearedAt)
	requirePostgresDevicePresenceReloadEquals(t, db, cleared)

	swappedAt := clearedAt.Add(time.Minute)
	service = newPostgresDevicePresenceService(
		t, db, swappedAt, repository, postgreschangelog.NewRecorder(db),
	)
	swapped, err := service.UpdateDevice(
		t.Context(), principal, applicationdcim.UpdateDeviceCommand{
			ID: device.ID(), Name: applicationdcim.FieldValue(""),
			AssetTag: applicationdcim.NullField[string](),
		},
	)
	require.NoError(t, err)
	requirePostgresDevicePresenceAggregate(t, swapped, postgresDevicePresenceExpected{
		deviceTypeID: dependencies.deviceTypeB, roleID: dependencies.role,
		name: &blank, siteID: dependencies.site, status: "planned", airflow: &blank,
		created: createdAt, lastUpdated: swappedAt,
	})
	swappedState := loadPostgresDevicePresenceState(t, db, device.ID())
	requirePostgresDevicePresenceRow(t, swappedState.row, dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: device.ID().Int64(), Created: createdAt, LastUpdated: swappedAt,
		},
		DeviceTypeID: dependencies.deviceTypeB.Int64(), RoleID: dependencies.role.Int64(),
		Name: &blank, SiteID: dependencies.site.Int64(), Status: "planned", Airflow: &blank,
	})
	requirePostgresDevicePresenceMutation(t, clearedState, swappedState, swappedAt)
	requirePostgresDevicePresenceReloadEquals(t, db, swapped)

	validationAt := swappedAt.Add(time.Minute)
	service = newPostgresDevicePresenceService(
		t, db, validationAt, repository, postgreschangelog.NewRecorder(db),
	)
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresDevicePresenceState(t, db, device.ID())
			operationErr := operation()
			require.Error(t, operationErr)
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(t, before, loadPostgresDevicePresenceState(t, db, device.ID()))
		})
	}
	assertRejected("POST missing required relationships", func() error {
		_, operationErr := service.CreateDevice(
			t.Context(), principal, applicationdcim.CreateDeviceCommand{},
		)
		return operationErr
	})
	assertRejected("PUT missing required relationships", func() error {
		_, operationErr := service.ReplaceDevice(
			t.Context(), principal, applicationdcim.ReplaceDeviceCommand{ID: device.ID()},
		)
		return operationErr
	})
	assertRejected("PATCH combined scalar and relationship violations", func() error {
		_, operationErr := service.UpdateDevice(
			t.Context(), principal, applicationdcim.UpdateDeviceCommand{
				ID: device.ID(), Role: applicationdcim.FieldValue(shared.ID(999999)),
				Status:  applicationdcim.FieldValue(" active "),
				Face:    applicationdcim.FieldValue(" front "),
				Airflow: applicationdcim.FieldValue(" rear-to-front "),
			},
		)
		return operationErr
	})

	repositoryFailureBefore := loadPostgresDevicePresenceState(t, db, device.ID())
	repositoryFailure := errors.New("forced Device repository failure after delegate")
	failingRepository := &delegateThenErrorDevicePresenceRepository{
		DeviceRepository: repository, err: repositoryFailure,
	}
	service = newPostgresDevicePresenceService(
		t, db, validationAt.Add(time.Minute), failingRepository,
		postgreschangelog.NewRecorder(db),
	)
	_, err = service.UpdateDevice(
		t.Context(), principal, applicationdcim.UpdateDeviceCommand{
			ID: device.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, repositoryFailure)
	require.Equal(t, 1, failingRepository.updateCalls)
	require.Equal(
		t, repositoryFailureBefore, loadPostgresDevicePresenceState(t, db, device.ID()),
		"delegate-then-error repository failure must roll back row and changes",
	)

	recorderFailureBefore := loadPostgresDevicePresenceState(t, db, device.ID())
	recorderFailure := errors.New("forced Device recorder failure after delegate")
	failingRecorder := &delegateThenErrorDevicePresenceRecorder{
		delegate: postgreschangelog.NewRecorder(db), err: recorderFailure,
	}
	service = newPostgresDevicePresenceService(
		t, db, validationAt.Add(2*time.Minute), repository, failingRecorder,
	)
	_, err = service.UpdateDevice(
		t.Context(), principal, applicationdcim.UpdateDeviceCommand{
			ID: device.ID(), Description: applicationdcim.FieldValue("must also roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(
		t, recorderFailureBefore, loadPostgresDevicePresenceState(t, db, device.ID()),
		"delegate-then-error recorder failure must roll back row and inserted change",
	)
}

type postgresDevicePresenceDependencies struct {
	manufacturer shared.ID
	site         shared.ID
	rack         shared.ID
	role         shared.ID
	deviceTypeA  shared.ID
	deviceTypeB  shared.ID
}

func seedPostgresDevicePresenceDependencies(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) postgresDevicePresenceDependencies {
	t.Helper()
	manufacturer := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Device Manufacturer", Slug: "presence-device-manufacturer",
	}
	require.NoError(t, db.Create(&manufacturer).Error)
	site := dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Device Site", Slug: "presence-device-site", Status: "active",
	}
	require.NoError(t, db.Create(&site).Error)
	rack := dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		SiteID:      site.ID, Name: "Presence Device Rack", Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}
	require.NoError(t, db.Create(&rack).Error)
	role := dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Device Role", Slug: "presence-device-role", Color: "00ff00",
	}
	require.NoError(t, db.Create(&role).Error)
	frontToRear := domaindcim.DeviceAirflowFrontToRear.String()
	deviceTypeA := dcimrow.DeviceTypeRow{
		RowMetadata:    dcimrow.RowMetadata{Created: now, LastUpdated: now},
		ManufacturerID: manufacturer.ID, Model: "Presence Device Type A",
		Slug: "presence-device-type-a", UHeight: 1, IsFullDepth: true,
		Airflow: &frontToRear,
	}
	require.NoError(t, db.Create(&deviceTypeA).Error)
	passive := domaindcim.DeviceAirflowPassive.String()
	deviceTypeB := dcimrow.DeviceTypeRow{
		RowMetadata:    dcimrow.RowMetadata{Created: now, LastUpdated: now},
		ManufacturerID: manufacturer.ID, Model: "Presence Device Type B",
		Slug: "presence-device-type-b", UHeight: 1, IsFullDepth: true,
		Airflow: &passive,
	}
	require.NoError(t, db.Create(&deviceTypeB).Error)
	return postgresDevicePresenceDependencies{
		manufacturer: shared.ID(manufacturer.ID), site: shared.ID(site.ID),
		rack: shared.ID(rack.ID), role: shared.ID(role.ID),
		deviceTypeA: shared.ID(deviceTypeA.ID), deviceTypeB: shared.ID(deviceTypeB.ID),
	}
}

type postgresDevicePresenceState struct {
	row                     dcimrow.DeviceRow
	manufacturerCount       int64
	siteCount               int64
	rackCount               int64
	deviceRoleCount         int64
	deviceTypeCount         int64
	interfaceTemplateCount  int64
	deviceCount             int64
	interfaceCount          int64
	targetDeviceChangeCount int64
	allDeviceChangeCount    int64
	totalChangeCount        int64
}

func loadPostgresDevicePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresDevicePresenceState {
	t.Helper()
	var state postgresDevicePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.ManufacturerRow{}).Count(&state.manufacturerCount).Error)
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&state.siteCount).Error)
	require.NoError(t, db.Model(&dcimrow.RackRow{}).Count(&state.rackCount).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceRoleRow{}).Count(&state.deviceRoleCount).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceTypeRow{}).Count(&state.deviceTypeCount).Error)
	require.NoError(
		t, db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&state.interfaceTemplateCount).Error,
	)
	require.NoError(t, db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, db.Model(&dcimrow.InterfaceRow{}).Count(&state.interfaceCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceObjectType, id.Int64(),
	).Count(&state.targetDeviceChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceObjectType,
	).Count(&state.allDeviceChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresDevicePresenceMutation(
	t *testing.T,
	before postgresDevicePresenceState,
	after postgresDevicePresenceState,
	expectedLastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, before.manufacturerCount, after.manufacturerCount)
	require.Equal(t, before.siteCount, after.siteCount)
	require.Equal(t, before.rackCount, after.rackCount)
	require.Equal(t, before.deviceRoleCount, after.deviceRoleCount)
	require.Equal(t, before.deviceTypeCount, after.deviceTypeCount)
	require.Equal(t, before.interfaceTemplateCount, after.interfaceTemplateCount)
	require.Zero(t, after.interfaceTemplateCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, int64(1), after.deviceCount)
	require.Equal(t, before.interfaceCount, after.interfaceCount)
	require.Zero(t, after.interfaceCount)
	require.Equal(t, before.targetDeviceChangeCount+1, after.targetDeviceChangeCount)
	require.Equal(t, before.allDeviceChangeCount+1, after.allDeviceChangeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.True(t, after.row.Created.Equal(before.row.Created))
	require.True(t, after.row.LastUpdated.Equal(expectedLastUpdated))
}

func requirePostgresDevicePresenceRow(
	t *testing.T,
	actual dcimrow.DeviceRow,
	expected dcimrow.DeviceRow,
) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	require.Equal(t, expected.DeviceTypeID, actual.DeviceTypeID)
	require.Equal(t, expected.RoleID, actual.RoleID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.SiteID, actual.SiteID)
	require.Equal(t, expected.RackID, actual.RackID)
	require.Equal(t, expected.Position, actual.Position)
	require.Equal(t, expected.Face, actual.Face)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.Serial, actual.Serial)
	require.Equal(t, expected.AssetTag, actual.AssetTag)
	require.Equal(t, expected.Airflow, actual.Airflow)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

type postgresDevicePresenceExpected struct {
	deviceTypeID shared.ID
	roleID       shared.ID
	name         *string
	siteID       shared.ID
	rackID       *shared.ID
	position     *string
	face         string
	status       string
	serial       string
	assetTag     *string
	airflow      *string
	description  string
	comments     string
	created      time.Time
	lastUpdated  time.Time
}

func requirePostgresDevicePresenceAggregate(
	t *testing.T,
	device *domaindcim.Device,
	expected postgresDevicePresenceExpected,
) {
	t.Helper()
	require.Equal(t, expected.deviceTypeID, device.DeviceType().ID())
	require.Equal(t, expected.roleID, device.Role().ID)
	requireDevicePresenceNullableString(t, device.Name(), expected.name)
	require.Equal(t, expected.siteID, device.Site().ID())
	if expected.rackID == nil {
		require.True(t, device.Rack().IsNull())
	} else {
		rack, present := device.Rack().Get()
		require.True(t, present)
		require.Equal(t, *expected.rackID, rack.ID())
	}
	if expected.position == nil {
		require.True(t, device.Position().IsNull())
	} else {
		position, present := device.Position().Get()
		require.True(t, present)
		require.Equal(t, *expected.position, position.String())
	}
	require.Equal(t, expected.face, device.Face().String())
	require.Equal(t, expected.status, device.Status().String())
	require.Equal(t, expected.serial, device.Serial())
	requireDevicePresenceNullableString(t, device.AssetTag(), expected.assetTag)
	if expected.airflow == nil {
		require.True(t, device.Airflow().IsNull())
	} else {
		airflow, present := device.Airflow().Get()
		require.True(t, present)
		require.Equal(t, *expected.airflow, airflow.String())
	}
	require.Equal(t, expected.description, device.Description())
	require.Equal(t, expected.comments, device.Comments())
	require.True(t, device.Created().Equal(expected.created))
	require.True(t, device.LastUpdated().Equal(expected.lastUpdated))
	require.Zero(t, device.InterfaceCount())
}

func requireDevicePresenceNullableString(
	t *testing.T,
	actual domaindcim.DeviceNullable[string],
	expected *string,
) {
	t.Helper()
	if expected == nil {
		require.True(t, actual.IsNull())
		return
	}
	value, present := actual.Get()
	require.True(t, present)
	require.Equal(t, *expected, value)
}

func requirePostgresDevicePresenceReloadEquals(
	t *testing.T,
	db *gorm.DB,
	expected *domaindcim.Device,
) {
	t.Helper()
	reloaded, err := NewDeviceRepository(db).Get(t.Context(), expected.ID())
	require.NoError(t, err)
	require.Equal(t, expected.State(), reloaded.State())
}

type delegateThenErrorDevicePresenceRepository struct {
	applicationdcim.DeviceRepository
	err         error
	updateCalls int
}

func (repository *delegateThenErrorDevicePresenceRepository) Update(
	ctx context.Context,
	device *domaindcim.Device,
) error {
	repository.updateCalls++
	if err := repository.DeviceRepository.Update(ctx, device); err != nil {
		return err
	}
	return repository.err
}

type delegateThenErrorDevicePresenceRecorder struct {
	delegate applicationchangelog.Recorder
	err      error
	calls    int
}

func (recorder *delegateThenErrorDevicePresenceRecorder) Record(
	ctx context.Context,
	change applicationchangelog.Change,
) error {
	recorder.calls++
	if err := recorder.delegate.Record(ctx, change); err != nil {
		return err
	}
	return recorder.err
}

type postgresDevicePresenceCascade struct{}

func (postgresDevicePresenceCascade) DeleteForDevice(
	context.Context,
	shared.ID,
	shared.Timestamp,
) ([]applicationdcim.InterfaceCascadeChange, error) {
	return nil, nil
}

func newPostgresDevicePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	repository applicationdcim.DeviceRepository,
	recorder applicationchangelog.Recorder,
) *applicationdcim.DeviceService {
	t.Helper()
	interfaces := NewInterfaceRepository(db)
	service, err := applicationdcim.NewDeviceService(
		repository, NewDeviceTypeRepository(db), NewDeviceRoleRepository(db),
		NewSiteRepository(db), NewRackRepository(db), NewInterfaceTemplateRepository(db),
		interfaces, postgresDevicePresenceCascade{}, postgresTransaction.NewUnitOfWork(db),
		recorder, authz.AllowAll{}, postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}

func postgresDevicePresenceInt64Pointer(value int64) *int64 { return &value }
