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

func TestReplaceInterfaceTemplatePreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend, repository, service := newInterfaceTemplatePresenceService(t, nil)
	seedInterfaceTemplatePresenceState(t, backend)
	before := backend.state.templates[41]

	template, err := service.ReplaceInterfaceTemplate(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceInterfaceTemplateCommand{
			ID: 41,
			CreateInterfaceTemplateCommand: appdcim.CreateInterfaceTemplateCommand{
				DeviceType: appdcim.FieldValue(shared.ID(9)),
				Name:       appdcim.FieldValue("  Ethernet2  "),
				Type:       appdcim.FieldValue("10gbase-sr"),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Ethernet2", template.Name())
	assert.Equal(t, dcimdomain.InterfaceType("10gbase-sr"), template.Type())
	assert.Equal(t, before.Label, template.Label())
	assert.Equal(t, before.Enabled, template.Enabled())
	assert.Equal(t, before.MgmtOnly, template.MgmtOnly())
	assert.Equal(t, before.Description, template.Description())
	assert.Equal(t, before.Created, template.Created())
	assert.Equal(t, updatedAt, template.LastUpdated())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.getForUpdateCalls)
	assert.Equal(t, 1, repository.updateCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, interfaceTemplatePresenceSnapshot(t, before), backend.state.changes[0].Before)
	assert.Equal(t, template.Snapshot(), backend.state.changes[0].After)
}

func TestInterfaceTemplateScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalidLabel := strings.Repeat("l", dcimdomain.InterfaceTemplateLabelMaxLength+1)
	invalidDescription := strings.Repeat(
		"d", dcimdomain.InterfaceTemplateDescriptionMaxLength+1,
	)
	expectedViolations := []shared.FieldViolation{
		{Field: "device_type", Reason: "null", Description: "This field may not be null."},
		{Field: "name", Reason: "null", Description: "This field may not be null."},
		{
			Field: "label", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
		{Field: "type", Reason: "invalid_choice", Description: "Select a valid choice."},
		{Field: "enabled", Reason: "null", Description: "This field may not be null."},
		{Field: "mgmt_only", Reason: "null", Description: "This field may not be null."},
		{
			Field: "description", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
	}

	for _, test := range []struct {
		name   string
		seed   bool
		mutate func(*appdcim.InterfaceTemplateService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.InterfaceTemplateService) error {
				_, err := service.CreateInterfaceTemplate(
					t.Context(), testPrincipal(), invalidCreateInterfaceTemplateCommand(
						invalidLabel, invalidDescription,
					),
				)
				return err
			},
		},
		{
			name: "PUT", seed: true,
			mutate: func(service *appdcim.InterfaceTemplateService) error {
				_, err := service.ReplaceInterfaceTemplate(
					t.Context(), testPrincipal(), appdcim.ReplaceInterfaceTemplateCommand{
						ID: 41,
						CreateInterfaceTemplateCommand: invalidCreateInterfaceTemplateCommand(
							invalidLabel, invalidDescription,
						),
					},
				)
				return err
			},
		},
		{
			name: "PATCH", seed: true,
			mutate: func(service *appdcim.InterfaceTemplateService) error {
				invalid := invalidCreateInterfaceTemplateCommand(
					invalidLabel, invalidDescription,
				)
				_, err := service.UpdateInterfaceTemplate(
					t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
						ID: 41, DeviceType: invalid.DeviceType, Name: invalid.Name,
						Label: invalid.Label, Type: invalid.Type, Enabled: invalid.Enabled,
						MgmtOnly: invalid.MgmtOnly, Description: invalid.Description,
					},
				)
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend, repository, service := newInterfaceTemplatePresenceService(t, nil)
			if test.seed {
				seedInterfaceTemplatePresenceState(t, backend)
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
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("repository failure rolls back state and audit", func(t *testing.T) {
		writeFailure := errors.New("forced InterfaceTemplate repository failure")
		backend, repository, service := newInterfaceTemplatePresenceService(t, nil)
		seedInterfaceTemplatePresenceState(t, backend)
		repository.updateErr = writeFailure
		before := backend.state.clone()

		_, err := service.UpdateInterfaceTemplate(
			t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
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
		recorderFailure := errors.New("forced InterfaceTemplate recorder failure")
		backend, repository, service := newInterfaceTemplatePresenceService(t, recorderFailure)
		seedInterfaceTemplatePresenceState(t, backend)
		before := backend.state.clone()

		_, err := service.UpdateInterfaceTemplate(
			t.Context(), testPrincipal(), appdcim.UpdateInterfaceTemplateCommand{
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

func invalidCreateInterfaceTemplateCommand(
	label string,
	description string,
) appdcim.CreateInterfaceTemplateCommand {
	return appdcim.CreateInterfaceTemplateCommand{
		DeviceType: appdcim.NullField[shared.ID](), Name: appdcim.NullField[string](),
		Label: appdcim.FieldValue(label), Type: appdcim.FieldValue(" bridge "),
		Enabled: appdcim.NullField[bool](), MgmtOnly: appdcim.NullField[bool](),
		Description: appdcim.FieldValue(description),
	}
}

type interfaceTemplatePresenceTransactionKey struct{}

type interfaceTemplatePresenceState struct {
	nextID    shared.ID
	templates map[shared.ID]dcimdomain.InterfaceTemplateState
	changes   []changelog.Change
}

func (state interfaceTemplatePresenceState) clone() interfaceTemplatePresenceState {
	cloned := interfaceTemplatePresenceState{
		nextID: state.nextID,
		templates: make(
			map[shared.ID]dcimdomain.InterfaceTemplateState,
			len(state.templates),
		),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, template := range state.templates {
		cloned.templates[id] = template
	}
	return cloned
}

type interfaceTemplatePresenceBackend struct {
	state            interfaceTemplatePresenceState
	transactionCalls int
}

func newInterfaceTemplatePresenceBackend() *interfaceTemplatePresenceBackend {
	return &interfaceTemplatePresenceBackend{state: interfaceTemplatePresenceState{
		nextID: 42, templates: make(map[shared.ID]dcimdomain.InterfaceTemplateState),
	}}
}

func (backend *interfaceTemplatePresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(
		ctx, interfaceTemplatePresenceTransactionKey{}, &working,
	)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *interfaceTemplatePresenceBackend) stateFor(
	ctx context.Context,
) (*interfaceTemplatePresenceState, bool) {
	state, transactional := ctx.Value(
		interfaceTemplatePresenceTransactionKey{},
	).(*interfaceTemplatePresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type interfaceTemplatePresenceRepository struct {
	backend           *interfaceTemplatePresenceBackend
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
	updateErr         error
}

func (*interfaceTemplatePresenceRepository) List(
	context.Context,
	appdcim.InterfaceTemplateListCriteria,
) (appdcim.InterfaceTemplatePage, error) {
	return appdcim.InterfaceTemplatePage{}, nil
}

func (repository *interfaceTemplatePresenceRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.InterfaceTemplate, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.templates[id]
	if !ok {
		return nil, shared.NotFound("InterfaceTemplate", id)
	}
	return dcimdomain.RestoreInterfaceTemplate(persisted)
}

func (repository *interfaceTemplatePresenceRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.InterfaceTemplate, error) {
	repository.getForUpdateCalls++
	return repository.Get(ctx, id)
}

func (repository *interfaceTemplatePresenceRepository) Create(
	ctx context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("InterfaceTemplate created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := template.AssignID(id); err != nil {
		return err
	}
	state.templates[id] = template.State()
	return nil
}

func (repository *interfaceTemplatePresenceRepository) Update(
	ctx context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("InterfaceTemplate updated outside transaction")
	}
	state.templates[template.ID()] = template.State()
	return repository.updateErr
}

func (repository *interfaceTemplatePresenceRepository) Delete(
	ctx context.Context,
	template *dcimdomain.InterfaceTemplate,
) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("InterfaceTemplate deleted outside transaction")
	}
	delete(state.templates, template.ID())
	return nil
}

type interfaceTemplatePresenceDeviceTypeReader struct {
	deviceType *dcimdomain.DeviceType
}

func (reader interfaceTemplatePresenceDeviceTypeReader) Get(
	_ context.Context,
	id shared.ID,
) (*dcimdomain.DeviceType, error) {
	if reader.deviceType == nil || reader.deviceType.ID() != id {
		return nil, shared.NotFound("DeviceType", id)
	}
	return reader.deviceType, nil
}

type interfaceTemplatePresenceRecorder struct {
	backend *interfaceTemplatePresenceBackend
	err     error
}

func (recorder interfaceTemplatePresenceRecorder) Record(
	ctx context.Context,
	change changelog.Change,
) error {
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("InterfaceTemplate change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

func newInterfaceTemplatePresenceService(
	t *testing.T,
	recorderErr error,
) (
	*interfaceTemplatePresenceBackend,
	*interfaceTemplatePresenceRepository,
	*appdcim.InterfaceTemplateService,
) {
	t.Helper()
	deviceType := newInterfaceTemplateDeviceTypeFixture(t, 9, "Router", "router")
	backend := newInterfaceTemplatePresenceBackend()
	repository := &interfaceTemplatePresenceRepository{backend: backend}
	service, err := appdcim.NewInterfaceTemplateService(
		repository,
		interfaceTemplatePresenceDeviceTypeReader{deviceType: deviceType},
		backend,
		interfaceTemplatePresenceRecorder{backend: backend, err: recorderErr},
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return backend, repository, service
}

func seedInterfaceTemplatePresenceState(
	t *testing.T,
	backend *interfaceTemplatePresenceBackend,
) {
	t.Helper()
	reference, err := dcimdomain.NewDeviceTypeReference(9, "Router", "router")
	require.NoError(t, err)
	backend.state.templates[41] = dcimdomain.InterfaceTemplateState{
		ID: 41, DeviceType: reference, Name: "Ethernet1", Label: "Original label",
		Type: "1000base-t", Enabled: false, MgmtOnly: true,
		Description: "Original description", Created: createdAt, LastUpdated: createdAt,
	}
}

func interfaceTemplatePresenceSnapshot(
	t *testing.T,
	state dcimdomain.InterfaceTemplateState,
) dcimdomain.InterfaceTemplateSnapshot {
	t.Helper()
	template, err := dcimdomain.RestoreInterfaceTemplate(state)
	require.NoError(t, err)
	return template.Snapshot()
}
