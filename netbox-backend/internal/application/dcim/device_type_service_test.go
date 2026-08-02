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

func TestCreateDeviceTypeAppliesDefaultsAndRecordsTypedChange(t *testing.T) {
	service, repository, recorder := newDeviceTypeApplicationService(t, nil)

	deviceType, err := service.CreateDeviceType(
		t.Context(), testPrincipal(), appdcim.CreateDeviceTypeCommand{
			Manufacturer: appdcim.FieldValue(shared.ID(9)),
			Model:        appdcim.FieldValue(" Router 9000 "),
			Slug:         appdcim.FieldValue("router-9000"),
		},
	)
	require.NoError(t, err)

	assert.Equal(t, shared.ID(41), deviceType.ID())
	assert.Equal(t, "1", deviceType.UHeight().String())
	assert.False(t, deviceType.ExcludeFromUtilization())
	assert.True(t, deviceType.IsFullDepth())
	assert.True(t, deviceType.Airflow().IsNull())
	assert.Equal(t, 1, repository.createCalls)
	require.Len(t, recorder.changes, 1)
	assert.Equal(t, changelog.ActionCreate, recorder.changes[0].Action)
	assert.IsType(t, dcimdomain.DeviceTypeSnapshot{}, recorder.changes[0].After)
}

func TestListDeviceTypesPinsRepeatedFiltersAndDefaultOrdering(t *testing.T) {
	service, repository, _ := newDeviceTypeApplicationService(t, nil)

	_, err := service.ListDeviceTypes(
		t.Context(), testPrincipal(), appdcim.ListDeviceTypesQuery{
			LimitPresent: true,
			IDs:          []int64{-1, 0, 41},
			ManufacturerIDs: []int64{
				-1, 9,
			},
			ManufacturerSlugs: []string{" acme ", " vendor "},
			Models:            []string{" Router ", " Switch "},
			Slugs:             []string{" router ", " switch "},
		},
	)
	require.NoError(t, err)

	criteria := repository.criteria
	assert.Equal(t, appdcim.MaximumDeviceTypePageLimit, criteria.Limit)
	assert.Equal(t, []int64{-1, 0, 41}, criteria.IDs)
	assert.Equal(t, []int64{-1, 9}, criteria.ManufacturerIDs)
	assert.Equal(t, []string{"acme", "vendor"}, criteria.ManufacturerSlugs)
	assert.Equal(t, []string{"Router", "Switch"}, criteria.Models)
	assert.Equal(t, []string{"router", "switch"}, criteria.Slugs)
	assert.Equal(t, []appdcim.DeviceTypeSort{
		{Field: appdcim.DeviceTypeSortManufacturer},
		{Field: appdcim.DeviceTypeSortModel},
		{Field: appdcim.DeviceTypeSortID},
	}, criteria.Ordering)
}

func TestDeviceTypeHeightTransitionProtection(t *testing.T) {
	for _, test := range []struct {
		name       string
		height     string
		fullDepth  *bool
		placements []appdcim.PositionedDevice
		want       string
	}{
		{
			name:   "positioned instance cannot become zero U",
			height: "0",
			placements: []appdcim.PositionedDevice{{
				ID: 51, DeviceTypeID: 41, RackID: 8, PositionHalfUnits: 2,
				Face: "front", StoredHeightHalfUnits: 2,
				RackStartingUnit: 1, RackUHeight: 42,
			}},
			want: "A DeviceType cannot become 0U while an instance is positioned.",
		},
		{
			name:   "increased height must fit rack",
			height: "2",
			placements: []appdcim.PositionedDevice{{
				ID: 51, DeviceTypeID: 41, RackID: 8, PositionHalfUnits: 83,
				Face: "front", StoredHeightHalfUnits: 2,
				RackStartingUnit: 1, RackUHeight: 42,
			}},
			want: "The new DeviceType height does not fit a positioned Device.",
		},
		{
			name:   "same face overlap is rejected",
			height: "2",
			placements: []appdcim.PositionedDevice{
				{
					ID: 51, DeviceTypeID: 41, RackID: 8, PositionHalfUnits: 2,
					Face: "front", StoredHeightHalfUnits: 2,
					RackStartingUnit: 1, RackUHeight: 42,
				},
				{
					ID: 52, DeviceTypeID: 42, RackID: 8, PositionHalfUnits: 4,
					Face: "front", StoredHeightHalfUnits: 2,
					RackStartingUnit: 1, RackUHeight: 42,
				},
			},
			want: "The new DeviceType height overlaps another positioned Device.",
		},
		{
			name:   "full depth overlaps opposite face",
			height: "2", fullDepth: boolPointer(true),
			placements: []appdcim.PositionedDevice{
				{
					ID: 51, DeviceTypeID: 41, RackID: 8, PositionHalfUnits: 2,
					Face: "front", StoredHeightHalfUnits: 2,
					RackStartingUnit: 1, RackUHeight: 42,
				},
				{
					ID: 52, DeviceTypeID: 42, RackID: 8, PositionHalfUnits: 4,
					Face: "rear", StoredHeightHalfUnits: 2,
					RackStartingUnit: 1, RackUHeight: 42,
				},
			},
			want: "The new DeviceType height overlaps another positioned Device.",
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			deviceType := newApplicationDeviceType(t, "1", false)
			service, repository, recorder := newDeviceTypeApplicationService(t, deviceType)
			repository.placements = test.placements
			command := appdcim.UpdateDeviceTypeCommand{
				ID:      deviceType.ID(),
				UHeight: appdcim.FieldValue(test.height),
			}
			if test.fullDepth != nil {
				command.IsFullDepth = appdcim.FieldValue(*test.fullDepth)
			}

			_, err := service.UpdateDeviceType(t.Context(), testPrincipal(), command)
			require.Error(t, err)
			assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
			assert.Contains(t, err.Error(), test.want)
			assert.Zero(t, repository.updateCalls)
			assert.Empty(t, recorder.changes)
		})
	}
}

