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

func TestPostgresDeviceRoleScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 25, 6, 0, 0, 0, time.UTC)
	service := newPostgresDeviceRolePresenceService(
		t, db, createdAt, postgreschangelog.NewRecorder(db),
	)

	root, err := service.CreateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.CreateDeviceRoleCommand{
			Name: applicationdcim.FieldValue("  Durable Root  "),
			Slug: applicationdcim.FieldValue("  durable-root  "),
		},
	)
	require.NoError(t, err)
	require.True(t, root.Parent().IsRoot())
	require.Equal(t, "Durable Root", root.Name())
	require.Equal(t, "durable-root", root.Slug().String())
	require.Equal(t, domaindcim.DeviceRoleDefaultColor, root.Color().String())
	require.True(t, root.VMRole())
	require.Empty(t, root.Description())
	require.Empty(t, root.Comments())
	rootState := loadPostgresDeviceRolePresenceState(t, db, root.ID())
	require.Equal(t, int64(1), rootState.roleCount)
	require.Equal(t, int64(1), rootState.changeCount)
	require.Equal(t, int64(1), rootState.totalChangeCount)
	requirePostgresDeviceRoleRow(t, rootState.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: root.ID().Int64(), Created: createdAt, LastUpdated: createdAt,
		},
		Name: "Durable Root", Slug: "durable-root", Color: domaindcim.DeviceRoleDefaultColor,
		VMRole: true,
	})

	child, err := service.CreateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.CreateDeviceRoleCommand{
			Parent:      applicationdcim.FieldValue(root.ID()),
			Name:        applicationdcim.FieldValue("  Durable Child  "),
			Slug:        applicationdcim.FieldValue("  durable-child  "),
			Color:       applicationdcim.FieldValue("  00ff00  "),
			VMRole:      applicationdcim.FieldValue(false),
			Description: applicationdcim.FieldValue("  durable description  "),
			Comments:    applicationdcim.FieldValue("  durable comments  "),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "00ff00", child.Color().String())
	require.False(t, child.VMRole())
	require.Equal(t, "durable description", child.Description())
	require.Equal(t, "durable comments", child.Comments())
	parentID, present := child.Parent().Get()
	require.True(t, present)
	require.Equal(t, root.ID(), parentID)

	created := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	require.Equal(t, int64(2), created.roleCount)
	require.Equal(t, int64(1), created.changeCount)
	require.Equal(t, int64(2), created.totalChangeCount)
	requirePostgresDeviceRoleRow(t, created.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: child.ID().Int64(), Created: createdAt, LastUpdated: createdAt,
		},
		ParentID: idPointer(root.ID().Int64()),
		Name:     "Durable Child", Slug: "durable-child", Color: "00ff00",
		VMRole: false, Description: "durable description", Comments: "durable comments",
	})

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresDeviceRolePresenceService(
		t, db, patchedAt, postgreschangelog.NewRecorder(db),
	)
	patched, err := service.UpdateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceRoleCommand{
			Color:       applicationdcim.FieldValue("  1122aa  "),
			Description: applicationdcim.FieldValue(""),
			Comments:    applicationdcim.FieldValue(""),
			ID:          child.ID(),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Durable Child", patched.Name())
	require.Equal(t, "durable-child", patched.Slug().String())
	require.Equal(t, "1122aa", patched.Color().String())
	require.False(t, patched.VMRole())
	require.Empty(t, patched.Description())
	require.Empty(t, patched.Comments())
	parentID, present = patched.Parent().Get()
	require.True(t, present)
	require.Equal(t, root.ID(), parentID)
	patchedState := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	requirePostgresDeviceRoleRow(t, patchedState.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: child.ID().Int64(), Created: createdAt, LastUpdated: patchedAt,
		},
		ParentID: idPointer(root.ID().Int64()),
		Name:     "Durable Child", Slug: "durable-child", Color: "1122aa",
		VMRole: false,
	})
	requireDeviceRolePostgresUpdateRecorded(t, created, patchedState, patchedAt)

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresDeviceRolePresenceService(
		t, db, replacedAt, postgreschangelog.NewRecorder(db),
	)
	replaced, err := service.ReplaceDeviceRole(
		t.Context(),
		principal,
		applicationdcim.ReplaceDeviceRoleCommand{
			ID:   child.ID(),
			Name: applicationdcim.FieldValue("  Durable Child Replaced  "),
			Slug: applicationdcim.FieldValue("  durable-child-replaced  "),
		},
	)
	require.NoError(t, err)
	require.True(t, replaced.Parent().IsRoot(), "PUT parent omission must reset to root")
	require.Equal(t, "1122aa", replaced.Color().String(), "PUT omission must preserve color")
	require.False(t, replaced.VMRole(), "PUT omission must preserve explicit false")
	require.Empty(t, replaced.Description(), "PUT omission must preserve description")
	require.Empty(t, replaced.Comments(), "PUT omission must preserve comments")
	replacedState := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	requirePostgresDeviceRoleRow(t, replacedState.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: child.ID().Int64(), Created: createdAt, LastUpdated: replacedAt,
		},
		Name: "Durable Child Replaced", Slug: "durable-child-replaced", Color: "1122aa",
		VMRole: false,
	})
	requireDeviceRolePostgresUpdateRecorded(t, patchedState, replacedState, replacedAt)

	parentedAt := replacedAt.Add(time.Minute)
	service = newPostgresDeviceRolePresenceService(
		t, db, parentedAt, postgreschangelog.NewRecorder(db),
	)
	parented, err := service.UpdateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceRoleCommand{
			ID: child.ID(), Parent: applicationdcim.FieldValue(root.ID()),
		},
	)
	require.NoError(t, err)
	parentID, present = parented.Parent().Get()
	require.True(t, present)
	require.Equal(t, root.ID(), parentID)
	parentedState := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	requirePostgresDeviceRoleRow(t, parentedState.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: child.ID().Int64(), Created: createdAt, LastUpdated: parentedAt,
		},
		ParentID: idPointer(root.ID().Int64()),
		Name:     "Durable Child Replaced", Slug: "durable-child-replaced", Color: "1122aa",
		VMRole: false,
	})
	requireDeviceRolePostgresUpdateRecorded(t, replacedState, parentedState, parentedAt)

	clearedAt := parentedAt.Add(time.Minute)
	service = newPostgresDeviceRolePresenceService(
		t, db, clearedAt, postgreschangelog.NewRecorder(db),
	)
	cleared, err := service.UpdateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceRoleCommand{
			ID: child.ID(), Parent: applicationdcim.NullField[shared.ID](),
		},
	)
	require.NoError(t, err)
	require.True(t, cleared.Parent().IsRoot())
	clearedState := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	requirePostgresDeviceRoleRow(t, clearedState.row, dcimrow.DeviceRoleRow{
		RowMetadata: dcimrow.RowMetadata{
			ID: child.ID().Int64(), Created: createdAt, LastUpdated: clearedAt,
		},
		Name: "Durable Child Replaced", Slug: "durable-child-replaced", Color: "1122aa",
		VMRole: false,
	})
	requireDeviceRolePostgresUpdateRecorded(t, parentedState, clearedState, clearedAt)

	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresDeviceRolePresenceState(t, db, child.ID())
			operationErr := operation()
			require.Error(t, operationErr)
			require.True(t, shared.HasReason(operationErr, shared.ErrorReasonValidation), operationErr)
			require.Equal(t, before, loadPostgresDeviceRolePresenceState(t, db, child.ID()))
		})
	}
	assertRejected("POST missing required identity", func() error {
		_, createErr := service.CreateDeviceRole(
			t.Context(), principal, applicationdcim.CreateDeviceRoleCommand{},
		)
		return createErr
	})
	assertRejected("PUT explicit null color", func() error {
		_, replaceErr := service.ReplaceDeviceRole(
			t.Context(),
			principal,
			applicationdcim.ReplaceDeviceRoleCommand{
				ID: child.ID(), Name: applicationdcim.FieldValue("Durable Child Replaced"),
				Slug:  applicationdcim.FieldValue("durable-child-replaced"),
				Color: applicationdcim.NullField[string](),
			},
		)
		return replaceErr
	})
	assertRejected("PATCH explicit null vm role", func() error {
		_, updateErr := service.UpdateDeviceRole(
			t.Context(), principal, applicationdcim.UpdateDeviceRoleCommand{
				ID: child.ID(), VMRole: applicationdcim.NullField[bool](),
			},
		)
		return updateErr
	})
	assertRejected("PATCH uppercase color", func() error {
		_, updateErr := service.UpdateDeviceRole(
			t.Context(), principal, applicationdcim.UpdateDeviceRoleCommand{
				ID: child.ID(), Color: applicationdcim.FieldValue("ABCDEF"),
			},
		)
		return updateErr
	})

	failureBefore := loadPostgresDeviceRolePresenceState(t, db, child.ID())
	recorderFailure := errors.New("forced PostgreSQL DeviceRole change recording failure")
	failingRecorder := &failingDeviceRolePresenceRecorder{err: recorderFailure}
	failingService := newPostgresDeviceRolePresenceService(
		t, db, clearedAt.Add(time.Minute), failingRecorder,
	)
	_, err = failingService.UpdateDeviceRole(
		t.Context(),
		principal,
		applicationdcim.UpdateDeviceRoleCommand{
			ID: child.ID(), Description: applicationdcim.FieldValue("must roll back"),
		},
	)
	require.ErrorIs(t, err, recorderFailure)
	require.Equal(t, 1, failingRecorder.calls)
	require.Equal(t, failureBefore, loadPostgresDeviceRolePresenceState(t, db, child.ID()))
}

