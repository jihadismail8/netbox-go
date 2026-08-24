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

func TestPostgresRackTypeScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 24, 20, 0, 0, 0, time.UTC)
	manufacturerA := seedRackTypePresenceManufacturer(t, db, createdAt, "Presence Manufacturer A", "presence-manufacturer-a")
	manufacturerB := seedRackTypePresenceManufacturer(t, db, createdAt, "Presence Manufacturer B", "presence-manufacturer-b")
	service := newPostgresRackTypePresenceService(t, db, createdAt)

	rackType, err := service.CreateRackType(
		t.Context(),
		principal,
		applicationdcim.CreateRackTypeCommand{
			Manufacturer: applicationdcim.FieldValue(manufacturerA),
			Model:        applicationdcim.FieldValue("  Durable Rack Type  "),
			Slug:         applicationdcim.FieldValue("  durable-rack-type  "),
			FormFactor:   applicationdcim.FieldValue("4-post-cabinet"),
		},
	)
	require.NoError(t, err)
	require.Equal(t, manufacturerA, rackType.Manufacturer().ID())
	require.Equal(t, "Durable Rack Type", rackType.Model())
	require.Equal(t, "durable-rack-type", rackType.Slug().String())
	require.Equal(t, domaindcim.RackFormFactorFourPostCabinet, rackType.FormFactor())
	require.Equal(t, domaindcim.RackWidth19, rackType.Width())
	require.Equal(t, uint32(42), rackType.UHeight())
	require.Equal(t, uint32(1), rackType.StartingUnit())
	require.False(t, rackType.DescUnits())
	require.Empty(t, rackType.Description())
	require.Empty(t, rackType.Comments())

	created := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, created.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerA.Int64(), Model: "Durable Rack Type", Slug: "durable-rack-type",
		FormFactor: "4-post-cabinet", Width: 19, UHeight: 42, StartingUnit: 1,
	})
	require.Equal(t, int64(1), created.rackTypeCount)
	require.Equal(t, int64(1), created.rackTypeChangeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	require.Zero(t, created.rackCount, "the scalar-presence proof requires zero referencing Racks")
	require.Zero(t, created.rackChangeCount, "the scalar-presence proof does not claim propagation")
	require.True(t, created.row.Created.Equal(createdAt))
	require.True(t, created.row.LastUpdated.Equal(createdAt))

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresRackTypePresenceService(t, db, patchedAt)
	patched, err := service.UpdateRackType(
		t.Context(), principal, applicationdcim.UpdateRackTypeCommand{
			ID:           rackType.ID(),
			Manufacturer: applicationdcim.FieldValue(manufacturerB),
			Model:        applicationdcim.FieldValue("  Durable Rack Type Patched  "),
			Slug:         applicationdcim.FieldValue("  durable-rack-type-patched  "),
			FormFactor:   applicationdcim.FieldValue("wall-cabinet"),
			Width:        applicationdcim.FieldValue(uint32(23)),
			UHeight:      applicationdcim.FieldValue(uint32(48)),
			StartingUnit: applicationdcim.FieldValue(uint32(2)),
			DescUnits:    applicationdcim.FieldValue(true),
			Description:  applicationdcim.FieldValue("  durable description  "),
			Comments:     applicationdcim.FieldValue("  durable comments  "),
		},
	)
	require.NoError(t, err)
	requirePostgresRackTypeAggregate(t, patched, manufacturerB, "Durable Rack Type Patched", "durable-rack-type-patched", "wall-cabinet", 23, 48, 2, true, "durable description", "durable comments")
	patchedState := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, patchedState.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerB.Int64(), Model: "Durable Rack Type Patched", Slug: "durable-rack-type-patched",
		FormFactor: "wall-cabinet", Width: 23, UHeight: 48, StartingUnit: 2, DescUnits: true,
		Description: "durable description", Comments: "durable comments",
	})
	requireRackTypePostgresUpdateRecorded(t, created, patchedState)
	require.True(t, patchedState.row.LastUpdated.Equal(patchedAt))

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresRackTypePresenceService(t, db, replacedAt)
	replaced, err := service.ReplaceRackType(
		t.Context(), principal, applicationdcim.ReplaceRackTypeCommand{
			ID: rackType.ID(), CreateRackTypeCommand: applicationdcim.CreateRackTypeCommand{
				Manufacturer: applicationdcim.FieldValue(manufacturerA),
				Model:        applicationdcim.FieldValue("  Durable Rack Type Replaced  "),
				Slug:         applicationdcim.FieldValue("  durable-rack-type-replaced  "),
			},
		},
	)
	require.NoError(t, err)
	requirePostgresRackTypeAggregate(t, replaced, manufacturerA, "Durable Rack Type Replaced", "durable-rack-type-replaced", "wall-cabinet", 23, 48, 2, true, "durable description", "durable comments")
	replacedState := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, replacedState.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerA.Int64(), Model: "Durable Rack Type Replaced", Slug: "durable-rack-type-replaced",
		FormFactor: "wall-cabinet", Width: 23, UHeight: 48, StartingUnit: 2, DescUnits: true,
		Description: "durable description", Comments: "durable comments",
	})
	requireRackTypePostgresUpdateRecorded(t, patchedState, replacedState)
	require.True(t, replacedState.row.LastUpdated.Equal(replacedAt))

	putClearedAt := replacedAt.Add(time.Minute)
	service = newPostgresRackTypePresenceService(t, db, putClearedAt)
	putCleared, err := service.ReplaceRackType(
		t.Context(), principal, applicationdcim.ReplaceRackTypeCommand{
			ID: rackType.ID(), CreateRackTypeCommand: applicationdcim.CreateRackTypeCommand{
				Manufacturer: applicationdcim.FieldValue(manufacturerA),
				Model:        applicationdcim.FieldValue("Durable Rack Type PUT Clear"),
				Slug:         applicationdcim.FieldValue("durable-rack-type-put-clear"),
				Description:  applicationdcim.FieldValue(""),
				Comments:     applicationdcim.FieldValue(""),
			},
		},
	)
	require.NoError(t, err)
	requirePostgresRackTypeAggregate(t, putCleared, manufacturerA, "Durable Rack Type PUT Clear", "durable-rack-type-put-clear", "wall-cabinet", 23, 48, 2, true, "", "")
	putClearedState := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, putClearedState.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerA.Int64(), Model: "Durable Rack Type PUT Clear", Slug: "durable-rack-type-put-clear",
		FormFactor: "wall-cabinet", Width: 23, UHeight: 48, StartingUnit: 2, DescUnits: true,
	})
	requireRackTypePostgresUpdateRecorded(t, replacedState, putClearedState)
	require.True(t, putClearedState.row.LastUpdated.Equal(putClearedAt))

	patchResetAt := putClearedAt.Add(time.Minute)
	service = newPostgresRackTypePresenceService(t, db, patchResetAt)
	_, err = service.UpdateRackType(
		t.Context(), principal, applicationdcim.UpdateRackTypeCommand{
			ID: rackType.ID(), DescUnits: applicationdcim.FieldValue(false),
			Description: applicationdcim.FieldValue("reset description"),
			Comments:    applicationdcim.FieldValue("reset comments"),
		},
	)
	require.NoError(t, err)
	patchResetState := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, patchResetState.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerA.Int64(), Model: "Durable Rack Type PUT Clear", Slug: "durable-rack-type-put-clear",
		FormFactor: "wall-cabinet", Width: 23, UHeight: 48, StartingUnit: 2,
		Description: "reset description", Comments: "reset comments",
	})
	requireRackTypePostgresUpdateRecorded(t, putClearedState, patchResetState)

	patchClearedAt := patchResetAt.Add(time.Minute)
	service = newPostgresRackTypePresenceService(t, db, patchClearedAt)
	patchCleared, err := service.UpdateRackType(
		t.Context(), principal, applicationdcim.UpdateRackTypeCommand{
			ID: rackType.ID(), Description: applicationdcim.FieldValue(""),
			Comments: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	require.False(t, patchCleared.DescUnits())
	require.Empty(t, patchCleared.Description())
	require.Empty(t, patchCleared.Comments())
	patchClearedState := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	requirePostgresRackTypeScalarRow(t, patchClearedState.row, dcimrow.RackTypeRow{
		ManufacturerID: manufacturerA.Int64(), Model: "Durable Rack Type PUT Clear", Slug: "durable-rack-type-put-clear",
		FormFactor: "wall-cabinet", Width: 23, UHeight: 48, StartingUnit: 2,
	})
	requireRackTypePostgresUpdateRecorded(t, patchResetState, patchClearedState)
	require.True(t, patchClearedState.row.LastUpdated.Equal(patchClearedAt))

	service = newPostgresRackTypePresenceService(t, db, patchClearedAt.Add(time.Minute))
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresRackTypePresenceState(t, db, rackType.ID())
			err := operation()
			require.Error(t, err)
			require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
			after := loadPostgresRackTypePresenceState(t, db, rackType.ID())
			require.Equal(t, before, after, "rejected mutation changed durable RackType state")
		})
	}
	assertRejected("POST missing effective form factor", func() error {
		_, err := service.CreateRackType(t.Context(), principal, applicationdcim.CreateRackTypeCommand{
			Manufacturer: applicationdcim.FieldValue(manufacturerA), Model: applicationdcim.FieldValue("Rejected Rack Type"),
			Slug: applicationdcim.FieldValue("rejected-rack-type"),
		})
		return err
	})
	assertRejected("POST explicit null width", func() error {
		_, err := service.CreateRackType(t.Context(), principal, applicationdcim.CreateRackTypeCommand{
			Manufacturer: applicationdcim.FieldValue(manufacturerA), Model: applicationdcim.FieldValue("Rejected Rack Type"),
			Slug: applicationdcim.FieldValue("rejected-rack-type"), FormFactor: applicationdcim.FieldValue("4-post-cabinet"),
			Width: applicationdcim.NullField[uint32](),
		})
		return err
	})
	assertRejected("PUT missing required identity", func() error {
		_, err := service.ReplaceRackType(t.Context(), principal, applicationdcim.ReplaceRackTypeCommand{ID: rackType.ID()})
		return err
	})
	assertRejected("PUT explicit null form factor", func() error {
		_, err := service.ReplaceRackType(t.Context(), principal, applicationdcim.ReplaceRackTypeCommand{
			ID: rackType.ID(), CreateRackTypeCommand: applicationdcim.CreateRackTypeCommand{
				Manufacturer: applicationdcim.FieldValue(manufacturerA), Model: applicationdcim.FieldValue("Rejected PUT"),
				Slug: applicationdcim.FieldValue("rejected-put"), FormFactor: applicationdcim.NullField[string](),
			},
		})
		return err
	})
	for _, rejected := range []struct {
		name      string
		operation func() error
	}{
		{"PATCH null manufacturer", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Manufacturer: applicationdcim.NullField[shared.ID]()})
			return err
		}},
		{"PATCH null model", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Model: applicationdcim.NullField[string]()})
			return err
		}},
		{"PATCH null slug", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Slug: applicationdcim.NullField[string]()})
			return err
		}},
		{"PATCH null form factor", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), FormFactor: applicationdcim.NullField[string]()})
			return err
		}},
		{"PATCH null width", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Width: applicationdcim.NullField[uint32]()})
			return err
		}},
		{"PATCH null u height", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), UHeight: applicationdcim.NullField[uint32]()})
			return err
		}},
		{"PATCH null starting unit", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), StartingUnit: applicationdcim.NullField[uint32]()})
			return err
		}},
		{"PATCH null desc units", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), DescUnits: applicationdcim.NullField[bool]()})
			return err
		}},
		{"PATCH null description", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Description: applicationdcim.NullField[string]()})
			return err
		}},
		{"PATCH null comments", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Comments: applicationdcim.NullField[string]()})
			return err
		}},
		{"PATCH whitespace form factor", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), FormFactor: applicationdcim.FieldValue(" wall-cabinet ")})
			return err
		}},
		{"PATCH zero width", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Width: applicationdcim.FieldValue(uint32(0))})
			return err
		}},
		{"PATCH zero u height", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), UHeight: applicationdcim.FieldValue(uint32(0))})
			return err
		}},
		{"PATCH zero starting unit", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), StartingUnit: applicationdcim.FieldValue(uint32(0))})
			return err
		}},
		{"PATCH unknown manufacturer", func() error {
			_, err := service.UpdateRackType(t.Context(), principal, applicationdcim.UpdateRackTypeCommand{ID: rackType.ID(), Manufacturer: applicationdcim.FieldValue(shared.ID(999999))})
			return err
		}},
	} {
		assertRejected(rejected.name, rejected.operation)
	}

	failureBefore := loadPostgresRackTypePresenceState(t, db, rackType.ID())
	recorderFailure := errors.New("forced PostgreSQL RackType change recording failure")
	failingRecorder := &failingRackTypeChangeRecorder{err: recorderFailure}
	failingService := newPostgresRackTypePresenceServiceWithRecorder(
		t, db, patchClearedAt.Add(2*time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateRackType(
		t.Context(), principal, applicationdcim.UpdateRackTypeCommand{
			ID: rackType.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(t, failureBefore, loadPostgresRackTypePresenceState(t, db, rackType.ID()))
}

type postgresRackTypePresenceState struct {
	row                 dcimrow.RackTypeRow
	rackTypeCount       int64
	rackCount           int64
	rackTypeChangeCount int64
	rackChangeCount     int64
	totalChangeCount    int64
}

func seedRackTypePresenceManufacturer(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	name string,
	slug string,
) shared.ID {
	t.Helper()
	row := dcimrow.ManufacturerRow{
		RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: name, Slug: slug,
	}
	require.NoError(t, db.Create(&row).Error)
	return shared.ID(row.ID)
}

func loadPostgresRackTypePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresRackTypePresenceState {
	t.Helper()
	var state postgresRackTypePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.RackTypeRow{}).Count(&state.rackTypeCount).Error)
	require.NoError(t, db.Model(&dcimrow.RackRow{}).Count(&state.rackCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackTypeObjectType, id.Int64(),
	).Count(&state.rackTypeChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.RackObjectType,
	).Count(&state.rackChangeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresRackTypeScalarRow(
	t *testing.T,
	actual dcimrow.RackTypeRow,
	expected dcimrow.RackTypeRow,
) {
	t.Helper()
	require.Equal(t, expected.ManufacturerID, actual.ManufacturerID)
	require.Equal(t, expected.Model, actual.Model)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.FormFactor, actual.FormFactor)
	require.Equal(t, expected.Width, actual.Width)
	require.Equal(t, expected.UHeight, actual.UHeight)
	require.Equal(t, expected.StartingUnit, actual.StartingUnit)
	require.Equal(t, expected.DescUnits, actual.DescUnits)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

func requirePostgresRackTypeAggregate(
	t *testing.T,
	rackType *domaindcim.RackType,
	manufacturer shared.ID,
	model string,
	slug string,
	formFactor string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	description string,
	comments string,
) {
	t.Helper()
	require.Equal(t, manufacturer, rackType.Manufacturer().ID())
	require.Equal(t, model, rackType.Model())
	require.Equal(t, slug, rackType.Slug().String())
	require.Equal(t, formFactor, rackType.FormFactor().String())
	require.Equal(t, width, rackType.Width().Uint32())
	require.Equal(t, uHeight, rackType.UHeight())
	require.Equal(t, startingUnit, rackType.StartingUnit())
	require.Equal(t, descUnits, rackType.DescUnits())
	require.Equal(t, description, rackType.Description())
	require.Equal(t, comments, rackType.Comments())
}

func requireRackTypePostgresUpdateRecorded(
	t *testing.T,
	before postgresRackTypePresenceState,
	after postgresRackTypePresenceState,
) {
	t.Helper()
	require.Equal(t, before.rackTypeCount, after.rackTypeCount)
	require.Equal(t, before.rackCount, after.rackCount)
	require.Zero(t, after.rackCount, "this proof deliberately has no referencing Racks")
	require.Equal(t, before.rackTypeChangeCount+1, after.rackTypeChangeCount)
	require.Equal(t, before.rackChangeCount, after.rackChangeCount)
	require.Zero(t, after.rackChangeCount, "this proof does not claim Rack propagation")
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

type failingRackTypeChangeRecorder struct {
	err   error
	calls int
}

func (recorder *failingRackTypeChangeRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresRackTypePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) *applicationdcim.RackTypeService {
	t.Helper()
	return newPostgresRackTypePresenceServiceWithRecorder(
		t, db, now, postgreschangelog.NewRecorder(db),
	)
}

func newPostgresRackTypePresenceServiceWithRecorder(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.RackTypeService {
	t.Helper()
	service, err := applicationdcim.NewRackTypeService(
		NewRackTypeRepository(db),
		NewManufacturerRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		recorder,
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}
