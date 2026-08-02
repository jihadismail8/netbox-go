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

func TestCreateRackAppliesDefaultsAndRackTypeOwnershipBeforePersistence(t *testing.T) {
	service, repository, recorder := newRackApplicationService(t)

	rack, err := service.CreateRack(t.Context(), testPrincipal(), appdcim.CreateRackCommand{
		Site:     appdcim.FieldValue(shared.ID(3)),
		Name:     appdcim.FieldValue(" A01 "),
		RackType: appdcim.FieldValue(shared.ID(8)),
		Width:    appdcim.FieldValue(uint32(19)),
		UHeight:  appdcim.FieldValue(uint32(48)),
	})
	require.NoError(t, err)

	assert.Equal(t, shared.ID(41), rack.ID())
	assert.Equal(t, dcimdomain.RackStatusActive, rack.Status())
	assert.Equal(t, dcimdomain.RackWidth23, rack.Width())
	assert.Equal(t, uint32(24), rack.UHeight())
	assert.Equal(t, uint32(3), rack.StartingUnit())
	assert.True(t, rack.DescUnits())
	factor, present := rack.FormFactor().Get()
	require.True(t, present)
	assert.Equal(t, dcimdomain.RackFormFactorWallFrame, factor)
	assert.Equal(t, 1, repository.createCalls)
	require.Len(t, recorder.changes, 1)
	assert.IsType(t, dcimdomain.RackSnapshot{}, recorder.changes[0].After)
}

func TestUpdateRackProtectsMountedDeviceCapacityAndNumbering(t *testing.T) {
	service, repository, recorder := newRackApplicationService(t)
	repository.rack = newApplicationRack(t, false)
	repository.placements = []appdcim.RackDevicePlacement{{
		ID: 71, PositionHalfUnits: 82, HeightHalfUnits: 4,
	}}

	_, err := service.UpdateRack(t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
		ID: repository.rack.ID(), UHeight: appdcim.FieldValue(uint32(40)),
	})
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "u_height", Reason: "invalid",
		Description: "Rack must be at least 42U tall to house currently installed devices.",
	}}, shared.ViolationsOf(err))
	assert.Zero(t, repository.updateCalls)
	assert.Empty(t, recorder.changes)

	repository.placements = []appdcim.RackDevicePlacement{{
		ID: 72, PositionHalfUnits: 2, HeightHalfUnits: 2,
	}}
	_, err = service.UpdateRack(t.Context(), testPrincipal(), appdcim.UpdateRackCommand{
		ID: repository.rack.ID(), StartingUnit: appdcim.FieldValue(uint32(2)),
	})
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "starting_unit", Reason: "invalid",
		Description: "Rack unit numbering must begin at 1 or less to house currently installed devices.",
	}}, shared.ViolationsOf(err))
	assert.Zero(t, repository.updateCalls)
}

func TestListRacksPinsProfileFiltersOrderingAndLimitZero(t *testing.T) {
	service, repository, _ := newRackApplicationService(t)

	_, err := service.ListRacks(t.Context(), testPrincipal(), appdcim.ListRacksQuery{
		LimitPresent: true,
		IDs:          []int64{-1, 41},
		SiteIDs:      []int64{-1, 3}, SiteSlugs: []string{" moscow "},
		Names: []string{" A01 "}, Statuses: []string{"active"},
		RoleIDs: []int64{9}, RoleSlugs: []string{" production "},
		RackTypeIDs: []int64{8}, RackTypeSlugs: []string{" r24 "},
	})
	require.NoError(t, err)

	assert.Equal(t, appdcim.MaximumRackPageLimit, repository.criteria.Limit)
	assert.Equal(t, []int64{-1, 41}, repository.criteria.IDs)
	assert.Equal(t, []int64{-1, 3}, repository.criteria.SiteIDs)
	assert.Equal(t, []string{"moscow"}, repository.criteria.SiteSlugs)
	assert.Equal(t, []string{"A01"}, repository.criteria.Names)
	assert.Equal(t, []dcimdomain.RackStatus{dcimdomain.RackStatusActive}, repository.criteria.Statuses)
	assert.Equal(t, []int64{9}, repository.criteria.RoleIDs)
	assert.Equal(t, []string{"production"}, repository.criteria.RoleSlugs)
	assert.Equal(t, []int64{8}, repository.criteria.RackTypeIDs)
	assert.Equal(t, []string{"r24"}, repository.criteria.RackTypeSlugs)
	assert.Equal(t, []appdcim.RackSort{
		{Field: appdcim.RackSortSite},
		{Field: appdcim.RackSortName},
		{Field: appdcim.RackSortID},
	}, repository.criteria.Ordering)
}

func TestCreateRackReportsExactMissingRelationshipViolation(t *testing.T) {
	service, repository, _ := newRackApplicationService(t)
	_, err := service.CreateRack(t.Context(), testPrincipal(), appdcim.CreateRackCommand{
		Site: appdcim.FieldValue(shared.ID(404)), Name: appdcim.FieldValue("A01"),
	})
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field: "site", Reason: "invalid_choice",
		Description: `Invalid pk "404" - object does not exist.`,
	}}, shared.ViolationsOf(err))
	assert.Zero(t, repository.createCalls)
}

