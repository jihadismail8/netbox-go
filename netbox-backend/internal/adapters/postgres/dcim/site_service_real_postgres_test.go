package dcim

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
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
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestSiteServiceConcurrentDuplicateCreatePostgres(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	service, err := applicationdcim.NewSiteService(
		NewSiteRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(time.Now().UTC())},
	)
	require.NoError(t, err)

	type result struct {
		site *domaindcim.Site
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for index := range 2 {
		index := index
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			site, createErr := service.CreateSite(
				t.Context(), principal, applicationdcim.CreateSiteCommand{
					Name: applicationdcim.FieldValue("Concurrent Site"),
					Slug: applicationdcim.FieldValue("concurrent-site"),
					Description: applicationdcim.FieldValue(
						fmt.Sprintf("writer-%d", index),
					),
				},
			)
			results <- result{site: site, err: createErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	successes, conflicts := 0, 0
	for createResult := range results {
		if createResult.err == nil {
			require.NotNil(t, createResult.site)
			successes++
			continue
		}
		require.True(t, shared.HasReason(createResult.err, shared.ErrorReasonConflict), createResult.err)
		conflicts++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)

	var sites int64
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&sites).Error)
	require.Equal(t, int64(1), sites)
	var changes int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).
		Where("kind = ? AND action = ?", domaindcim.SiteObjectType, "create").
		Count(&changes).Error)
	require.Equal(t, int64(1), changes, "the rejected duplicate must not leave a change row")
}

