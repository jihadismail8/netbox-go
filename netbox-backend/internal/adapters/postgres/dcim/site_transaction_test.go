package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	applicationchangelog "netbox-go/internal/application/changelog"
	domaindcim "netbox-go/internal/domain/dcim"
)

func TestSiteMutationAndTypedChangeRollbackTogether(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository := NewSiteRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	site := newSiteFixture(t, "Rollback", "rollback", "M9")
	sentinel := errors.New("fail after both writes")

	err := unitOfWork.WithinTransaction(t.Context(), func(transactionContext context.Context) error {
		require.NoError(t, repository.Create(transactionContext, site))
		change, changeErr := applicationchangelog.NewChange(
			17,
			domaindcim.SiteObjectType,
			site.ID(),
			site.Display(),
			applicationchangelog.ActionCreate,
			nil,
			site.Snapshot(),
			repositoryCreatedAt,
		)
		require.NoError(t, changeErr)
		require.NoError(t, recorder.Record(transactionContext, change))
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	var sites, changes int64
	require.NoError(t, db.Model(&dcimrow.SiteRow{}).Count(&sites).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changes).Error)
	assert.Zero(t, sites)
	assert.Zero(t, changes)
}

func TestGetForUpdateAndUpdateShareTheBoundTransaction(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	repository := NewSiteRepository(db)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	site := newSiteFixture(t, "Locked", "locked", "M9")
	require.NoError(t, repository.Create(t.Context(), site))

	require.NoError(t, unitOfWork.WithinTransaction(
		t.Context(),
		func(transactionContext context.Context) error {
			loaded, err := repository.GetForUpdate(transactionContext, site.ID())
			if err != nil {
				return err
			}
			if err := loaded.ApplyPatch(domaindcim.SitePatch{
				Description: stringPointer("locked update"),
			}, repositoryUpdatedAt); err != nil {
				return err
			}
			return repository.Update(transactionContext, loaded)
		},
	))

	updated, err := repository.Get(t.Context(), site.ID())
	require.NoError(t, err)
	assert.Equal(t, "locked update", updated.Description())
	assert.Equal(t, repositoryUpdatedAt, updated.LastUpdated())
}

func TestChangeRecorderSerializesTypedSiteSnapshotOnlyAtAuditBoundary(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository := NewSiteRepository(db)
	recorder := postgreschangelog.NewRecorder(db)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	site := newSiteFixture(t, "Audited", "audited", "M9")

	require.NoError(t, unitOfWork.WithinTransaction(
		t.Context(),
		func(transactionContext context.Context) error {
			if err := repository.Create(transactionContext, site); err != nil {
				return err
			}
			change, err := applicationchangelog.NewChange(
				17,
				domaindcim.SiteObjectType,
				site.ID(),
				site.Display(),
				applicationchangelog.ActionCreate,
				nil,
				site.Snapshot(),
				repositoryCreatedAt,
			)
			if err != nil {
				return err
			}
			_, typedBeforePersistence := change.After.(domaindcim.SiteSnapshot)
			assert.True(t, typedBeforePersistence)
			return recorder.Record(transactionContext, change)
		},
	))

	var row postgreschangelog.ChangeRow
	require.NoError(t, db.Take(&row).Error)
	assert.Equal(t, int64(17), row.ActorID)
	assert.Equal(t, domaindcim.SiteObjectType, row.Kind)
	assert.Equal(t, string(applicationchangelog.ActionCreate), row.Action)
	assert.Equal(t, site.ID().Int64(), row.ObjectID)
	assert.Empty(t, row.BeforeData)

	var after map[string]any
	require.NoError(t, json.Unmarshal(row.AfterData, &after))
	assert.Equal(t, map[string]any{
		"name":        "Audited",
		"slug":        "audited",
		"status":      "active",
		"facility":    "M9",
		"description": "Site description",
		"comments":    "Site comments",
	}, after)
}

func seedChangeActor(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Create(&identityrows.UserRow{
		ID:           17,
		Username:     "site-auditor",
		PasswordHash: "unused-in-adapter-test",
		IsActive:     true,
		Permissions:  datatypes.JSON([]byte("[]")),
		Created:      repositoryCreatedAt.Time,
		Updated:      repositoryCreatedAt.Time,
	}).Error)
}
