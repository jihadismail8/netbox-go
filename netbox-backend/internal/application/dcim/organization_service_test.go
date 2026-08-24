package dcim_test

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestOrganizationServicesRequireAuthorizationAndTransactionDependencies(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	_, manufacturerErr := appdcim.NewManufacturerService(
		&manufacturerMemoryRepository{backend: backend}, backend, &organizationRecorder{backend: backend},
		nil, fixedClock{now: updatedAt},
	)
	_, rackRoleErr := appdcim.NewRackRoleService(
		&rackRoleMemoryRepository{backend: backend}, backend, &organizationRecorder{backend: backend},
		nil, fixedClock{now: updatedAt},
	)
	for _, err := range []error{manufacturerErr, rackRoleErr} {
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
		assert.Contains(t, err.Error(), "authorizer")
	}
}

func TestCreateManufacturerAndRackRoleApplyPinnedDefaultsAndRecordTypedChanges(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	recorder := &organizationRecorder{backend: backend}
	authorizer := &trackingAuthorizer{}
	manufacturerService := newManufacturerTestService(t, backend, recorder, authorizer)
	rackRoleService := newRackRoleTestService(t, backend, recorder, authorizer)

	manufacturer, err := manufacturerService.CreateManufacturer(t.Context(), testPrincipal(), appdcim.CreateManufacturerCommand{
		Name: appdcim.FieldValue("  Acme  "), Slug: appdcim.FieldValue("acme"),
	})
	require.NoError(t, err)
	role, err := rackRoleService.CreateRackRole(t.Context(), testPrincipal(), appdcim.CreateRackRoleCommand{
		Name: appdcim.FieldValue("  Distribution  "), Slug: appdcim.FieldValue("distribution"),
	})
	require.NoError(t, err)

	assert.Empty(t, manufacturer.Description())
	assert.Equal(t, dcimdomain.RackRoleDefaultColor, role.Color().String())
	assert.Empty(t, role.Description())
	assert.Equal(t, 2, backend.transactionCalls)
	require.Len(t, backend.state.changes, 2)
	manufacturerAfter, ok := backend.state.changes[0].After.(dcimdomain.ManufacturerSnapshot)
	require.True(t, ok)
	assert.Equal(t, "Acme", manufacturerAfter.Name)
	roleAfter, ok := backend.state.changes[1].After.(dcimdomain.RackRoleSnapshot)
	require.True(t, ok)
	assert.Equal(t, dcimdomain.RackRoleDefaultColor, roleAfter.Color)
	assert.Equal(t, dcimdomain.ManufacturerObjectType, backend.state.changes[0].ObjectType)
	assert.Equal(t, dcimdomain.RackRoleObjectType, backend.state.changes[1].ObjectType)
}

func TestOrganizationCommandsDistinguishNullOmittedAndBlank(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	service := newManufacturerTestService(
		t, backend, &organizationRecorder{backend: backend}, &trackingAuthorizer{},
	)
	_, err := service.CreateManufacturer(t.Context(), testPrincipal(), appdcim.CreateManufacturerCommand{
		Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("vendor"),
	})
	require.Error(t, err)
	assert.Equal(t, "null", shared.ViolationsOf(err)[0].Reason)

	_, err = service.CreateManufacturer(t.Context(), testPrincipal(), appdcim.CreateManufacturerCommand{
		Name: appdcim.FieldValue(""), Slug: appdcim.FieldValue("vendor"),
	})
	require.Error(t, err)
	assert.Equal(t, "required", shared.ViolationsOf(err)[0].Reason)

	_, err = service.CreateManufacturer(t.Context(), testPrincipal(), appdcim.CreateManufacturerCommand{
		Name: appdcim.FieldValue("Vendor"),
	})
	require.Error(t, err)
	assert.Equal(t, "slug", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "required", shared.ViolationsOf(err)[0].Reason)
}

