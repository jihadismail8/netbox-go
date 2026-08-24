package dcim

import (
	"testing"
	"time"

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

func TestPostgresManufacturerScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 24, 10, 0, 0, 0, time.UTC)
	service := newPostgresManufacturerPresenceService(t, db, createdAt)

	manufacturer, err := service.CreateManufacturer(
		t.Context(),
		principal,
		applicationdcim.CreateManufacturerCommand{
			Name: applicationdcim.FieldValue("  Durable Manufacturer  "),
			Slug: applicationdcim.FieldValue("  durable-manufacturer  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Manufacturer", manufacturer.Name())
	require.Equal(t, "durable-manufacturer", manufacturer.Slug().String())
	require.Empty(t, manufacturer.Description())

	created := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, created.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer", Slug: "durable-manufacturer",
	})
	require.Equal(t, int64(1), created.manufacturerCount)
	require.Equal(t, int64(1), created.changeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	require.True(t, created.row.Created.Equal(createdAt))
	require.True(t, created.row.LastUpdated.Equal(createdAt))

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresManufacturerPresenceService(t, db, patchedAt)
	patched, err := service.UpdateManufacturer(
		t.Context(),
		principal,
		applicationdcim.UpdateManufacturerCommand{
			ID:          manufacturer.ID(),
			Description: applicationdcim.FieldValue("  durable description  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Manufacturer", patched.Name())
	require.Equal(t, "durable-manufacturer", patched.Slug().String())
	require.Equal(t, "durable description", patched.Description())

	patchedState := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, patchedState.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer", Slug: "durable-manufacturer",
		Description: "durable description",
	})
	require.Equal(t, created.manufacturerCount, patchedState.manufacturerCount)
	require.Equal(t, created.changeCount+1, patchedState.changeCount)
	require.Equal(t, created.totalChangeCount+1, patchedState.totalChangeCount)
	require.Equal(t, created.row.Created, patchedState.row.Created)
	require.True(t, patchedState.row.LastUpdated.Equal(patchedAt))

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresManufacturerPresenceService(t, db, replacedAt)
	replaced, err := service.ReplaceManufacturer(
		t.Context(),
		principal,
		applicationdcim.ReplaceManufacturerCommand{
			ID:   manufacturer.ID(),
			Name: applicationdcim.FieldValue("  Durable Manufacturer Renamed  "),
			Slug: applicationdcim.FieldValue("  durable-manufacturer-renamed  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Manufacturer Renamed", replaced.Name())
	require.Equal(t, "durable-manufacturer-renamed", replaced.Slug().String())
	require.Equal(
		t,
		"durable description",
		replaced.Description(),
		"PUT omission must preserve description",
	)

	replacedState := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, replacedState.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer Renamed", Slug: "durable-manufacturer-renamed",
		Description: "durable description",
	})
	require.Equal(t, patchedState.manufacturerCount, replacedState.manufacturerCount)
	require.Equal(t, patchedState.changeCount+1, replacedState.changeCount)
	require.Equal(t, patchedState.totalChangeCount+1, replacedState.totalChangeCount)
	require.Equal(t, created.row.Created, replacedState.row.Created)
	require.True(t, replacedState.row.LastUpdated.Equal(replacedAt))

	putClearedAt := replacedAt.Add(time.Minute)
	service = newPostgresManufacturerPresenceService(t, db, putClearedAt)
	putCleared, err := service.ReplaceManufacturer(
		t.Context(),
		principal,
		applicationdcim.ReplaceManufacturerCommand{
			ID:          manufacturer.ID(),
			Name:        applicationdcim.FieldValue("Durable Manufacturer PUT Cleared"),
			Slug:        applicationdcim.FieldValue("durable-manufacturer-put-cleared"),
			Description: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	require.Empty(t, putCleared.Description())
	putClearedState := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, putClearedState.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer PUT Cleared", Slug: "durable-manufacturer-put-cleared",
	})
	require.Equal(t, replacedState.manufacturerCount, putClearedState.manufacturerCount)
	require.Equal(t, replacedState.changeCount+1, putClearedState.changeCount)
	require.Equal(t, replacedState.totalChangeCount+1, putClearedState.totalChangeCount)
	require.Equal(t, created.row.Created, putClearedState.row.Created)
	require.True(t, putClearedState.row.LastUpdated.Equal(putClearedAt))

	patchResetAt := putClearedAt.Add(time.Minute)
	service = newPostgresManufacturerPresenceService(t, db, patchResetAt)
	patchReset, err := service.UpdateManufacturer(
		t.Context(),
		principal,
		applicationdcim.UpdateManufacturerCommand{
			ID:          manufacturer.ID(),
			Description: applicationdcim.FieldValue("  reset before PATCH clear  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "reset before PATCH clear", patchReset.Description())
	patchResetState := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, patchResetState.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer PUT Cleared", Slug: "durable-manufacturer-put-cleared",
		Description: "reset before PATCH clear",
	})
	require.Equal(t, putClearedState.manufacturerCount, patchResetState.manufacturerCount)
	require.Equal(t, putClearedState.changeCount+1, patchResetState.changeCount)
	require.Equal(t, putClearedState.totalChangeCount+1, patchResetState.totalChangeCount)
	require.Equal(t, created.row.Created, patchResetState.row.Created)
	require.True(t, patchResetState.row.LastUpdated.Equal(patchResetAt))

	patchClearedAt := patchResetAt.Add(time.Minute)
	service = newPostgresManufacturerPresenceService(t, db, patchClearedAt)
	patchCleared, err := service.UpdateManufacturer(
		t.Context(),
		principal,
		applicationdcim.UpdateManufacturerCommand{
			ID: manufacturer.ID(), Description: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	require.Empty(t, patchCleared.Description())
	patchClearedState := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
	requirePostgresManufacturerScalarRow(t, patchClearedState.row, dcimrow.ManufacturerRow{
		Name: "Durable Manufacturer PUT Cleared", Slug: "durable-manufacturer-put-cleared",
	})
	require.Equal(t, patchResetState.manufacturerCount, patchClearedState.manufacturerCount)
	require.Equal(t, patchResetState.changeCount+1, patchClearedState.changeCount)
	require.Equal(t, patchResetState.totalChangeCount+1, patchClearedState.totalChangeCount)
	require.Equal(t, created.row.Created, patchClearedState.row.Created)
	require.True(t, patchClearedState.row.LastUpdated.Equal(patchClearedAt))

	service = newPostgresManufacturerPresenceService(t, db, patchClearedAt.Add(time.Minute))
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
			err := operation()
			require.Error(t, err)
			require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
			after := loadPostgresManufacturerPresenceState(t, db, manufacturer.ID())
			require.Equal(t, before, after, "rejected mutation changed durable Manufacturer state")
		})
	}
	assertRejected("POST missing required name", func() error {
		_, err := service.CreateManufacturer(
			t.Context(),
			principal,
			applicationdcim.CreateManufacturerCommand{
				Slug: applicationdcim.FieldValue("rejected-manufacturer"),
			},
		)
		return err
	})
	assertRejected("POST explicit null description", func() error {
		_, err := service.CreateManufacturer(
			t.Context(),
			principal,
			applicationdcim.CreateManufacturerCommand{
				Name:        applicationdcim.FieldValue("Rejected Manufacturer"),
				Slug:        applicationdcim.FieldValue("rejected-manufacturer"),
				Description: applicationdcim.NullField[string](),
			},
		)
		return err
	})
	assertRejected("PUT missing required identity", func() error {
		_, err := service.ReplaceManufacturer(
			t.Context(),
			principal,
			applicationdcim.ReplaceManufacturerCommand{ID: manufacturer.ID()},
		)
		return err
	})
	assertRejected("PUT explicit null description", func() error {
		_, err := service.ReplaceManufacturer(
			t.Context(),
			principal,
			applicationdcim.ReplaceManufacturerCommand{
				ID:          manufacturer.ID(),
				Name:        applicationdcim.FieldValue("Durable Manufacturer PUT Cleared"),
				Slug:        applicationdcim.FieldValue("durable-manufacturer-put-cleared"),
				Description: applicationdcim.NullField[string](),
			},
		)
		return err
	})
	assertRejected("PATCH explicit null name", func() error {
		_, err := service.UpdateManufacturer(
			t.Context(),
			principal,
			applicationdcim.UpdateManufacturerCommand{
				ID: manufacturer.ID(), Name: applicationdcim.NullField[string](),
			},
		)
		return err
	})
	assertRejected("PATCH blank slug", func() error {
		_, err := service.UpdateManufacturer(
			t.Context(),
			principal,
			applicationdcim.UpdateManufacturerCommand{
				ID: manufacturer.ID(), Slug: applicationdcim.FieldValue(""),
			},
		)
		return err
	})
}

type postgresManufacturerPresenceState struct {
	row               dcimrow.ManufacturerRow
	manufacturerCount int64
	changeCount       int64
	totalChangeCount  int64
}

func requirePostgresManufacturerScalarRow(
	t *testing.T,
	actual dcimrow.ManufacturerRow,
	expected dcimrow.ManufacturerRow,
) {
	t.Helper()
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.Description, actual.Description)
}

func loadPostgresManufacturerPresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresManufacturerPresenceState {
	t.Helper()
	var state postgresManufacturerPresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.ManufacturerRow{}).Count(&state.manufacturerCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.ManufacturerObjectType, id.Int64(),
	).Count(&state.changeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func newPostgresManufacturerPresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) *applicationdcim.ManufacturerService {
	t.Helper()
	service, err := applicationdcim.NewManufacturerService(
		NewManufacturerRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}