type postgresDeviceRolePresenceState struct {
	row              dcimrow.DeviceRoleRow
	roleCount        int64
	changeCount      int64
	totalChangeCount int64
}

func loadPostgresDeviceRolePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresDeviceRolePresenceState {
	t.Helper()
	var state postgresDeviceRolePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.DeviceRoleRow{}).Count(&state.roleCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceRoleObjectType, id.Int64(),
	).Count(&state.changeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requirePostgresDeviceRoleRow(
	t *testing.T,
	actual dcimrow.DeviceRoleRow,
	expected dcimrow.DeviceRoleRow,
) {
	t.Helper()
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, actual.Created.Equal(expected.Created), "unexpected persisted created timestamp")
	require.True(
		t, actual.LastUpdated.Equal(expected.LastUpdated),
		"unexpected persisted last_updated timestamp",
	)
	require.Equal(t, expected.ParentID, actual.ParentID)
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.Color, actual.Color)
	require.Equal(t, expected.VMRole, actual.VMRole)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

func requireDeviceRolePostgresUpdateRecorded(
	t *testing.T,
	before postgresDeviceRolePresenceState,
	after postgresDeviceRolePresenceState,
	expectedLastUpdated time.Time,
) {
	t.Helper()
	require.Equal(t, before.roleCount, after.roleCount)
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.True(t, after.row.Created.Equal(before.row.Created))
	require.True(t, after.row.LastUpdated.Equal(expectedLastUpdated))
}

type failingDeviceRolePresenceRecorder struct {
	err   error
	calls int
}

func (recorder *failingDeviceRolePresenceRecorder) Record(
	context.Context,
	applicationchangelog.Change,
) error {
	recorder.calls++
	return recorder.err
}

func newPostgresDeviceRolePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
	recorder applicationchangelog.Recorder,
) *applicationdcim.DeviceRoleService {
	t.Helper()
	service, err := applicationdcim.NewDeviceRoleService(
		NewDeviceRoleRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		recorder,
		authz.AllowAll{},
		fixedDeviceRoleClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}
