package dcim

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var (
	repositoryCreatedAt = shared.NewTimestamp(time.Date(2026, time.July, 18, 8, 30, 0, 0, time.UTC))
	repositoryUpdatedAt = shared.NewTimestamp(time.Date(2026, time.July, 18, 9, 45, 0, 0, time.UTC))
)

func TestSiteRepositoryMapsExistingSiteRowBothWays(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewSiteRepository(db)
	site := newSiteFixture(t, "Moscow", "moscow", "M9")

	require.NoError(t, repository.Create(t.Context(), site))
	require.True(t, site.ID().IsValid())
	require.NoError(t, db.Create(&dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			Created:     repositoryCreatedAt.Time,
			LastUpdated: repositoryCreatedAt.Time,
		},
		SiteID:       site.ID().Int64(),
		Name:         "A01",
		Status:       "active",
		Serial:       "",
		Width:        19,
		UHeight:      42,
		StartingUnit: 1,
		Description:  "",
		Comments:     "",
	}).Error)

	loaded, err := repository.Get(t.Context(), site.ID())
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, site.ID(), loaded.ID())
	assert.Equal(t, "Moscow", loaded.Name())
	assert.Equal(t, "moscow", loaded.Slug().String())
	assert.Equal(t, domaindcim.SiteStatusActive, loaded.Status())
	assert.Equal(t, "M9", loaded.Facility())
	assert.Equal(t, repositoryCreatedAt, loaded.Created())
	assert.Equal(t, repositoryCreatedAt, loaded.LastUpdated())
	assert.Zero(t, loaded.DeviceCount())
	assert.Zero(t, loaded.PrefixCount())
	assert.Equal(t, uint64(1), loaded.RackCount())

	require.NoError(t, loaded.ApplyPatch(domaindcim.SitePatch{
		Facility: stringPointer(""),
		Comments: stringPointer("updated comment"),
	}, repositoryUpdatedAt))
	require.NoError(t, repository.Update(t.Context(), loaded))

	updated, err := repository.Get(t.Context(), site.ID())
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Empty(t, updated.Facility(), "explicit blank fields must be persisted")
	assert.Equal(t, "updated comment", updated.Comments())
	assert.Equal(t, repositoryCreatedAt, updated.Created())
	assert.Equal(t, repositoryUpdatedAt, updated.LastUpdated())
	assert.Equal(t, uint64(1), updated.RackCount())

	_, err = repository.Get(t.Context(), shared.ID(9999))
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonNotFound))
}

func TestSiteRepositoryTranslatesProtectedDeleteAndPreservesSite(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewSiteRepository(db)
	site := newSiteFixture(t, "Protected", "protected", "M9")
	require.NoError(t, repository.Create(t.Context(), site))
	require.NoError(t, db.Create(&dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			Created:     repositoryCreatedAt.Time,
			LastUpdated: repositoryCreatedAt.Time,
		},
		SiteID: site.ID().Int64(), Name: "R1", Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}).Error)

	err := repository.Delete(t.Context(), site)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	loaded, getErr := repository.Get(t.Context(), site.ID())
	require.NoError(t, getErr)
	assert.Equal(t, site.ID(), loaded.ID())
}

func TestSiteRepositoryTranslatesNameAndSlugUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewSiteRepository(db)
	require.NoError(t, repository.Create(
		t.Context(),
		newSiteFixture(t, "Moscow", "moscow", "M9"),
	))

	for _, test := range []struct {
		field     string
		message   string
		duplicate *domaindcim.Site
	}{
		{
			field:     "name",
			message:   "site with this name already exists.",
			duplicate: newSiteFixture(t, "Moscow", "moscow-two", "M10"),
		},
		{
			field:     "slug",
			message:   "site with this slug already exists.",
			duplicate: newSiteFixture(t, "Moscow Two", "moscow", "M10"),
		},
	} {
		err := repository.Create(t.Context(), test.duplicate)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, test.message, err.Error())
		assert.Equal(t, []shared.FieldViolation{{
			Field:       test.field,
			Reason:      "unique",
			Description: test.message,
		}}, shared.ViolationsOf(err))
		assert.False(t, test.duplicate.ID().IsValid(), "failed inserts must not assign an ID")
	}

	var count int64
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestSiteMutationErrorUsesPostgresConstraintName(t *testing.T) {
	for _, test := range []struct {
		constraint string
		field      string
	}{
		{constraint: "uq_go_site_name", field: "name"},
		{constraint: "uq_go_site_slug", field: "slug"},
	} {
		t.Run(test.field, func(t *testing.T) {
			driverError := errors.New(
				`ERROR: duplicate key value violates unique constraint "` +
					test.constraint +
					`" (SQLSTATE 23505)`,
			)
			err := translateSiteMutationError("create Site", driverError)
			require.Error(t, err)
			assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
			assert.Equal(t, []shared.FieldViolation{{
				Field:       test.field,
				Reason:      "unique",
				Description: "site with this " + test.field + " already exists.",
			}}, shared.ViolationsOf(err))
			assert.ErrorIs(t, err, driverError)
		})
	}
}

