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

func TestPostgresDeviceTypeScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	manufacturerA := seedDeviceTypePresenceManufacturer(
		t, db, createdAt, "Presence Device Manufacturer A", "presence-device-manufacturer-a",
	)
	manufacturerB := seedDeviceTypePresenceManufacturer(
		t, db, createdAt, "Presence Device Manufacturer B", "presence-device-manufacturer-b",
	)
	service := newPostgresDeviceTypePresenceService(
		t, db, createdAt, postgreschangelog.NewRecorder(db),
	)

	deviceType, err := service.CreateDeviceType(
		t.Context(),
		principal,
		applicationdcim.CreateDeviceTypeCommand{
			Manufacturer: applicationdcim.FieldValue(manufacturerA),
			Model:        applicationdcim.FieldValue("  Durable Device Type Defaults  "),
			Slug:         applicationdcim.FieldValue("  durable-device-type-defaults  "),
		},
	)
	require.NoError(t, err)
	requirePostgresDeviceTypeAggregate(
		t, deviceType,
		manufacturerA, "Presence Device Manufacturer A", "presence-device-manufacturer-a",
		"Durable Device Type Defaults", "durable-device-type-defaults", "", "1",
		false, true, nil, "", "", createdAt, createdAt,
	)

	created := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	requirePostgresDeviceTypeScalarRow(t, created.row, dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: deviceType.ID().Int64(), Created: createdAt, LastUpdated: createdAt,
		},
		ManufacturerID: manufacturerA.Int64(),
		Model:          "Durable Device Type Defaults",
		Slug:           "durable-device-type-defaults",
		UHeight:        1,
		IsFullDepth:    true,
	})
	require.Equal(t, int64(1), created.deviceTypeCount)
	require.Equal(t, int64(1), created.deviceTypeChangeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	require.Zero(t, created.deviceCount, "the scalar-presence proof requires zero Devices")
	require.Zero(
		t, created.interfaceTemplateCount,
		"the scalar-presence proof requires zero InterfaceTemplates",
	)
	requirePostgresDeviceTypeReloadedManufacturer(
		t, db, deviceType.ID(), manufacturerA,
		"Presence Device Manufacturer A", "presence-device-manufacturer-a",
	)

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresDeviceTypePresenceService(
		t, db, patchedAt, postgreschangelog.NewRecorder(db),
	)
	patched, err := service.UpdateDeviceType(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceTypeCommand{
			ID:                     deviceType.ID(),
			Manufacturer:           applicationdcim.FieldValue(manufacturerB),
			Model:                  applicationdcim.FieldValue("  Durable Device Type Concrete  "),
			Slug:                   applicationdcim.FieldValue("  durable-device-type-concrete  "),
			PartNumber:             applicationdcim.FieldValue("  PN-DURABLE-0  "),
			UHeight:                applicationdcim.FieldValue("0"),
			ExcludeFromUtilization: applicationdcim.FieldValue(true),
			IsFullDepth:            applicationdcim.FieldValue(false),
			Airflow:                applicationdcim.FieldValue("front-to-rear"),
			Description:            applicationdcim.FieldValue("  durable description  "),
			Comments:               applicationdcim.FieldValue("  durable comments  "),
		},
	)
	require.NoError(t, err)
	concreteAirflow := "front-to-rear"
	requirePostgresDeviceTypeAggregate(
		t, patched,
		manufacturerB, "Presence Device Manufacturer B", "presence-device-manufacturer-b",
		"Durable Device Type Concrete", "durable-device-type-concrete", "PN-DURABLE-0", "0",
		true, false, &concreteAirflow, "durable description", "durable comments",
		createdAt, patchedAt,
	)
	patchedState := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	requirePostgresDeviceTypeScalarRow(t, patchedState.row, dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: deviceType.ID().Int64(), Created: createdAt, LastUpdated: patchedAt,
		},
		ManufacturerID:         manufacturerB.Int64(),
		Model:                  "Durable Device Type Concrete",
		Slug:                   "durable-device-type-concrete",
		PartNumber:             "PN-DURABLE-0",
		UHeight:                0,
		ExcludeFromUtilization: true,
		IsFullDepth:            false,
		Airflow:                deviceTypePresenceStringPointer("front-to-rear"),
		Description:            "durable description",
		Comments:               "durable comments",
	})
	requireDeviceTypePostgresUpdateRecorded(t, created, patchedState, patchedAt)
	requirePostgresDeviceTypeReloadedManufacturer(
		t, db, deviceType.ID(), manufacturerB,
		"Presence Device Manufacturer B", "presence-device-manufacturer-b",
	)

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresDeviceTypePresenceService(
		t, db, replacedAt, postgreschangelog.NewRecorder(db),
	)
	replaced, err := service.ReplaceDeviceType(
		t.Context(),
		principal,
		applicationdcim.ReplaceDeviceTypeCommand{
			ID: deviceType.ID(),
			CreateDeviceTypeCommand: applicationdcim.CreateDeviceTypeCommand{
				Manufacturer: applicationdcim.FieldValue(manufacturerA),
				Model:        applicationdcim.FieldValue("  Durable Device Type Replaced  "),
				Slug:         applicationdcim.FieldValue("  durable-device-type-replaced  "),
			},
		},
	)
	require.NoError(t, err)
	requirePostgresDeviceTypeAggregate(
		t, replaced,
		manufacturerA, "Presence Device Manufacturer A", "presence-device-manufacturer-a",
		"Durable Device Type Replaced", "durable-device-type-replaced", "PN-DURABLE-0", "1",
		true, false, &concreteAirflow, "durable description", "durable comments",
		createdAt, replacedAt,
	)
	replacedState := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	requirePostgresDeviceTypeScalarRow(t, replacedState.row, dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: deviceType.ID().Int64(), Created: createdAt, LastUpdated: replacedAt,
		},
		ManufacturerID:         manufacturerA.Int64(),
		Model:                  "Durable Device Type Replaced",
		Slug:                   "durable-device-type-replaced",
		PartNumber:             "PN-DURABLE-0",
		UHeight:                1,
		ExcludeFromUtilization: true,
		IsFullDepth:            false,
		Airflow:                deviceTypePresenceStringPointer("front-to-rear"),
		Description:            "durable description",
		Comments:               "durable comments",
	})
	requireDeviceTypePostgresUpdateRecorded(t, patchedState, replacedState, replacedAt)
	requirePostgresDeviceTypeReloadedManufacturer(
		t, db, deviceType.ID(), manufacturerA,
		"Presence Device Manufacturer A", "presence-device-manufacturer-a",
	)

	blankedAt := replacedAt.Add(time.Minute)
	service = newPostgresDeviceTypePresenceService(
		t, db, blankedAt, postgreschangelog.NewRecorder(db),
	)
	blanked, err := service.UpdateDeviceType(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceTypeCommand{
			ID:                     deviceType.ID(),
			PartNumber:             applicationdcim.FieldValue(""),
			UHeight:                applicationdcim.FieldValue("0.5"),
			ExcludeFromUtilization: applicationdcim.FieldValue(false),
			Airflow:                applicationdcim.FieldValue(""),
			Description:            applicationdcim.FieldValue(""),
			Comments:               applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	blankAirflow := ""
	requirePostgresDeviceTypeAggregate(
		t, blanked,
		manufacturerA, "Presence Device Manufacturer A", "presence-device-manufacturer-a",
		"Durable Device Type Replaced", "durable-device-type-replaced", "", "0.5",
		false, false, &blankAirflow, "", "", createdAt, blankedAt,
	)
	blankedState := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	requirePostgresDeviceTypeScalarRow(t, blankedState.row, dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: deviceType.ID().Int64(), Created: createdAt, LastUpdated: blankedAt,
		},
		ManufacturerID: manufacturerA.Int64(),
		Model:          "Durable Device Type Replaced",
		Slug:           "durable-device-type-replaced",
		UHeight:        0.5,
		IsFullDepth:    false,
		Airflow:        deviceTypePresenceStringPointer(""),
	})
	requireDeviceTypePostgresUpdateRecorded(t, replacedState, blankedState, blankedAt)

	nulledAt := blankedAt.Add(time.Minute)
	service = newPostgresDeviceTypePresenceService(
		t, db, nulledAt, postgreschangelog.NewRecorder(db),
	)
	nulled, err := service.UpdateDeviceType(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceTypeCommand{
			ID: deviceType.ID(), Airflow: applicationdcim.NullField[string](),
		},
	)
	require.NoError(t, err)
	requirePostgresDeviceTypeAggregate(
		t, nulled,
		manufacturerA, "Presence Device Manufacturer A", "presence-device-manufacturer-a",
		"Durable Device Type Replaced", "durable-device-type-replaced", "", "0.5",
		false, false, nil, "", "", createdAt, nulledAt,
	)
	nulledState := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	requirePostgresDeviceTypeScalarRow(t, nulledState.row, dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: deviceType.ID().Int64(), Created: createdAt, LastUpdated: nulledAt,
		},
		ManufacturerID: manufacturerA.Int64(),
		Model:          "Durable Device Type Replaced",
		Slug:           "durable-device-type-replaced",
		UHeight:        0.5,
		IsFullDepth:    false,
	})
	requireDeviceTypePostgresUpdateRecorded(t, blankedState, nulledState, nulledAt)

	service = newPostgresDeviceTypePresenceService(
		t, db, nulledAt.Add(time.Minute), postgreschangelog.NewRecorder(db),
	)
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
			operationErr := operation()
			require.Error(t, operationErr)
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(
				t, before, loadPostgresDeviceTypePresenceState(t, db, deviceType.ID()),
				"rejected mutation changed durable DeviceType state",
			)
		})
	}
	assertRejected("POST invalid height", func() error {
		_, createErr := service.CreateDeviceType(
			t.Context(),
			principal,
			applicationdcim.CreateDeviceTypeCommand{
				Manufacturer: applicationdcim.FieldValue(manufacturerA),
				Model:        applicationdcim.FieldValue("Rejected Unique Device Type"),
				Slug:         applicationdcim.FieldValue("rejected-unique-device-type"),
				UHeight:      applicationdcim.FieldValue("1.1"),
			},
		)
		return createErr
	})
	assertRejected("PUT missing required identity", func() error {
		_, replaceErr := service.ReplaceDeviceType(
			t.Context(), principal,
			applicationdcim.ReplaceDeviceTypeCommand{ID: deviceType.ID()},
		)
		return replaceErr
	})
	assertRejected("PATCH unknown manufacturer", func() error {
		_, updateErr := service.UpdateDeviceType(
			t.Context(), principal,
			applicationdcim.UpdateDeviceTypeCommand{
				ID: deviceType.ID(), Manufacturer: applicationdcim.FieldValue(shared.ID(999999)),
			},
		)
		return updateErr
	})

	failureBefore := loadPostgresDeviceTypePresenceState(t, db, deviceType.ID())
	recorderFailure := errors.New("forced PostgreSQL DeviceType change recording failure")
	failingRecorder := &failingDeviceTypePresenceRecorder{err: recorderFailure}
	failingService := newPostgresDeviceTypePresenceService(
		t, db, nulledAt.Add(2*time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateDeviceType(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceTypeCommand{
			ID: deviceType.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(
		t, failureBefore, loadPostgresDeviceTypePresenceState(t, db, deviceType.ID()),
		"recorder failure must roll back state, timestamp, and audit intent",
	)
}

type postgresDeviceTypePresenceState struct {
	row                    dcimrow.DeviceTypeRow
	deviceTypeCount        int64
	deviceCount            int64
	interfaceTemplateCount int64
	deviceTypeChangeCount  int64
	totalChangeCount       int64
}

func seedDeviceTypePresenceManufacturer(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	name string,
	slug string,
) shared.ID {
	t.Helper()
	row := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        name,
		Slug:        slug,
	}
	require.NoError(t, db.Create(&row).Error)
	return shared.ID(row.ID)
}

func loadPostgresDeviceTypePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresDeviceTypePresenceState {
	t.Helper()
	var state postgresDeviceTypePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceTypeRow{}).Count(&state.deviceTypeCount).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(
		t,
		db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&state.interfaceTemplateCount).Error,
	)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceTypeObjectType, id.Int64(),
	).Count(&state.deviceTypeChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresDeviceTypeScalarRow(
	t *testing.T,
	actual dcimrow.DeviceTypeRow,
	expected dcimrow.DeviceTypeRow,
) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, actual.Created.Equal(expected.Created), "unexpected persisted created timestamp")
	require.True(
		t, actual.LastUpdated.Equal(expected.LastUpdated),
		"unexpected persisted last_updated timestamp",
	)
	require.Equal(t, expected.ManufacturerID, actual.ManufacturerID)
	require.Equal(t, expected.Model, actual.Model)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.PartNumber, actual.PartNumber)
	require.Equal(t, expected.UHeight, actual.UHeight)
	require.Equal(t, expected.ExcludeFromUtilization, actual.ExcludeFromUtilization)
	require.Equal(t, expected.IsFullDepth, actual.IsFullDepth)
	require.Equal(t, expected.Airflow, actual.Airflow)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

