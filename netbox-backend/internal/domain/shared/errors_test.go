package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/shared"
)

func TestAdapterInputErrorsUseCanonicalTaxonomy(t *testing.T) {
	t.Parallel()

	invalid := shared.Invalid("limit", "A valid integer is required.")
	assert.Equal(t, shared.ErrorReasonValidation, shared.ReasonOf(invalid))
	require.Equal(t, []shared.FieldViolation{{
		Field:       "limit",
		Reason:      "invalid",
		Description: "A valid integer is required.",
	}}, shared.ViolationsOf(invalid))

	unauthenticated := shared.Unauthenticated()
	assert.Equal(
		t,
		shared.ErrorReasonUnauthenticated,
		shared.ReasonOf(unauthenticated),
	)
	assert.Equal(
		t,
		"Authentication credentials were not provided.",
		unauthenticated.Error(),
	)
}