func TestPostgresSiteScalarPresenceDurability(t *testing.T) {
	db, principal := newSiteConcurrencyPostgres(t)
	createdAt := time.Date(2026, time.August, 23, 10, 0, 0, 0, time.UTC)
	service := newPostgresSitePresenceService(t, db, createdAt)

	site, err := service.CreateSite(t.Context(), principal, applicationdcim.CreateSiteCommand{
		Name: applicationdcim.FieldValue("  Durable Site  "),
		Slug: applicationdcim.FieldValue("  durable-site  "),
	})
	require.NoError(t, err)
	require.Equal(t, "Durable Site", site.Name())
	require.Equal(t, "durable-site", site.Slug().String())
	require.Equal(t, domaindcim.SiteStatusActive, site.Status())
	require.Empty(t, site.Facility())
	require.Empty(t, site.Description())
	require.Empty(t, site.Comments())

	created := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, created.row, dcimrow.SiteRow{
		Name: "Durable Site", Slug: "durable-site", Status: "active",
	})
	require.Equal(t, int64(1), created.siteCount)
	require.Equal(t, int64(1), created.changeCount)
	require.Equal(t, int64(1), created.totalChangeCount)
	require.True(t, created.row.Created.Equal(createdAt))
	require.True(t, created.row.LastUpdated.Equal(createdAt))

	patchedAt := createdAt.Add(time.Minute)
	service = newPostgresSitePresenceService(t, db, patchedAt)
	patched, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
		ID:          site.ID(),
		Status:      applicationdcim.FieldValue("staging"),
		Facility:    applicationdcim.FieldValue("  D1  "),
		Description: applicationdcim.FieldValue("  durable description  "),
		Comments:    applicationdcim.FieldValue("  durable comments  "),
	})
	require.NoError(t, err)
	require.Equal(t, "staging", patched.Status().String())
	require.Equal(t, "D1", patched.Facility())
	require.Equal(t, "durable description", patched.Description())
	require.Equal(t, "durable comments", patched.Comments())

	patchedState := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, patchedState.row, dcimrow.SiteRow{
		Name: "Durable Site", Slug: "durable-site", Status: "staging",
		Facility: "D1", Description: "durable description", Comments: "durable comments",
	})
	require.Equal(t, int64(1), patchedState.siteCount)
	require.Equal(t, created.changeCount+1, patchedState.changeCount)
	require.Equal(t, created.totalChangeCount+1, patchedState.totalChangeCount)
	require.Equal(t, created.row.Created, patchedState.row.Created)
	require.True(t, patchedState.row.LastUpdated.Equal(patchedAt))

	replacedAt := patchedAt.Add(time.Minute)
	service = newPostgresSitePresenceService(t, db, replacedAt)
	replaced, err := service.ReplaceSite(t.Context(), principal, applicationdcim.ReplaceSiteCommand{
		ID:   site.ID(),
		Name: applicationdcim.FieldValue("  Durable Site Renamed  "),
		Slug: applicationdcim.FieldValue("  durable-site-renamed  "),
	})
	require.NoError(t, err)
	require.Equal(t, "Durable Site Renamed", replaced.Name())
	require.Equal(t, "durable-site-renamed", replaced.Slug().String())
	require.Equal(t, "staging", replaced.Status().String(), "PUT omission must preserve status")
	require.Equal(t, "D1", replaced.Facility(), "PUT omission must preserve facility")
	require.Equal(t, "durable description", replaced.Description(), "PUT omission must preserve description")
	require.Equal(t, "durable comments", replaced.Comments(), "PUT omission must preserve comments")

	replacedState := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, replacedState.row, dcimrow.SiteRow{
		Name: "Durable Site Renamed", Slug: "durable-site-renamed", Status: "staging",
		Facility: "D1", Description: "durable description", Comments: "durable comments",
	})
	require.Equal(t, int64(1), replacedState.siteCount)
	require.Equal(t, patchedState.changeCount+1, replacedState.changeCount)
	require.Equal(t, patchedState.totalChangeCount+1, replacedState.totalChangeCount)
	require.Equal(t, created.row.Created, replacedState.row.Created)
	require.True(t, replacedState.row.LastUpdated.Equal(replacedAt))

	putClearedAt := replacedAt.Add(time.Minute)
	service = newPostgresSitePresenceService(t, db, putClearedAt)
	putCleared, err := service.ReplaceSite(t.Context(), principal, applicationdcim.ReplaceSiteCommand{
		ID:          site.ID(),
		Name:        applicationdcim.FieldValue("Durable Site PUT Cleared"),
		Slug:        applicationdcim.FieldValue("durable-site-put-cleared"),
		Facility:    applicationdcim.FieldValue(""),
		Description: applicationdcim.FieldValue(""),
		Comments:    applicationdcim.FieldValue(""),
	})
	require.NoError(t, err)
	require.Equal(t, "staging", putCleared.Status().String(), "PUT status omission must preserve state")
	require.Empty(t, putCleared.Facility())
	require.Empty(t, putCleared.Description())
	require.Empty(t, putCleared.Comments())
	putClearedState := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, putClearedState.row, dcimrow.SiteRow{
		Name: "Durable Site PUT Cleared", Slug: "durable-site-put-cleared", Status: "staging",
	})
	require.Equal(t, replacedState.siteCount, putClearedState.siteCount)
	require.Equal(t, replacedState.changeCount+1, putClearedState.changeCount)
	require.Equal(t, replacedState.totalChangeCount+1, putClearedState.totalChangeCount)
	require.Equal(t, created.row.Created, putClearedState.row.Created)
	require.True(t, putClearedState.row.LastUpdated.Equal(putClearedAt))

	patchResetAt := putClearedAt.Add(time.Minute)
	service = newPostgresSitePresenceService(t, db, patchResetAt)
	patchReset, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
		ID:          site.ID(),
		Facility:    applicationdcim.FieldValue("  P2  "),
		Description: applicationdcim.FieldValue("  reset before PATCH clear  "),
		Comments:    applicationdcim.FieldValue("  reset comments  "),
	})
	require.NoError(t, err)
	require.Equal(t, "P2", patchReset.Facility())
	require.Equal(t, "reset before PATCH clear", patchReset.Description())
	require.Equal(t, "reset comments", patchReset.Comments())
	patchResetState := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, patchResetState.row, dcimrow.SiteRow{
		Name: "Durable Site PUT Cleared", Slug: "durable-site-put-cleared", Status: "staging",
		Facility: "P2", Description: "reset before PATCH clear", Comments: "reset comments",
	})
	require.Equal(t, putClearedState.siteCount, patchResetState.siteCount)
	require.Equal(t, putClearedState.changeCount+1, patchResetState.changeCount)
	require.Equal(t, putClearedState.totalChangeCount+1, patchResetState.totalChangeCount)
	require.Equal(t, created.row.Created, patchResetState.row.Created)
	require.True(t, patchResetState.row.LastUpdated.Equal(patchResetAt))

	patchClearedAt := patchResetAt.Add(time.Minute)
	service = newPostgresSitePresenceService(t, db, patchClearedAt)
	patchCleared, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
		ID:          site.ID(),
		Facility:    applicationdcim.FieldValue(""),
		Description: applicationdcim.FieldValue(""),
		Comments:    applicationdcim.FieldValue(""),
	})
	require.NoError(t, err)
	require.Empty(t, patchCleared.Facility())
	require.Empty(t, patchCleared.Description())
	require.Empty(t, patchCleared.Comments())
	patchClearedState := loadPostgresSitePresenceState(t, db, site.ID())
	requirePostgresSiteScalarRow(t, patchClearedState.row, dcimrow.SiteRow{
		Name: "Durable Site PUT Cleared", Slug: "durable-site-put-cleared", Status: "staging",
	})
	require.Equal(t, patchResetState.siteCount, patchClearedState.siteCount)
	require.Equal(t, patchResetState.changeCount+1, patchClearedState.changeCount)
	require.Equal(t, patchResetState.totalChangeCount+1, patchClearedState.totalChangeCount)
	require.Equal(t, created.row.Created, patchClearedState.row.Created)
	require.True(t, patchClearedState.row.LastUpdated.Equal(patchClearedAt))

	service = newPostgresSitePresenceService(t, db, patchClearedAt.Add(time.Minute))
	assertRejected := func(name string, operation func() error) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := loadPostgresSitePresenceState(t, db, site.ID())
			err := operation()
			require.Error(t, err)
			require.True(t, shared.HasReason(err, shared.ErrorReasonValidation), err)
			after := loadPostgresSitePresenceState(t, db, site.ID())
			require.Equal(t, before, after, "rejected mutation changed durable Site state")
		})
	}
	assertRejected("POST missing required name", func() error {
		_, err := service.CreateSite(t.Context(), principal, applicationdcim.CreateSiteCommand{
			Slug: applicationdcim.FieldValue("rejected-site"),
		})
		return err
	})
	assertRejected("PUT explicit null name", func() error {
		_, err := service.ReplaceSite(t.Context(), principal, applicationdcim.ReplaceSiteCommand{
			ID: site.ID(), Name: applicationdcim.NullField[string](),
			Slug: applicationdcim.FieldValue("durable-site-renamed"),
		})
		return err
	})
	assertRejected("PATCH explicit null status", func() error {
		_, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
			ID: site.ID(), Status: applicationdcim.NullField[string](),
		})
		return err
	})
	assertRejected("PATCH explicit null optional text", func() error {
		_, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
			ID: site.ID(), Description: applicationdcim.NullField[string](),
		})
		return err
	})
	assertRejected("PATCH spaced status domain validation", func() error {
		_, err := service.UpdateSite(t.Context(), principal, applicationdcim.UpdateSiteCommand{
			ID: site.ID(), Status: applicationdcim.FieldValue(" active "),
		})
		return err
	})
}