func requirePostgresDeviceTypeAggregate(
	t *testing.T,
	deviceType *domaindcim.DeviceType,
	manufacturerID shared.ID,
	manufacturerName string,
	manufacturerSlug string,
	model string,
	slug string,
	partNumber string,
	uHeight string,
	excludeFromUtilization bool,
	isFullDepth bool,
	airflow *string,
	description string,
	comments string,
	created time.Time,
	lastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, manufacturerID, deviceType.Manufacturer().ID())
	require.Equal(t, manufacturerName, deviceType.Manufacturer().Name())
	require.Equal(t, manufacturerSlug, deviceType.Manufacturer().Slug().String())
	require.Equal(t, model, deviceType.Model())
	require.Equal(t, slug, deviceType.Slug().String())
	require.Equal(t, partNumber, deviceType.PartNumber())
	require.Equal(t, uHeight, deviceType.UHeight().String())
	require.Equal(t, excludeFromUtilization, deviceType.ExcludeFromUtilization())
	require.Equal(t, isFullDepth, deviceType.IsFullDepth())
	if airflow == nil {
		require.True(t, deviceType.Airflow().IsNull())
	} else {
		actualAirflow, present := deviceType.Airflow().Get()
		require.True(t, present)
		require.Equal(t, *airflow, actualAirflow.String())
	}
	require.Equal(t, description, deviceType.Description())
	require.Equal(t, comments, deviceType.Comments())
	require.True(t, deviceType.Created().Equal(created))
	require.True(t, deviceType.LastUpdated().Equal(lastUpdated))
	require.Zero(t, deviceType.DeviceCount())
	require.Zero(t, deviceType.InterfaceTemplateCount())
}

