package dcim_test

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
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

var (
	createdAt = shared.NewTimestamp(time.Date(2026, time.July, 16, 8, 0, 0, 0, time.UTC))
	updatedAt = shared.NewTimestamp(time.Date(2026, time.July, 16, 9, 0, 0, 0, time.UTC))
)

func TestNewSiteServiceRequiresEverySecurityAndTransactionDependency(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	_, err := appdcim.NewSiteService(
		&memoryRepository{backend: backend},
		backend,
		&memoryRecorder{backend: backend},
		nil,
		fixedClock{now: updatedAt},
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
	assert.Contains(t, err.Error(), "authorizer")
}

func TestCreateSiteRecordsOneChangeInTheSiteTransaction(t *testing.T) {
	t.Parallel()

	service, backend, repository, recorder, authorizer := newTestService(t)
	site, err := service.CreateSite(context.Background(), testPrincipal(), appdcim.CreateSiteCommand{
		Name:     appdcim.FieldValue("  Moscow DC  "),
		Slug:     appdcim.FieldValue("moscow-dc"),
		Facility: appdcim.FieldValue("  M9  "),
	})
	require.NoError(t, err)

	assert.Equal(t, shared.ID(1), site.ID())
	assert.Equal(t, dcimdomain.SiteStatusActive, site.Status())
	assert.Equal(t, "M9", site.Facility())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.createCalls)
	assert.False(t, repository.mutationOutsideTransaction)
	assert.False(t, recorder.outsideTransaction)
	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	assert.Equal(t, changelog.ActionCreate, change.Action)
	assert.Nil(t, change.Before)
	after, ok := change.After.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	assert.Equal(t, "Moscow DC", after.Name)
	assert.Equal(t, testPrincipal().ID, change.ActorID)
	assert.Equal(t, dcimdomain.SiteObjectType, change.ObjectType)
	require.Len(t, authorizer.calls, 2)
	assert.Nil(t, authorizer.calls[0].resource)
	assert.Nil(t, authorizer.calls[1].resource)
}

func TestDeniedCreateStopsBeforeValidationAndTransaction(t *testing.T) {
	t.Parallel()

	backend := newMemoryBackend()
	repository := &memoryRepository{backend: backend}
	recorder := &memoryRecorder{backend: backend}
	authorizer := &trackingAuthorizer{denyAction: authz.Add}
	service, err := appdcim.NewSiteService(
		repository,
		backend,
		recorder,
		authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	_, err = service.CreateSite(context.Background(), testPrincipal(), appdcim.CreateSiteCommand{})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonForbidden))
	assert.Zero(t, backend.transactionCalls)
	assert.Zero(t, repository.createCalls)
	assert.Empty(t, backend.state.changes)
}

func TestCreateDistinguishesNullFromOmittedAndBlank(t *testing.T) {
	t.Parallel()

	service, backend, _, _, _ := newTestService(t)
	_, err := service.CreateSite(context.Background(), testPrincipal(), appdcim.CreateSiteCommand{
		Name: appdcim.NullField[string](),
		Slug: appdcim.FieldValue("site"),
	})
	require.Error(t, err)
	require.Len(t, shared.ViolationsOf(err), 1)
	assert.Equal(t, "name", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "null", shared.ViolationsOf(err)[0].Reason)
	assert.Equal(t, 1, backend.transactionCalls)

	_, err = service.CreateSite(context.Background(), testPrincipal(), appdcim.CreateSiteCommand{
		Name: appdcim.FieldValue(""),
		Slug: appdcim.FieldValue("site"),
	})
	require.Error(t, err)
	require.Len(t, shared.ViolationsOf(err), 1)
	assert.Equal(t, "required", shared.ViolationsOf(err)[0].Reason)
	assert.Equal(t, 2, backend.transactionCalls)
}

func TestUnauthenticatedGetDoesNotReachPersistence(t *testing.T) {
	t.Parallel()

	service, _, repository, _, authorizer := newTestService(t)
	_, err := service.GetSite(context.Background(), identity.Principal{}, appdcim.GetSiteQuery{ID: 1})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonUnauthenticated))
	assert.Zero(t, repository.getCalls)
	assert.Empty(t, authorizer.calls, "unauthenticated input fails before the authorizer port")
}

