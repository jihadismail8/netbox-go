package dcim

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestRackTypeRepositoryMapsManufacturerAndRepeatedFilters(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	rackTypeRepository := NewRackTypeRepository(db)
	alpha := newManufacturerFixture(t, "Alpha", "alpha")
	beta := newManufacturerFixture(t, "Beta", "beta")
	require.NoError(t, manufacturerRepository.Create(t.Context(), alpha))
	require.NoError(t, manufacturerRepository.Create(t.Context(), beta))

	fixtures := []*domaindcim.RackType{
		newRackTypeFixture(t, beta, "Zulu", "zulu", 42),
		newRackTypeFixture(t, alpha, "Zulu", "alpha-zulu", 45),
		newRackTypeFixture(t, alpha, "Alpha", "alpha-alpha", 48),
	}
	for _, rackType := range fixtures {
		require.NoError(t, rackTypeRepository.Create(t.Context(), rackType))
	}

	loaded, err := rackTypeRepository.Get(t.Context(), fixtures[1].ID())
	require.NoError(t, err)
	assert.Equal(t, alpha.ID(), loaded.Manufacturer().ID())
	assert.Equal(t, "Alpha", loaded.Manufacturer().Name())
	assert.Equal(t, "alpha", loaded.Manufacturer().Slug().String())
	assert.Equal(t, "Alpha Zulu", loaded.FullName())

	page, err := rackTypeRepository.List(t.Context(), applicationdcim.RackTypeListCriteria{
		Limit: 50, IDs: []int64{-1, fixtures[1].ID().Int64(), fixtures[2].ID().Int64()},
		ManufacturerIDs:   []int64{-1, alpha.ID().Int64()},
		ManufacturerSlugs: []string{"missing", "alpha"}, Models: []string{"Alpha", "Zulu"},
		Slugs: []string{"alpha-alpha", "alpha-zulu"},
		Ordering: []applicationdcim.RackTypeSort{
			{Field: applicationdcim.RackTypeSortManufacturer},
			{Field: applicationdcim.RackTypeSortModel},
			{Field: applicationdcim.RackTypeSortID},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, []string{"Alpha", "Zulu"}, []string{page.Results[0].Model(), page.Results[1].Model()})
}

func TestRackTypeRepositoryUsesStableDefaultOrderingAndSearch(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewRackTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	for _, rackType := range []*domaindcim.RackType{
		newRackTypeFixture(t, manufacturer, "B", "b", 42),
		newRackTypeFixture(t, manufacturer, "A", "a-one", 42),
		newRackTypeFixture(t, manufacturer, "A2", "a-two", 42),
	} {
		require.NoError(t, repository.Create(t.Context(), rackType))
	}

	criteria, err := applicationRackTypeDefaultCriteria()
	require.NoError(t, err)
	page, err := repository.List(t.Context(), criteria)
	require.NoError(t, err)
	require.Len(t, page.Results, 3)
	assert.Equal(t, []string{"A", "A2", "B"}, []string{
		page.Results[0].Model(), page.Results[1].Model(), page.Results[2].Model(),
	})

	search, err := repository.List(t.Context(), applicationdcim.RackTypeListCriteria{
		Limit: 50, Query: "TYPE COMMENT", Ordering: []applicationdcim.RackTypeSort{{Field: applicationdcim.RackTypeSortID}},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(3), search.Count)
}

func TestRackTypeRepositoryTranslatesExactUniquenessAndProtectedDelete(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewRackTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	original := newRackTypeFixture(t, manufacturer, "R42", "r42", 42)
	require.NoError(t, repository.Create(t.Context(), original))

	for _, test := range []struct {
		candidate *domaindcim.RackType
		want      shared.FieldViolation
	}{
		{
			candidate: newRackTypeFixture(t, manufacturer, "Other", "r42", 42),
			want:      shared.FieldViolation{Field: "slug", Reason: "unique", Description: "rack type with this slug already exists."},
		},
		{
			candidate: newRackTypeFixture(t, manufacturer, "R42", "r42-other", 42),
			want:      shared.FieldViolation{Field: "non_field_errors", Reason: "unique_together", Description: "Rack type with this Manufacturer and Model already exists."},
		},
	} {
		err := repository.Create(t.Context(), test.candidate)
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
		assert.Equal(t, []shared.FieldViolation{test.want}, shared.ViolationsOf(err))
		assert.False(t, test.candidate.ID().IsValid())
	}

	siteID := seedOrganizationSite(t, db)
	rack := seedRackForType(t, db, siteID, original.ID(), "Protected Rack")
	err := repository.Delete(t.Context(), original)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	assert.NotZero(t, rack.ID)
	_, getErr := repository.Get(t.Context(), original.ID())
	require.NoError(t, getErr)
}

func TestRackTypePropagationReturnsTypedSnapshotsAndCopiesOnlyInheritedFields(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewRackTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	rackType := newRackTypeFixture(t, manufacturer, "R42", "r42", 42)
	require.NoError(t, repository.Create(t.Context(), rackType))
	siteID := seedOrganizationSite(t, db)
	rack := seedRackForType(t, db, siteID, rackType.ID(), "R1")

	changes, err := repository.PropagateToRacks(t.Context(), rackType.ID(), domaindcim.RackPhysicalAttributes{
		FormFactor: domaindcim.RackFormFactorWallCabinet, Width: domaindcim.RackWidth23,
		UHeight: 48, StartingUnit: 5, DescUnits: true,
	}, repositoryUpdatedAt)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, shared.ID(rack.ID), changes[0].ID)
	assert.Equal(t, uint32(19), changes[0].Before.Width)
	assert.Equal(t, uint32(23), changes[0].After.Width)
	assert.Equal(t, "R1", changes[0].After.Name)
	assert.Equal(t, "rack serial", changes[0].After.Serial)

	var persisted dcimrow.RackRow
	require.NoError(t, db.First(&persisted, rack.ID).Error)
	require.NotNil(t, persisted.FormFactor)
	assert.Equal(t, "wall-cabinet", *persisted.FormFactor)
	assert.Equal(t, int64(23), persisted.Width)
	assert.Equal(t, int64(48), persisted.UHeight)
	assert.Equal(t, int64(5), persisted.StartingUnit)
	assert.True(t, persisted.DescUnits)
	assert.Equal(t, "rack serial", persisted.Serial, "non-inherited Rack fields must be preserved")
}

func TestRackTypeServiceRollsBackProfileRackPropagationAndAuditTogether(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	manufacturerRepository := NewManufacturerRepository(db)
	repository := NewRackTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	rackType := newRackTypeFixture(t, manufacturer, "R42", "r42", 42)
	require.NoError(t, repository.Create(t.Context(), rackType))
	siteID := seedOrganizationSite(t, db)
	rack := seedRackForType(t, db, siteID, rackType.ID(), "Rollback Rack")
	realRecorder := postgreschangelog.NewRecorder(db)
	sentinel := errors.New("fail on propagated Rack change")
	recorder := &failOnRecorder{delegate: realRecorder, failAt: 2, err: sentinel}
	service, err := applicationdcim.NewRackTypeService(
		repository, manufacturerRepository, postgresTransaction.NewUnitOfWork(db), recorder,
		authz.AllowAll{}, rackTypeClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)

	_, err = service.UpdateRackType(t.Context(), rackTypePrincipal(), applicationdcim.UpdateRackTypeCommand{
		ID: rackType.ID(), Width: applicationdcim.FieldValue(uint32(23)),
	})
	require.ErrorIs(t, err, sentinel)

	rolledBackType, getErr := repository.Get(t.Context(), rackType.ID())
	require.NoError(t, getErr)
	assert.Equal(t, domaindcim.RackWidth19, rolledBackType.Width())
	var rolledBackRack dcimrow.RackRow
	require.NoError(t, db.First(&rolledBackRack, rack.ID).Error)
	assert.Equal(t, int64(19), rolledBackRack.Width)
	assertTableCount(t, db, &postgreschangelog.ChangeRow{}, 0)

	success, err := applicationdcim.NewRackTypeService(
		repository, manufacturerRepository, postgresTransaction.NewUnitOfWork(db), realRecorder,
		authz.AllowAll{}, rackTypeClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	updated, err := success.UpdateRackType(t.Context(), rackTypePrincipal(), applicationdcim.UpdateRackTypeCommand{
		ID: rackType.ID(), Width: applicationdcim.FieldValue(uint32(23)),
	})
	require.NoError(t, err)
	assert.Equal(t, domaindcim.RackWidth23, updated.Width())

	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&changes).Error)
	require.Len(t, changes, 2)
	assert.Equal(t, domaindcim.RackTypeObjectType, changes[0].Kind)
	assert.Equal(t, domaindcim.RackObjectType, changes[1].Kind)
	var rackAfter map[string]any
	require.NoError(t, json.Unmarshal(changes[1].AfterData, &rackAfter))
	assert.Equal(t, float64(23), rackAfter["width"])
	assert.Equal(t, float64(rackType.ID().Int64()), rackAfter["rack_type"])
}

func newRackTypeFixture(
	t *testing.T,
	manufacturer *domaindcim.Manufacturer,
	model string,
	slug string,
	height uint32,
) *domaindcim.RackType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(
		manufacturer.ID(), manufacturer.Name(), manufacturer.Slug().String(),
	)
	require.NoError(t, err)
	rackType, err := domaindcim.NewRackType(domaindcim.RackTypeValues{
		Manufacturer: reference, Model: model, Slug: slug, FormFactor: "4-post-cabinet",
		Width: 19, UHeight: height, StartingUnit: 1,
		Description: "Type description", Comments: "Type comment",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return rackType
}

func seedRackForType(
	t *testing.T,
	db *gorm.DB,
	siteID int64,
	rackTypeID shared.ID,
	name string,
) dcimrow.RackRow {
	t.Helper()
	factor := "4-post-frame"
	id := rackTypeID.Int64()
	rack := dcimrow.RackRow{
		RowMetadata: dcimrow.RowMetadata{Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time},
		SiteID:      siteID, Name: name, RackTypeID: &id, Status: "active", Serial: "rack serial",
		FormFactor: &factor, Width: 19, UHeight: 40, StartingUnit: 2,
		Description: "rack description", Comments: "rack comments",
	}
	require.NoError(t, db.Create(&rack).Error)
	return rack
}

// This uses the public application validator so the persistence test exercises
// the exact default ordering contract without reaching into package internals.
func applicationRackTypeDefaultCriteria() (applicationdcim.RackTypeListCriteria, error) {
	serviceRepository := &criteriaCaptureRackTypeRepository{}
	manufacturer := &criteriaManufacturerReader{}
	service, err := applicationdcim.NewRackTypeService(
		serviceRepository, manufacturer, noOpUnitOfWork{}, noOpRecorder{}, authz.AllowAll{},
		rackTypeClock{now: repositoryCreatedAt},
	)
	if err != nil {
		return applicationdcim.RackTypeListCriteria{}, err
	}
	_, err = service.ListRackTypes(context.Background(), rackTypePrincipal(), applicationdcim.ListRackTypesQuery{})
	return serviceRepository.criteria, err
}

type criteriaCaptureRackTypeRepository struct {
	criteria applicationdcim.RackTypeListCriteria
}

func (repository *criteriaCaptureRackTypeRepository) List(_ context.Context, criteria applicationdcim.RackTypeListCriteria) (applicationdcim.RackTypePage, error) {
	repository.criteria = criteria
	return applicationdcim.RackTypePage{}, nil
}
func (*criteriaCaptureRackTypeRepository) Get(context.Context, shared.ID) (*domaindcim.RackType, error) {
	return nil, errors.New("unused")
}
func (*criteriaCaptureRackTypeRepository) GetForUpdate(context.Context, shared.ID) (*domaindcim.RackType, error) {
	return nil, errors.New("unused")
}
func (*criteriaCaptureRackTypeRepository) Create(context.Context, *domaindcim.RackType) error {
	return errors.New("unused")
}
func (*criteriaCaptureRackTypeRepository) Update(context.Context, *domaindcim.RackType) error {
	return errors.New("unused")
}
func (*criteriaCaptureRackTypeRepository) Delete(context.Context, *domaindcim.RackType) error {
	return errors.New("unused")
}
func (*criteriaCaptureRackTypeRepository) PropagateToRacks(context.Context, shared.ID, domaindcim.RackPhysicalAttributes, shared.Timestamp) ([]applicationdcim.RackPropagationChange, error) {
	return nil, errors.New("unused")
}

type criteriaManufacturerReader struct{}

func (*criteriaManufacturerReader) Get(context.Context, shared.ID) (*domaindcim.Manufacturer, error) {
	return nil, errors.New("unused")
}

type noOpUnitOfWork struct{}

func (noOpUnitOfWork) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

type noOpRecorder struct{}

func (noOpRecorder) Record(context.Context, applicationchangelog.Change) error { return nil }

type rackTypeClock struct{ now shared.Timestamp }

func (clock rackTypeClock) Now() shared.Timestamp { return clock.now }

type failOnRecorder struct {
	delegate applicationchangelog.Recorder
	failAt   int
	calls    int
	err      error
}

func (recorder *failOnRecorder) Record(ctx context.Context, change applicationchangelog.Change) error {
	recorder.calls++
	if recorder.calls == recorder.failAt {
		return recorder.err
	}
	return recorder.delegate.Record(ctx, change)
}

func rackTypePrincipal() identity.Principal {
	return identity.Principal{ID: 17, Username: "rack-type-tester", IsSuperuser: true}
}
