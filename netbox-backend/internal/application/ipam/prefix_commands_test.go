package ipam

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/shared"
)

func TestPrefixScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	defaults, err := (CreatePrefixCommand{Prefix: FieldValue("192.0.2.0/24")}).values()
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.0/24", defaults.prefix)
	assert.Nil(t, defaults.vrfID)
	assert.Equal(t, "active", defaults.status)
	assert.False(t, defaults.isPool)
	assert.False(t, defaults.markUtilized)
	assert.Empty(t, defaults.description)
	assert.Empty(t, defaults.comments)

	for _, operation := range []string{"POST", "PUT"} {
		t.Run(operation+" required Prefix", func(t *testing.T) {
			command := CreatePrefixCommand{Comments: FieldValue("retained")}
			var values prefixCommandValues
			if operation == "POST" {
				values, err = command.values()
			} else {
				values, err = (ReplacePrefixCommand{ID: 3, CreatePrefixCommand: command}).values()
			}
			assert.Equal(t, []shared.FieldViolation{{
				Field: "prefix", Reason: "required", Description: "This field is required.",
			}}, shared.ViolationsOf(err))
			assert.Equal(t, "retained", values.comments, "presence failures must retain independent input")
		})
	}

	for _, id := range []int64{0, -9, 7} {
		command := CreatePrefixCommand{
			Prefix: FieldValue("192.0.2.0/24"), VRF: FieldValue(id), Status: FieldValue("reserved"),
			IsPool: FieldValue(false), MarkUtilized: FieldValue(false),
			Description: FieldValue(""), Comments: FieldValue(""),
		}
		for _, operation := range []string{"POST", "PUT"} {
			var values prefixCommandValues
			if operation == "POST" {
				values, err = command.values()
			} else {
				values, err = (ReplacePrefixCommand{ID: 3, CreatePrefixCommand: command}).values()
			}
			require.NoError(t, err)
			require.NotNil(t, values.vrfID)
			assert.Equal(t, id, values.vrfID.Int64(), operation)
			assert.Equal(t, "reserved", values.status)
			assert.False(t, values.isPool)
			assert.False(t, values.markUtilized)
			assert.Empty(t, values.description)
			assert.Empty(t, values.comments)
		}
		patch, err := prefixPresenceUpdate(command).patch()
		require.NoError(t, err)
		assert.True(t, patch.vrfSet)
		require.NotNil(t, patch.vrfID)
		assert.Equal(t, id, patch.vrfID.Int64())
		require.NotNil(t, patch.isPool)
		require.NotNil(t, patch.markUtilized)
		assert.False(t, *patch.isPool)
		assert.False(t, *patch.markUtilized)
		require.NotNil(t, patch.description)
		require.NotNil(t, patch.comments)
		assert.Empty(t, *patch.description)
		assert.Empty(t, *patch.comments)
	}

	for _, operation := range []string{"POST", "PUT", "PATCH"} {
		for _, field := range []string{"prefix", "vrf", "status", "is_pool", "mark_utilized", "description", "comments"} {
			t.Run(operation+"/"+field+" null", func(t *testing.T) {
				command := CreatePrefixCommand{
					Prefix: FieldValue("192.0.2.0/24"), Status: FieldValue("reserved"),
					IsPool: FieldValue(false), MarkUtilized: FieldValue(false),
					Description: FieldValue("retained description"), Comments: FieldValue("retained comments"),
				}
				switch field {
				case "prefix":
					command.Prefix = NullField[string]()
				case "vrf":
					command.VRF = NullField[int64]()
				case "status":
					command.Status = NullField[string]()
				case "is_pool":
					command.IsPool = NullField[bool]()
				case "mark_utilized":
					command.MarkUtilized = NullField[bool]()
				case "description":
					command.Description = NullField[string]()
				case "comments":
					command.Comments = NullField[string]()
				}
				var values prefixCommandValues
				var patch prefixCommandPatch
				var err error
				switch operation {
				case "POST":
					values, err = command.values()
				case "PUT":
					values, err = (ReplacePrefixCommand{ID: 3, CreatePrefixCommand: command}).values()
				case "PATCH":
					patch, err = prefixPresenceUpdate(command).patch()
				}
				if field == "vrf" {
					require.NoError(t, err)
					if operation == "PATCH" {
						assert.True(t, patch.vrfSet)
						assert.Nil(t, patch.vrfID)
					} else {
						assert.Nil(t, values.vrfID)
					}
				} else {
					reason, description := "null", "This field may not be null."
					if field == "status" {
						reason, description = "invalid", "This field may not be blank."
					}
					assert.Equal(t, []shared.FieldViolation{{
						Field: field, Reason: reason, Description: description,
					}}, shared.ViolationsOf(err))
				}
				if field != "comments" {
					if operation == "PATCH" {
						require.NotNil(t, patch.comments)
						assert.Equal(t, "retained comments", *patch.comments)
					} else {
						assert.Equal(t, "retained comments", values.comments)
					}
				} else if operation == "PATCH" {
					require.NotNil(t, patch.description)
					assert.Equal(t, "retained description", *patch.description)
				} else {
					assert.Equal(t, "retained description", values.description)
				}
			})
		}
	}

	patch, err := (UpdatePrefixCommand{ID: 3, Description: FieldValue("changed")}).patch()
	require.NoError(t, err)
	assert.Nil(t, patch.prefix)
	assert.False(t, patch.vrfSet)
	assert.Nil(t, patch.status)
	assert.Nil(t, patch.isPool)
	assert.Nil(t, patch.markUtilized)
	assert.Nil(t, patch.comments)
	_, err = (UpdatePrefixCommand{ID: 3}).patch()
	assert.Equal(t, []shared.FieldViolation{{
		Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
	}}, shared.ViolationsOf(err))
	_, err = (UpdatePrefixCommand{ID: 3, IsPool: NullField[bool](), MarkUtilized: NullField[bool]()}).patch()
	assert.Equal(t, []shared.FieldViolation{
		{Field: "is_pool", Reason: "null", Description: "This field may not be null."},
		{Field: "mark_utilized", Reason: "null", Description: "This field may not be null."},
	}, shared.ViolationsOf(err), "invalid supplied fields must not fabricate an empty-update violation")

	prefixNull := shared.FieldViolation{Field: "prefix", Reason: "null", Description: "This field may not be null."}
	vrfInvalid := shared.FieldViolation{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: -9"}
	statusInvalid := shared.FieldViolation{Field: "status", Reason: "invalid", Description: "This field may not be blank."}
	length := shared.FieldViolation{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."}
	nul := shared.FieldViolation{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}
	commentsNUL := shared.FieldViolation{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}
	merged := mergePrefixMutationErrors(
		shared.NewValidationError(prefixNull, statusInvalid, commentsNUL),
		shared.NewValidationError(vrfInvalid),
		shared.NewValidationError(
			shared.FieldViolation{Field: "prefix", Reason: "blank", Description: "This field may not be blank."},
			statusInvalid, length, nul, commentsNUL,
		),
	)
	assert.Equal(t, []shared.FieldViolation{prefixNull, vrfInvalid, statusInvalid, length, nul, commentsNUL}, shared.ViolationsOf(merged))
	internal := shared.NewError(shared.ErrorReasonInternal, "reader failed")
	assert.Same(t, internal, mergePrefixMutationErrors(merged, internal))
}

func prefixPresenceUpdate(command CreatePrefixCommand) UpdatePrefixCommand {
	return UpdatePrefixCommand{
		ID: 3, Prefix: command.Prefix, VRF: command.VRF, Status: command.Status,
		IsPool: command.IsPool, MarkUtilized: command.MarkUtilized,
		Description: command.Description, Comments: command.Comments,
	}
}
