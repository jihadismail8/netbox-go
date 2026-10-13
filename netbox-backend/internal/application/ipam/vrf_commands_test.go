package ipam

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/shared"
)

func TestVRFScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	defaults, err := (CreateVRFCommand{Name: FieldValue("valid")}).values()
	require.NoError(t, err)
	assert.True(t, defaults.RD.IsNull())
	assert.True(t, defaults.EnforceUnique)
	assert.Empty(t, defaults.Description)
	assert.Empty(t, defaults.Comments)
	_, err = (CreateVRFCommand{}).values()
	assert.Equal(t, []shared.FieldViolation{{Field: "name", Reason: "required", Description: "This field is required."}}, shared.ViolationsOf(err))
	_, err = (ReplaceVRFCommand{ID: 1}).patch()
	assert.Equal(t, []shared.FieldViolation{{Field: "name", Reason: "required", Description: "This field is required."}}, shared.ViolationsOf(err))

	optionalPUT, err := (ReplaceVRFCommand{ID: 1, Name: FieldValue("valid")}).patch()
	require.NoError(t, err)
	require.NotNil(t, optionalPUT.Name)
	assert.Nil(t, optionalPUT.RD)
	assert.Nil(t, optionalPUT.EnforceUnique)
	assert.Nil(t, optionalPUT.Description)
	assert.Nil(t, optionalPUT.Comments)

	for _, rd := range []Field[string]{NullField[string](), FieldValue(""), FieldValue("arbitrary text")} {
		values, err := (CreateVRFCommand{Name: FieldValue("valid"), RD: rd, EnforceUnique: FieldValue(false)}).values()
		require.NoError(t, err)
		assert.False(t, values.EnforceUnique)
		assert.Equal(t, rd.State() == FieldNull, values.RD.IsNull())
		patch, err := (UpdateVRFCommand{ID: 1, RD: rd, EnforceUnique: FieldValue(false)}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.RD)
		assert.Equal(t, values.RD, *patch.RD)
		require.NotNil(t, patch.EnforceUnique)
		assert.False(t, *patch.EnforceUnique)
		assert.Nil(t, patch.Name)
		assert.Nil(t, patch.Description)
		assert.Nil(t, patch.Comments)
		put, err := (ReplaceVRFCommand{ID: 1, Name: FieldValue("valid"), RD: rd, EnforceUnique: FieldValue(false)}).patch()
		require.NoError(t, err)
		assert.Equal(t, patch.RD, put.RD)
		assert.Equal(t, patch.EnforceUnique, put.EnforceUnique)
	}

	for _, field := range []string{"name", "enforce_unique", "description", "comments"} {
		t.Run(field+" null", func(t *testing.T) {
			create := CreateVRFCommand{Name: FieldValue("valid"), Comments: FieldValue("retained concrete")}
			switch field {
			case "name":
				create.Name = NullField[string]()
			case "enforce_unique":
				create.EnforceUnique = NullField[bool]()
			case "description":
				create.Description = NullField[string]()
			case "comments":
				create.Comments = NullField[string]()
			}
			values, createErr := create.values()
			put, putErr := (ReplaceVRFCommand{ID: 1, Name: create.Name, RD: create.RD,
				EnforceUnique: create.EnforceUnique, Description: create.Description, Comments: create.Comments}).patch()
			patch, patchErr := (UpdateVRFCommand{ID: 1, Name: create.Name, RD: create.RD,
				EnforceUnique: create.EnforceUnique, Description: create.Description, Comments: create.Comments}).patch()
			for _, err := range []error{createErr, putErr, patchErr} {
				require.Error(t, err)
				assert.Equal(t, []shared.FieldViolation{{Field: field, Reason: "null", Description: "This field may not be null."}}, shared.ViolationsOf(err))
			}
			if field != "comments" {
				assert.Equal(t, "retained concrete", values.Comments)
				require.NotNil(t, put.Comments)
				assert.Equal(t, "retained concrete", *put.Comments)
				assert.Equal(t, put.Comments, patch.Comments)
			}
		})
	}
	blankTexts, err := (ReplaceVRFCommand{ID: 1, Name: FieldValue("valid"),
		Description: FieldValue(""), Comments: FieldValue("")}).patch()
	require.NoError(t, err)
	require.NotNil(t, blankTexts.Description)
	require.NotNil(t, blankTexts.Comments)
	assert.Empty(t, *blankTexts.Description)
	assert.Empty(t, *blankTexts.Comments)

	overlengthNUL := FieldValue(strings.Repeat("界", 21) + "\x00")
	want := []shared.FieldViolation{
		{Field: "rd", Reason: "max_length", Description: "Ensure this field has no more than 21 characters."},
		{Field: "rd", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
	}
	_, err = (CreateVRFCommand{Name: FieldValue("valid"), RD: overlengthNUL}).values()
	require.Error(t, err)
	assert.Equal(t, want, shared.ViolationsOf(err))
	_, err = (UpdateVRFCommand{ID: 1, RD: overlengthNUL}).patch()
	require.Error(t, err)
	assert.Equal(t, want, shared.ViolationsOf(err))
	_, err = (ReplaceVRFCommand{ID: 1, Name: FieldValue("valid"), RD: overlengthNUL}).patch()
	require.Error(t, err)
	assert.Equal(t, want, shared.ViolationsOf(err))

	merged := mergeVRFMutationErrors(
		shared.NewValidationError(shared.FieldViolation{Field: "name", Reason: "null", Description: "This field may not be null."}, want[0], want[1]),
		shared.NewValidationError(shared.FieldViolation{Field: "name", Reason: "blank", Description: "This field may not be blank."}, want[1]),
	)
	assert.Equal(t, append([]shared.FieldViolation{{Field: "name", Reason: "null", Description: "This field may not be null."}}, want...), shared.ViolationsOf(merged))
	internal := shared.NewError(shared.ErrorReasonInternal, "clock failed")
	assert.Same(t, internal, mergeVRFMutationErrors(merged, internal))
}
