package dcim_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestReplaceRackRolePreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend := newOrganizationBackend()
	seedRackRole(t, backend, 1, "Original Role")
	state := backend.state.rackRoles[1]
	state.RackCount = 7
	backend.state.rackRoles[1] = state
	repository := newTrackingRackRoleRepository(backend)
	service, err := appdcim.NewRackRoleService(
		repository,
		backend,
		&organizationRecorder{backend: backend},
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	before := backend.state.rackRoles[1]

	role, err := service.ReplaceRackRole(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceRackRoleCommand{
			ID: 1, Name: appdcim.FieldValue("  Replacement Role  "),
			Slug: appdcim.FieldValue("  replacement-role  "),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Replacement Role", role.Name())
	assert.Equal(t, "replacement-role", role.Slug().String())
	assert.Equal(t, before.Color, role.Color().String())
	assert.Equal(t, before.Description, role.Description())
	assert.Equal(t, before.Created, role.Created())
	assert.Equal(t, updatedAt, role.LastUpdated())
	assert.Equal(t, before.RackCount, role.RackCount())
	assert.Equal(t, 1, backend.transactionCalls)
	assert.Equal(t, 1, repository.getForUpdateCalls)
	assert.Equal(t, 1, repository.updateCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, dcimdomain.RackRoleSnapshot{
		Name: before.Name, Slug: before.Slug, Color: before.Color, Description: before.Description,
	}, backend.state.changes[0].Before)
	assert.Equal(t, role.Snapshot(), backend.state.changes[0].After)
}

func TestRackRoleScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	invalidDescription := strings.Repeat("d", dcimdomain.RackRoleDescriptionMaxLength+1)
	expectedViolations := []shared.FieldViolation{
		{Field: "name", Reason: "null", Description: "This field may not be null."},
		{
			Field: "slug", Reason: "invalid",
			Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
		},
		{
			Field: "color", Reason: "invalid",
			Description: "Enter a valid hexadecimal RGB color code.",
		},
		{
			Field: "description", Reason: "max_length",
			Description: "Ensure this field has no more than the supported number of characters.",
		},
	}

	for _, test := range []struct {
		name            string
		seed            bool
		wantLockedReads int
		mutate          func(*appdcim.RackRoleService) error
	}{
		{
			name: "POST",
			mutate: func(service *appdcim.RackRoleService) error {
				_, err := service.CreateRackRole(t.Context(), testPrincipal(), appdcim.CreateRackRoleCommand{
					Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug!"),
					Color:       appdcim.FieldValue("ABCDEF"),
					Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
		{
			name: "PUT", seed: true, wantLockedReads: 1,
			mutate: func(service *appdcim.RackRoleService) error {
				_, err := service.ReplaceRackRole(t.Context(), testPrincipal(), appdcim.ReplaceRackRoleCommand{
					ID: 1, Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug!"),
					Color:       appdcim.FieldValue("ABCDEF"),
					Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
		{
			name: "PATCH", seed: true, wantLockedReads: 1,
			mutate: func(service *appdcim.RackRoleService) error {
				_, err := service.UpdateRackRole(t.Context(), testPrincipal(), appdcim.UpdateRackRoleCommand{
					ID: 1, Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug!"),
					Color:       appdcim.FieldValue("ABCDEF"),
					Description: appdcim.FieldValue(invalidDescription),
				})
				return err
			},
		},
	} {
		test := test
		t.Run(test.name+" validation", func(t *testing.T) {
			backend := newOrganizationBackend()
			if test.seed {
				seedRackRoleWithCount(t, backend, 1, 7)
			}
			repository := newTrackingRackRoleRepository(backend)
			service := newTrackedRackRoleService(t, backend, repository, &organizationRecorder{backend: backend})
			before := backend.state.clone()

			err := test.mutate(service)
			require.Error(t, err)
			assert.Equal(t, expectedViolations, shared.ViolationsOf(err))
			assert.Equal(t, before, backend.state)
			assert.Equal(t, 1, backend.transactionCalls)
			assert.Equal(t, test.wantLockedReads, repository.getForUpdateCalls)
			assert.Zero(t, repository.createCalls)
			assert.Zero(t, repository.updateCalls)
			assert.Empty(t, backend.state.changes)
		})
	}

	t.Run("change recorder failure rolls back an attempted update", func(t *testing.T) {
		backend := newOrganizationBackend()
		seedRackRoleWithCount(t, backend, 1, 7)
		repository := newTrackingRackRoleRepository(backend)
		recorderFailure := errors.New("forced RackRole change recording failure")
		service := newTrackedRackRoleService(
			t, backend, repository, &organizationRecorder{backend: backend, err: recorderFailure},
		)
		before := backend.state.clone()

		_, err := service.UpdateRackRole(t.Context(), testPrincipal(), appdcim.UpdateRackRoleCommand{
			ID: 1, Description: appdcim.FieldValue("changed but rolled back"),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, recorderFailure)
		assert.Equal(t, 1, backend.transactionCalls)
		assert.Equal(t, 1, repository.getForUpdateCalls)
		assert.Equal(t, 1, repository.updateCalls, "failure is injected after the repository write")
		assert.Equal(t, before, backend.state)
		assert.Equal(t, before.rackRoles[1].Created, backend.state.rackRoles[1].Created)
		assert.Equal(t, before.rackRoles[1].LastUpdated, backend.state.rackRoles[1].LastUpdated)
		assert.Equal(t, before.rackRoles[1].RackCount, backend.state.rackRoles[1].RackCount)
		assert.Empty(t, backend.state.changes)
	})
}

type trackingRackRoleRepository struct {
	*rackRoleMemoryRepository
	createCalls       int
	updateCalls       int
	getForUpdateCalls int
}

func newTrackingRackRoleRepository(backend *organizationBackend) *trackingRackRoleRepository {
	return &trackingRackRoleRepository{
		rackRoleMemoryRepository: &rackRoleMemoryRepository{backend: backend},
	}
}

func (repository *trackingRackRoleRepository) GetForUpdate(
	ctx context.Context,
	id shared.ID,
) (*dcimdomain.RackRole, error) {
	repository.getForUpdateCalls++
	return repository.rackRoleMemoryRepository.GetForUpdate(ctx, id)
}

func (repository *trackingRackRoleRepository) Create(
	ctx context.Context,
	role *dcimdomain.RackRole,
) error {
	repository.createCalls++
	return repository.rackRoleMemoryRepository.Create(ctx, role)
}

func (repository *trackingRackRoleRepository) Update(
	ctx context.Context,
	role *dcimdomain.RackRole,
) error {
	repository.updateCalls++
	return repository.rackRoleMemoryRepository.Update(ctx, role)
}

func newTrackedRackRoleService(
	t *testing.T,
	backend *organizationBackend,
	repository *trackingRackRoleRepository,
	recorder *organizationRecorder,
) *appdcim.RackRoleService {
	t.Helper()
	service, err := appdcim.NewRackRoleService(
		repository, backend, recorder, authz.AllowAll{}, fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	return service
}

func seedRackRoleWithCount(
	t *testing.T,
	backend *organizationBackend,
	id shared.ID,
	rackCount uint64,
) {
	t.Helper()
	seedRackRole(t, backend, id, "Original Role")
	state := backend.state.rackRoles[id]
	state.RackCount = rackCount
	backend.state.rackRoles[id] = state
}