func TestDeviceTypeHeightDecreaseSkipsPlacementLock(t *testing.T) {
	deviceType := newApplicationDeviceType(t, "2", false)
	service, repository, _ := newDeviceTypeApplicationService(t, deviceType)

	updated, err := service.UpdateDeviceType(
		t.Context(), testPrincipal(), appdcim.UpdateDeviceTypeCommand{
			ID: deviceType.ID(), UHeight: appdcim.FieldValue("1.5"),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "1.5", updated.UHeight().String())
	assert.Zero(t, repository.placementCalls)
	assert.Equal(t, 1, repository.updateCalls)
}

func TestDeleteDeviceTypeProtectsDeviceBeforeTemplateCascade(t *testing.T) {
	deviceType := newApplicationDeviceType(t, "1", true)
	service, repository, recorder := newDeviceTypeApplicationService(t, deviceType)
	repository.dependent = &appdcim.DeviceTypeDependent{
		ID: 77, Display: "edge-1 (ASSET-1)",
	}
	repository.templates = []appdcim.InterfaceTemplateDeletion{{
		ID: 91, Representation: "eth0",
		Snapshot: dcimdomain.InterfaceTemplateSnapshot{
			DeviceTypeID: deviceType.ID(), Name: "eth0", Type: "1000base-t",
		},
	}}

	err := service.DeleteDeviceType(
		t.Context(), testPrincipal(), appdcim.DeleteDeviceTypeCommand{ID: deviceType.ID()},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
	assert.Contains(
		t, err.Error(),
		"Unable to delete object. 1 dependent objects were found: edge-1 (ASSET-1) (77)",
	)
	assert.Zero(t, repository.templateListCalls)
	assert.Empty(t, repository.deletedTemplateIDs)
	assert.Zero(t, repository.deleteCalls)
	assert.Empty(t, recorder.changes)
}

func TestDeleteDeviceTypeCascadesTemplatesAndRecordsChildrenBeforeParent(t *testing.T) {
	deviceType := newApplicationDeviceType(t, "1", true)
	service, repository, recorder := newDeviceTypeApplicationService(t, deviceType)
	repository.templates = []appdcim.InterfaceTemplateDeletion{
		{
			ID: 91, Representation: "eth0",
			Snapshot: dcimdomain.InterfaceTemplateSnapshot{
				DeviceTypeID: deviceType.ID(), Name: "eth0", Type: "1000base-t",
				Enabled: true,
			},
		},
		{
			ID: 92, Representation: "eth1 (uplink)",
			Snapshot: dcimdomain.InterfaceTemplateSnapshot{
				DeviceTypeID: deviceType.ID(), Name: "eth1", Label: "uplink",
				Type: "10gbase-t", Enabled: true,
			},
		},
	}

	err := service.DeleteDeviceType(
		t.Context(), testPrincipal(), appdcim.DeleteDeviceTypeCommand{ID: deviceType.ID()},
	)
	require.NoError(t, err)

	assert.Equal(t, []shared.ID{91, 92}, repository.deletedTemplateIDs)
	assert.Equal(t, 1, repository.deleteCalls)
	require.Len(t, recorder.changes, 3)
	assert.Equal(t, []string{
		dcimdomain.InterfaceTemplateObjectType,
		dcimdomain.InterfaceTemplateObjectType,
		dcimdomain.DeviceTypeObjectType,
	}, []string{
		recorder.changes[0].ObjectType,
		recorder.changes[1].ObjectType,
		recorder.changes[2].ObjectType,
	})
	assert.Equal(t, []shared.ID{91, 92, 41}, []shared.ID{
		recorder.changes[0].ObjectID,
		recorder.changes[1].ObjectID,
		recorder.changes[2].ObjectID,
	})
	for _, change := range recorder.changes {
		assert.Equal(t, changelog.ActionDelete, change.Action)
		assert.NotNil(t, change.Before)
		assert.Nil(t, change.After)
	}
}

func boolPointer(value bool) *bool { return &value }

func newApplicationDeviceType(
	t *testing.T,
	height string,
	fullDepth bool,
) *dcimdomain.DeviceType {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	deviceType, err := dcimdomain.NewDeviceType(dcimdomain.DeviceTypeValues{
		Manufacturer: manufacturer, Model: "Router 9000", Slug: "router-9000",
		UHeight: height, IsFullDepth: fullDepth,
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, deviceType.AssignID(41))
	return deviceType
}

func newDeviceTypeApplicationService(
	t *testing.T,
	deviceType *dcimdomain.DeviceType,
) (*appdcim.DeviceTypeService, *deviceTypeApplicationRepository, *deviceTypeApplicationRecorder) {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: "Acme", Slug: "acme",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, manufacturer.AssignID(9))
	repository := &deviceTypeApplicationRepository{deviceType: deviceType}
	recorder := &deviceTypeApplicationRecorder{}
	service, err := appdcim.NewDeviceTypeService(
		repository,
		&deviceTypeManufacturerReader{manufacturer: manufacturer},
		deviceTypeApplicationUnitOfWork{},
		recorder,
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, repository, recorder
}

type deviceTypeManufacturerReader struct {
	manufacturer *dcimdomain.Manufacturer
}

func (reader *deviceTypeManufacturerReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Manufacturer, error) {
	if reader.manufacturer == nil || reader.manufacturer.ID() != id {
		return nil, shared.NotFound("Manufacturer", id)
	}
	return reader.manufacturer, nil
}

type deviceTypeApplicationRepository struct {
	deviceType *dcimdomain.DeviceType
	criteria   appdcim.DeviceTypeListCriteria
	placements []appdcim.PositionedDevice
	dependent  *appdcim.DeviceTypeDependent
	templates  []appdcim.InterfaceTemplateDeletion

	createCalls        int
	updateCalls        int
	placementCalls     int
	templateListCalls  int
	deletedTemplateIDs []shared.ID
	deleteCalls        int
}

func (repository *deviceTypeApplicationRepository) List(
	_ context.Context,
	criteria appdcim.DeviceTypeListCriteria,
) (appdcim.DeviceTypePage, error) {
	repository.criteria = criteria
	return appdcim.DeviceTypePage{}, nil
}

func (repository *deviceTypeApplicationRepository) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	if repository.deviceType == nil || repository.deviceType.ID() != id {
		return nil, shared.NotFound("DeviceType", id)
	}
	return repository.deviceType, nil
}

func (repository *deviceTypeApplicationRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	return repository.Get(ctx, id)
}

func (repository *deviceTypeApplicationRepository) Create(
	_ context.Context,
	deviceType *dcimdomain.DeviceType,
) error {
	repository.createCalls++
	if err := deviceType.AssignID(41); err != nil {
		return err
	}
	repository.deviceType = deviceType
	return nil
}

func (repository *deviceTypeApplicationRepository) Update(
	_ context.Context,
	deviceType *dcimdomain.DeviceType,
) error {
	repository.updateCalls++
	repository.deviceType = deviceType
	return nil
}

func (repository *deviceTypeApplicationRepository) ListPositionedDevicesForUpdate(
	context.Context,
) ([]appdcim.PositionedDevice, error) {
	repository.placementCalls++
	return append([]appdcim.PositionedDevice(nil), repository.placements...), nil
}

func (repository *deviceTypeApplicationRepository) FindDeviceUsingDeviceType(
	context.Context,
	shared.ID,
) (*appdcim.DeviceTypeDependent, error) {
	return repository.dependent, nil
}

func (repository *deviceTypeApplicationRepository) ListInterfaceTemplatesForUpdate(
	context.Context,
	shared.ID,
) ([]appdcim.InterfaceTemplateDeletion, error) {
	repository.templateListCalls++
	return append([]appdcim.InterfaceTemplateDeletion(nil), repository.templates...), nil
}

func (repository *deviceTypeApplicationRepository) DeleteInterfaceTemplate(
	_ context.Context,
	id shared.ID,
) error {
	repository.deletedTemplateIDs = append(repository.deletedTemplateIDs, id)
	return nil
}

func (repository *deviceTypeApplicationRepository) Delete(
	context.Context,
	*dcimdomain.DeviceType,
) error {
	repository.deleteCalls++
	return nil
}

type deviceTypeApplicationUnitOfWork struct{}

func (deviceTypeApplicationUnitOfWork) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	return operation(ctx)
}

type deviceTypeApplicationRecorder struct{ changes []changelog.Change }

func (recorder *deviceTypeApplicationRecorder) Record(
	_ context.Context,
	change changelog.Change,
) error {
	recorder.changes = append(recorder.changes, change)
	return nil
}
