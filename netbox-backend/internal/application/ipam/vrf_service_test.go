package ipam_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var (
	applicationCreatedAt = shared.NewTimestamp(time.Date(2026, time.July, 18, 8, 0, 0, 0, time.UTC))
	applicationUpdatedAt = shared.NewTimestamp(time.Date(2026, time.July, 18, 9, 0, 0, 0, time.UTC))
)

func TestNewVRFServiceRequiresSecurityAndTransactionDependencies(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	_, err := applicationipam.NewVRFService(
		&memoryRepository{backend: backend},
		backend,
		&memoryRecorder{backend: backend},
		nil,
		fixedClock{now: applicationUpdatedAt},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
	assert.Contains(t, err.Error(), "authorizer")
}

func TestCreateVRFAppliesDefaultsAndRecordsTypedChangeInTransaction(t *testing.T) {
	t.Parallel()

	service, backend, repository, recorder := newTestService(t, &trackingAuthorizer{})
	vrf, err := service.CreateVRF(t.Context(), testPrincipal(), applicationipam.CreateVRFCommand{
		Name:        applicationipam.FieldValue("  Tenant Blue  "),
		Description: applicationipam.FieldValue("  production  "),
	})
	require.NoError(t, err)

	assert.Equal(t, shared.ID(1), vrf.ID())
	assert.True(t, vrf.RD().IsNull())
	assert.True(t, vrf.EnforceUnique())
	assert.Equal(t, "production", vrf.Description())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.createCalls)
	assert.False(t, repository.mutationOutsideTransaction)
	assert.False(t, recorder.outsideTransaction)
	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	assert.Equal(t, changelog.ActionCreate, change.Action)
	assert.Equal(t, domainipam.VRFObjectType, change.ObjectType)
	after, ok := change.After.(domainipam.VRFSnapshot)
	require.True(t, ok)
	assert.Equal(t, "Tenant Blue", after.Name)
	assert.True(t, after.RD.IsNull())
}

func TestCreateVRFPreservesPresentBlankRDAndRejectsDuplicateNullsOnlyWhereRequired(t *testing.T) {
	t.Parallel()

	service, backend, _, _ := newTestService(t, &trackingAuthorizer{})
	vrf, err := service.CreateVRF(t.Context(), testPrincipal(), applicationipam.CreateVRFCommand{
		Name: applicationipam.FieldValue("Blank RD"),
		RD:   applicationipam.FieldValue("  "),
	})
	require.NoError(t, err)
	rd, present := vrf.RD().Get()
	assert.True(t, present)
	assert.Empty(t, rd.String())

	_, err = service.CreateVRF(t.Context(), testPrincipal(), applicationipam.CreateVRFCommand{
		Name:          applicationipam.FieldValue("Null bool"),
		EnforceUnique: applicationipam.NullField[bool](),
	})
	require.Error(t, err)
	assert.Equal(t, "enforce_unique", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "null", shared.ViolationsOf(err)[0].Reason)

	_, err = service.CreateVRF(t.Context(), testPrincipal(), applicationipam.CreateVRFCommand{
		Name: applicationipam.NullField[string](),
	})
	require.Error(t, err)
	assert.Equal(t, "name", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, 3, backend.transactionCalls)
}

func TestDeniedCreateStopsBeforeValidationAndTransaction(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	repository := &memoryRepository{backend: backend}
	service, err := applicationipam.NewVRFService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		&trackingAuthorizer{denyAction: authz.Add},
		fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)

	_, err = service.CreateVRF(t.Context(), testPrincipal(), applicationipam.CreateVRFCommand{})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonForbidden))
	assert.Zero(t, backend.transactionCalls)
	assert.Zero(t, repository.createCalls)
}