func TestUpdateRollsBackSiteWhenChangeRecordingFails(t *testing.T) {
	t.Parallel()

	service, backend, repository, recorder, _ := newTestService(t)
	seedSite(t, backend)
	recorderError := errors.New("forced change failure")
	recorder.err = recorderError
	description := "Changed but rolled back"

	_, err := service.UpdateSite(context.Background(), testPrincipal(), appdcim.UpdateSiteCommand{
		ID:          1,
		Description: appdcim.FieldValue(description),
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, recorderError))
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
	assert.Equal(t, 1, repository.updateCalls, "the failure is injected after the Site write")
	assert.Empty(t, backend.state.changes)

	persisted, restoreErr := dcimdomain.RestoreSite(backend.state.sites[1])
	require.NoError(t, restoreErr)
	assert.Equal(t, "Original description", persisted.Description())
	assert.Equal(t, createdAt, persisted.LastUpdated())
}

func TestUpdatePreservesPresenceAndRecordsBeforeAfter(t *testing.T) {
	t.Parallel()

	service, backend, _, _, _ := newTestService(t)
	seedSite(t, backend)
	site, err := service.UpdateSite(context.Background(), testPrincipal(), appdcim.UpdateSiteCommand{
		ID:       1,
		Facility: appdcim.FieldValue(""),
	})
	require.NoError(t, err)
	assert.Empty(t, site.Facility())
	assert.Equal(t, "Original description", site.Description())
	assert.Equal(t, createdAt, site.Created())
	assert.Equal(t, updatedAt, site.LastUpdated())

	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	assert.Equal(t, changelog.ActionUpdate, change.Action)
	before, ok := change.Before.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	after, ok := change.After.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	assert.Equal(t, "M9", before.Facility)
	assert.Equal(t, "", after.Facility)
}

func TestEmptyUpdateMaskDoesNotWrite(t *testing.T) {
	t.Parallel()

	service, backend, repository, _, _ := newTestService(t)
	seedSite(t, backend)

	_, err := service.UpdateSite(context.Background(), testPrincipal(), appdcim.UpdateSiteCommand{ID: 1})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	assert.Zero(t, repository.updateCalls)
	assert.Empty(t, backend.state.changes)
}

func TestReplaceSitePreservesOmittedState(t *testing.T) {
	t.Parallel()

	service, backend, _, _, _ := newTestService(t)
	seedSite(t, backend)
	state := backend.state.sites[1]
	state.Status = dcimdomain.SiteStatusPlanned.String()
	backend.state.sites[1] = state

	site, err := service.ReplaceSite(context.Background(), testPrincipal(), appdcim.ReplaceSiteCommand{
		ID:   1,
		Name: appdcim.FieldValue("Replacement"),
		Slug: appdcim.FieldValue("replacement"),
	})
	require.NoError(t, err)
	assert.Equal(t, "Replacement", site.Name())
	assert.Equal(t, dcimdomain.SiteStatusPlanned, site.Status())
	assert.Equal(t, "M9", site.Facility())
	assert.Equal(t, "Original description", site.Description())
	assert.Equal(t, "Original comments", site.Comments())
	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	before, ok := change.Before.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	after, ok := change.After.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	assert.Equal(t, dcimdomain.SiteStatusPlanned.String(), before.Status)
	assert.Equal(t, before.Status, after.Status)
	assert.Equal(t, before.Facility, after.Facility)
	assert.Equal(t, before.Description, after.Description)
	assert.Equal(t, before.Comments, after.Comments)
}

func TestSiteScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	t.Run("create presence validation", func(t *testing.T) {
		service, backend, repository, _, _ := newTestService(t)
		_, err := service.CreateSite(t.Context(), testPrincipal(), appdcim.CreateSiteCommand{
			Name: appdcim.FieldValue("Site"), Slug: appdcim.FieldValue("site"),
			Status: appdcim.NullField[string](),
		})
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{{
			Field: "status", Reason: "blank", Description: "This field may not be blank.",
		}}, shared.ViolationsOf(err))
		assert.Empty(t, backend.state.sites)
		assert.Empty(t, backend.state.changes)
		assert.Zero(t, repository.createCalls)
	})

	t.Run("create mixed presence and domain validation", func(t *testing.T) {
		service, backend, repository, _, _ := newTestService(t)
		_, err := service.CreateSite(t.Context(), testPrincipal(), appdcim.CreateSiteCommand{
			Name: appdcim.FieldValue(""),
		})
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field may not be blank."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.Empty(t, backend.state.sites)
		assert.Empty(t, backend.state.changes)
		assert.Zero(t, repository.createCalls)
	})

	t.Run("create command violation wins same-field domain violation", func(t *testing.T) {
		service, backend, repository, _, _ := newTestService(t)
		_, err := service.CreateSite(t.Context(), testPrincipal(), appdcim.CreateSiteCommand{
			Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("site"),
		})
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{{
			Field: "name", Reason: "null", Description: "This field may not be null.",
		}}, shared.ViolationsOf(err))
		assert.Empty(t, backend.state.sites)
		assert.Empty(t, backend.state.changes)
		assert.Zero(t, repository.createCalls)
	})

	for _, test := range []struct {
		name       string
		mutate     func(*appdcim.SiteService) error
		violations []shared.FieldViolation
	}{
		{
			name: "replace presence validation",
			mutate: func(service *appdcim.SiteService) error {
				_, err := service.ReplaceSite(t.Context(), testPrincipal(), appdcim.ReplaceSiteCommand{
					ID: 1, Name: appdcim.FieldValue("Replacement"),
					Slug: appdcim.FieldValue("replacement"), Status: appdcim.NullField[string](),
				})
				return err
			},
			violations: []shared.FieldViolation{{
				Field: "status", Reason: "blank", Description: "This field may not be blank.",
			}},
		},
		{
			name: "update presence validation",
			mutate: func(service *appdcim.SiteService) error {
				_, err := service.UpdateSite(t.Context(), testPrincipal(), appdcim.UpdateSiteCommand{
					ID: 1, Status: appdcim.NullField[string](),
				})
				return err
			},
			violations: []shared.FieldViolation{{
				Field: "status", Reason: "blank", Description: "This field may not be blank.",
			}},
		},
		{
			name: "replace domain invalid and overlength",
			mutate: func(service *appdcim.SiteService) error {
				_, err := service.ReplaceSite(t.Context(), testPrincipal(), appdcim.ReplaceSiteCommand{
					ID:       1,
					Name:     appdcim.FieldValue("Replacement"),
					Slug:     appdcim.FieldValue("not valid!"),
					Facility: appdcim.FieldValue(strings.Repeat("f", dcimdomain.SiteFacilityMaxLength+1)),
				})
				return err
			},
			violations: []shared.FieldViolation{
				{
					Field: "slug", Reason: "invalid",
					Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
				},
				{
					Field: "facility", Reason: "max_length",
					Description: "Ensure this field has no more than the supported number of characters.",
				},
			},
		},
		{
			name: "update domain invalid and overlength",
			mutate: func(service *appdcim.SiteService) error {
				_, err := service.UpdateSite(t.Context(), testPrincipal(), appdcim.UpdateSiteCommand{
					ID:          1,
					Status:      appdcim.FieldValue(" active "),
					Description: appdcim.FieldValue(strings.Repeat("d", dcimdomain.SiteDescriptionMaxLength+1)),
				})
				return err
			},
			violations: []shared.FieldViolation{
				{
					Field: "status", Reason: "invalid_choice",
					Description: " active  is not a valid choice.",
				},
				{
					Field: "description", Reason: "max_length",
					Description: "Ensure this field has no more than the supported number of characters.",
				},
			},
		},
		{
			name: "replace mixed presence and domain validation",
			mutate: func(service *appdcim.SiteService) error {
				_, err := service.ReplaceSite(t.Context(), testPrincipal(), appdcim.ReplaceSiteCommand{
					ID:       1,
					Name:     appdcim.NullField[string](),
					Facility: appdcim.FieldValue(strings.Repeat("f", dcimdomain.SiteFacilityMaxLength+1)),
				})
				return err
			},
			violations: []shared.FieldViolation{
				{Field: "name", Reason: "null", Description: "This field may not be null."},
				{Field: "slug", Reason: "required", Description: "This field is required."},
				{
					Field: "facility", Reason: "max_length",
					Description: "Ensure this field has no more than the supported number of characters.",
				},
			},
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			service, backend, repository, _, _ := newTestService(t)
			seedSite(t, backend)
			before := backend.state.sites[1]

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, test.violations, shared.ViolationsOf(err))
			assert.Equal(t, before, backend.state.sites[1])
			assert.Equal(t, before.Created, backend.state.sites[1].Created)
			assert.Equal(t, before.LastUpdated, backend.state.sites[1].LastUpdated)
			assert.Empty(t, backend.state.changes)
			assert.Zero(t, repository.updateCalls)
		})
	}

	t.Run("internal clock failure wins presence validation", func(t *testing.T) {
		backend := newMemoryBackend()
		repository := &memoryRepository{backend: backend}
		service, err := appdcim.NewSiteService(
			repository,
			backend,
			&memoryRecorder{backend: backend},
			&trackingAuthorizer{},
			fixedClock{},
		)
		require.NoError(t, err)
		seedSite(t, backend)
		before := backend.state.sites[1]

		_, err = service.UpdateSite(t.Context(), testPrincipal(), appdcim.UpdateSiteCommand{
			ID: 1, Status: appdcim.NullField[string](),
		})
		require.Error(t, err)
		assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
		assert.Equal(t, before, backend.state.sites[1])
		assert.Zero(t, repository.updateCalls)
		assert.Empty(t, backend.state.changes)
	})
}

