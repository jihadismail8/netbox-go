package dcim_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/application/authz"
	appdcim "netbox-go/internal/application/dcim"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestReplaceManufacturerPreservesOmittedState(t *testing.T) {
	t.Parallel()

	backend := newOrganizationBackend()
	seedManufacturer(t, backend, 1, "Original")
	repository := &manufacturerMemoryRepository{backend: backend}
	recorder := &organizationRecorder{backend: backend}
	service, err := appdcim.NewManufacturerService(
		repository,
		backend,
		recorder,
		&trackingAuthorizer{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)
	before := backend.state.manufacturers[1]

	manufacturer, err := service.ReplaceManufacturer(
		t.Context(),
		testPrincipal(),
		appdcim.ReplaceManufacturerCommand{
			ID: 1, Name: appdcim.FieldValue("  Replacement  "),
			Slug: appdcim.FieldValue("  replacement  "),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "Replacement", manufacturer.Name())
	assert.Equal(t, "replacement", manufacturer.Slug().String())
	assert.Equal(t, before.Description, manufacturer.Description())
	assert.Equal(t, before.Created, manufacturer.Created())
	assert.Equal(t, updatedAt, manufacturer.LastUpdated())
	assert.Equal(t, 1, repository.updateCalls)
	require.Len(t, backend.state.changes, 1)
	assert.Equal(t, dcimdomain.ManufacturerSnapshot{
		Name: before.Name, Slug: before.Slug, Description: before.Description,
	}, backend.state.changes[0].Before)
	assert.Equal(t, manufacturer.Snapshot(), backend.state.changes[0].After)
}

func TestManufacturerScalarValidationLeavesStateUnchanged(t *testing.T) {
	t.Parallel()

	backend := newOrganizationBackend()
	seedManufacturer(t, backend, 1, "Original")
	repository := &manufacturerMemoryRepository{backend: backend}
	service, err := appdcim.NewManufacturerService(
		repository,
		backend,
		&organizationRecorder{backend: backend},
		authz.AllowAll{},
		fixedClock{now: updatedAt},
	)
	require.NoError(t, err)

	invalidDescription := strings.Repeat("d", dcimdomain.ManufacturerDescriptionMaxLength+1)
	tests := []struct {
		name   string
		mutate func() error
	}{
		{
			name: "POST",
			mutate: func() error {
				_, createErr := service.CreateManufacturer(
					t.Context(), testPrincipal(), appdcim.CreateManufacturerCommand{
						Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug"),
						Description: appdcim.FieldValue(invalidDescription),
					},
				)
				return createErr
			},
		},
		{
			name: "PUT",
			mutate: func() error {
				_, replaceErr := service.ReplaceManufacturer(
					t.Context(), testPrincipal(), appdcim.ReplaceManufacturerCommand{
						ID: 1, Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug"),
						Description: appdcim.FieldValue(invalidDescription),
					},
				)
				return replaceErr
			},
		},
		{
			name: "PATCH",
			mutate: func() error {
				_, updateErr := service.UpdateManufacturer(
					t.Context(), testPrincipal(), appdcim.UpdateManufacturerCommand{
						ID: 1, Name: appdcim.NullField[string](), Slug: appdcim.FieldValue("invalid slug"),
						Description: appdcim.FieldValue(invalidDescription),
					},
				)
				return updateErr
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			before := backend.state.clone()
			updatesBefore := repository.updateCalls
			err := test.mutate()
			require.Error(t, err)
			assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
			violations := shared.ViolationsOf(err)
			require.Len(t, violations, 3)
			assert.Equal(t, []string{"name", "slug", "description"}, []string{
				violations[0].Field, violations[1].Field, violations[2].Field,
			})
			assert.Equal(t, []string{"null", "invalid", "max_length"}, []string{
				violations[0].Reason, violations[1].Reason, violations[2].Reason,
			})
			assert.Equal(t, before, backend.state)
			assert.Equal(t, updatesBefore, repository.updateCalls)
		})
	}
}