type postgresSitePresenceState struct {
	row              dcimrow.SiteRow
	siteCount        int64
	changeCount      int64
	totalChangeCount int64
}

func requirePostgresSiteScalarRow(
	t *testing.T,
	actual dcimrow.SiteRow,
	expected dcimrow.SiteRow,
) {
	t.Helper()
	require.Equal(t, expected.Name, actual.Name)
	require.Equal(t, expected.Slug, actual.Slug)
	require.Equal(t, expected.Status, actual.Status)
	require.Equal(t, expected.Facility, actual.Facility)
	require.Equal(t, expected.Description, actual.Description)
	require.Equal(t, expected.Comments, actual.Comments)
}

func loadPostgresSitePresenceState(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) postgresSitePresenceState {
	t.Helper()
	var state postgresSitePresenceState
	require.NoError(t, db.Take(&state.row, "id = ?", id.Int64()).Error)
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&state.siteCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.SiteObjectType, id.Int64(),
	).Count(&state.changeCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func newPostgresSitePresenceService(
	t *testing.T,
	db *gorm.DB,
	now time.Time,
) *applicationdcim.SiteService {
	t.Helper()
	service, err := applicationdcim.NewSiteService(
		NewSiteRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		postgresConcurrencyClock{now: shared.NewTimestamp(now)},
	)
	require.NoError(t, err)
	return service
}

func newSiteConcurrencyPostgres(t *testing.T) (*gorm.DB, identity.Principal) {
	t.Helper()
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	base, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	schema := fmt.Sprintf("dcim_site_concurrency_%d", time.Now().UnixNano())
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)

	db, err := gorm.Open(
		postgres.Open(dcimPostgresDSNWithSearchPath(t, dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = base.Exec(`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`).Error
		_ = baseSQL.Close()
	})

	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	changeDependencies := []string{"go_identity_users"}
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
		changeDependencies = append(changeDependencies, descriptor.Name)
	}
	entries = append(entries, bootstrap.Entry{
		Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{},
		Dependencies: changeDependencies,
	})
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	_, err = bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "postgres-test", Email: "postgres-test@invalid.example",
		PasswordHash: "not-used-by-tests", IsStaff: true, IsSuperuser: true,
		IsActive: true, Permissions: []byte(`[]`), Created: now, Updated: now,
	}).Error)
	return db, identity.Principal{ID: 1, Username: "postgres-test", IsSuperuser: true}
}

func dcimPostgresDSNWithSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err == nil && (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	if strings.Contains(dsn, "=") {
		return strings.TrimSpace(dsn) + " search_path=" + schema
	}
	t.Fatalf("NETBOX_TEST_POSTGRES_DSN must be a postgres URL or keyword DSN")
	return ""
}

type postgresConcurrencyClock struct{ now shared.Timestamp }

func (clock postgresConcurrencyClock) Now() shared.Timestamp { return clock.now }