func newRackApplicationService(
	t *testing.T,
) (*appdcim.RackService, *rackApplicationRepository, *rackApplicationRecorder) {
	t.Helper()
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name: "Moscow", Slug: "moscow", Status: "active",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, site.AssignID(3))

	manufacturer, err := dcimdomain.NewManufacturerReference(5, "Acme", "acme")
	require.NoError(t, err)
	rackType, err := dcimdomain.NewRackType(dcimdomain.RackTypeValues{
		Manufacturer: manufacturer, Model: "R24", Slug: "r24",
		FormFactor: "wall-frame", Width: 23, UHeight: 24, StartingUnit: 3, DescUnits: true,
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, rackType.AssignID(8))

	role, err := dcimdomain.NewRackRole(dcimdomain.RackRoleValues{
		Name: "Production", Slug: "production", Color: "00ff00",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, role.AssignID(9))

	repository := &rackApplicationRepository{}
	recorder := &rackApplicationRecorder{}
	service, err := appdcim.NewRackService(
		repository,
		&rackSiteReader{site: site},
		&rackTypeReader{rackType: rackType},
		&rackRoleReader{role: role},
		rackApplicationUnitOfWork{},
		recorder,
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, repository, recorder
}

func newApplicationRack(t *testing.T, withRackType bool) *dcimdomain.Rack {
	t.Helper()
	site, err := dcimdomain.NewSiteReference(3, "Moscow", "moscow")
	require.NoError(t, err)
	values := dcimdomain.RackValues{
		Site: site, Name: "A01", Status: "active", Width: 19, UHeight: 42, StartingUnit: 1,
	}
	if withRackType {
		reference, referenceErr := dcimdomain.NewRackTypeReference(
			8, "R24", "r24", dcimdomain.RackPhysicalAttributes{
				FormFactor: dcimdomain.RackFormFactorWallFrame,
				Width:      dcimdomain.RackWidth23, UHeight: 24, StartingUnit: 3, DescUnits: true,
			},
		)
		require.NoError(t, referenceErr)
		values.RackType = dcimdomain.NonNullRackValue(reference)
	}
	rack, err := dcimdomain.NewRack(values, createdAt)
	require.NoError(t, err)
	require.NoError(t, rack.AssignID(41))
	return rack
}

type rackSiteReader struct{ site *dcimdomain.Site }

func (reader *rackSiteReader) Get(_ context.Context, id shared.ID) (*dcimdomain.Site, error) {
	if reader.site == nil || reader.site.ID() != id {
		return nil, shared.NotFound("Site", id)
	}
	return reader.site, nil
}

type rackTypeReader struct{ rackType *dcimdomain.RackType }

func (reader *rackTypeReader) Get(_ context.Context, id shared.ID) (*dcimdomain.RackType, error) {
	if reader.rackType == nil || reader.rackType.ID() != id {
		return nil, shared.NotFound("RackType", id)
	}
	return reader.rackType, nil
}

type rackRoleReader struct{ role *dcimdomain.RackRole }

func (reader *rackRoleReader) Get(_ context.Context, id shared.ID) (*dcimdomain.RackRole, error) {
	if reader.role == nil || reader.role.ID() != id {
		return nil, shared.NotFound("RackRole", id)
	}
	return reader.role, nil
}

type rackApplicationRepository struct {
	rack         *dcimdomain.Rack
	placements   []appdcim.RackDevicePlacement
	criteria     appdcim.RackListCriteria
	createCalls  int
	updateCalls  int
	propagations int
}

func (repository *rackApplicationRepository) List(
	_ context.Context,
	criteria appdcim.RackListCriteria,
) (appdcim.RackPage, error) {
	repository.criteria = criteria
	return appdcim.RackPage{}, nil
}

func (repository *rackApplicationRepository) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Rack, error) {
	return repository.GetForUpdate(context.Background(), id)
}

func (repository *rackApplicationRepository) GetForUpdate(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Rack, error) {
	if repository.rack == nil || repository.rack.ID() != id {
		return nil, shared.NotFound("Rack", id)
	}
	return repository.rack, nil
}

func (repository *rackApplicationRepository) Create(
	_ context.Context,
	rack *dcimdomain.Rack,
) error {
	repository.createCalls++
	repository.rack = rack
	return rack.AssignID(41)
}

func (repository *rackApplicationRepository) Update(
	_ context.Context,
	rack *dcimdomain.Rack,
) error {
	repository.updateCalls++
	repository.rack = rack
	return nil
}

func (*rackApplicationRepository) Delete(context.Context, *dcimdomain.Rack) error { return nil }

func (repository *rackApplicationRepository) MountedDevices(
	context.Context,
	shared.ID,
) ([]appdcim.RackDevicePlacement, error) {
	return append([]appdcim.RackDevicePlacement(nil), repository.placements...), nil
}

func (repository *rackApplicationRepository) PropagateSiteToDevices(
	context.Context,
	shared.ID,
	shared.ID,
	shared.Timestamp,
) ([]appdcim.RackSitePropagationChange, error) {
	repository.propagations++
	return nil, nil
}

type rackApplicationUnitOfWork struct{}

func (rackApplicationUnitOfWork) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	return operation(ctx)
}

type rackApplicationRecorder struct{ changes []changelog.Change }

func (recorder *rackApplicationRecorder) Record(
	_ context.Context,
	change changelog.Change,
) error {
	recorder.changes = append(recorder.changes, change)
	return nil
}
