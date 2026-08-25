package dcim_test

import (
	"context"
	"errors"
	"sort"
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

func TestReplaceDeviceRolePreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend := newDeviceRolePresenceBackend()
	seedPresenceDeviceRole(t, backend, 1, 0, "Parent", 0, 0)
	seedPresenceDeviceRole(t, backend, 2, 1, "Original Role", 1, 7)
	repository := newTrackingDeviceRoleRepository(backend)
	service := newTrackedDeviceRoleService(
		t, backend, repository, &deviceRolePresenceRecorder{backend: backend},
	)
	before := backend.state.roles[2]

	role, err := service.ReplaceDeviceRole(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceDeviceRoleCommand{
			ID: 2, Name: appdcim.FieldValue("  Replacement Role  "),
			Slug: appdcim.FieldValue("  replacement-role  "),
		},
	)
	require.NoError(t, err)
	assert.True(t, role.Parent().IsRoot(), "PUT omission resets parent to root")
	assert.Equal(t, "Replacement Role", role.Name())
	assert.Equal(t, "replacement-role", role.Slug().String())
	assert.Equal(t, before.Color, role.Color().String())
	assert.Equal(t, before.VMRole, role.VMRole())
	assert.Equal(t, before.Description, role.Description())
	assert.Equal(t, before.Comments, role.Comments())
	assert.Equal(t, before.Created, role.Created())
	assert.Equal(t, updatedAt, role.LastUpdated())
	assert.Equal(t, before.DeviceCount, role.DeviceCount())
	assert.Zero(t, role.Depth())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.hierarchyCalls)
	assert.Equal(t, 1, repository.updateCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, deviceRoleSnapshotFromState(before), backend.state.changes[0].Before)
	assert.Equal(t, role.Snapshot(), backend.state.changes[0].After)
}

func TestDeviceRoleScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalidDescription := strings.Repeat("界", dcimdomain.DeviceRoleDescriptionMaxLength+1)
	expectedViolations := []shared.FieldViolation{
		{Field: "parent", Reason: "invalid", Description: "A valid object ID is required."},
		{Field: "name", Reason: "null", Description: "This field may not be null."},
		{
			Field: "slug", Reason: "invalid",
			Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
		},
		{
			Field: "color", Reason: "invalid",
			Description: "Enter a valid hexadecimal RGB color code.",
		},
		{Field: "vm_role", Reason: "null", Description: "This field may not be null."},
		{
			Field: "description", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
	}

	for _, test := range []struct {
		name          string
		seed          bool
		wantHierarchy int
		mutate        func(*appdcim.DeviceRoleService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.DeviceRoleService) error {
				_, err := service.CreateDeviceRole(t.Context(), testPrincipal(), appdcim.CreateDeviceRoleCommand{
					Parent: appdcim.FieldValue(shared.ID(0)), Name: appdcim.NullField[string](),
					Slug: appdcim.FieldValue("invalid slug!"), Color: appdcim.FieldValue("ABCDEF"),
					VMRole: appdcim.NullField[bool](), Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
		{
			name: "PUT", seed: true, wantHierarchy: 1,
			mutate: func(service *appdcim.DeviceRoleService) error {
				_, err := service.ReplaceDeviceRole(t.Context(), testPrincipal(), appdcim.ReplaceDeviceRoleCommand{
					ID: 2, Parent: appdcim.FieldValue(shared.ID(0)), Name: appdcim.NullField[string](),
					Slug: appdcim.FieldValue("invalid slug!"), Color: appdcim.FieldValue("ABCDEF"),
					VMRole: appdcim.NullField[bool](), Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
		{
			name: "PATCH", seed: true, wantHierarchy: 1,
			mutate: func(service *appdcim.DeviceRoleService) error {
				_, err := service.UpdateDeviceRole(t.Context(), testPrincipal(), appdcim.UpdateDeviceRoleCommand{
					ID: 2, Parent: appdcim.FieldValue(shared.ID(0)), Name: appdcim.NullField[string](),
					Slug: appdcim.FieldValue("invalid slug!"), Color: appdcim.FieldValue("ABCDEF"),
					VMRole: appdcim.NullField[bool](), Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend := newDeviceRolePresenceBackend()
			if test.seed {
				seedPresenceDeviceRole(t, backend, 1, 0, "Parent", 0, 0)
				seedPresenceDeviceRole(t, backend, 2, 1, "Original Role", 1, 7)
			}
			repository := newTrackingDeviceRoleRepository(backend)
			service := newTrackedDeviceRoleService(
				t, backend, repository, &deviceRolePresenceRecorder{backend: backend},
			)
			before := backend.state.clone()

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, expectedViolations, shared.ViolationsOf(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			assert.Equal(t, test.wantHierarchy, repository.hierarchyCalls)
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("missing positive parent ID leaves state unchanged", func(t *testing.T) {
		backend := newDeviceRolePresenceBackend()
		seedPresenceDeviceRole(t, backend, 1, 0, "Parent", 0, 0)
		seedPresenceDeviceRole(t, backend, 2, 1, "Original Role", 1, 7)
		repository := newTrackingDeviceRoleRepository(backend)
		recorder := &deviceRolePresenceRecorder{backend: backend}
		service := newTrackedDeviceRoleService(t, backend, repository, recorder)
		before := backend.state.clone()

		_, err := service.UpdateDeviceRole(t.Context(), testPrincipal(), appdcim.UpdateDeviceRoleCommand{
			ID: 2, Parent: appdcim.FieldValue(shared.ID(999)),
			Description: appdcim.FieldValue("valid sibling"),
		})
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{{
			Field: "parent", Reason: "does_not_exist",
			Description: "The related object does not exist.",
		}}, shared.ViolationsOf(err))
		assert.Equal(t, before, backend.state)
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.hierarchyCalls)
		assert.Zero(t, repository.createCalls)
		assert.Zero(t, repository.updateCalls)
		assert.Zero(t, recorder.calls)
		assert.Len(t, backend.state.changes, len(before.changes))
	})

	t.Run("repository update failure rolls back mutated working state", func(t *testing.T) {
		backend := newDeviceRolePresenceBackend()
		seedPresenceDeviceRole(t, backend, 1, 0, "Parent", 0, 0)
		seedPresenceDeviceRole(t, backend, 2, 1, "Original Role", 1, 7)
		repositoryFailure := errors.New("forced DeviceRole repository update failure")
		repository := newTrackingDeviceRoleRepository(backend)
		repository.updateErr = repositoryFailure
		recorder := &deviceRolePresenceRecorder{backend: backend}
		service := newTrackedDeviceRoleService(t, backend, repository, recorder)
		before := backend.state.clone()

		_, err := service.UpdateDeviceRole(t.Context(), testPrincipal(), appdcim.UpdateDeviceRoleCommand{
			ID: 2, Description: appdcim.FieldValue("changed in working state only"),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, repositoryFailure)
		assert.True(t, repository.updateMutatedWorkingState)
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.hierarchyCalls)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Zero(t, recorder.calls)
		assert.Equal(t, before, backend.state)
		assert.Len(t, backend.state.changes, len(before.changes))
	})

	t.Run("change recorder failure rolls back attempted update", func(t *testing.T) {
		backend := newDeviceRolePresenceBackend()
		seedPresenceDeviceRole(t, backend, 1, 0, "Parent", 0, 0)
		seedPresenceDeviceRole(t, backend, 2, 1, "Original Role", 1, 7)
		repository := newTrackingDeviceRoleRepository(backend)
		recorderFailure := errors.New("forced DeviceRole change recording failure")
		recorder := &deviceRolePresenceRecorder{backend: backend, err: recorderFailure}
		service := newTrackedDeviceRoleService(
			t, backend, repository, recorder,
		)
		before := backend.state.clone()

		_, err := service.UpdateDeviceRole(t.Context(), testPrincipal(), appdcim.UpdateDeviceRoleCommand{
			ID: 2, Description: appdcim.FieldValue("changed but rolled back"),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.hierarchyCalls)
		assert.Equal(t, 1, repository.updateCalls)
		assert.Equal(t, 1, recorder.calls)
		assert.Equal(t, before, backend.state)
		assert.Empty(t, backend.state.changes)
	})
}

type deviceRolePresenceTransactionKey struct{}

type deviceRolePresenceState struct {
	nextID  shared.ID
	roles   map[shared.ID]dcimdomain.DeviceRoleState
	changes []changelog.Change
}

func (state deviceRolePresenceState) clone() deviceRolePresenceState {
	cloned := deviceRolePresenceState{
		nextID: state.nextID, roles: make(map[shared.ID]dcimdomain.DeviceRoleState, len(state.roles)),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, role := range state.roles {
		cloned.roles[id] = role
	}
	return cloned
}

type deviceRolePresenceBackend struct {
	state            deviceRolePresenceState
	transactionCalls int
}

func newDeviceRolePresenceBackend() *deviceRolePresenceBackend {
	return &deviceRolePresenceBackend{state: deviceRolePresenceState{
		nextID: 1, roles: make(map[shared.ID]dcimdomain.DeviceRoleState),
	}}
}

func (backend *deviceRolePresenceBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, deviceRolePresenceTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *deviceRolePresenceBackend) stateFor(
	ctx context.Context,
) (*deviceRolePresenceState, bool) {
	state, transactional := ctx.Value(deviceRolePresenceTransactionKey{}).(*deviceRolePresenceState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type deviceRolePresenceRecorder struct {
	backend *deviceRolePresenceBackend
	err     error
	calls   int
}

func (recorder *deviceRolePresenceRecorder) Record(
	ctx context.Context,
	change changelog.Change,
) error {
	recorder.calls++
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceRole change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

type trackingDeviceRoleRepository struct {
	backend                   *deviceRolePresenceBackend
	createCalls               int
	updateCalls               int
	hierarchyCalls            int
	updateErr                 error
	updateMutatedWorkingState bool
}

func newTrackingDeviceRoleRepository(
	backend *deviceRolePresenceBackend,
) *trackingDeviceRoleRepository {
	return &trackingDeviceRoleRepository{backend: backend}
}

func (repository *trackingDeviceRoleRepository) List(
	context.Context,
	appdcim.DeviceRoleListCriteria,
) (appdcim.DeviceRolePage, error) {
	return appdcim.DeviceRolePage{}, errors.New("unexpected DeviceRole list")
}

func (repository *trackingDeviceRoleRepository) Get(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.DeviceRole, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, found := state.roles[id]
	if !found {
		return nil, shared.NotFound("DeviceRole", id)
	}
	return dcimdomain.RestoreDeviceRole(persisted)
}

func (repository *trackingDeviceRoleRepository) ListHierarchyForUpdate(
	ctx context.Context,
) ([]*dcimdomain.DeviceRole, error) {
	repository.hierarchyCalls++
	state, _ := repository.backend.stateFor(ctx)
	ids := make([]shared.ID, 0, len(state.roles))
	for id := range state.roles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	roles := make([]*dcimdomain.DeviceRole, 0, len(ids))
	for _, id := range ids {
		role, err := dcimdomain.RestoreDeviceRole(state.roles[id])
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (repository *trackingDeviceRoleRepository) Create(
	ctx context.Context,
	role *dcimdomain.DeviceRole,
) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceRole created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := role.AssignID(id); err != nil {
		return err
	}
	projected, err := projectPresenceDeviceRoleState(role.State(), state.roles)
	if err != nil {
		return err
	}
	state.roles[id] = projected
	return nil
}

func (repository *trackingDeviceRoleRepository) Update(
	ctx context.Context,
	role *dcimdomain.DeviceRole,
) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("DeviceRole updated outside transaction")
	}
	projected, err := projectPresenceDeviceRoleState(role.State(), state.roles)
	if err != nil {
		return err
	}
	state.roles[role.ID()] = projected
	repository.updateMutatedWorkingState = true
	if repository.updateErr != nil {
		return repository.updateErr
	}
	return nil
}

func (*trackingDeviceRoleRepository) FindDeviceUsingRoles(
	context.Context,
	[]shared.ID,
) (*appdcim.DeviceRoleDependent, error) {
	return nil, nil
}

func (*trackingDeviceRoleRepository) Delete(
	context.Context,
	*dcimdomain.DeviceRole,
) error {
	return errors.New("unexpected DeviceRole delete")
}

func newTrackedDeviceRoleService(
	t *testing.T,
	backend *deviceRolePresenceBackend,
	repository *trackingDeviceRoleRepository,
	recorder *deviceRolePresenceRecorder,
) *appdcim.DeviceRoleService {
	t.Helper()
	service, err := appdcim.NewDeviceRoleService(
		repository, backend, recorder, authz.AllowAll{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service
}

func seedPresenceDeviceRole(
	t *testing.T,
	backend *deviceRolePresenceBackend,
	id shared.ID,
	parentID shared.ID,
	name string,
	depth uint32,
	deviceCount uint64,
) {
	t.Helper()
	parent := dcimdomain.RootDeviceRoleParent()
	parentDisplay := ""
	if parentID.IsValid() {
		parent = dcimdomain.NonRootDeviceRoleParent(parentID)
		parentDisplay = backend.state.roles[parentID].Name
	}
	role, err := dcimdomain.NewDeviceRole(dcimdomain.DeviceRoleValues{
		Parent: parent, Name: name, Slug: strings.ToLower(strings.ReplaceAll(name, " ", "-")),
		Color: "123abc", VMRole: false, Description: "Original description",
		Comments: "Original comments",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, role.AssignID(id))
	state := role.State()
	state.ParentDisplay = parentDisplay
	state.Depth = depth
	state.DeviceCount = deviceCount
	backend.state.roles[id] = state
	if backend.state.nextID <= id {
		backend.state.nextID = id + 1
	}
}

func projectPresenceDeviceRoleState(
	state dcimdomain.DeviceRoleState,
	roles map[shared.ID]dcimdomain.DeviceRoleState,
) (dcimdomain.DeviceRoleState, error) {
	parentID, present := state.Parent.Get()
	if !present {
		state.ParentDisplay = ""
		state.Depth = 0
		return state, nil
	}
	parent, found := roles[parentID]
	if !found {
		return dcimdomain.DeviceRoleState{}, shared.NotFound("DeviceRole", parentID)
	}
	state.ParentDisplay = parent.Name
	state.Depth = parent.Depth + 1
	return state, nil
}

func deviceRoleSnapshotFromState(state dcimdomain.DeviceRoleState) dcimdomain.DeviceRoleSnapshot {
	var parentID *int64
	if id, present := state.Parent.Get(); present {
		value := id.Int64()
		parentID = &value
	}
	return dcimdomain.DeviceRoleSnapshot{
		ParentID: parentID, Name: state.Name, Slug: state.Slug, Color: state.Color,
		VMRole: state.VMRole, Description: state.Description, Comments: state.Comments,
	}
}
