package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestManufacturerRepositoryMapsExistingRowCounterAndMutations(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewManufacturerRepository(db)
	manufacturer := newManufacturerFixture(t, "Acme", "acme")
	require.NoError(t, repository.Create(t.Context(), manufacturer))
	require.NoError(t, db.Create(&dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		ManufacturerID: manufacturer.ID().Int64(), Model: "Router 1", Slug: "router-1",
		UHeight: 1, IsFullDepth: true,
	}).Error)

	loaded, err := repository.Get(t.Context(), manufacturer.ID())
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, "Acme", loaded.Name())
	assert.Equal(t, "acme", loaded.Slug().String())
	assert.Equal(t, "Manufacturer description", loaded.Description())
	assert.Equal(t, uint64(1), loaded.DeviceTypeCount())
	assert.Equal(t, repositoryCreatedAt, loaded.Created())

	empty := ""
	require.NoError(t, loaded.ApplyPatch(
		domaindcim.ManufacturerPatch{Description: &empty}, repositoryUpdatedAt,
	))
	require.NoError(t, repository.Update(t.Context(), loaded))
	updated, err := repository.Get(t.Context(), manufacturer.ID())
	require.NoError(t, err)
	assert.Empty(t, updated.Description())
	assert.Equal(t, repositoryUpdatedAt, updated.LastUpdated())
	assert.Equal(t, uint64(1), updated.DeviceTypeCount())

	_, err = repository.Get(t.Context(), 9999)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonNotFound))
}

func TestRackRoleRepositoryMapsExistingRowCounterAndMutations(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewRackRoleRepository(db)
	role := newRackRoleFixture(t, "Core", "core", "123abc")
	require.NoError(t, repository.Create(t.Context(), role))
	siteID := seedOrganizationSite(t, db)
	require.NoError(t, db.Create(&dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		SiteID: siteID, Name: "R1", RoleID: idPointer(role.ID().Int64()), Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}).Error)

	loaded, err := repository.Get(t.Context(), role.ID())
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, "123abc", loaded.Color().String())
	assert.Equal(t, uint64(1), loaded.RackCount())

	color := "00ff00"
	require.NoError(t, loaded.ApplyPatch(domaindcim.RackRolePatch{Color: &color}, repositoryUpdatedAt))
	require.NoError(t, repository.Update(t.Context(), loaded))
	updated, err := repository.Get(t.Context(), role.ID())
	require.NoError(t, err)
	assert.Equal(t, "00ff00", updated.Color().String())
	assert.Equal(t, repositoryUpdatedAt, updated.LastUpdated())
}

func TestOrganizationRepositoriesTranslateNameAndSlugUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	rackRoleRepository := NewRackRoleRepository(db)
	require.NoError(t, manufacturerRepository.Create(
		t.Context(), newManufacturerFixture(t, "Acme", "acme"),
	))
	require.NoError(t, rackRoleRepository.Create(
		t.Context(), newRackRoleFixture(t, "Core", "core", "123abc"),
	))

	manufacturerDuplicates := []struct {
		manufacturer *domaindcim.Manufacturer
		field        string
	}{
		{manufacturer: newManufacturerFixture(t, "Acme", "acme-two"), field: "name"},
		{manufacturer: newManufacturerFixture(t, "Acme Two", "acme"), field: "slug"},
	}
	for _, test := range manufacturerDuplicates {
		err := manufacturerRepository.Create(t.Context(), test.manufacturer)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, []shared.FieldViolation{{
			Field: test.field, Reason: "unique",
			Description: "manufacturer with this " + test.field + " already exists.",
		}}, shared.ViolationsOf(err))
		assert.False(t, test.manufacturer.ID().IsValid())
	}

	rackRoleDuplicates := []struct {
		role  *domaindcim.RackRole
		field string
	}{
		{role: newRackRoleFixture(t, "Core", "core-two", "123abc"), field: "name"},
		{role: newRackRoleFixture(t, "Core Two", "core", "123abc"), field: "slug"},
	}
	for _, test := range rackRoleDuplicates {
		err := rackRoleRepository.Create(t.Context(), test.role)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, []shared.FieldViolation{{
			Field: test.field, Reason: "unique",
			Description: "rack role with this " + test.field + " already exists.",
		}}, shared.ViolationsOf(err))
		assert.False(t, test.role.ID().IsValid())
	}
}

func TestOrganizationRepositoriesTranslateProtectedDeletes(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	rackRoleRepository := NewRackRoleRepository(db)
	manufacturer := newManufacturerFixture(t, "Protected Vendor", "protected-vendor")
	role := newRackRoleFixture(t, "Protected Role", "protected-role", "123abc")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	require.NoError(t, rackRoleRepository.Create(t.Context(), role))
	require.NoError(t, db.Create(&dcimrow.DeviceTypeRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		ManufacturerID: manufacturer.ID().Int64(), Model: "Protected", Slug: "protected",
		UHeight: 1, IsFullDepth: true,
	}).Error)
	siteID := seedOrganizationSite(t, db)
	require.NoError(t, db.Create(&dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		SiteID: siteID, Name: "Protected Rack", RoleID: idPointer(role.ID().Int64()), Status: "active",
		Width: 19, UHeight: 42, StartingUnit: 1,
	}).Error)

	for _, err := range []error{
		manufacturerRepository.Delete(t.Context(), manufacturer),
		rackRoleRepository.Delete(t.Context(), role),
	} {
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	}
	_, manufacturerErr := manufacturerRepository.Get(t.Context(), manufacturer.ID())
	_, roleErr := rackRoleRepository.Get(t.Context(), role.ID())
	require.NoError(t, manufacturerErr)
	require.NoError(t, roleErr)
}