func TestManufacturerUpdateRollsBackWhenTypedChangeRecordingFails(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	seedManufacturer(t, backend, 1, "Original")
	recorderFailure := errors.New("forced change failure")
	recorder := &organizationRecorder{backend: backend, err: recorderFailure}
	repository := &manufacturerMemoryRepository{backend: backend}
	service, err := appdcim.NewManufacturerService(
		repository, backend, recorder, &trackingAuthorizer{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	_, err = service.UpdateManufacturer(t.Context(), testPrincipal(), appdcim.UpdateManufacturerCommand{
		ID: 1, Description: appdcim.FieldValue("changed"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, recorderFailure)
	assert.Equal(t, 1, repository.updateCalls)
	assert.Equal(t, "Original description", backend.state.manufacturers[1].Description)
	assert.Empty(t, backend.state.changes)
}

func TestReplaceRackRoleResetsOmittedOptionals(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	seedRackRole(t, backend, 2, "Original Role")
	recorder := &organizationRecorder{backend: backend}
	rackRoleService := newRackRoleTestService(t, backend, recorder, &trackingAuthorizer{})

	role, err := rackRoleService.ReplaceRackRole(t.Context(), testPrincipal(), appdcim.ReplaceRackRoleCommand{
		ID: 2, Name: appdcim.FieldValue("Replacement Role"), Slug: appdcim.FieldValue("replacement-role"),
	})
	require.NoError(t, err)
	assert.Equal(t, dcimdomain.RackRoleDefaultColor, role.Color().String())
	assert.Empty(t, role.Description())
}

func TestOrganizationListValidatesCriteriaAndPushesCompleteVisibilityScope(t *testing.T) {
	backend := newOrganizationBackend()
	seedManufacturer(t, backend, 1, "First")
	seedManufacturer(t, backend, 2, "Second")
	seedManufacturer(t, backend, 3, "Third")
	repository := &manufacturerMemoryRepository{backend: backend}
	authorizer := &scopedTrackingAuthorizer{scope: authz.ListScope{
		ObjectIDs: []int64{2, 3}, Constrained: true,
	}}
	service, err := appdcim.NewManufacturerService(
		repository, backend, &organizationRecorder{backend: backend}, authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	page, err := service.ListManufacturers(t.Context(), testPrincipal(), appdcim.ListManufacturersQuery{
		Limit: 1, Ordering: []string{"-created,name"},
		Names: []string{"  Second  ", " Third "}, Slugs: []string{"  second  ", " third "},
		IDs: []int64{-1, 2},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	require.NotNil(t, repository.lastCriteria)
	assert.True(t, repository.lastCriteria.VisibilityConstrained)
	assert.False(t, repository.lastCriteria.DeferPagination)
	assert.Equal(t, []shared.ID{2, 3}, repository.lastCriteria.VisibleObjectIDs)
	assert.Equal(t, []int64{-1, 2}, repository.lastCriteria.IDs)
	assert.Equal(t, []string{"Second", "Third"}, repository.lastCriteria.Names)
	assert.Equal(t, []string{"second", "third"}, repository.lastCriteria.Slugs)
	assert.Equal(t, []appdcim.ManufacturerSort{
		{Field: appdcim.ManufacturerSortCreated, Descending: true},
		{Field: appdcim.ManufacturerSortName},
	}, repository.lastCriteria.Ordering)

	_, err = service.ListManufacturers(t.Context(), testPrincipal(), appdcim.ListManufacturersQuery{
		Limit: appdcim.MaximumManufacturerPageLimit + 1, IDs: []int64{0, -1}, Ordering: []string{"unknown"},
	})
	require.Error(t, err)
	assert.Len(t, shared.ViolationsOf(err), 2)
	assert.Equal(t, 1, repository.listCalls)
}

func TestOrganizationListDistinguishesOmittedAndExplicitZeroLimit(t *testing.T) {
	backend := newOrganizationBackend()
	manufacturerRepository := &manufacturerMemoryRepository{backend: backend}
	manufacturerService, err := appdcim.NewManufacturerService(
		manufacturerRepository, backend, &organizationRecorder{backend: backend},
		&trackingAuthorizer{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	_, err = manufacturerService.ListManufacturers(
		t.Context(), testPrincipal(), appdcim.ListManufacturersQuery{},
	)
	require.NoError(t, err)
	assert.Equal(t, appdcim.DefaultManufacturerPageLimit, manufacturerRepository.lastCriteria.Limit)

	_, err = manufacturerService.ListManufacturers(
		t.Context(), testPrincipal(), appdcim.ListManufacturersQuery{LimitPresent: true},
	)
	require.NoError(t, err)
	assert.Equal(t, appdcim.MaximumManufacturerPageLimit, manufacturerRepository.lastCriteria.Limit)

	rackRoleRepository := &rackRoleMemoryRepository{backend: backend}
	rackRoleService, err := appdcim.NewRackRoleService(
		rackRoleRepository, backend, &organizationRecorder{backend: backend},
		&trackingAuthorizer{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	_, err = rackRoleService.ListRackRoles(
		t.Context(), testPrincipal(), appdcim.ListRackRolesQuery{LimitPresent: true},
	)
	require.NoError(t, err)
	assert.Equal(t, appdcim.MaximumRackRolePageLimit, rackRoleRepository.lastCriteria.Limit)
}

func TestRackRoleListWithoutCompleteScopeAuthorizesBeforeCountAndPage(t *testing.T) {
	backend := newOrganizationBackend()
	seedRackRole(t, backend, 1, "First")
	seedRackRole(t, backend, 2, "Second")
	seedRackRole(t, backend, 3, "Third")
	repository := &rackRoleMemoryRepository{backend: backend}
	authorizer := &trackingAuthorizer{denyObjectIDs: map[int64]struct{}{1: {}}}
	service, err := appdcim.NewRackRoleService(
		repository, backend, &organizationRecorder{backend: backend}, authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	page, err := service.ListRackRoles(t.Context(), testPrincipal(), appdcim.ListRackRolesQuery{
		Limit: 1, Ordering: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	require.NotNil(t, repository.lastCriteria)
	assert.True(t, repository.lastCriteria.DeferPagination)
}

func TestOrganizationDeleteRecordsTypedPrechangeSnapshots(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	seedManufacturer(t, backend, 1, "Vendor")
	seedRackRole(t, backend, 2, "Core")
	recorder := &organizationRecorder{backend: backend}
	manufacturerService := newManufacturerTestService(t, backend, recorder, &trackingAuthorizer{})
	rackRoleService := newRackRoleTestService(t, backend, recorder, &trackingAuthorizer{})

	require.NoError(t, manufacturerService.DeleteManufacturer(
		t.Context(), testPrincipal(), appdcim.DeleteManufacturerCommand{ID: 1},
	))
	require.NoError(t, rackRoleService.DeleteRackRole(
		t.Context(), testPrincipal(), appdcim.DeleteRackRoleCommand{ID: 2},
	))
	assert.Empty(t, backend.state.manufacturers)
	assert.Empty(t, backend.state.rackRoles)
	require.Len(t, backend.state.changes, 2)
	_, manufacturerSnapshot := backend.state.changes[0].Before.(dcimdomain.ManufacturerSnapshot)
	_, rackRoleSnapshot := backend.state.changes[1].Before.(dcimdomain.RackRoleSnapshot)
	assert.True(t, manufacturerSnapshot)
	assert.True(t, rackRoleSnapshot)
	assert.Nil(t, backend.state.changes[0].After)
	assert.Nil(t, backend.state.changes[1].After)
}

func TestOrganizationServiceRejectsUnauthenticatedBeforePersistence(t *testing.T) {
	t.Parallel()
	backend := newOrganizationBackend()
	repository := &manufacturerMemoryRepository{backend: backend}
	service := newManufacturerTestService(
		t, backend, &organizationRecorder{backend: backend}, &trackingAuthorizer{},
	)
	_, err := service.GetManufacturer(t.Context(), identity.Principal{}, appdcim.GetManufacturerQuery{ID: 1})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonUnauthenticated))
	assert.Zero(t, repository.getCalls)
}

func newManufacturerTestService(
	t *testing.T,
	backend *organizationBackend,
	recorder *organizationRecorder,
	authorizer authz.ResourceAuthorizer,
) *appdcim.ManufacturerService {
	t.Helper()
	service, err := appdcim.NewManufacturerService(
		&manufacturerMemoryRepository{backend: backend}, backend, recorder, authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service
}

func newRackRoleTestService(
	t *testing.T,
	backend *organizationBackend,
	recorder *organizationRecorder,
	authorizer authz.ResourceAuthorizer,
) *appdcim.RackRoleService {
	t.Helper()
	service, err := appdcim.NewRackRoleService(
		&rackRoleMemoryRepository{backend: backend}, backend, recorder, authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service
}

type organizationTransactionKey struct{}

type organizationState struct {
	nextID        shared.ID
	manufacturers map[shared.ID]dcimdomain.ManufacturerState
	rackRoles     map[shared.ID]dcimdomain.RackRoleState
	changes       []changelog.Change
}

func (state organizationState) clone() organizationState {
	cloned := organizationState{
		nextID:        state.nextID,
		manufacturers: make(map[shared.ID]dcimdomain.ManufacturerState, len(state.manufacturers)),
		rackRoles:     make(map[shared.ID]dcimdomain.RackRoleState, len(state.rackRoles)),
		changes:       append([]changelog.Change(nil), state.changes...),
	}
	for id, manufacturer := range state.manufacturers {
		cloned.manufacturers[id] = manufacturer
	}
	for id, role := range state.rackRoles {
		cloned.rackRoles[id] = role
	}
	return cloned
}

type organizationBackend struct {
	state            organizationState
	transactionCalls int
}

func newOrganizationBackend() *organizationBackend {
	return &organizationBackend{state: organizationState{
		nextID:        1,
		manufacturers: make(map[shared.ID]dcimdomain.ManufacturerState),
		rackRoles:     make(map[shared.ID]dcimdomain.RackRoleState),
	}}
}

func (backend *organizationBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, organizationTransactionKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *organizationBackend) stateFor(ctx context.Context) (*organizationState, bool) {
	state, transactional := ctx.Value(organizationTransactionKey{}).(*organizationState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type organizationRecorder struct {
	backend *organizationBackend
	err     error
}

func (recorder *organizationRecorder) Record(ctx context.Context, change changelog.Change) error {
	if recorder.err != nil {
		return recorder.err
	}
	state, transactional := recorder.backend.stateFor(ctx)
	if !transactional {
		return errors.New("change recorded outside transaction")
	}
	state.changes = append(state.changes, change)
	return nil
}

type manufacturerMemoryRepository struct {
	backend      *organizationBackend
	lastCriteria *appdcim.ManufacturerListCriteria
	listCalls    int
	getCalls     int
	updateCalls  int
}

func (repository *manufacturerMemoryRepository) List(
	ctx context.Context,
	criteria appdcim.ManufacturerListCriteria,
) (appdcim.ManufacturerPage, error) {
	repository.listCalls++
	copyCriteria := criteria
	copyCriteria.IDs = append([]int64(nil), criteria.IDs...)
	copyCriteria.Names = append([]string(nil), criteria.Names...)
	copyCriteria.Slugs = append([]string(nil), criteria.Slugs...)
	copyCriteria.Ordering = append([]appdcim.ManufacturerSort(nil), criteria.Ordering...)
	copyCriteria.VisibleObjectIDs = append([]shared.ID(nil), criteria.VisibleObjectIDs...)
	repository.lastCriteria = &copyCriteria
	state, _ := repository.backend.stateFor(ctx)
	ids := manufacturerIDs(state.manufacturers)
	visible := idSet(criteria.VisibleObjectIDs)
	results := make([]*dcimdomain.Manufacturer, 0, len(ids))
	for _, id := range ids {
		if criteria.VisibilityConstrained && !visible[id] {
			continue
		}
		manufacturer, err := dcimdomain.RestoreManufacturer(state.manufacturers[id])
		if err != nil {
			return appdcim.ManufacturerPage{}, err
		}
		results = append(results, manufacturer)
	}
	count := uint64(len(results))
	if !criteria.DeferPagination {
		results = pageManufacturers(results, criteria.Offset, criteria.Limit)
	}
	return appdcim.ManufacturerPage{Count: count, Results: results}, nil
}

func (repository *manufacturerMemoryRepository) Get(ctx context.Context, id shared.ID) (*dcimdomain.Manufacturer, error) {
	repository.getCalls++
	return repository.get(ctx, id)
}

func (repository *manufacturerMemoryRepository) GetForUpdate(ctx context.Context, id shared.ID) (*dcimdomain.Manufacturer, error) {
	return repository.get(ctx, id)
}

func (repository *manufacturerMemoryRepository) get(ctx context.Context, id shared.ID) (*dcimdomain.Manufacturer, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.manufacturers[id]
	if !ok {
		return nil, shared.NotFound("Manufacturer", id)
	}
	return dcimdomain.RestoreManufacturer(persisted)
}

func (repository *manufacturerMemoryRepository) Create(ctx context.Context, manufacturer *dcimdomain.Manufacturer) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("manufacturer created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := manufacturer.AssignID(id); err != nil {
		return err
	}
	state.manufacturers[id] = manufacturer.State()
	return nil
}

func (repository *manufacturerMemoryRepository) Update(ctx context.Context, manufacturer *dcimdomain.Manufacturer) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("manufacturer updated outside transaction")
	}
	state.manufacturers[manufacturer.ID()] = manufacturer.State()
	return nil
}

func (repository *manufacturerMemoryRepository) Delete(ctx context.Context, manufacturer *dcimdomain.Manufacturer) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("manufacturer deleted outside transaction")
	}
	delete(state.manufacturers, manufacturer.ID())
	return nil
}

type rackRoleMemoryRepository struct {
	backend      *organizationBackend
	lastCriteria *appdcim.RackRoleListCriteria
}

func (repository *rackRoleMemoryRepository) List(
	ctx context.Context,
	criteria appdcim.RackRoleListCriteria,
) (appdcim.RackRolePage, error) {
	copyCriteria := criteria
	copyCriteria.IDs = append([]int64(nil), criteria.IDs...)
	copyCriteria.Names = append([]string(nil), criteria.Names...)
	copyCriteria.Slugs = append([]string(nil), criteria.Slugs...)
	copyCriteria.Ordering = append([]appdcim.RackRoleSort(nil), criteria.Ordering...)
	copyCriteria.VisibleObjectIDs = append([]shared.ID(nil), criteria.VisibleObjectIDs...)
	repository.lastCriteria = &copyCriteria
	state, _ := repository.backend.stateFor(ctx)
	ids := rackRoleIDs(state.rackRoles)
	visible := idSet(criteria.VisibleObjectIDs)
	results := make([]*dcimdomain.RackRole, 0, len(ids))
	for _, id := range ids {
		if criteria.VisibilityConstrained && !visible[id] {
			continue
		}
		role, err := dcimdomain.RestoreRackRole(state.rackRoles[id])
		if err != nil {
			return appdcim.RackRolePage{}, err
		}
		results = append(results, role)
	}
	count := uint64(len(results))
	if !criteria.DeferPagination {
		results = pageRackRoles(results, criteria.Offset, criteria.Limit)
	}
	return appdcim.RackRolePage{Count: count, Results: results}, nil
}

func (repository *rackRoleMemoryRepository) Get(ctx context.Context, id shared.ID) (*dcimdomain.RackRole, error) {
	return repository.get(ctx, id)
}

func (repository *rackRoleMemoryRepository) GetForUpdate(ctx context.Context, id shared.ID) (*dcimdomain.RackRole, error) {
	return repository.get(ctx, id)
}

func (repository *rackRoleMemoryRepository) get(ctx context.Context, id shared.ID) (*dcimdomain.RackRole, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, ok := state.rackRoles[id]
	if !ok {
		return nil, shared.NotFound("RackRole", id)
	}
	return dcimdomain.RestoreRackRole(persisted)
}

func (repository *rackRoleMemoryRepository) Create(ctx context.Context, role *dcimdomain.RackRole) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("rack role created outside transaction")
	}
	id := state.nextID
	state.nextID++
	if err := role.AssignID(id); err != nil {
		return err
	}
	state.rackRoles[id] = role.State()
	return nil
}

func (repository *rackRoleMemoryRepository) Update(ctx context.Context, role *dcimdomain.RackRole) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("rack role updated outside transaction")
	}
	state.rackRoles[role.ID()] = role.State()
	return nil
}

func (repository *rackRoleMemoryRepository) Delete(ctx context.Context, role *dcimdomain.RackRole) error {
	state, transactional := repository.backend.stateFor(ctx)
	if !transactional {
		return errors.New("rack role deleted outside transaction")
	}
	delete(state.rackRoles, role.ID())
	return nil
}

func seedManufacturer(t *testing.T, backend *organizationBackend, id shared.ID, name string) {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: name, Slug: slugForName(name), Description: "Original description",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, manufacturer.AssignID(id))
	backend.state.manufacturers[id] = manufacturer.State()
	if backend.state.nextID <= id {
		backend.state.nextID = id + 1
	}
}

func seedRackRole(t *testing.T, backend *organizationBackend, id shared.ID, name string) {
	t.Helper()
	role, err := dcimdomain.NewRackRole(dcimdomain.RackRoleValues{
		Name: name, Slug: slugForName(name), Color: "123abc", Description: "Original description",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, role.AssignID(id))
	backend.state.rackRoles[id] = role.State()
	if backend.state.nextID <= id {
		backend.state.nextID = id + 1
	}
}

func slugForName(value string) string {
	result := make([]byte, 0, len(value))
	for _, character := range []byte(value) {
		if character == ' ' {
			result = append(result, '-')
		} else if character >= 'A' && character <= 'Z' {
			result = append(result, character+('a'-'A'))
		} else {
			result = append(result, character)
		}
	}
	return string(result)
}

func manufacturerIDs(values map[shared.ID]dcimdomain.ManufacturerState) []shared.ID {
	ids := make([]shared.ID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	return ids
}

func rackRoleIDs(values map[shared.ID]dcimdomain.RackRoleState) []shared.ID {
	ids := make([]shared.ID, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	return ids
}

func idSet(ids []shared.ID) map[shared.ID]bool {
	set := make(map[shared.ID]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func pageManufacturers(
	values []*dcimdomain.Manufacturer,
	offset uint32,
	limit uint32,
) []*dcimdomain.Manufacturer {
	start := min(int(offset), len(values))
	end := min(start+int(limit), len(values))
	return values[start:end]
}

func pageRackRoles(values []*dcimdomain.RackRole, offset uint32, limit uint32) []*dcimdomain.RackRole {
	start := min(int(offset), len(values))
	end := min(start+int(limit), len(values))
	return values[start:end]
}
