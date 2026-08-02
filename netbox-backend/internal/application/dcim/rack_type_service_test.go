package dcim_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestCreateRackTypeAppliesProfileDefaultsAndRecordsTypedChange(t *testing.T) {
	service, repository, recorder := newRackTypeApplicationService(t, false)
	rackType, err := service.CreateRackType(t.Context(), testPrincipal(), appdcim.CreateRackTypeCommand{
		Manufacturer: appdcim.FieldValue(shared.ID(9)),
		Model:        appdcim.FieldValue(" R42 "), Slug: appdcim.FieldValue("r42"),
		FormFactor: appdcim.FieldValue("4-post-cabinet"),
		DescUnits:  appdcim.FieldValue(false),
	})
	require.NoError(t, err)
	assert.Equal(t, shared.ID(41), rackType.ID())
	assert.Equal(t, dcimdomain.RackWidth19, rackType.Width())
	assert.Equal(t, uint32(42), rackType.UHeight())
	assert.Equal(t, uint32(1), rackType.StartingUnit())
	assert.False(t, rackType.DescUnits())
	assert.Equal(t, "Acme", rackType.Manufacturer().Display())
	assert.Equal(t, 1, repository.createCalls)
	require.Len(t, recorder.changes, 1)
	assert.IsType(t, dcimdomain.RackTypeSnapshot{}, recorder.changes[0].After)
}

func TestCreateRackTypeReportsExactMissingManufacturerViolation(t *testing.T) {
	service, repository, _ := newRackTypeApplicationService(t, true)
	_, err := service.CreateRackType(t.Context(), testPrincipal(), appdcim.CreateRackTypeCommand{
		Manufacturer: appdcim.FieldValue(shared.ID(77)),
		Model:        appdcim.FieldValue("R42"), Slug: appdcim.FieldValue("r42"),
		FormFactor: appdcim.FieldValue("4-post-cabinet"),
	})
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "manufacturer", Reason: "invalid_choice",
		Description: `Invalid pk "77" - object does not exist.`,
	}}, shared.ViolationsOf(err))
	assert.Zero(t, repository.createCalls)
}

func TestListRackTypesPinsLimitsRepeatedFiltersAndDefaultOrdering(t *testing.T) {
	service, repository, _ := newRackTypeApplicationService(t, false)
	_, err := service.ListRackTypes(t.Context(), testPrincipal(), appdcim.ListRackTypesQuery{
		LimitPresent: true, IDs: []int64{-7, 0, 41},
		ManufacturerIDs: []int64{-1, 9}, ManufacturerSlugs: []string{" acme ", " vendor "},
		Models: []string{" R42 ", " R48 "}, Slugs: []string{" r42 ", " r48 "},
	})
	require.NoError(t, err)
	criteria := repository.criteria
	assert.Equal(t, appdcim.MaximumRackTypePageLimit, criteria.Limit)
	assert.Equal(t, []int64{-7, 0, 41}, criteria.IDs)
	assert.Equal(t, []int64{-1, 9}, criteria.ManufacturerIDs)
	assert.Equal(t, []string{"acme", "vendor"}, criteria.ManufacturerSlugs)
	assert.Equal(t, []string{"R42", "R48"}, criteria.Models)
	assert.Equal(t, []string{"r42", "r48"}, criteria.Slugs)
	assert.Equal(t, []appdcim.RackTypeSort{
		{Field: appdcim.RackTypeSortManufacturer},
		{Field: appdcim.RackTypeSortModel},
		{Field: appdcim.RackTypeSortID},
	}, criteria.Ordering)
}

func newRackTypeApplicationService(
	t *testing.T,
	missingManufacturer bool,
) (*appdcim.RackTypeService, *rackTypeApplicationRepository, *rackTypeApplicationRecorder) {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: "Acme", Slug: "acme", Description: "Vendor",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, manufacturer.AssignID(9))
	repository := &rackTypeApplicationRepository{}
	recorder := &rackTypeApplicationRecorder{}
	service, err := appdcim.NewRackTypeService(
		repository,
		&rackTypeManufacturerReader{manufacturer: manufacturer, missing: missingManufacturer},
		rackTypeApplicationUnitOfWork{}, recorder, authz.AllowAll{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, repository, recorder
}

type rackTypeManufacturerReader struct {
	manufacturer *dcimdomain.Manufacturer
	missing      bool
}

func (reader *rackTypeManufacturerReader) Get(_ context.Context, id shared.ID) (*dcimdomain.Manufacturer, error) {
	if reader.missing || reader.manufacturer == nil || reader.manufacturer.ID() != id {
		return nil, shared.NotFound("Manufacturer", id)
	}
	return reader.manufacturer, nil
}

type rackTypeApplicationRepository struct {
	criteria    appdcim.RackTypeListCriteria
	createCalls int
}

func (repository *rackTypeApplicationRepository) List(_ context.Context, criteria appdcim.RackTypeListCriteria) (appdcim.RackTypePage, error) {
	repository.criteria = criteria
	return appdcim.RackTypePage{}, nil
}
func (*rackTypeApplicationRepository) Get(context.Context, shared.ID) (*dcimdomain.RackType, error) {
	return nil, shared.NotFound("RackType", 0)
}
func (*rackTypeApplicationRepository) GetForUpdate(context.Context, shared.ID) (*dcimdomain.RackType, error) {
	return nil, shared.NotFound("RackType", 0)
}
func (repository *rackTypeApplicationRepository) Create(_ context.Context, rackType *dcimdomain.RackType) error {
	repository.createCalls++
	return rackType.AssignID(41)
}
func (*rackTypeApplicationRepository) Update(context.Context, *dcimdomain.RackType) error { return nil }
func (*rackTypeApplicationRepository) Delete(context.Context, *dcimdomain.RackType) error { return nil }
func (*rackTypeApplicationRepository) PropagateToRacks(context.Context, shared.ID, dcimdomain.RackPhysicalAttributes, shared.Timestamp) ([]appdcim.RackPropagationChange, error) {
	return nil, nil
}

type rackTypeApplicationUnitOfWork struct{}

func (rackTypeApplicationUnitOfWork) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

type rackTypeApplicationRecorder struct{ changes []changelog.Change }

func (recorder *rackTypeApplicationRecorder) Record(_ context.Context, change changelog.Change) error {
	recorder.changes = append(recorder.changes, change)
	return nil
}