func TestOrganizationRepositoriesPushSearchVisibilityCountOrderingAndPageIntoSQL(t *testing.T) {
	capture := &queryCapture{}
	db, capture := newSiteTestDatabase(t, capture)
	now := repositoryCreatedAt.Time
	manufacturers := []dcimrow.ManufacturerRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Alpha", Slug: "alpha", Description: "primary"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Zulu", Slug: "zulu", Description: "alpha staging"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Beta", Slug: "beta"},
	}
	require.NoError(t, db.Create(&manufacturers).Error)
	roles := []dcimrow.RackRoleRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "First", Slug: "first", Color: "123abc"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Second", Slug: "second", Color: "123abc"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Third", Slug: "third", Color: "123abc"},
	}
	require.NoError(t, db.Create(&roles).Error)
	capture.Reset()

	manufacturerPage, err := NewManufacturerRepository(db).List(
		t.Context(),
		applicationdcim.ManufacturerListCriteria{
			Limit: 1, Offset: 1, Query: "ALPHA",
			Ordering: []applicationdcim.ManufacturerSort{{Field: applicationdcim.ManufacturerSortName}},
		},
	)
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, uint64(2), manufacturerPage.Count)
	require.Len(t, manufacturerPage.Results, 1)
	assert.Equal(t, "Zulu", manufacturerPage.Results[0].Name())

	rackRolePage, err := NewRackRoleRepository(db).List(
		t.Context(),
		applicationdcim.RackRoleListCriteria{
			Limit: 1, Ordering: []applicationdcim.RackRoleSort{{Field: applicationdcim.RackRoleSortID}},
			VisibleObjectIDs:      []shared.ID{shared.ID(roles[1].ID), shared.ID(roles[2].ID)},
			VisibilityConstrained: true,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), rackRolePage.Count)
	require.Len(t, rackRolePage.Results, 1)
	assert.Equal(t, shared.ID(roles[1].ID), rackRolePage.Results[0].ID())

	empty, err := NewRackRoleRepository(db).List(t.Context(), applicationdcim.RackRoleListCriteria{
		Limit: 50, Ordering: []applicationdcim.RackRoleSort{{Field: applicationdcim.RackRoleSortID}},
		VisibilityConstrained: true,
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)

	statements := strings.ToLower(strings.Join(capture.Statements(), "\n"))
	assert.Contains(t, statements, "select count(*)")
	assert.Contains(t, statements, "lower(manufacturers.name)")
	assert.Contains(t, statements, "order by")
	assert.Contains(t, statements, "limit 1 offset 1")
	assert.Contains(t, statements, "rack_roles.id in")
}

