package dcim

import (
	"context"
	"encoding/json"
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

func TestPostgresRackRoleScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 24, 17, 0, 0, 0, time.UTC)
	service := newPostgresRackRolePresenceService(t, db, createdAt)

	role, err := service.CreateRackRole(
		t.Context(),
		principal,
		applicationdcim.CreateRackRoleCommand{
			Name: applicationdcim.FieldValue("  Durable Rack Role  "),
			Slug: applicationdcim.FieldValue("  durable-rack-role  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Rack Role", role.Name())
	require.Equal(t, "durable-rack-role", role.Slug().String())
	require.Equal(t, domaindcim.RackRoleDefaultColor, role.Color().String())
	require.Empty(t, role.Description())

	created := loadPostgresRackRolePresenceState(t, db, role.ID())
	requirePostgresRackRoleScalarRow(t, created.row, dcimrow.RackRoleRow{
		Name: "Durable Rack Role", Slug: "durable-rack-role",
		Color: domaindcim.RackRoleDefaultColor,
	})
	require.Equal(t, int64(1), created.roleCount)
	require.Equal(t, int64(1), created.changeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	require.Zero(t, created.rackCount)
	require.True(t, created.row.Created.Equal(createdAt))
	require.True(t, created.row.LastUpdated.Equal(createdAt))

	siteID := seedOrganizationSite(t, db)
	require.NoError(t, db.Create(&dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{Created: createdAt, LastUpdated: createdAt},
		SiteID:      siteID, Name: "Durability Rack", RoleID: idPointer(role.ID().Int64()),
		Status: "active", Width: 19, UHeight: 42, StartingUnit: 1,
	}).Error)
	withRack := loadPostgresRackRolePresenceState(t, db, role.ID())
	require.Equal(t, uint64(1), withRack.rackCount)

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresRackRolePresenceService(t, db, patchedAt)
	patched, err := service.UpdateRackRole(
		t.Context(),
		principal,
		applicationdcim.UpdateRackRoleCommand{
			ID: role.ID(), Color: applicationdcim.FieldValue("  00ff00  "),
			Description: applicationdcim.FieldValue("  durable description  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Rack Role", patched.Name())
	require.Equal(t, "durable-rack-role", patched.Slug().String())
	require.Equal(t, "00ff00", patched.Color().String())
	require.Equal(t, "durable description", patched.Description())
	require.Equal(t, uint64(1), patched.RackCount())

	patchedState := loadPostgresRackRolePresenceState(t, db, role.ID())
	requirePostgresRackRoleScalarRow(t, patchedState.row, dcimrow.RackRoleRow{
		Name: "Durable Rack Role", Slug: "durable-rack-role", Color: "00ff00",
		Description: "durable description",
	})
	requireRackRolePostgresUpdateRecorded(t, withRack, patchedState)
	require.True(t, patchedState.row.LastUpdated.Equal(patchedAt))

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresRackRolePresenceService(t, db, replacedAt)
	replaced, err := service.ReplaceRackRole(
		t.Context(),
		principal,
		applicationdcim.ReplaceRackRoleCommand{
			ID: role.ID(), Name: applicationdcim.FieldValue("  Durable Rack Role Renamed  "),
			Slug: applicationdcim.FieldValue("  durable-rack-role-renamed  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Rack Role Renamed", replaced.Name())
	require.Equal(t, "durable-rack-role-renamed", replaced.Slug().String())
	require.Equal(t, "00ff00", replaced.Color().String(), "PUT omission must preserve color")
	require.Equal(t, "durable description", replaced.Description(), "PUT omission must preserve description")
	require.Equal(t, uint64(1), replaced.RackCount())

	replacedState := loadPostgresRackRolePresenceState(t, db, role.ID())
	requirePostgresRackRoleScalarRow(t, replacedState.row, dcimrow.RackRoleRow{
		Name: "Durable Rack Role Renamed", Slug: "durable-rack-role-renamed",
		Color: "00ff00", Description: "durable description",
	})
	requireRackRolePostgresUpdateRecorded(t, patchedState, replacedState)
	require.True(t, replacedState.row.LastUpdated.Equal(replacedAt))

	putClearedAt := replacedAt.Add(time.Minute)
	service = newPostgresRackRolePresenceService(t, db, putClearedAt)
	putCleared, err := service.ReplaceRackRole(
		t.Context(),
		principal,
		applicationdcim.ReplaceRackRoleCommand{
			ID: role.ID(), Name: applicationdcim.FieldValue("Durable Rack Role PUT Cleared"),
			Slug:        applicationdcim.FieldValue("durable-rack-role-put-cleared"),
			Description: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	require.Empty(t, putCleared.Description())
	require.Equal(t, "00ff00", putCleared.Color().String())
	putClearedState := loadPostgresRackRolePresenceState(t, db, role.ID())
	requireRackRolePostgresUpdateRecorded(t, replacedState, putClearedState)
	require.Equal(t, uint64(1), putClearedState.rackCount)
	require.True(t, putClearedState.row.LastUpdated.Equal(putClearedAt))

	patchResetAt := putClearedAt.Add(time.Minute)
	service = newPostgresRackRolePresenceService(t, db, patchResetAt)
	_, err = service.UpdateRackRole(
		t.Context(), principal, applicationdcim.UpdateRackRoleCommand{
			ID: role.ID(), Description: applicationdcim.FieldValue("reset before PATCH clear"),
		},
	)
	require.NoError(t, err)
	patchResetState := loadPostgresRackRolePresenceState(t, db, role.ID())
	requireRackRolePostgresUpdateRecorded(t, putClearedState, patchResetState)
	require.True(t, patchResetState.row.LastUpdated.Equal(patchResetAt))

	patchClearedAt := patchResetAt.Add(time.Minute)
	service = newPostgresRackRolePresenceService(t, db, patchClearedAt)
	patchCleared, err := service.UpdateRackRole(
		t.Context(), principal, applicationdcim.UpdateRackRoleCommand{
			ID: role.ID(), Description: applicationdcim.FieldValue(""),
		},
	)
	require.NoError(t, err)
	require.Empty(t, patchCleared.Description())
	require.Equal(t, uint64(1), patchCleared.RackCount())
	patchClearedState := loadPostgresRackRolePresenceState(t, db, role.ID())
	requireRackRolePostgresUpdateRecorded(t, patchResetState, patchClearedState)
	require.True(t, patchClearedState.row.LastUpdated.Equal(patchClearedAt))

	service = newPostgresRackRolePresenceService(t, db, patchClearedAt.Add(time.Minute))
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresRackRolePresenceState(t, db, role.ID())
			err := operation()
			require.Error(t, err)
			require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
			after := loadPostgresRackRolePresenceState(t, db, role.ID())
			require.Equal(t, before, after, "rejected mutation changed durable RackRole state")
		})
	}
	assertRejected("POST missing required identity", func() error {
		_, err := service.CreateRackRole(
			t.Context(), principal, applicationdcim.CreateRackRoleCommand{
				Slug: applicationdcim.FieldValue("rejected-rack-role"),
			},
		)
		return err
	})
	assertRejected("POST explicit null color", func() error {
		_, err := service.CreateRackRole(
			t.Context(), principal, applicationdcim.CreateRackRoleCommand{
				Name:  applicationdcim.FieldValue("Rejected Rack Role"),
				Slug:  applicationdcim.FieldValue("rejected-rack-role"),
				Color: applicationdcim.NullField[string](),
			},
		)
		return err
	})
	assertRejected("PUT missing required identity", func() error {
		_, err := service.ReplaceRackRole(
			t.Context(), principal, applicationdcim.ReplaceRackRoleCommand{ID: role.ID()},
		)
		return err
	})
	assertRejected("PUT blank color", func() error {
		_, err := service.ReplaceRackRole(
			t.Context(), principal, applicationdcim.ReplaceRackRoleCommand{
				ID: role.ID(), Name: applicationdcim.FieldValue("Durable Rack Role PUT Cleared"),
				Slug:  applicationdcim.FieldValue("durable-rack-role-put-cleared"),
				Color: applicationdcim.FieldValue(""),
			},
		)
		return err
	})
	assertRejected("PATCH explicit null description", func() error {
		_, err := service.UpdateRackRole(
			t.Context(), principal, applicationdcim.UpdateRackRoleCommand{
				ID: role.ID(), Description: applicationdcim.NullField[string](),
			},
		)
		return err
	})
	assertRejected("PATCH uppercase color", func() error {
		_, err := service.UpdateRackRole(
			t.Context(), principal, applicationdcim.UpdateRackRoleCommand{
				ID: role.ID(), Color: applicationdcim.FieldValue("ABCDEF"),
			},
		)
		return err
	})

	failureBefore := loadPostgresRackRolePresenceState(t, db, role.ID())
	recorderFailure := errors.New("forced PostgreSQL RackRole change recording failure")
	failingRecorder := &failingRackRoleChangeRecorder{err: recorderFailure}
	failingService := newPostgresRackRolePresenceServiceWithRecorder(
		t, db, patchClearedAt.Add(2*time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateRackRole(
		t.Context(), principal, applicationdcim.UpdateRackRoleCommand{
			ID: role.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(t, failureBefore, loadPostgresRackRolePresenceState(t, db, role.ID()))

	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Where(
		"kind = ? AND object_id = ?", domaindcim.RackRoleObjectType, role.ID().Int64(),
	).Order("id").Find(&changes).Error)
	require.NotEmpty(t, changes)
	for _, change := range changes {
		requireRackRoleChangeSnapshotExcludesCounter(t, change.BeforeData)
		requireRackRoleChangeSnapshotExcludesCounter(t, change.AfterData)
	}
}

type postgresRackRolePresenceState struct {
	row              dcimrow.RackRoleRow
	roleCount        int64
	rackCount        uint64
	changeCount      int64
	totalChangeCount int64
}

func requirePostgresRackRoleScalarRow(
	t *testing.T,
	actual dcimrow.RackRoleRow,
	expected dcimrow.RackRoleRow,
) {
	t.Helper()
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.Color, actual.Color)
	require.Equal(t, expected.Description, actual.Description)
}

func loadPostgresRackRolePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresRackRolePresenceState {
	t.Helper()
	var state postgresRackRolePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.RackRoleRow{}).Count(&state.roleCount).Error)
	loaded, err := NewRackRoleRepository(db).Get(t.Context(), id)
	require.NoError(t, err)
	state.rackCount = loaded.RackCount()
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackRoleObjectType, id.Int64(),
	).Count(&state.changeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireRackRolePostgresUpdateRecorded(
	t *testing.T,
	before postgresRackRolePresenceState,
	after postgresRackRolePresenceState,
) {
	t.Helper()
	require.Equal(t, before.roleCount, after.roleCount)
	require.Equal(t, before.rackCount, after.rackCount)
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireRackRoleChangeSnapshotExcludesCounter(t *testing.T, data []byte) {
	t.Helper()
	if len(data) == 0 || string(data) == "null" {
		return
	}
	var snapshot map[string]any
	require.NoError(t, json.Unmarshal(data, &snapshot))
	require.NotContains(t, snapshot, "rack_count")
}

type failingRackRoleChangeRecorder struct {
	err   error
	calls int
}

func (recorder *failingRackRoleChangeRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresRackRolePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) *applicationdcim.RackRoleService {
	t.Helper()
	return newPostgresRackRolePresenceServiceWithRecorder(
		t, db, now, postgreschangelog.NewRecorder(db),
	)
}

func newPostgresRackRolePresenceServiceWithRecorder(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.RackRoleService {
	t.Helper()
	service, err := applicationdcim.NewRackRoleService(
		NewRackRoleRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		recorder,
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}
