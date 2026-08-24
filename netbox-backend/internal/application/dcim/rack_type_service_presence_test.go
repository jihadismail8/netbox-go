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

func TestReplaceRackTypePreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend, repository, service := newRackTypePresenceService(t, nil)
	seedRackTypePresenceState(t, backend)
	before := backend.state.rackTypes[1]

	rackType, err := service.ReplaceRackType(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceRackTypeCommand{
			ID: 1,
			CreateRackTypeCommand: appdcim.CreateRackTypeCommand{
				Manufacturer: appdcim.FieldValue(shared.ID(9)),
				Model:        appdcim.FieldValue("  Replacement  "),
				Slug:         appdcim.FieldValue("  replacement  "),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Replacement", rackType.Model())
	assert.Equal(t, "replacement", rackType.Slug().String())
	assert.Equal(t, before.FormFactor, rackType.FormFactor().String())
	assert.Equal(t, before.Width, rackType.Width().Uint32())
	assert.Equal(t, before.UHeight, rackType.UHeight())
	assert.Equal(t, before.StartingUnit, rackType.StartingUnit())
	assert.Equal(t, before.DescUnits, rackType.DescUnits())
	assert.Equal(t, before.Description, rackType.Description())
	assert.Equal(t, before.Comments, rackType.Comments())
	assert.Equal(t, before.Created, rackType.Created())
	assert.Equal(t, updatedAt, rackType.LastUpdated())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.getForUpdateCalls)
	assert.Equal(t, 1, repository.updateCalls)
	assert.Equal(t, 1, repository.propagationCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, rackTypePresenceSnapshot(before), backend.state.changes[0].Before)
	assert.Equal(t, rackType.Snapshot(), backend.state.changes[0].After)
}

func TestRackTypeScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalidDescription := strings.Repeat("d", dcimdomain.RackTypeDescriptionMaxLength+1)
	expectedViolations := []shared.FieldViolation{
		{Field: "manufacturer", Reason: "null", Description: "This field may not be null."},
		{Field: "model", Reason: "null", Description: "This field may not be null."},
		{
			Field: "slug", Reason: "invalid",
			Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
		},
		{Field: "form_factor", Reason: "invalid_choice", Description: "Select a valid choice."},
		{Field: "width", Reason: "blank", Description: "This field may not be blank."},
		{Field: "u_height", Reason: "range", Description: "Ensure this value is between 1 and 100."},
		{
			Field: "starting_unit", Reason: "range",
			Description: "Ensure this value is greater than or equal to 1.",
		},
		{Field: "desc_units", Reason: "null", Description: "This field may not be null."},
		{
			Field: "description", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
		{Field: "comments", Reason: "null", Description: "This field may not be null."},
	}

	for _, test := range []struct {
		name            string
		seed            bool
		wantLockedReads int
		mutate          func(*appdcim.RackTypeService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.RackTypeService) error {
				_, err := service.CreateRackType(
					t.Context(), testPrincipal(), invalidCreateRackTypeCommand(invalidDescription),
				)
				return err
			},
		},
		{
			name: "PUT", seed: true, wantLockedReads: 1,
			mutate: func(service *appdcim.RackTypeService) error {
				_, err := service.ReplaceRackType(t.Context(), testPrincipal(), appdcim.ReplaceRackTypeCommand{
					ID: 1, CreateRackTypeCommand: invalidCreateRackTypeCommand(invalidDescription),
				})
				return err
			},
		},
		{
			name: "PATCH", seed: true, wantLockedReads: 1,
			mutate: func(service *appdcim.RackTypeService) error {
				invalid := invalidCreateRackTypeCommand(invalidDescription)
				_, err := service.UpdateRackType(t.Context(), testPrincipal(), appdcim.UpdateRackTypeCommand{
					ID:           1,
					Manufacturer: invalid.Manufacturer, Model: invalid.Model, Slug: invalid.Slug,
					FormFactor: invalid.FormFactor, Width: invalid.Width, UHeight: invalid.UHeight,
					StartingUnit: invalid.StartingUnit, DescUnits: invalid.DescUnits,
					Description: invalid.Description, Comments: invalid.Comments,
				})
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend, repository, service := newRackTypePresenceService(t, nil)
			if test.seed {
				seedRackTypePresenceState(t, backend)
			}
			before := backend.state.clone()

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, expectedViolations, shared.ViolationsOf(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			assert.Equal(t, test.wantLockedReads, repository.getForUpdateCalls)
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Zero(t, repository.propagationCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("repository failure rolls back an attempted update", func(t *testing.T) {
		writeFailure := errors.New("forced RackType repository failure")
		backend, repository, service := newRackTypePresenceService(t, nil)
		seedRackTypePresenceState(t, backend)
		repository.updateErr = writeFailure
		before := backend.state.clone()

		_, err := service.UpdateRackType(t.Context(), testPrincipal(), appdcim.UpdateRackTypeCommand{
			ID: 1, Description: appdcim.FieldValue("changed but rolled back"),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, writeFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Zero(t, repository.propagationCalls)
		assert.Empty(t, backend.state.changes)
	})

	t.Run("change recorder failure rolls back an attempted update", func(t *testing.T) {
		recorderFailure := errors.New("forced RackType change recording failure")
		backend, repository, service := newRackTypePresenceService(t, recorderFailure)
		seedRackTypePresenceState(t, backend)
		before := backend.state.clone()

		_, err := service.UpdateRackType(t.Context(), testPrincipal(), appdcim.UpdateRackTypeCommand{
			ID: 1, Description: appdcim.FieldValue("changed but rolled back"),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Zero(t, repository.propagationCalls)
		assert.Empty(t, backend.state.changes)
	})
}

func invalidCreateRackTypeCommand(description string) appdcim.CreateRackTypeCommand {
	return appdcim.CreateRackTypeCommand{
		Manufacturer: appdcim.NullField[shared.ID](), Model: appdcim.NullField[string](),
		Slug:       appdcim.FieldValue("invalid slug!"),
		FormFactor: appdcim.FieldValue(" 4-post-cabinet "),
		Width:      appdcim.NullField[uint32](), UHeight: appdcim.FieldValue(uint32(0)),
		StartingUnit: appdcim.FieldValue(uint32(0)), DescUnits: appdcim.NullField[bool](),
		Description: appdcim.FieldValue(description), Comments: appdcim.NullField[string](),
	}
}

type rackTypePresenceTransactionKey struct{}

type rackTypePresenceState struct {
	nextID    shared.ID
	rackTypes map[shared.ID]dcimdomain.RackTypeState
	changes   []changelog.Change
}

func (state rackTypePresenceState) clone() rackTypePresenceState {
	cloned := rackTypePresenceState{
		nextID:    state.nextID,
		rackTypes: make(map[shared.ID]dcimdomain.RackTypeState, len(state.rackTypes)),
		changes:   append([]changelog.Change(nil), state.changes...),
	}
	for id, rackType := range state.rackTypes {
		cloned.rackTypes[id] = rackType
	}
	return cloned
}

type rackTypePresenceBackend struct {
	state            rackTypePresenceState
	transactionCalls int
}

func newRackTypePresenceBackend() *rackTypePresenceBackend {
	return &rackTypePresenceBackend{state: rackTypePresenceState{
		nextID: 2, rackTypes: make(map[shared.ID]dcimdomain.RackTypeState),
	}}
}

func (backend *rackTypePresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, rackTypePresenceTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *rackTypePresenceBackend) stateFor(ctx context.Context) (*rackTypePresenceState, bool) {
	state, transactional := ctx.Value(rackTypePresenceTransactionKey{}).(*rackTypePresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type rackTypePresenceRepository struct {
	backend           *rackTypePresenceBackend
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
	propagationCalls  int
	updateErr         error
}

func (repository *rackTypePresenceRepository) List(
	context.Context,
	appdcim.RackTypeListCriteria,
) (appdcim.RackTypePage, error) {
	return appdcim.RackTypePage{}, nil
}

func (repository *rackTypePresenceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.RackType, error) {
	return repository.get(ctx, id)
}

func (repository *rackTypePresenceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.RackType, error) {
	repository.getForUpdateCalls++
	return repository.get(ctx, id)
}

func (repository *rackTypePresenceRepository) get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.RackType, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.rackTypes[id]
	if !ok {
		return nil, shared.NotFound("RackType", id)
	}
	return dcimdomain.RestoreRackType(persisted)
}

func (repository *rackTypePresenceRepository) Create(
	ctx context.Context,
	rackType *dcimdomain.RackType,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("RackType created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := rackType.AssignID(id); err != nil {
		return err
	}
	state.rackTypes[id] = rackType.State()
	return nil
}

func (repository *rackTypePresenceRepository) Update(
	ctx context.Context,
	rackType *dcimdomain.RackType,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("RackType updated outside transaction")
	}
	state.rackTypes[rackType.ID()] = rackType.State()
	return repository.updateErr
}

func (repository *rackTypePresenceRepository) Delete(
	ctx context.Context,
	rackType *dcimdomain.RackType,
) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("RackType deleted outside transaction")
	}
	delete(state.rackTypes, rackType.ID())
	return nil
}

func (repository *rackTypePresenceRepository) PropagateToRacks(
	ctx context.Context,
	_ shared.ID,
	_ dcimdomain.RackPhysicalAttributes,
	_ shared.Timestamp,
) ([]appdcim.RackPropagationChange, error) {
	repository.propagationCalls++
	if _, transactional := repository.backend.stateFor(ctx); !transactional {
		return nil, errors.New("RackType propagated outside transaction")
	}
	return nil, nil
}

type rackTypePresenceManufacturerReader struct {
	manufacturer *dcimdomain.Manufacturer
}

func (reader rackTypePresenceManufacturerReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.Manufacturer, error) {
	if reader.manufacturer == nil || reader.manufacturer.ID() != id {
		return nil, shared.NotFound("Manufacturer", id)
	}
	return reader.manufacturer, nil
}

type rackTypePresenceRecorder struct {
	backend *rackTypePresenceBackend
	err     error
}

func (recorder rackTypePresenceRecorder) Record(
	ctx context.Context,
	change changelog.Change,
) error {
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("RackType change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

func newRackTypePresenceService(
	t *testing.T,
	recorderErr error,
) (*rackTypePresenceBackend, *rackTypePresenceRepository, *appdcim.RackTypeService) {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: "Acme", Slug: "acme", Description: "Vendor",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, manufacturer.AssignID(9))
	backend := newRackTypePresenceBackend()
	repository := &rackTypePresenceRepository{backend: backend}
	service, err := appdcim.NewRackTypeService(
		repository,
		rackTypePresenceManufacturerReader{manufacturer: manufacturer},
		backend,
		rackTypePresenceRecorder{backend: backend, err: recorderErr},
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return backend, repository, service
}

func seedRackTypePresenceState(t *testing.T, backend *rackTypePresenceBackend) {
	t.Helper()
	reference, err := dcimdomain.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	backend.state.rackTypes[1] = dcimdomain.RackTypeState{
		ID: 1, Manufacturer: reference, Model: "Original", Slug: "original",
		FormFactor: "2-post-frame", Width: 23, UHeight: 80, StartingUnit: 5,
		DescUnits: true, Description: "Original description", Comments: "Original comments",
		Created: createdAt, LastUpdated: createdAt,
	}
}

func rackTypePresenceSnapshot(state dcimdomain.RackTypeState) dcimdomain.RackTypeSnapshot {
	return dcimdomain.RackTypeSnapshot{
		ManufacturerID: state.Manufacturer.ID(), Model: state.Model, Slug: state.Slug,
		FormFactor: state.FormFactor, Width: state.Width, UHeight: state.UHeight,
		StartingUnit: state.StartingUnit, DescUnits: state.DescUnits,
		Description: state.Description, Comments: state.Comments,
	}
}
