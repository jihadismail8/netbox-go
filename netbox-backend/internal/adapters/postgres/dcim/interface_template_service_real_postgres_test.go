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

func TestPostgresInterfaceTemplateScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 25, 19, 0, 0, 0, time.UTC)
	manufacturer := seedInterfaceTemplatePresenceManufacturer(t, db, createdAt)
	deviceTypeA := seedInterfaceTemplatePresenceDeviceType(
		t, db, createdAt, manufacturer,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
	)
	deviceTypeB := seedInterfaceTemplatePresenceDeviceType(
		t, db, createdAt, manufacturer,
		"Presence Interface Device Type B", "presence-interface-device-type-b",
	)
	service := newPostgresInterfaceTemplatePresenceService(
		t, db, createdAt, postgreschangelog.NewRecorder(db),
	)

	template, err := service.CreateInterfaceTemplate(
		t.Context(),
		principal,
		applicationdcim.CreateInterfaceTemplateCommand{
			DeviceType: applicationdcim.FieldValue(deviceTypeA),
			Name:       applicationdcim.FieldValue("  Durable Interface Template  "),
			Type:       applicationdcim.FieldValue("1000base-t"),
		},
	)
	require.NoError(t, err)
	requirePostgresInterfaceTemplateAggregate(
		t, template, deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
		"Durable Interface Template", "", "1000base-t", true, false, "",
		createdAt, createdAt,
	)

	created := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
	requirePostgresInterfaceTemplateScalarRow(t, created.row, dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: template.ID().Int64(), Created: createdAt, LastUpdated: createdAt,
		},
		DeviceTypeID: deviceTypeA.Int64(), Name: "Durable Interface Template",
		Type: "1000base-t", Enabled: true,
	})
	require.Equal(t, int64(1), created.manufacturerCount)
	require.Equal(t, int64(2), created.deviceTypeCount)
	require.Equal(t, int64(1), created.interfaceTemplateCount)
	require.Zero(t, created.deviceCount, "scalar presence requires zero Devices")
	require.Zero(t, created.interfaceCount, "scalar presence requires zero Interfaces")
	require.Equal(t, int64(1), created.interfaceTemplateChangeCount)
	require.Zero(t, created.deviceTypeChangeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	requirePostgresInterfaceTemplateReloadedDeviceType(
		t, db, template.ID(), deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
	)

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresInterfaceTemplatePresenceService(
		t, db, patchedAt, postgreschangelog.NewRecorder(db),
	)
	patched, err := service.UpdateInterfaceTemplate(
		t.Context(),
		principal,
		applicationdcim.UpdateInterfaceTemplateCommand{
			ID:          template.ID(),
			DeviceType:  applicationdcim.FieldValue(deviceTypeA),
			Name:        applicationdcim.FieldValue("  Durable Interface Concrete  "),
			Label:       applicationdcim.FieldValue("  WAN  "),
			Type:        applicationdcim.FieldValue("10gbase-sr"),
			Enabled:     applicationdcim.FieldValue(false),
			MgmtOnly:    applicationdcim.FieldValue(true),
			Description: applicationdcim.FieldValue("  durable description  "),
		},
	)
	require.NoError(t, err)
	requirePostgresInterfaceTemplateAggregate(
		t, patched, deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
		"Durable Interface Concrete", "WAN", "10gbase-sr", false, true,
		"durable description", createdAt, patchedAt,
	)
	patchedState := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
	requirePostgresInterfaceTemplateScalarRow(t, patchedState.row, dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: template.ID().Int64(), Created: createdAt, LastUpdated: patchedAt,
		},
		DeviceTypeID: deviceTypeA.Int64(), Name: "Durable Interface Concrete",
		Label: "WAN", Type: "10gbase-sr", MgmtOnly: true,
		Description: "durable description",
	})
	requireInterfaceTemplatePostgresMutationRecorded(t, created, patchedState, patchedAt)

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresInterfaceTemplatePresenceService(
		t, db, replacedAt, postgreschangelog.NewRecorder(db),
	)
	replaced, err := service.ReplaceInterfaceTemplate(
		t.Context(),
		principal,
		applicationdcim.ReplaceInterfaceTemplateCommand{
			ID: template.ID(),
			CreateInterfaceTemplateCommand: applicationdcim.CreateInterfaceTemplateCommand{
				DeviceType: applicationdcim.FieldValue(deviceTypeA),
				Name:       applicationdcim.FieldValue("  Durable Interface Replaced  "),
				Type:       applicationdcim.FieldValue("25gbase-sr"),
			},
		},
	)
	require.NoError(t, err)
	requirePostgresInterfaceTemplateAggregate(
		t, replaced, deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
		"Durable Interface Replaced", "WAN", "25gbase-sr", false, true,
		"durable description", createdAt, replacedAt,
	)
	replacedState := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
	requirePostgresInterfaceTemplateScalarRow(t, replacedState.row, dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: template.ID().Int64(), Created: createdAt, LastUpdated: replacedAt,
		},
		DeviceTypeID: deviceTypeA.Int64(), Name: "Durable Interface Replaced",
		Label: "WAN", Type: "25gbase-sr", MgmtOnly: true,
		Description: "durable description",
	})
	requireInterfaceTemplatePostgresMutationRecorded(t, patchedState, replacedState, replacedAt)
	requirePostgresInterfaceTemplateReloadedDeviceType(
		t, db, template.ID(), deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
	)

	clearedAt := replacedAt.Add(time.Minute)
	service = newPostgresInterfaceTemplatePresenceService(
		t, db, clearedAt, postgreschangelog.NewRecorder(db),
	)
	cleared, err := service.UpdateInterfaceTemplate(
		t.Context(),
		principal,
		applicationdcim.UpdateInterfaceTemplateCommand{
			ID:          template.ID(),
			Label:       applicationdcim.FieldValue(""),
			Enabled:     applicationdcim.FieldValue(false),
			MgmtOnly:    applicationdcim.FieldValue(false),
			Description: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	requirePostgresInterfaceTemplateAggregate(
		t, cleared, deviceTypeA,
		"Presence Interface Device Type A", "presence-interface-device-type-a",
		"Durable Interface Replaced", "", "25gbase-sr", false, false, "",
		createdAt, clearedAt,
	)
	clearedState := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
	requirePostgresInterfaceTemplateScalarRow(t, clearedState.row, dcimrow.InterfaceTemplateRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: template.ID().Int64(), Created: createdAt, LastUpdated: clearedAt,
		},
		DeviceTypeID: deviceTypeA.Int64(), Name: "Durable Interface Replaced",
		Type: "25gbase-sr",
	})
	requireInterfaceTemplatePostgresMutationRecorded(t, replacedState, clearedState, clearedAt)

	service = newPostgresInterfaceTemplatePresenceService(
		t, db, clearedAt.Add(time.Minute), postgreschangelog.NewRecorder(db),
	)
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
			operationErr := operation()
			require.Error(t, operationErr)
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(
				t, before, loadPostgresInterfaceTemplatePresenceState(t, db, template.ID()),
				"rejected mutation changed durable InterfaceTemplate state",
			)
		})
	}
	assertRejected("POST missing required fields", func() error {
		_, createErr := service.CreateInterfaceTemplate(
			t.Context(), principal, applicationdcim.CreateInterfaceTemplateCommand{},
		)
		return createErr
	})
	assertRejected("POST zero DeviceType", func() error {
		_, createErr := service.CreateInterfaceTemplate(
			t.Context(), principal, applicationdcim.CreateInterfaceTemplateCommand{
				DeviceType: applicationdcim.FieldValue(shared.ID(0)),
				Name:       applicationdcim.FieldValue("Rejected Zero Device Type"),
				Type:       applicationdcim.FieldValue("1000base-t"),
			},
		)
		return createErr
	})
	assertRejected("PUT missing required fields", func() error {
		_, replaceErr := service.ReplaceInterfaceTemplate(
			t.Context(), principal,
			applicationdcim.ReplaceInterfaceTemplateCommand{ID: template.ID()},
		)
		return replaceErr
	})
	for _, rejected := range []struct {
		name      string
		operation func() error
	}{
		{"PATCH null DeviceType", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), DeviceType: applicationdcim.NullField[shared.ID]()})
			return updateErr
		}},
		{"PATCH null name", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Name: applicationdcim.NullField[string]()})
			return updateErr
		}},
		{"PATCH null label", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Label: applicationdcim.NullField[string]()})
			return updateErr
		}},
		{"PATCH null type", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Type: applicationdcim.NullField[string]()})
			return updateErr
		}},
		{"PATCH null enabled", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Enabled: applicationdcim.NullField[bool]()})
			return updateErr
		}},
		{"PATCH null management-only", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), MgmtOnly: applicationdcim.NullField[bool]()})
			return updateErr
		}},
		{"PATCH null description", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Description: applicationdcim.NullField[string]()})
			return updateErr
		}},
		{"PATCH zero DeviceType", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), DeviceType: applicationdcim.FieldValue(shared.ID(0))})
			return updateErr
		}},
		{"PATCH blank name", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Name: applicationdcim.FieldValue("   ")})
			return updateErr
		}},
		{"PATCH untrimmed type", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), Type: applicationdcim.FieldValue(" 25gbase-sr ")})
			return updateErr
		}},
		{"PATCH unknown DeviceType", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), DeviceType: applicationdcim.FieldValue(shared.ID(999999))})
			return updateErr
		}},
		{"PATCH different known DeviceType", func() error {
			_, updateErr := service.UpdateInterfaceTemplate(t.Context(), principal, applicationdcim.UpdateInterfaceTemplateCommand{ID: template.ID(), DeviceType: applicationdcim.FieldValue(deviceTypeB)})
			return updateErr
		}},
	} {
		assertRejected(rejected.name, rejected.operation)
	}

	failureBefore := loadPostgresInterfaceTemplatePresenceState(t, db, template.ID())
	recorderFailure := errors.New("forced PostgreSQL InterfaceTemplate change recording failure")
	failingRecorder := &failingInterfaceTemplatePresenceRecorder{err: recorderFailure}
	failingService := newPostgresInterfaceTemplatePresenceService(
		t, db, clearedAt.Add(2*time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateInterfaceTemplate(
		t.Context(),
		principal,
		applicationdcim.UpdateInterfaceTemplateCommand{
			ID: template.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(
		t, failureBefore, loadPostgresInterfaceTemplatePresenceState(t, db, template.ID()),
		"recorder failure must roll back state, timestamp, and object change",
	)
}

type postgresInterfaceTemplatePresenceState struct {
	row                          dcimrow.InterfaceTemplateRow
	manufacturerCount            int64
	deviceTypeCount              int64
	interfaceTemplateCount       int64
	deviceCount                  int64
	interfaceCount               int64
	interfaceTemplateChangeCount int64
	deviceTypeChangeCount        int64
	totalChangeCount             int64
}

func seedInterfaceTemplatePresenceManufacturer(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) shared.ID {
	t.Helper()
	row := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Presence Interface Manufacturer",
		Slug:        "presence-interface-manufacturer",
	}
	require.NoError(t, db.Create(&row).Error)
	return shared.ID(row.ID)
}

func seedInterfaceTemplatePresenceDeviceType(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	manufacturer shared.ID,
	model string,
	slug string,
) shared.ID {
	t.Helper()
	row := dcimrow.DeviceTypeRow{
		RowMetadata:    dcimrow.RowMetadata{Created: now, LastUpdated: now},
		ManufacturerID: manufacturer.Int64(),
		Model:          model,
		Slug:           slug,
		UHeight:        1,
		IsFullDepth:    true,
	}
	require.NoError(t, db.Create(&row).Error)
	return shared.ID(row.ID)
}

func loadPostgresInterfaceTemplatePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresInterfaceTemplatePresenceState {
	t.Helper()
	var state postgresInterfaceTemplatePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.ManufacturerRow{}).Count(&state.manufacturerCount).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceTypeRow{}).Count(&state.deviceTypeCount).Error)
	require.NoError(
		t,
		db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&state.interfaceTemplateCount).Error,
	)
	require.NoError(t, db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, db.Model(&dcimrow.InterfaceRow{}).Count(&state.interfaceCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.InterfaceTemplateObjectType, id.Int64(),
	).Count(&state.interfaceTemplateChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceTypeObjectType,
	).Count(&state.deviceTypeChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresInterfaceTemplateScalarRow(
	t *testing.T,
	actual dcimrow.InterfaceTemplateRow,
	expected dcimrow.InterfaceTemplateRow,
) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, actual.Created.Equal(expected.Created), "unexpected persisted created timestamp")
	require.True(
		t, actual.LastUpdated.Equal(expected.LastUpdated),
		"unexpected persisted last_updated timestamp",
	)
	require.Equal(t, expected.DeviceTypeID, actual.DeviceTypeID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Label, actual.Label)
	require.Equal(t, expected.Type, actual.Type)
	require.Equal(t, expected.Enabled, actual.Enabled)
	require.Equal(t, expected.MgmtOnly, actual.MgmtOnly)
	require.Equal(t, expected.Description, actual.Description)
}

