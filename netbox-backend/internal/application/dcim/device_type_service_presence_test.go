package dcim_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestReplaceDeviceTypePreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend, repository, service := newDeviceTypePresenceService(t, nil)
	seedDeviceTypePresenceState(t, backend)
	before := backend.state.deviceTypes[41]

	deviceType, err := service.ReplaceDeviceType(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceDeviceTypeCommand{
			ID: 41,
			CreateDeviceTypeCommand: appdcim.CreateDeviceTypeCommand{
				Manufacturer: appdcim.FieldValue(shared.ID(9)),
				Model:        appdcim.FieldValue("  Replacement Router  "),
				Slug:         appdcim.FieldValue("  replacement-router  "),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Replacement Router", deviceType.Model())
	assert.Equal(t, "replacement-router", deviceType.Slug().String())
	assert.Equal(t, dcimdomain.DeviceTypeDefaultHeight, deviceType.UHeight().String(),
		"PUT omission is the pinned height reset exception")
	assert.Equal(t, before.PartNumber, deviceType.PartNumber())
	assert.Equal(t, before.ExcludeFromUtilization, deviceType.ExcludeFromUtilization())
	assert.Equal(t, before.IsFullDepth, deviceType.IsFullDepth())
	assert.Equal(t, before.Airflow, deviceType.Airflow())
	assert.Equal(t, before.Description, deviceType.Description())
	assert.Equal(t, before.Comments, deviceType.Comments())
	assert.Equal(t, before.Created, deviceType.Created())
	assert.Equal(t, updatedAt, deviceType.LastUpdated())
	assert.Equal(t, before.DeviceCount, deviceType.DeviceCount())
	assert.Equal(t, before.InterfaceTemplateCount, deviceType.InterfaceTemplateCount())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.getForUpdateCalls)
	assert.Equal(t, 1, repository.updateCalls)
	assert.Zero(t, repository.placementCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, deviceTypePresenceSnapshot(t, before), backend.state.changes[0].Before)
	assert.Equal(t, deviceType.Snapshot(), backend.state.changes[0].After)
}

func TestDeviceTypeScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalidPartNumber := strings.Repeat("p", dcimdomain.DeviceTypePartNumberMaxLength+1)
	invalidDescription := strings.Repeat("d", dcimdomain.DeviceTypeDescriptionMaxLength+1)
	expectedViolations := []shared.FieldViolation{
		{Field: "manufacturer", Reason: "null", Description: "This field may not be null."},
		{Field: "model", Reason: "null", Description: "This field may not be null."},
		{
			Field: "slug", Reason: "invalid",
			Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
		},
		{
			Field: "part_number", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
		{
			Field: "u_height", Reason: "invalid",
			Description: "Ensure that there are no more than 1 decimal places.",
		},
		{
			Field: "exclude_from_utilization", Reason: "null",
			Description: "This field may not be null.",
		},
		{Field: "is_full_depth", Reason: "null", Description: "This field may not be null."},
		{
			Field: "airflow", Reason: "invalid_choice",
			Description: "Unsupported airflow direction.",
		},
		{
			Field: "description", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
		{Field: "comments", Reason: "null", Description: "This field may not be null."},
	}

	for _, test := range []struct {
		name   string
		seed   bool
		mutate func(*appdcim.DeviceTypeService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.DeviceTypeService) error {
				_, err := service.CreateDeviceType(
					t.Context(), testPrincipal(), invalidCreateDeviceTypeCommand(
						invalidPartNumber, invalidDescription,
					),
				)
				return err
			},
		},
		{
			name: "PUT", seed: true,
			mutate: func(service *appdcim.DeviceTypeService) error {
				_, err := service.ReplaceDeviceType(
					t.Context(), testPrincipal(), appdcim.ReplaceDeviceTypeCommand{
						ID: 41,
						CreateDeviceTypeCommand: invalidCreateDeviceTypeCommand(
							invalidPartNumber, invalidDescription,
						),
					},
				)
				return err
			},
		},
		{
			name: "PATCH", seed: true,
			mutate: func(service *appdcim.DeviceTypeService) error {
				invalid := invalidCreateDeviceTypeCommand(invalidPartNumber, invalidDescription)
				_, err := service.UpdateDeviceType(
					t.Context(), testPrincipal(), appdcim.UpdateDeviceTypeCommand{
						ID: 41, Manufacturer: invalid.Manufacturer, Model: invalid.Model,
						Slug: invalid.Slug, PartNumber: invalid.PartNumber,
						UHeight:                invalid.UHeight,
						ExcludeFromUtilization: invalid.ExcludeFromUtilization,
						IsFullDepth:            invalid.IsFullDepth, Airflow: invalid.Airflow,
						Description: invalid.Description, Comments: invalid.Comments,
					},
				)
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend, repository, service := newDeviceTypePresenceService(t, nil)
			if test.seed {
				seedDeviceTypePresenceState(t, backend)
			}
			before := backend.state.clone()

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, expectedViolations, shared.ViolationsOf(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			if test.seed {
				assert.Equal(t, 1, repository.getForUpdateCalls)
			} else {
				assert.Zero(t, repository.getForUpdateCalls)
			}
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Zero(t, repository.placementCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("repository failure rolls back state and audit", func(t *testing.T) {
		writeFailure := errors.New("forced DeviceType repository failure")
		backend, repository, service := newDeviceTypePresenceService(t, nil)
		seedDeviceTypePresenceState(t, backend)
		repository.updateErr = writeFailure
		before := backend.state.clone()

		_, err := service.UpdateDeviceType(
			t.Context(), testPrincipal(), appdcim.UpdateDeviceTypeCommand{
				ID: 41, Description: appdcim.FieldValue("changed but rolled back"),
			},
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, writeFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Empty(t, backend.state.changes)
	})

	t.Run("change recorder failure rolls back state and audit", func(t *testing.T) {
		recorderFailure := errors.New("forced DeviceType recorder failure")
		backend, repository, service := newDeviceTypePresenceService(t, recorderFailure)
		seedDeviceTypePresenceState(t, backend)
		before := backend.state.clone()

		_, err := service.UpdateDeviceType(
			t.Context(), testPrincipal(), appdcim.UpdateDeviceTypeCommand{
				ID: 41, Description: appdcim.FieldValue("changed but rolled back"),
			},
		)
		require.Error(t, err)
		assert.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Empty(t, backend.state.changes)
	})
}

func invalidCreateDeviceTypeCommand(
	partNumber string,
	description string,
) appdcim.CreateDeviceTypeCommand {
	return appdcim.CreateDeviceTypeCommand{
		Manufacturer: appdcim.NullField[shared.ID](), Model: appdcim.NullField[string](),
		Slug: appdcim.FieldValue("invalid slug!"), PartNumber: appdcim.FieldValue(partNumber),
		UHeight:                appdcim.FieldValue("1.00"),
		ExcludeFromUtilization: appdcim.NullField[bool](),
		IsFullDepth:            appdcim.NullField[bool](), Airflow: appdcim.FieldValue(" front-to-rear "),
		Description: appdcim.FieldValue(description), Comments: appdcim.NullField[string](),
	}
}

type deviceTypePresenceTransactionKey struct{}

type deviceTypePresenceState struct {
	nextID      shared.ID
	deviceTypes map[shared.ID]dcimdomain.DeviceTypeState
	changes     []changelog.Change
}

func (state deviceTypePresenceState) clone() deviceTypePresenceState {
	cloned := deviceTypePresenceState{
		nextID: state.nextID,
		deviceTypes: make(
			map[shared.ID]dcimdomain.DeviceTypeState,
			len(state.deviceTypes),
		),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, deviceType := range state.deviceTypes {
		cloned.deviceTypes[id] = deviceType
	}
	return cloned
}

type deviceTypePresenceBackend struct {
	state            deviceTypePresenceState
	transactionCalls int
}

func newDeviceTypePresenceBackend() *deviceTypePresenceBackend {
	return &deviceTypePresenceBackend{state: deviceTypePresenceState{
		nextID: 42, deviceTypes: make(map[shared.ID]dcimdomain.DeviceTypeState),
	}}
}

func (backend *deviceTypePresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, deviceTypePresenceTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *deviceTypePresenceBackend) stateFor(
	ctx context.Context,
) (*deviceTypePresenceState, bool) {
	state, transactional := ctx.Value(deviceTypePresenceTransactionKey{}).(*deviceTypePresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type deviceTypePresenceRepository struct {
	backend           *deviceTypePresenceBackend
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
	placementCalls    int
	updateErr         error
}

func (repository *deviceTypePresenceRepository) List(
	context.Context,
	appdcim.DeviceTypeListCriteria,
) (appdcim.DeviceTypePage, error) {
	return appdcim.DeviceTypePage{}, nil
}

func (repository *deviceTypePresenceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.deviceTypes[id]
	if !ok {
		return nil, shared.NotFound("DeviceType", id)
	}
	return dcimdomain.RestoreDeviceType(persisted)
}

func (repository *deviceTypePresenceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	repository.getForUpdateCalls++
	return repository.Get(ctx, id)
}

func (repository *deviceTypePresenceRepository) Create(
	ctx context.Context,
	deviceType *dcimdomain.DeviceType,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceType created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := deviceType.AssignID(id); err != nil {
		return err
	}
	state.deviceTypes[id] = deviceType.State()
	return nil
}

func (repository *deviceTypePresenceRepository) Update(
	ctx context.Context,
	deviceType *dcimdomain.DeviceType,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceType updated outside transaction")
	}
	state.deviceTypes[deviceType.ID()] = deviceType.State()
	return repository.updateErr
}

func (repository *deviceTypePresenceRepository) ListPositionedDevicesForUpdate(
	context.Context,
) ([]appdcim.PositionedDevice, error) {
	repository.placementCalls++
	return nil, nil
}

func (*deviceTypePresenceRepository) FindDeviceUsingDeviceType(
	context.Context,
	shared.ID,
) (*appdcim.DeviceTypeDependent, error) {
	return nil, nil
}

func (*deviceTypePresenceRepository) ListInterfaceTemplatesForUpdate(
	context.Context,
	shared.ID,
) ([]appdcim.InterfaceTemplateDeletion, error) {
	return nil, nil
}

func (*deviceTypePresenceRepository) DeleteInterfaceTemplate(
	context.Context,
	shared.ID,
) error {
	return nil
}

func (repository *deviceTypePresenceRepository) Delete(
	ctx context.Context,
	deviceType *dcimdomain.DeviceType,
) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceType deleted outside transaction")
	}
	delete(state.deviceTypes, deviceType.ID())
	return nil
}

type deviceTypePresenceManufacturerReader struct {
	manufacturer *dcimdomain.Manufacturer
}

func (reader deviceTypePresenceManufacturerReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Manufacturer, error) {
	if reader.manufacturer == nil || reader.manufacturer.ID() != id {
		return nil, shared.NotFound("Manufacturer", id)
	}
	return reader.manufacturer, nil
}

type deviceTypePresenceRecorder struct {
	backend *deviceTypePresenceBackend
	err     error
}

func (recorder deviceTypePresenceRecorder) Record(
	ctx context.Context,
	change changelog.Change,
) error {
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceType change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

func newDeviceTypePresenceService(
	t *testing.T,
	recorderErr error,
) (*deviceTypePresenceBackend, *deviceTypePresenceRepository, *appdcim.DeviceTypeService) {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: "Acme", Slug: "acme", Description: "Vendor",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, manufacturer.AssignID(9))
	backend := newDeviceTypePresenceBackend()
	repository := &deviceTypePresenceRepository{backend: backend}
	service, err := appdcim.NewDeviceTypeService(
		repository,
		deviceTypePresenceManufacturerReader{manufacturer: manufacturer},
		backend,
		deviceTypePresenceRecorder{backend: backend, err: recorderErr},
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return backend, repository, service
}

func seedDeviceTypePresenceState(t *testing.T, backend *deviceTypePresenceBackend) {
	t.Helper()
	reference, err := dcimdomain.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	backend.state.deviceTypes[41] = dcimdomain.DeviceTypeState{
		ID: 41, Manufacturer: reference, Model: "Original Router", Slug: "original-router",
		PartNumber: "PN-ORIGINAL", UHeight: "2.5", ExcludeFromUtilization: true,
		IsFullDepth: false,
		Airflow:     dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflowRearToFront),
		Description: "Original description", Comments: "Original comments",
		Created: createdAt, LastUpdated: createdAt,
		DeviceCount: 4, InterfaceTemplateCount: 6,
	}
}

func deviceTypePresenceSnapshot(
	t *testing.T,
	state dcimdomain.DeviceTypeState,
) dcimdomain.DeviceTypeSnapshot {
	t.Helper()
	deviceType, err := dcimdomain.RestoreDeviceType(state)
	require.NoError(t, err)
	return deviceType.Snapshot()
}