func requirePostgresDeviceTypeReloadedManufacturer(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
	manufacturerID shared.ID,
	manufacturerName string,
	manufacturerSlug string,
) {
	t.Helper()
	reloaded, err := NewDeviceTypeRepository(db).Get(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, manufacturerID, reloaded.Manufacturer().ID())
	require.Equal(t, manufacturerName, reloaded.Manufacturer().Name())
	require.Equal(t, manufacturerSlug, reloaded.Manufacturer().Slug().String())
	require.Zero(t, reloaded.DeviceCount())
	require.Zero(t, reloaded.InterfaceTemplateCount())
}

func requireDeviceTypePostgresUpdateRecorded(
	t *testing.T,
	before postgresDeviceTypePresenceState,
	after postgresDeviceTypePresenceState,
	expectedLastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, before.deviceTypeCount, after.deviceTypeCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, before.interfaceTemplateCount, after.interfaceTemplateCount)
	require.Zero(t, after.deviceCount, "this proof deliberately has no Devices")
	require.Zero(
		t, after.interfaceTemplateCount,
		"this proof deliberately has no InterfaceTemplates",
	)
	require.Equal(t, before.deviceTypeChangeCount+1, after.deviceTypeChangeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.True(t, after.row.Created.Equal(before.row.Created))
	require.True(t, after.row.LastUpdated.Equal(expectedLastUpdated))
}

type failingDeviceTypePresenceRecorder struct {
	err   error
	calls int
}

func (recorder *failingDeviceTypePresenceRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresDeviceTypePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.DeviceTypeService {
	t.Helper()
	service, err := applicationdcim.NewDeviceTypeService(
		NewDeviceTypeRepository(db),
		NewManufacturerRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		recorder,
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}

func deviceTypePresenceStringPointer(value string) *string { return &value }