func requirePostgresInterfaceTemplateAggregate(
	t *testing.T,
	template *domaindcim.InterfaceTemplate,
	deviceTypeID shared.ID,
	deviceTypeModel string,
	deviceTypeSlug string,
	name string,
	label string,
	interfaceType string,
	enabled bool,
	mgmtOnly bool,
	description string,
	created time.Time,
	lastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, deviceTypeID, template.DeviceType().ID())
	require.Equal(t, deviceTypeModel, template.DeviceType().Model())
	require.Equal(t, deviceTypeSlug, template.DeviceType().Slug().String())
	require.Equal(t, deviceTypeModel, template.DeviceType().Display())
	require.Equal(t, name, template.Name())
	require.Equal(t, label, template.Label())
	require.Equal(t, interfaceType, template.Type().String())
	require.Equal(t, enabled, template.Enabled())
	require.Equal(t, mgmtOnly, template.MgmtOnly())
	require.Equal(t, description, template.Description())
	require.True(t, template.Created().Equal(created))
	require.True(t, template.LastUpdated().Equal(lastUpdated))
}

func requirePostgresInterfaceTemplateReloadedDeviceType(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
	deviceTypeID shared.ID,
	model string,
	slug string,
) {
	t.Helper()
	reloaded, err := NewInterfaceTemplateRepository(db).Get(t.Context(), id)
	require.NoError(t, err)
	require.Equal(t, deviceTypeID, reloaded.DeviceType().ID())
	require.Equal(t, model, reloaded.DeviceType().Model())
	require.Equal(t, slug, reloaded.DeviceType().Slug().String())
	require.Equal(t, model, reloaded.DeviceType().Display())
}