func TestSiteRepositoryORsRepeatedFiltersAndAcceptsSignedIDs(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	now := repositoryCreatedAt.Time
	rows := []dcimrow.SiteRow{
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Alpha", Slug: "alpha", Status: "active",
		},
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Beta", Slug: "beta", Status: "planned",
		},
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Gamma", Slug: "gamma", Status: "retired",
		},
	}
	require.NoError(t, db.Create(&rows).Error)

	page, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:    50,
		IDs:      []int64{-1, rows[0].ID, rows[1].ID},
		Names:    []string{"Alpha", "Beta"},
		Slugs:    []string{"alpha", "beta"},
		Statuses: []domaindcim.SiteStatus{domaindcim.SiteStatusActive, domaindcim.SiteStatusPlanned},
		Ordering: []applicationdcim.SiteSort{{Field: applicationdcim.SiteSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, []string{"Alpha", "Beta"}, []string{
		page.Results[0].Name(),
		page.Results[1].Name(),
	})

	empty, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:    50,
		IDs:      []int64{-1},
		Ordering: []applicationdcim.SiteSort{{Field: applicationdcim.SiteSortID}},
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)
}

func TestSiteRepositoryPushesSearchFilterCountOrderingAndPageIntoSQL(t *testing.T) {
	capture := &queryCapture{}
	db, capture := newSiteTestDatabase(t, capture)
	now := repositoryCreatedAt.Time
	require.NoError(t, db.Create(&[]dcimrow.SiteRow{
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Alpha North", Slug: "alpha-north", Status: "active",
		},
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Zulu", Slug: "zulu", Status: "active", Description: "alpha staging area",
		},
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Alpha Retired", Slug: "alpha-retired", Status: "retired",
		},
		{
			RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now},
			Name:        "Beta", Slug: "beta", Status: "active",
		},
	}).Error)
	capture.Reset()

	page, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:  1,
		Offset: 1,
		Query:  "ALPHA",
		Ordering: []applicationdcim.SiteSort{
			{Field: applicationdcim.SiteSortName},
		},
		Statuses: []domaindcim.SiteStatus{domaindcim.SiteStatusActive},
	})
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, uint64(2), page.Count, "count must be computed before pagination")
	require.Len(t, page.Results, 1)
	assert.Equal(t, "Zulu", page.Results[0].Name())

	statements := strings.ToLower(strings.Join(capture.Statements(), "\n"))
	assert.Contains(t, statements, "select count(*)")
	assert.Contains(t, statements, "lower(sites.name)")
	assert.Contains(t, statements, "sites.status")
	assert.Contains(t, statements, "order by")
	assert.Contains(t, statements, "limit 1 offset 1")
}

func TestSiteRepositoryAppliesVisibilityBeforeCountAndPage(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	now := repositoryCreatedAt.Time
	rows := []dcimrow.SiteRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "First", Slug: "first", Status: "active"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Second", Slug: "second", Status: "active"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Third", Slug: "third", Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)

	page, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:                 1,
		Ordering:              []applicationdcim.SiteSort{{Field: applicationdcim.SiteSortID}},
		VisibleObjectIDs:      []shared.ID{shared.ID(rows[1].ID), shared.ID(rows[2].ID)},
		VisibilityConstrained: true,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(rows[1].ID), page.Results[0].ID())

	empty, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:                 50,
		Ordering:              []applicationdcim.SiteSort{{Field: applicationdcim.SiteSortID}},
		VisibilityConstrained: true,
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)
}

func TestSiteRepositorySearchTreatsSQLWildcardsLiterally(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	now := repositoryCreatedAt.Time
	require.NoError(t, db.Create(&[]dcimrow.SiteRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Percent % Site", Slug: "percent", Status: "active"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Ordinary Site", Slug: "ordinary", Status: "active"},
	}).Error)

	page, err := NewSiteRepository(db).List(t.Context(), applicationdcim.SiteListCriteria{
		Limit:    50,
		Query:    "%",
		Ordering: []applicationdcim.SiteSort{{Field: applicationdcim.SiteSortID}},
	})
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, uint64(1), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, "Percent % Site", page.Results[0].Name())
}

func newSiteFixture(t *testing.T, name, slug, facility string) *domaindcim.Site {
	t.Helper()
	site, err := domaindcim.NewSite(domaindcim.SiteValues{
		Name:        name,
		Slug:        slug,
		Status:      domaindcim.SiteStatusActive.String(),
		Facility:    facility,
		Description: "Site description",
		Comments:    "Site comments",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return site
}

func stringPointer(value string) *string { return &value }

func newSiteTestDatabase(t *testing.T, capture *queryCapture) (*gorm.DB, *queryCapture) {
	t.Helper()
	if capture == nil {
		capture = &queryCapture{}
	}
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", name)),
		&gorm.Config{Logger: capture},
	)
	require.NoError(t, err)
	models := append(identityrows.Models(), dcimrow.Models()...)
	models = append(models, ipamrow.Models()...)
	models = append(models, &postgreschangelog.ChangeRow{})
	require.NoError(t, db.AutoMigrate(models...))
	return db, capture
}

type queryCapture struct {
	mu         sync.Mutex
	statements []string
}

func (capture *queryCapture) LogMode(logger.LogLevel) logger.Interface { return capture }
func (capture *queryCapture) Info(context.Context, string, ...any)     {}
func (capture *queryCapture) Warn(context.Context, string, ...any)     {}
func (capture *queryCapture) Error(context.Context, string, ...any)    {}

func (capture *queryCapture) Trace(
	_ context.Context,
	_ time.Time,
	statement func() (string, int64),
	_ error,
) {
	sql, _ := statement()
	capture.mu.Lock()
	defer capture.mu.Unlock()
	capture.statements = append(capture.statements, sql)
}

func (capture *queryCapture) Reset() {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	capture.statements = nil
}

func (capture *queryCapture) Statements() []string {
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return append([]string(nil), capture.statements...)
}