func TestOrganizationRepositoriesUseORWithinRepeatedFiltersAndAcceptSignedIDs(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	now := repositoryCreatedAt.Time
	manufacturers := []dcimrow.ManufacturerRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Alpha", Slug: "alpha"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Beta", Slug: "beta"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Gamma", Slug: "gamma"},
	}
	require.NoError(t, db.Create(&manufacturers).Error)

	page, err := NewManufacturerRepository(db).List(t.Context(), applicationdcim.ManufacturerListCriteria{
		Limit: 50, IDs: []int64{-1, manufacturers[0].ID, manufacturers[1].ID},
		Names: []string{"Alpha", "Beta"}, Slugs: []string{"alpha", "beta"},
		Ordering: []applicationdcim.ManufacturerSort{{Field: applicationdcim.ManufacturerSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, "Alpha", page.Results[0].Name())
	assert.Equal(t, "Beta", page.Results[1].Name())

	empty, err := NewManufacturerRepository(db).List(t.Context(), applicationdcim.ManufacturerListCriteria{
		Limit: 50, IDs: []int64{-1, 0},
		Ordering: []applicationdcim.ManufacturerSort{{Field: applicationdcim.ManufacturerSortID}},
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)
}

func TestManufacturerRepositorySearchTreatsSQLWildcardsLiterally(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	now := repositoryCreatedAt.Time
	require.NoError(t, db.Create(&[]dcimrow.ManufacturerRow{
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Percent % Vendor", Slug: "percent"},
		{RowMetadata: dcimrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Ordinary Vendor", Slug: "ordinary"},
	}).Error)
	page, err := NewManufacturerRepository(db).List(t.Context(), applicationdcim.ManufacturerListCriteria{
		Limit: 50, Query: "%", Ordering: []applicationdcim.ManufacturerSort{{Field: applicationdcim.ManufacturerSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, "Percent % Vendor", page.Results[0].Name())
}

func TestOrganizationMutationAndTypedChangesShareOneTransaction(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	rackRoleRepository := NewRackRoleRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	sentinel := errors.New("force rollback")

	err := unitOfWork.WithinTransaction(t.Context(), func(transactionContext context.Context) error {
		manufacturer := newManufacturerFixture(t, "Rollback Vendor", "rollback-vendor")
		role := newRackRoleFixture(t, "Rollback Role", "rollback-role", "123abc")
		require.NoError(t, manufacturerRepository.Create(transactionContext, manufacturer))
		require.NoError(t, rackRoleRepository.Create(transactionContext, role))
		require.NoError(t, recordManufacturerCreate(transactionContext, recorder, manufacturer))
		require.NoError(t, recordRackRoleCreate(transactionContext, recorder, role))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)
	assertTableCount(t, db, &dcimrow.ManufacturerRow{}, 0)
	assertTableCount(t, db, &dcimrow.RackRoleRow{}, 0)
	assertTableCount(t, db, &postgreschangelog.ChangeRow{}, 0)

	require.NoError(t, unitOfWork.WithinTransaction(
		t.Context(),
		func(transactionContext context.Context) error {
			manufacturer := newManufacturerFixture(t, "Audited Vendor", "audited-vendor")
			role := newRackRoleFixture(t, "Audited Role", "audited-role", "00ff00")
			if err := manufacturerRepository.Create(transactionContext, manufacturer); err != nil {
				return err
			}
			if err := rackRoleRepository.Create(transactionContext, role); err != nil {
				return err
			}
			if err := recordManufacturerCreate(transactionContext, recorder, manufacturer); err != nil {
				return err
			}
			return recordRackRoleCreate(transactionContext, recorder, role)
		},
	))

	var rows []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&rows).Error)
	require.Len(t, rows, 2)
	assert.Equal(t, domaindcim.ManufacturerObjectType, rows[0].Kind)
	assert.Equal(t, domaindcim.RackRoleObjectType, rows[1].Kind)
	var manufacturerJSON manufacturerAuditJSON
	var rackRoleJSON rackRoleAuditJSON
	require.NoError(t, json.Unmarshal(rows[0].AfterData, &manufacturerJSON))
	require.NoError(t, json.Unmarshal(rows[1].AfterData, &rackRoleJSON))
	assert.Equal(t, manufacturerAuditJSON{
		Name: "Audited Vendor", Slug: "audited-vendor", Description: "Manufacturer description",
	}, manufacturerJSON)
	assert.Equal(t, rackRoleAuditJSON{
		Name: "Audited Role", Slug: "audited-role", Color: "00ff00", Description: "Rack role description",
	}, rackRoleJSON)
}

type manufacturerAuditJSON struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type rackRoleAuditJSON struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

func newManufacturerFixture(t *testing.T, name, slug string) *domaindcim.Manufacturer {
	t.Helper()
	manufacturer, err := domaindcim.NewManufacturer(domaindcim.ManufacturerValues{
		Name: name, Slug: slug, Description: "Manufacturer description",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return manufacturer
}

func newRackRoleFixture(t *testing.T, name, slug, color string) *domaindcim.RackRole {
	t.Helper()
	role, err := domaindcim.NewRackRole(domaindcim.RackRoleValues{
		Name: name, Slug: slug, Color: color, Description: "Rack role description",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return role
}

func seedOrganizationSite(t *testing.T, db interface{ Create(any) *gorm.DB }) int64 {
	t.Helper()
	row := dcimrow.SiteRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Name: "Organization Test Site", Slug: "organization-test-site", Status: "active",
	}
	require.NoError(t, db.Create(&row).Error)
	return row.ID
}

func idPointer(value int64) *int64 { return &value }

func recordManufacturerCreate(
	ctx context.Context,
	recorder *postgreschangelog.Recorder,
	manufacturer *domaindcim.Manufacturer,
) error {
	change, err := applicationchangelog.NewChange(
		17, domaindcim.ManufacturerObjectType, manufacturer.ID(), manufacturer.Display(),
		applicationchangelog.ActionCreate, nil, manufacturer.Snapshot(), repositoryCreatedAt,
	)
	if err != nil {
		return err
	}
	return recorder.Record(ctx, change)
}

func recordRackRoleCreate(
	ctx context.Context,
	recorder *postgreschangelog.Recorder,
	role *domaindcim.RackRole,
) error {
	change, err := applicationchangelog.NewChange(
		17, domaindcim.RackRoleObjectType, role.ID(), role.Display(),
		applicationchangelog.ActionCreate, nil, role.Snapshot(), repositoryCreatedAt,
	)
	if err != nil {
		return err
	}
	return recorder.Record(ctx, change)
}

func assertTableCount(t *testing.T, db *gorm.DB, model any, want int64) {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(model).Count(&count).Error)
	assert.Equal(t, want, count)
}