func requireInterfaceTemplatePostgresMutationRecorded(
	t *testing.T,
	before postgresInterfaceTemplatePresenceState,
	after postgresInterfaceTemplatePresenceState,
	expectedLastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, before.manufacturerCount, after.manufacturerCount)
	require.Equal(t, before.deviceTypeCount, after.deviceTypeCount)
	require.Equal(t, before.interfaceTemplateCount, after.interfaceTemplateCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, before.interfaceCount, after.interfaceCount)
	require.Equal(t, int64(2), after.deviceTypeCount)
	require.Equal(t, int64(1), after.interfaceTemplateCount)
	require.Zero(t, after.deviceCount, "scalar presence requires zero Devices")
	require.Zero(t, after.interfaceCount, "scalar presence requires zero Interfaces")
	require.Equal(
		t, before.interfaceTemplateChangeCount+1, after.interfaceTemplateChangeCount,
	)
	require.Equal(t, before.deviceTypeChangeCount, after.deviceTypeChangeCount)
	require.Zero(t, after.deviceTypeChangeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.True(t, after.row.Created.Equal(before.row.Created))
	require.True(t, after.row.LastUpdated.Equal(expectedLastUpdated))
}

type failingInterfaceTemplatePresenceRecorder struct {
	err   error
	calls int
}

func (recorder *failingInterfaceTemplatePresenceRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresInterfaceTemplatePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.InterfaceTemplateService {
	t.Helper()
	service, err := applicationdcim.NewInterfaceTemplateService(
		NewInterfaceTemplateRepository(db),
		NewDeviceTypeRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		recorder,
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}