func TestUpdateCanClearRDAndRollsBackWhenChangeRecordingFails(t *testing.T) {
	t.Parallel()

	service, backend, repository, recorder := newTestService(t, &trackingAuthorizer{})
	seedVRF(t, backend, 1, "Original", "65000:10")
	recorder.err = errors.New("forced change failure")

	_, err := service.UpdateVRF(t.Context(), testPrincipal(), applicationipam.UpdateVRFCommand{
		ID: 1,
		RD: applicationipam.NullField[string](),
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
	assert.Equal(t, 1, repository.updateCalls)
	assert.Empty(t, backend.state.changes)
	persisted, restoreErr := domainipam.RestoreVRF(backend.state.vrfs[1])
	require.NoError(t, restoreErr)
	assert.Equal(t, "65000:10", requiredRD(t, persisted.RD()))
	assert.Equal(t, applicationCreatedAt, persisted.LastUpdated())

	recorder.err = nil
	updated, err := service.UpdateVRF(t.Context(), testPrincipal(), applicationipam.UpdateVRFCommand{
		ID: 1,
		RD: applicationipam.NullField[string](),
	})
	require.NoError(t, err)
	assert.True(t, updated.RD().IsNull())
	assert.Equal(t, applicationUpdatedAt, updated.LastUpdated())
	change := backend.state.changes[0]
	before := change.Before.(domainipam.VRFSnapshot)
	after := change.After.(domainipam.VRFSnapshot)
	assert.Equal(t, "65000:10", requiredRD(t, before.RD))
	assert.True(t, after.RD.IsNull())
}

func TestReplaceVRFResetsOmittedOptionalFieldsToDefaults(t *testing.T) {
	t.Parallel()

	service, backend, _, _ := newTestService(t, &trackingAuthorizer{})
	seedVRF(t, backend, 1, "Original", "65000:10")
	vrf, err := service.ReplaceVRF(t.Context(), testPrincipal(), applicationipam.ReplaceVRFCommand{
		ID:   1,
		Name: applicationipam.FieldValue("Replacement"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Replacement", vrf.Name())
	assert.True(t, vrf.RD().IsNull())
	assert.True(t, vrf.EnforceUnique())
	assert.Empty(t, vrf.Description())
	assert.Empty(t, vrf.Comments())
}

func TestEmptyUpdateAndInvalidIDDoNotWrite(t *testing.T) {
	t.Parallel()

	service, backend, repository, _ := newTestService(t, &trackingAuthorizer{})
	seedVRF(t, backend, 1, "Original", "65000:10")

	_, err := service.UpdateVRF(t.Context(), testPrincipal(), applicationipam.UpdateVRFCommand{ID: 1})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	assert.Zero(t, repository.updateCalls)

	_, err = service.GetVRF(t.Context(), testPrincipal(), applicationipam.GetVRFQuery{})
	require.Error(t, err)
	assert.Zero(t, repository.getCalls)
}

func TestDeleteVRFRecordsPrechangeSnapshot(t *testing.T) {
	t.Parallel()

	service, backend, repository, _ := newTestService(t, &trackingAuthorizer{})
	seedVRF(t, backend, 1, "Original", "65000:10")
	require.NoError(t, service.DeleteVRF(t.Context(), testPrincipal(), applicationipam.DeleteVRFCommand{ID: 1}))
	assert.Empty(t, backend.state.vrfs)
	assert.Equal(t, 1, repository.deleteCalls)
	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	assert.Equal(t, changelog.ActionDelete, change.Action)
	assert.Equal(t, "Original", change.Before.(domainipam.VRFSnapshot).Name)
	assert.Nil(t, change.After)
}

func TestListVRFsValidatesAndNormalizesCriteria(t *testing.T) {
	t.Parallel()

	service, _, repository, _ := newTestService(t, &trackingAuthorizer{})
	enforceUnique := false
	_, err := service.ListVRFs(t.Context(), testPrincipal(), applicationipam.ListVRFsQuery{
		Ordering:      []string{"-rd,name"},
		Names:         []string{"  Tenant Blue  ", "Tenant Green"},
		RDs:           []string{"  65000:100  ", "65000:200"},
		EnforceUnique: &enforceUnique,
	})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	criteria := *repository.lastCriteria
	assert.Equal(t, applicationipam.DefaultVRFPageLimit, criteria.Limit)
	assert.Equal(t, []string{"Tenant Blue", "Tenant Green"}, criteria.Names)
	require.Len(t, criteria.RDs, 2)
	assert.Equal(t, "65000:100", criteria.RDs[0].String())
	assert.Equal(t, "65000:200", criteria.RDs[1].String())
	assert.Equal(t, []applicationipam.VRFSort{
		{Field: applicationipam.VRFSortRD, Descending: true},
		{Field: applicationipam.VRFSortName},
	}, criteria.Ordering)

	tooLongRD := strings.Repeat("r", domainipam.VRFRouteDistinguisherMaxLength+1)
	_, err = service.ListVRFs(t.Context(), testPrincipal(), applicationipam.ListVRFsQuery{
		Limit:    applicationipam.MaximumVRFPageLimit + 1,
		Ordering: []string{"unknown"},
		RDs:      []string{tooLongRD},
	})
	require.Error(t, err)
	assert.Len(t, shared.ViolationsOf(err), 3)
	assert.Equal(t, 1, repository.listCalls)
}

func TestListVRFsPreservesPinnedLimitFiltersAndDefaultOrdering(t *testing.T) {
	t.Parallel()

	service, _, repository, _ := newTestService(t, &trackingAuthorizer{})
	_, err := service.ListVRFs(t.Context(), testPrincipal(), applicationipam.ListVRFsQuery{
		LimitPresent: true,
		IDs:          []int64{-7, 0, 11},
	})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	assert.Equal(t, applicationipam.MaximumVRFPageLimit, repository.lastCriteria.Limit)
	assert.Equal(t, []int64{-7, 0, 11}, repository.lastCriteria.IDs)
	assert.Equal(t, []applicationipam.VRFSort{
		{Field: applicationipam.VRFSortName},
		{Field: applicationipam.VRFSortRD},
	}, repository.lastCriteria.Ordering)
}

func TestListVRFsUsesCompleteVisibilityBeforeCountAndPage(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	seedVRF(t, backend, 1, "First", "1:1")
	seedVRF(t, backend, 2, "Second", "1:2")
	seedVRF(t, backend, 3, "Third", "1:3")
	repository := &memoryRepository{backend: backend}
	authorizer := &scopedAuthorizer{scope: authz.ListScope{
		ObjectIDs:   []int64{2, 3},
		Constrained: true,
	}}
	service, err := applicationipam.NewVRFService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		authorizer,
		fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)

	page, err := service.ListVRFs(t.Context(), testPrincipal(), applicationipam.ListVRFsQuery{
		Limit: 1, Ordering: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	assert.True(t, repository.lastCriteria.VisibilityConstrained)
	assert.False(t, repository.lastCriteria.DeferPagination)
}

func TestListVRFsWithoutCompleteScopeAuthorizesBeforePaging(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	seedVRF(t, backend, 1, "First", "1:1")
	seedVRF(t, backend, 2, "Second", "1:2")
	seedVRF(t, backend, 3, "Third", "1:3")
	repository := &memoryRepository{backend: backend}
	authorizer := &trackingAuthorizer{deniedObjectIDs: map[int64]struct{}{1: {}}}
	service, err := applicationipam.NewVRFService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		authorizer,
		fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)

	page, err := service.ListVRFs(t.Context(), testPrincipal(), applicationipam.ListVRFsQuery{
		Limit: 1, Ordering: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	assert.True(t, repository.lastCriteria.DeferPagination)
}

func newTestService(
	t *testing.T,
	authorizer authz.ResourceAuthorizer,
) (*applicationipam.VRFService, *memoryBackend, *memoryRepository, *memoryRecorder) {
	t.Helper()
	backend := newMemoryBackend()
	repository := &memoryRepository{backend: backend}
	recorder := &memoryRecorder{backend: backend}
	service, err := applicationipam.NewVRFService(
		repository,
		backend,
		recorder,
		authorizer,
		fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)
	return service, backend, repository, recorder
}

func testPrincipal() identity.Principal {
	return identity.Principal{ID: 17, Username: "operator"}
}

func seedVRF(t *testing.T, backend *memoryBackend, id shared.ID, name, rdValue string) {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	vrf, err := domainipam.NewVRF(domainipam.VRFValues{
		Name:          name,
		RD:            domainipam.NonNullRouteDistinguisher(rd),
		EnforceUnique: true,
		Description:   "Original description",
		Comments:      "Original comments",
	}, applicationCreatedAt)
	require.NoError(t, err)
	require.NoError(t, vrf.AssignID(id))
	backend.state.vrfs[id] = vrf.State()
	if backend.state.nextID <= id {
		backend.state.nextID = id + 1
	}
}

func requiredRD(t *testing.T, nullable domainipam.NullableRouteDistinguisher) string {
	t.Helper()
	rd, present := nullable.Get()
	require.True(t, present)
	return rd.String()
}

type fixedClock struct{ now shared.Timestamp }

func (clock fixedClock) Now() shared.Timestamp { return clock.now }

type trackingAuthorizer struct {
	denyAction      authz.Action
	deniedObjectIDs map[int64]struct{}
}

func (authorizer *trackingAuthorizer) AuthorizeResource(
	_ context.Context,
	_ identity.Principal,
	action authz.Action,
	_ authz.ResourceType,
	resource *authz.Object,
) error {
	if action == authorizer.denyAction {
		return shared.NewError(
			shared.ErrorReasonForbidden,
			"You do not have permission to perform this action.",
		)
	}
	if resource != nil {
		if _, denied := authorizer.deniedObjectIDs[resource.ID]; denied {
			return shared.NewError(
				shared.ErrorReasonForbidden,
				"You do not have permission to perform this action.",
			)
		}
	}
	return nil
}

type scopedAuthorizer struct{ scope authz.ListScope }

func (authorizer *scopedAuthorizer) ResourceListScope(
	context.Context,
	identity.Principal,
	authz.Action,
	authz.ResourceType,
) authz.ListScope {
	return authorizer.scope
}

func (authorizer *scopedAuthorizer) AuthorizeResource(
	_ context.Context,
	_ identity.Principal,
	_ authz.Action,
	_ authz.ResourceType,
	resource *authz.Object,
) error {
	if resource == nil || !authorizer.scope.Constrained {
		return nil
	}
	for _, id := range authorizer.scope.ObjectIDs {
		if resource.ID == id {
			return nil
		}
	}
	return shared.NewError(
		shared.ErrorReasonForbidden,
		"You do not have permission to perform this action.",
	)
}

type transactionContextKey struct{}

type memoryState struct {
	nextID  shared.ID
	vrfs    map[shared.ID]domainipam.VRFState
	changes []changelog.Change
}

func (state memoryState) clone() memoryState {
	cloned := memoryState{
		nextID:  state.nextID,
		vrfs:    make(map[shared.ID]domainipam.VRFState, len(state.vrfs)),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, vrf := range state.vrfs {
		cloned.vrfs[id] = vrf
	}
	return cloned
}

type memoryBackend struct {
	state            memoryState
	transactionCalls int
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{state: memoryState{
		nextID: 1,
		vrfs:   make(map[shared.ID]domainipam.VRFState),
	}}
}

func (backend *memoryBackend) WithinTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	backend.transactionCalls++
	working := backend.state.clone()
	transactionContext := context.WithValue(ctx, transactionContextKey{}, &working)
	if err := operation(transactionContext); err != nil {
		return err
	}
	backend.state = working
	return nil
}

func (backend *memoryBackend) stateFor(ctx context.Context) (*memoryState, bool) {
	state, transactional := ctx.Value(transactionContextKey{}).(*memoryState)
	if transactional {
		return state, true
	}
	return &backend.state, false
}

type memoryRepository struct {
	backend                    *memoryBackend
	listCalls                  int
	getCalls                   int
	createCalls                int
	updateCalls                int
	deleteCalls                int
	mutationOutsideTransaction bool
	lastCriteria               *applicationipam.VRFListCriteria
}

func (repository *memoryRepository) List(
	ctx context.Context,
	criteria applicationipam.VRFListCriteria,
) (applicationipam.VRFPage, error) {
	repository.listCalls++
	criteria.IDs = append([]int64(nil), criteria.IDs...)
	criteria.Names = append([]string(nil), criteria.Names...)
	criteria.RDs = append([]domainipam.RouteDistinguisher(nil), criteria.RDs...)
	criteria.Ordering = append([]applicationipam.VRFSort(nil), criteria.Ordering...)
	repository.lastCriteria = &criteria
	state, _ := repository.backend.stateFor(ctx)
	ids := make([]int, 0, len(state.vrfs))
	for id := range state.vrfs {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	visible := make(map[shared.ID]struct{}, len(criteria.VisibleObjectIDs))
	for _, id := range criteria.VisibleObjectIDs {
		visible[id] = struct{}{}
	}
	results := make([]*domainipam.VRF, 0, len(ids))
	for _, primitiveID := range ids {
		id := shared.ID(primitiveID)
		if criteria.VisibilityConstrained {
			if _, allowed := visible[id]; !allowed {
				continue
			}
		}
		vrf, err := domainipam.RestoreVRF(state.vrfs[id])
		if err != nil {
			return applicationipam.VRFPage{}, err
		}
		results = append(results, vrf)
	}
	count := uint64(len(results))
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(results))
		end := min(start+int(criteria.Limit), len(results))
		results = results[start:end]
	}
	return applicationipam.VRFPage{Count: count, Results: results}, nil
}

func (repository *memoryRepository) Get(ctx context.Context, id shared.ID) (*domainipam.VRF, error) {
	repository.getCalls++
	return repository.get(ctx, id)
}

func (repository *memoryRepository) GetForUpdate(ctx context.Context, id shared.ID) (*domainipam.VRF, error) {
	return repository.get(ctx, id)
}

func (repository *memoryRepository) get(ctx context.Context, id shared.ID) (*domainipam.VRF, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, found := state.vrfs[id]
	if !found {
		return nil, shared.NotFound("VRF", id)
	}
	return domainipam.RestoreVRF(persisted)
}

func (repository *memoryRepository) Create(ctx context.Context, vrf *domainipam.VRF) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	id := state.nextID
	state.nextID++
	if err := vrf.AssignID(id); err != nil {
		return err
	}
	state.vrfs[id] = vrf.State()
	return nil
}

func (repository *memoryRepository) Update(ctx context.Context, vrf *domainipam.VRF) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	state.vrfs[vrf.ID()] = vrf.State()
	return nil
}

func (repository *memoryRepository) Delete(ctx context.Context, vrf *domainipam.VRF) error {
	repository.deleteCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	delete(state.vrfs, vrf.ID())
	return nil
}

type memoryRecorder struct {
	backend            *memoryBackend
	err                error
	outsideTransaction bool
}

func (recorder *memoryRecorder) Record(ctx context.Context, change changelog.Change) error {
	state, transactional := recorder.backend.stateFor(ctx)
	recorder.outsideTransaction = recorder.outsideTransaction || !transactional
	if recorder.err != nil {
		return recorder.err
	}
	state.changes = append(state.changes, change)
	return nil
}
