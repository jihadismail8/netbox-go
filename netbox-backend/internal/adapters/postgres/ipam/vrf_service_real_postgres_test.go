package ipam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrow "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestPostgresVRFScalarPresenceDurability(t *testing.T) {
	db, principal := newVRFScalarPostgres(t)
	repository := NewVRFRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	now := time.Date(2026, time.October, 13, 0, 40, 0, 0, time.UTC)
	serviceAt := func(at time.Time) *applicationipam.VRFService {
		return newPostgresVRFPresenceService(t, db, at, repository, recorder)
	}
	state := loadPostgresVRFPresenceState(t, db)
	require.Empty(t, state.vrfs)
	require.Empty(t, state.changes)

	for _, test := range []struct {
		name string
		rd   applicationipam.Field[string]
	}{
		{"omitted RD", applicationipam.OmittedField[string]()},
		{"null RD", applicationipam.NullField[string]()},
	} {
		t.Run("create defaults/"+test.name, func(t *testing.T) {
			created, err := serviceAt(now).CreateVRF(t.Context(), principal, applicationipam.CreateVRFCommand{
				Name: applicationipam.FieldValue("  " + test.name + "  "), RD: test.rd,
			})
			require.NoError(t, err)
			expected := ipamrow.VRFRow{
				RowMetadata: ipamrow.RowMetadata{ID: created.ID().Int64(), Created: now, LastUpdated: now},
				Name:        test.name, EnforceUnique: true,
			}
			state = requirePostgresVRFPresenceMutation(t, db, state, expected, true, principal.ID)
			requirePostgresVRFPresenceReload(t, repository, created)
		})
	}

	created, err := serviceAt(now).CreateVRF(t.Context(), principal, applicationipam.CreateVRFCommand{
		Name: applicationipam.FieldValue("  Durable VRF  "), RD: applicationipam.FieldValue("  arbitrary RD text  "),
		EnforceUnique: applicationipam.FieldValue(false), Description: applicationipam.FieldValue("  durable description  "),
		Comments: applicationipam.FieldValue("  " + strings.Repeat("c", 240) + "  "),
	})
	require.NoError(t, err)
	expected := ipamrow.VRFRow{
		RowMetadata: ipamrow.RowMetadata{ID: created.ID().Int64(), Created: now, LastUpdated: now},
		Name:        "Durable VRF", RD: stringPointer("arbitrary RD text"), EnforceUnique: false,
		Description: "durable description", Comments: strings.Repeat("c", 240),
	}
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, true, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, created)

	putAt := now.Add(time.Minute)
	replaced, err := serviceAt(putAt).ReplaceVRF(t.Context(), principal, applicationipam.ReplaceVRFCommand{
		ID: created.ID(), Name: applicationipam.FieldValue("  Replaced VRF  "),
	})
	require.NoError(t, err)
	expected.Name, expected.LastUpdated = "Replaced VRF", putAt
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, replaced)

	patchAt := putAt.Add(time.Minute)
	patched, err := serviceAt(patchAt).UpdateVRF(t.Context(), principal, applicationipam.UpdateVRFCommand{
		ID: created.ID(), Name: applicationipam.FieldValue("  Patched VRF  "),
	})
	require.NoError(t, err)
	expected.Name, expected.LastUpdated = "Patched VRF", patchAt
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, patched)

	clearAt := patchAt.Add(time.Minute)
	cleared, err := serviceAt(clearAt).ReplaceVRF(t.Context(), principal, applicationipam.ReplaceVRFCommand{
		ID: created.ID(), Name: applicationipam.FieldValue("Cleared VRF"), RD: applicationipam.NullField[string](),
		EnforceUnique: applicationipam.FieldValue(false), Description: applicationipam.FieldValue("  "), Comments: applicationipam.FieldValue("  "),
	})
	require.NoError(t, err)
	expected.Name, expected.RD, expected.Description, expected.Comments, expected.LastUpdated = "Cleared VRF", nil, "", "", clearAt
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, cleared)

	blankAt := clearAt.Add(time.Minute)
	blank, err := serviceAt(blankAt).UpdateVRF(t.Context(), principal, applicationipam.UpdateVRFCommand{
		ID: created.ID(), RD: applicationipam.FieldValue("  "), EnforceUnique: applicationipam.FieldValue(true),
	})
	require.NoError(t, err)
	expected.RD, expected.EnforceUnique, expected.LastUpdated = stringPointer(""), true, blankAt
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, blank)

	nullAt := blankAt.Add(time.Minute)
	null, err := serviceAt(nullAt).UpdateVRF(t.Context(), principal, applicationipam.UpdateVRFCommand{
		ID: created.ID(), RD: applicationipam.NullField[string](), EnforceUnique: applicationipam.FieldValue(false),
	})
	require.NoError(t, err)
	expected.RD, expected.EnforceUnique, expected.LastUpdated = nil, false, nullAt
	state = requirePostgresVRFPresenceMutation(t, db, state, expected, false, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, null)

	blankCreated, err := serviceAt(nullAt).CreateVRF(t.Context(), principal, applicationipam.CreateVRFCommand{
		Name: applicationipam.FieldValue("Blank RD create"), RD: applicationipam.FieldValue("  "), EnforceUnique: applicationipam.FieldValue(false),
	})
	require.NoError(t, err)
	state = requirePostgresVRFPresenceMutation(t, db, state, ipamrow.VRFRow{
		RowMetadata: ipamrow.RowMetadata{ID: blankCreated.ID().Int64(), Created: nullAt, LastUpdated: nullAt},
		Name:        "Blank RD create", RD: stringPointer(""), EnforceUnique: false,
	}, true, principal.ID)
	requirePostgresVRFPresenceReload(t, repository, blankCreated)

	for _, operation := range []string{"create", "replace", "update"} {
		for _, test := range []struct {
			name    string
			fields  applicationipam.CreateVRFCommand
			field   string
			reasons []string
		}{
			{"null name", applicationipam.CreateVRFCommand{Name: applicationipam.NullField[string]()}, "name", []string{"null"}},
			{"blank name", applicationipam.CreateVRFCommand{Name: applicationipam.FieldValue("  ")}, "name", []string{"blank"}},
			{"null boolean", applicationipam.CreateVRFCommand{EnforceUnique: applicationipam.NullField[bool]()}, "enforce_unique", []string{"null"}},
			{"null description", applicationipam.CreateVRFCommand{Description: applicationipam.NullField[string]()}, "description", []string{"null"}},
			{"null comments", applicationipam.CreateVRFCommand{Comments: applicationipam.NullField[string]()}, "comments", []string{"null"}},
			{"name NUL", applicationipam.CreateVRFCommand{Name: applicationipam.FieldValue("bad\x00name")}, "name", []string{"null_characters_not_allowed"}},
			{"RD NUL", applicationipam.CreateVRFCommand{RD: applicationipam.FieldValue("bad\x00RD")}, "rd", []string{"null_characters_not_allowed"}},
			{"description NUL", applicationipam.CreateVRFCommand{Description: applicationipam.FieldValue("bad\x00description")}, "description", []string{"null_characters_not_allowed"}},
			{"comments NUL", applicationipam.CreateVRFCommand{Comments: applicationipam.FieldValue("bad\x00comments")}, "comments", []string{"null_characters_not_allowed"}},
			{"overlength name and NUL", applicationipam.CreateVRFCommand{Name: applicationipam.FieldValue(strings.Repeat("n", 100) + "\x00")}, "name", []string{"max_length", "null_characters_not_allowed"}},
			{"overlength RD and NUL", applicationipam.CreateVRFCommand{RD: applicationipam.FieldValue(strings.Repeat("r", 21) + "\x00")}, "rd", []string{"max_length", "null_characters_not_allowed"}},
			{"overlength description and NUL", applicationipam.CreateVRFCommand{Description: applicationipam.FieldValue(strings.Repeat("d", 200) + "\x00")}, "description", []string{"max_length", "null_characters_not_allowed"}},
		} {
			t.Run("validation non-effects/"+operation+"/"+test.name, func(t *testing.T) {
				before := loadPostgresVRFPresenceState(t, db)
				fields := test.fields
				if operation != "update" && fields.Name.State() == applicationipam.FieldOmitted {
					fields.Name = applicationipam.FieldValue("Rejected VRF")
				}
				result, err := mutatePostgresVRFPresence(t.Context(), serviceAt(nullAt.Add(time.Minute)), principal, operation, created.ID(), fields)
				require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
				require.Nil(t, result)
				var reasons []string
				for _, violation := range shared.ViolationsOf(err) {
					require.Equal(t, test.field, violation.Field)
					reasons = append(reasons, violation.Reason)
				}
				require.Equal(t, test.reasons, reasons)
				require.Equal(t, before, loadPostgresVRFPresenceState(t, db))
			})
		}
		if operation != "update" {
			t.Run("required omission non-effects/"+operation, func(t *testing.T) {
				before := loadPostgresVRFPresenceState(t, db)
				_, err := mutatePostgresVRFPresence(t.Context(), serviceAt(nullAt.Add(time.Minute)), principal, operation, created.ID(), applicationipam.CreateVRFCommand{})
				require.Equal(t, []shared.FieldViolation{{Field: "name", Reason: "required", Description: "This field is required."}}, shared.ViolationsOf(err))
				require.Equal(t, before, loadPostgresVRFPresenceState(t, db))
			})
		}
		for _, name := range []applicationipam.Field[string]{applicationipam.NullField[string](), applicationipam.OmittedField[string]()} {
			t.Run(fmt.Sprintf("mixed validation non-effects/%s/name-state-%d", operation, name.State()), func(t *testing.T) {
				before := loadPostgresVRFPresenceState(t, db)
				result, err := mutatePostgresVRFPresence(t.Context(), serviceAt(nullAt.Add(time.Minute)), principal, operation, created.ID(), applicationipam.CreateVRFCommand{
					Name: name, RD: applicationipam.FieldValue(strings.Repeat("界", 21) + "\x00"),
					EnforceUnique: applicationipam.NullField[bool](), Description: applicationipam.FieldValue(strings.Repeat("界", 200) + "\x00"),
					Comments: applicationipam.FieldValue("c\x00"),
				})
				require.Error(t, err)
				require.Nil(t, result)
				var expected []shared.FieldViolation
				if name.State() == applicationipam.FieldNull {
					expected = append(expected, shared.FieldViolation{Field: "name", Reason: "null", Description: "This field may not be null."})
				} else if operation != "update" {
					expected = append(expected, shared.FieldViolation{Field: "name", Reason: "required", Description: "This field is required."})
				}
				expected = append(expected,
					shared.FieldViolation{Field: "rd", Reason: "max_length", Description: "Ensure this field has no more than 21 characters."},
					shared.FieldViolation{Field: "rd", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
					shared.FieldViolation{Field: "enforce_unique", Reason: "null", Description: "This field may not be null."},
					shared.FieldViolation{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
					shared.FieldViolation{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
					shared.FieldViolation{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
				)
				require.Equal(t, expected, shared.ViolationsOf(err), "all independent failures must survive without fallback blanks or mask errors")
				require.Equal(t, before, loadPostgresVRFPresenceState(t, db))
			})
		}
	}

	for _, failure := range []string{"repository", "recorder"} {
		for _, operation := range []string{"create", "replace", "update"} {
			t.Run("delegate then error rollback/"+failure+"/"+operation, func(t *testing.T) {
				before := loadPostgresVRFPresenceState(t, db)
				sentinel := errors.New("forced VRF failure after actual delegate")
				observed := 0
				observe := func(ctx context.Context, item *domainipam.VRF) {
					observed++
					tx, bound := postgresTransaction.FromContext(ctx)
					require.True(t, bound, "the delegate must use the application-owned transaction")
					inside := loadPostgresVRFPresenceState(t, tx)
					current := postgresVRFPresenceRow(t, inside, item.ID().Int64())
					require.Equal(t, "rollback "+operation, current.Name, "observe the actual persisted mutation before injecting failure")
					require.False(t, current.EnforceUnique)
					require.Equal(t, stringPointer("rollback RD"), current.RD)
					require.Equal(t, "rollback description", current.Description)
					require.Equal(t, "rollback comments", current.Comments)
					require.True(t, current.LastUpdated.Equal(nullAt.Add(2*time.Minute)))
					if failure == "recorder" {
						require.Len(t, inside.changes, len(before.changes)+1, "observe actual transaction-local change insertion before injecting failure")
						requirePostgresVRFPresenceChange(t, inside, before, current, operation == "create", principal.ID)
					} else {
						require.Equal(t, before.changes, inside.changes)
					}
				}
				failingRepository := &vrfPresenceFailingRepository{VRFRepository: repository, observe: observe, err: sentinel}
				failingRecorder := &vrfPresenceFailingRecorder{delegate: recorder, repository: repository, observe: observe, err: sentinel}
				var activeRepository applicationipam.VRFRepository = repository
				var activeRecorder applicationchangelog.Recorder = recorder
				if failure == "repository" {
					activeRepository = failingRepository
				} else {
					activeRecorder = failingRecorder
				}
				service := newPostgresVRFPresenceService(t, db, nullAt.Add(2*time.Minute), activeRepository, activeRecorder)
				result, err := mutatePostgresVRFPresence(t.Context(), service, principal, operation, created.ID(), applicationipam.CreateVRFCommand{
					Name: applicationipam.FieldValue("rollback " + operation), RD: applicationipam.FieldValue("rollback RD"),
					EnforceUnique: applicationipam.FieldValue(false), Description: applicationipam.FieldValue("rollback description"),
					Comments: applicationipam.FieldValue("rollback comments"),
				})
				require.ErrorIs(t, err, sentinel)
				require.Nil(t, result)
				require.Equal(t, 1, observed)
				require.Equal(t, before, loadPostgresVRFPresenceState(t, db), "rows, timestamps and append-only audit state must all roll back")
			})
		}
	}
}

func mutatePostgresVRFPresence(ctx context.Context, service *applicationipam.VRFService, principal identity.Principal, operation string, id shared.ID, fields applicationipam.CreateVRFCommand) (*domainipam.VRF, error) {
	switch operation {
	case "create":
		return service.CreateVRF(ctx, principal, fields)
	case "replace":
		return service.ReplaceVRF(ctx, principal, applicationipam.ReplaceVRFCommand{ID: id, Name: fields.Name, RD: fields.RD, EnforceUnique: fields.EnforceUnique, Description: fields.Description, Comments: fields.Comments})
	default:
		return service.UpdateVRF(ctx, principal, applicationipam.UpdateVRFCommand{ID: id, Name: fields.Name, RD: fields.RD, EnforceUnique: fields.EnforceUnique, Description: fields.Description, Comments: fields.Comments})
	}
}

type postgresVRFPresenceState struct {
	vrfs    []ipamrow.VRFRow
	changes []postgreschangelog.ChangeRow
}

func loadPostgresVRFPresenceState(t *testing.T, db *gorm.DB) postgresVRFPresenceState {
	t.Helper()
	var state postgresVRFPresenceState
	require.NoError(t, db.Order("id").Find(&state.vrfs).Error)
	require.NoError(t, db.Order("id").Find(&state.changes).Error)
	for _, model := range []any{&ipamrow.PrefixRow{}, &ipamrow.IPAddressRow{}} {
		var count int64
		require.NoError(t, db.Model(model).Count(&count).Error)
		require.Zero(t, count, "I12 does not exercise downstream uniqueness or reference policy")
	}
	return state
}

func postgresVRFPresenceRow(t *testing.T, state postgresVRFPresenceState, id int64) ipamrow.VRFRow {
	t.Helper()
	for _, row := range state.vrfs {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("VRF %d is absent", id)
	return ipamrow.VRFRow{}
}

func requirePostgresVRFPresenceMutation(t *testing.T, db *gorm.DB, before postgresVRFPresenceState, expected ipamrow.VRFRow, create bool, actorID int64) postgresVRFPresenceState {
	t.Helper()
	after := loadPostgresVRFPresenceState(t, db)
	count := len(before.vrfs)
	if create {
		count++
	}
	require.Len(t, after.vrfs, count)
	for _, row := range before.vrfs {
		if row.ID != expected.ID {
			require.Equal(t, row, postgresVRFPresenceRow(t, after, row.ID))
		}
	}
	actual := postgresVRFPresenceRow(t, after, expected.ID)
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	actual.Created, actual.LastUpdated = expected.Created, expected.LastUpdated
	require.Equal(t, expected, actual, "raw nullable RD, false, blank and concrete values must survive persistence")
	requirePostgresVRFPresenceChange(t, after, before, expected, create, actorID)
	return after
}

func requirePostgresVRFPresenceChange(t *testing.T, after, before postgresVRFPresenceState, current ipamrow.VRFRow, create bool, actorID int64) {
	t.Helper()
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	change := after.changes[len(after.changes)-1]
	require.Positive(t, change.ID)
	require.Equal(t, actorID, change.ActorID)
	require.Equal(t, domainipam.VRFObjectType, change.Kind)
	require.Equal(t, current.ID, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(current.LastUpdated))
	if create {
		require.Equal(t, "create", change.Action)
		require.Empty(t, change.BeforeData)
	} else {
		require.Equal(t, "update", change.Action)
		prior := postgresVRFPresenceRow(t, before, current.ID)
		require.True(t, prior.Created.Equal(current.Created))
		encoded, err := json.Marshal(prior)
		require.NoError(t, err)
		require.JSONEq(t, string(encoded), string(change.BeforeData))
	}
	encoded, err := json.Marshal(current)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(change.AfterData))
}

func requirePostgresVRFPresenceReload(t *testing.T, repository *VRFRepository, expected *domainipam.VRF) {
	t.Helper()
	actual, err := repository.Get(t.Context(), expected.ID())
	require.NoError(t, err)
	actualState, expectedState := actual.State(), expected.State()
	require.True(t, actualState.Created.Equal(expectedState.Created.Time))
	require.True(t, actualState.LastUpdated.Equal(expectedState.LastUpdated.Time))
	actualState.Created, actualState.LastUpdated = expectedState.Created, expectedState.LastUpdated
	require.Equal(t, expectedState, actualState)
	require.Equal(t, expected.Snapshot(), actual.Snapshot())
	require.Zero(t, actual.IPAddressCount())
	require.Zero(t, actual.PrefixCount())
}

type vrfPresenceFailingRepository struct {
	applicationipam.VRFRepository
	observe func(context.Context, *domainipam.VRF)
	err     error
}

func (repository *vrfPresenceFailingRepository) Create(ctx context.Context, item *domainipam.VRF) error {
	if err := repository.VRFRepository.Create(ctx, item); err != nil {
		return err
	}
	repository.observe(ctx, item)
	return repository.err
}

func (repository *vrfPresenceFailingRepository) Update(ctx context.Context, item *domainipam.VRF) error {
	if err := repository.VRFRepository.Update(ctx, item); err != nil {
		return err
	}
	repository.observe(ctx, item)
	return repository.err
}

type vrfPresenceFailingRecorder struct {
	delegate   applicationchangelog.Recorder
	repository *VRFRepository
	observe    func(context.Context, *domainipam.VRF)
	err        error
}

func (recorder *vrfPresenceFailingRecorder) Record(ctx context.Context, change applicationchangelog.Change) error {
	if err := recorder.delegate.Record(ctx, change); err != nil {
		return err
	}
	item, err := recorder.repository.Get(ctx, change.ObjectID)
	if err != nil {
		return err
	}
	recorder.observe(ctx, item)
	return recorder.err
}

func newPostgresVRFPresenceService(t *testing.T, db *gorm.DB, now time.Time, repository applicationipam.VRFRepository, recorder applicationchangelog.Recorder) *applicationipam.VRFService {
	t.Helper()
	service, err := applicationipam.NewVRFService(repository, postgresTransaction.NewUnitOfWork(db), recorder, authz.AllowAll{}, ipamPostgresConcurrencyClock{now: shared.NewTimestamp(now)})
	require.NoError(t, err)
	return service
}

func newVRFScalarPostgres(t *testing.T) (*gorm.DB, identity.Principal) {
	t.Helper()
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	base, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	schema := fmt.Sprintf("ipam_vrf_scalar_presence_%d", time.Now().UnixNano())
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)
	t.Cleanup(func() {
		if err := base.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error("could not drop the owned VRF scalar-presence schema")
		}
		var remaining int64
		if err := base.Raw("SELECT count(*) FROM pg_namespace WHERE nspname = ?", schema).Scan(&remaining).Error; err != nil || remaining != 0 {
			t.Error("the owned VRF scalar-presence schema remained after cleanup")
		}
		if err := baseSQL.Close(); err != nil {
			t.Error("could not close the PostgreSQL cleanup connection")
		}
	})
	db, err := gorm.Open(postgres.Open(ipamPostgresDSNWithSearchPath(t, dsn, schema)), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error("could not close the scoped PostgreSQL VRF connection")
		}
	})
	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	// IPAddress's existing FK requires these empty DCIM tables; no DCIM fixtures are seeded.
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	for _, descriptor := range ipamrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	entries = append(entries, bootstrap.Entry{Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{}, Dependencies: []string{"go_identity_users"}})
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(entries))
	require.Empty(t, result.Existing)
	now := time.Date(2026, time.October, 13, 0, 40, 0, 0, time.UTC)
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "vrf-presence-test", PasswordHash: "not-used-by-tests", IsActive: true, IsStaff: true, IsSuperuser: true,
		Permissions: []byte(`[]`), Created: now, Updated: now,
	}).Error)
	return db, identity.Principal{ID: 1, Username: "vrf-presence-test", IsSuperuser: true}
}