func TestDeleteRecordsThePrechangeSnapshot(t *testing.T) {
	t.Parallel()

	service, backend, _, _, _ := newTestService(t)
	seedSite(t, backend)

	err := service.DeleteSite(context.Background(), testPrincipal(), appdcim.DeleteSiteCommand{ID: 1})
	require.NoError(t, err)
	assert.Empty(t, backend.state.sites)
	require.Len(t, backend.state.changes, 1)
	change := backend.state.changes[0]
	assert.Equal(t, changelog.ActionDelete, change.Action)
	before, ok := change.Before.(dcimdomain.SiteSnapshot)
	require.True(t, ok)
	assert.Equal(t, "Original", before.Name)
	assert.Nil(t, change.After)
}

func TestListValidatesAndNormalizesRepositoryCriteria(t *testing.T) {
	t.Parallel()

	service, _, repository, _, _ := newTestService(t)
	_, err := service.ListSites(context.Background(), testPrincipal(), appdcim.ListSitesQuery{
		Ordering: []string{"-status,name"},
		Names:    []string{"  Moscow DC  ", " St Petersburg "},
		Slugs:    []string{"  moscow-dc  ", " spb "},
		Statuses: []string{" active ", "planned"},
	})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	criteria := *repository.lastCriteria
	assert.Equal(t, appdcim.DefaultSitePageLimit, criteria.Limit)
	assert.Equal(t, []string{"Moscow DC", "St Petersburg"}, criteria.Names)
	assert.Equal(t, []string{"moscow-dc", "spb"}, criteria.Slugs)
	assert.Equal(t, []dcimdomain.SiteStatus{
		dcimdomain.SiteStatusActive,
		dcimdomain.SiteStatusPlanned,
	}, criteria.Statuses)
	assert.Equal(t, []appdcim.SiteSort{
		{Field: appdcim.SiteSortStatus, Descending: true},
		{Field: appdcim.SiteSortName},
	}, criteria.Ordering)

	_, err = service.ListSites(context.Background(), testPrincipal(), appdcim.ListSitesQuery{
		Limit:    appdcim.MaximumSitePageLimit + 1,
		Ordering: []string{"unknown"},
		Statuses: []string{"not-a-status"},
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	assert.Equal(t, 1, repository.listCalls, "invalid criteria must not reach persistence")
	assert.Len(t, shared.ViolationsOf(err), 3)
}

func TestListUsesPinnedPaginationLimitsAndPreservesSignedRepeatedFilters(t *testing.T) {
	t.Parallel()

	service, _, repository, _, _ := newTestService(t)

	_, err := service.ListSites(context.Background(), testPrincipal(), appdcim.ListSitesQuery{})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	assert.Equal(t, appdcim.DefaultSitePageLimit, repository.lastCriteria.Limit)

	_, err = service.ListSites(context.Background(), testPrincipal(), appdcim.ListSitesQuery{
		LimitPresent: true,
	})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	assert.Equal(t, appdcim.MaximumSitePageLimit, repository.lastCriteria.Limit)

	_, err = service.ListSites(context.Background(), testPrincipal(), appdcim.ListSitesQuery{
		IDs:      []int64{-7, 0, 42},
		Names:    []string{" Alpha ", " Beta "},
		Slugs:    []string{" alpha ", " beta "},
		Statuses: []string{" active ", " planned "},
	})
	require.NoError(t, err)
	require.NotNil(t, repository.lastCriteria)
	assert.Equal(t, []int64{-7, 0, 42}, repository.lastCriteria.IDs)
	assert.Equal(t, []string{"Alpha", "Beta"}, repository.lastCriteria.Names)
	assert.Equal(t, []string{"alpha", "beta"}, repository.lastCriteria.Slugs)
	assert.Equal(t, []dcimdomain.SiteStatus{
		dcimdomain.SiteStatusActive,
		dcimdomain.SiteStatusPlanned,
	}, repository.lastCriteria.Statuses)
}

func TestListAppliesCompleteVisibilityScopeBeforeCountAndPage(t *testing.T) {
	backend := newMemoryBackend()
	seedNamedSite(t, backend, 1, "First")
	seedNamedSite(t, backend, 2, "Second")
	seedNamedSite(t, backend, 3, "Third")
	repository := &memoryRepository{backend: backend}
	authorizer := &scopedTrackingAuthorizer{
		trackingAuthorizer: trackingAuthorizer{},
		scope: authz.ListScope{
			ObjectIDs:   []int64{2, 3},
			Constrained: true,
		},
	}
	service, err := appdcim.NewSiteService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	page, err := service.ListSites(t.Context(), testPrincipal(), appdcim.ListSitesQuery{
		Limit:    1,
		Ordering: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	require.NotNil(t, repository.lastCriteria)
	assert.True(t, repository.lastCriteria.VisibilityConstrained)
	assert.False(t, repository.lastCriteria.DeferPagination)
	assert.Equal(t, []shared.ID{2, 3}, repository.lastCriteria.VisibleObjectIDs)
	require.Len(t, authorizer.calls, 2)
	assert.Nil(t, authorizer.calls[0].resource)
	assert.Equal(t, int64(2), authorizer.calls[1].resource.ID)
}

func TestListRejectsRepositoryRowsOutsideCompleteVisibilityScope(t *testing.T) {
	backend := newMemoryBackend()
	seedNamedSite(t, backend, 1, "First")
	seedNamedSite(t, backend, 2, "Second")
	repository := &memoryRepository{backend: backend, ignoreVisibility: true}
	authorizer := &scopedTrackingAuthorizer{
		scope: authz.ListScope{ObjectIDs: []int64{2}, Constrained: true},
	}
	service, err := appdcim.NewSiteService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	_, err = service.ListSites(t.Context(), testPrincipal(), appdcim.ListSitesQuery{
		Limit:    1,
		Ordering: []string{"id"},
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func TestListWithoutCompleteScopeDefersPageUntilAfterObjectAuthorization(t *testing.T) {
	backend := newMemoryBackend()
	seedNamedSite(t, backend, 1, "First")
	seedNamedSite(t, backend, 2, "Second")
	seedNamedSite(t, backend, 3, "Third")
	repository := &memoryRepository{backend: backend}
	authorizer := &trackingAuthorizer{denyObjectIDs: map[int64]struct{}{1: {}}}
	service, err := appdcim.NewSiteService(
		repository,
		backend,
		&memoryRecorder{backend: backend},
		authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	page, err := service.ListSites(t.Context(), testPrincipal(), appdcim.ListSitesQuery{
		Limit:    1,
		Ordering: []string{"id"},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, shared.ID(2), page.Results[0].ID())
	require.NotNil(t, repository.lastCriteria)
	assert.True(t, repository.lastCriteria.DeferPagination)
}

func newTestService(
	t *testing.T,
) (*appdcim.SiteService, *memoryBackend, *memoryRepository, *memoryRecorder, *trackingAuthorizer) {
	t.Helper()
	backend := newMemoryBackend()
	repository := &memoryRepository{backend: backend}
	recorder := &memoryRecorder{backend: backend}
	authorizer := &trackingAuthorizer{}
	service, err := appdcim.NewSiteService(
		repository,
		backend,
		recorder,
		authorizer,
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service, backend, repository, recorder, authorizer
}

func testPrincipal() identity.Principal {
	return identity.Principal{ID: 17, Username: "operator"}
}

func seedSite(t *testing.T, backend *memoryBackend) {
	t.Helper()
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name:        "Original",
		Slug:        "original",
		Status:      dcimdomain.SiteStatusActive.String(),
		Facility:    "M9",
		Description: "Original description",
		Comments:    "Original comments",
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, site.AssignID(1))
	backend.state.sites[1] = site.State()
	backend.state.nextID = 2
}

func seedNamedSite(t *testing.T, backend *memoryBackend, id shared.ID, name string) {
	t.Helper()
	site, err := dcimdomain.NewSite(dcimdomain.SiteValues{
		Name:   name,
		Slug:   strings.ToLower(name),
		Status: dcimdomain.SiteStatusActive.String(),
	}, createdAt)
	require.NoError(t, err)
	require.NoError(t, site.AssignID(id))
	backend.state.sites[id] = site.State()
	if backend.state.nextID <= id {
		backend.state.nextID = id + 1
	}
}

type fixedClock struct {
	now shared.Timestamp
}

func (clock fixedClock) Now() shared.Timestamp { return clock.now }

type authorizationCall struct {
	action   authz.Action
	resource *authz.Object
}

type trackingAuthorizer struct {
	denyAction    authz.Action
	denyObjectIDs map[int64]struct{}
	calls         []authorizationCall
}

func (authorizer *trackingAuthorizer) AuthorizeResource(
	_ context.Context,
	_ identity.Principal,
	action authz.Action,
	_ authz.ResourceType,
	resource *authz.Object,
) error {
	authorizer.calls = append(authorizer.calls, authorizationCall{action: action, resource: resource})
	if action == authorizer.denyAction {
		return shared.NewError(
			shared.ErrorReasonForbidden,
			"You do not have permission to perform this action.",
		)
	}
	if resource != nil {
		if _, denied := authorizer.denyObjectIDs[resource.ID]; denied {
			return shared.NewError(
				shared.ErrorReasonForbidden,
				"You do not have permission to perform this action.",
			)
		}
	}
	return nil
}

type scopedTrackingAuthorizer struct {
	trackingAuthorizer
	scope authz.ListScope
}

func (authorizer *scopedTrackingAuthorizer) ResourceListScope(
	context.Context,
	identity.Principal,
	authz.Action,
	authz.ResourceType,
) authz.ListScope {
	return authorizer.scope
}

func (authorizer *scopedTrackingAuthorizer) AuthorizeResource(
	ctx context.Context,
	principal identity.Principal,
	action authz.Action,
	resourceType authz.ResourceType,
	resource *authz.Object,
) error {
	if err := authorizer.trackingAuthorizer.AuthorizeResource(
		ctx,
		principal,
		action,
		resourceType,
		resource,
	); err != nil {
		return err
	}
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
	sites   map[shared.ID]dcimdomain.SiteState
	changes []changelog.Change
}

func (state memoryState) clone() memoryState {
	clone := memoryState{
		nextID:  state.nextID,
		sites:   make(map[shared.ID]dcimdomain.SiteState, len(state.sites)),
		changes: append([]changelog.Change(nil), state.changes...),
	}
	for id, site := range state.sites {
		clone.sites[id] = site
	}
	return clone
}

type memoryBackend struct {
	state            memoryState
	transactionCalls int
}

func newMemoryBackend() *memoryBackend {
	return &memoryBackend{state: memoryState{nextID: 1, sites: make(map[shared.ID]dcimdomain.SiteState)}}
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
	lastCriteria               *appdcim.SiteListCriteria
	ignoreVisibility           bool
}

func (repository *memoryRepository) List(
	ctx context.Context,
	criteria appdcim.SiteListCriteria,
) (appdcim.SitePage, error) {
	repository.listCalls++
	criteria.IDs = append([]int64(nil), criteria.IDs...)
	criteria.Names = append([]string(nil), criteria.Names...)
	criteria.Slugs = append([]string(nil), criteria.Slugs...)
	criteria.Statuses = append([]dcimdomain.SiteStatus(nil), criteria.Statuses...)
	criteria.Ordering = append([]appdcim.SiteSort(nil), criteria.Ordering...)
	repository.lastCriteria = &criteria
	state, _ := repository.backend.stateFor(ctx)
	ids := make([]int, 0, len(state.sites))
	for id := range state.sites {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	results := make([]*dcimdomain.Site, 0, len(ids))
	visible := make(map[shared.ID]struct{}, len(criteria.VisibleObjectIDs))
	for _, id := range criteria.VisibleObjectIDs {
		visible[id] = struct{}{}
	}
	for _, primitiveID := range ids {
		id := shared.ID(primitiveID)
		if criteria.VisibilityConstrained && !repository.ignoreVisibility {
			if _, allowed := visible[id]; !allowed {
				continue
			}
		}
		site, err := dcimdomain.RestoreSite(state.sites[id])
		if err != nil {
			return appdcim.SitePage{}, err
		}
		results = append(results, site)
	}
	count := uint64(len(results))
	if !criteria.DeferPagination {
		start := min(int(criteria.Offset), len(results))
		end := min(start+int(criteria.Limit), len(results))
		results = results[start:end]
	}
	return appdcim.SitePage{Count: count, Results: results}, nil
}

func (repository *memoryRepository) Get(ctx context.Context, id shared.ID) (*dcimdomain.Site, error) {
	repository.getCalls++
	return repository.getFrom(ctx, id)
}

func (repository *memoryRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.Site, error) {
	return repository.getFrom(ctx, id)
}

func (repository *memoryRepository) getFrom(ctx context.Context, id shared.ID) (*dcimdomain.Site, error) {
	state, _ := repository.backend.stateFor(ctx)
	persisted, found := state.sites[id]
	if !found {
		return nil, shared.NotFound("Site", id)
	}
	return dcimdomain.RestoreSite(persisted)
}

func (repository *memoryRepository) Create(ctx context.Context, site *dcimdomain.Site) error {
	repository.createCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	id := state.nextID
	state.nextID++
	if err := site.AssignID(id); err != nil {
		return err
	}
	state.sites[id] = site.State()
	return nil
}

func (repository *memoryRepository) Update(ctx context.Context, site *dcimdomain.Site) error {
	repository.updateCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	state.sites[site.ID()] = site.State()
	return nil
}

func (repository *memoryRepository) Delete(ctx context.Context, site *dcimdomain.Site) error {
	repository.deleteCalls++
	state, transactional := repository.backend.stateFor(ctx)
	repository.mutationOutsideTransaction = repository.mutationOutsideTransaction || !transactional
	delete(state.sites, site.ID())
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
