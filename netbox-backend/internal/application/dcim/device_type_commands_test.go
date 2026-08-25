package dcim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceTypeScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST required fields and defaults", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceTypeCommand{}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "manufacturer", Reason: "required", Description: "This field is required."},
			{Field: "model", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.Equal(t, dcimdomain.DeviceTypeDefaultHeight, values.uHeight)
		assert.False(t, values.excludeFromUtilization)
		assert.True(t, values.isFullDepth)
		assert.True(t, values.airflow.IsNull())
		assert.Empty(t, values.partNumber)
		assert.Empty(t, values.description)
		assert.Empty(t, values.comments)
	})

	t.Run("POST nulls aggregate while airflow alone accepts null", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceTypeCommand{
			Manufacturer: NullField[shared.ID](), Model: NullField[string](),
			Slug: NullField[string](), PartNumber: NullField[string](),
			UHeight: NullField[string](), ExcludeFromUtilization: NullField[bool](),
			IsFullDepth: NullField[bool](), Airflow: NullField[string](),
			Description: NullField[string](), Comments: NullField[string](),
		}).values()
		assert.Equal(t, []shared.FieldViolation{
			nullDeviceTypeViolation("manufacturer"),
			nullDeviceTypeViolation("model"),
			nullDeviceTypeViolation("slug"),
			nullDeviceTypeViolation("part_number"),
			nullDeviceTypeViolation("u_height"),
			nullDeviceTypeViolation("exclude_from_utilization"),
			nullDeviceTypeViolation("is_full_depth"),
			nullDeviceTypeViolation("description"),
			nullDeviceTypeViolation("comments"),
		}, shared.ViolationsOf(err))
		assert.True(t, values.airflow.IsNull())
	})

	t.Run("POST concrete false zero blank and airflow remain present", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceTypeCommand{
			Manufacturer: FieldValue(shared.ID(9)), Model: FieldValue("  Router  "),
			Slug: FieldValue("  router  "), PartNumber: FieldValue(""),
			UHeight: FieldValue("0"), ExcludeFromUtilization: FieldValue(false),
			IsFullDepth: FieldValue(false), Airflow: FieldValue(""),
			Description: FieldValue(""), Comments: FieldValue(""),
		}).values()
		require.NoError(t, err)
		assert.Equal(t, shared.ID(9), values.manufacturerID)
		assert.Equal(t, "0", values.uHeight)
		assert.False(t, values.excludeFromUtilization)
		assert.False(t, values.isFullDepth)
		airflow, present := values.airflow.Get()
		assert.True(t, present)
		assert.Empty(t, airflow)
	})

	t.Run("PUT requires identity resets height and preserves other omissions", func(t *testing.T) {
		t.Parallel()
		patch, err := (ReplaceDeviceTypeCommand{
			ID: 41,
			CreateDeviceTypeCommand: CreateDeviceTypeCommand{
				Manufacturer: FieldValue(shared.ID(9)), Model: FieldValue("Replacement"),
				Slug: FieldValue("replacement"),
			},
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.manufacturerID)
		assert.Equal(t, shared.ID(9), *patch.manufacturerID)
		require.NotNil(t, patch.model)
		require.NotNil(t, patch.slug)
		require.NotNil(t, patch.uHeight)
		assert.Equal(t, dcimdomain.DeviceTypeDefaultHeight, *patch.uHeight)
		assert.Nil(t, patch.partNumber)
		assert.Nil(t, patch.excludeFromUtilization)
		assert.Nil(t, patch.isFullDepth)
		assert.Nil(t, patch.airflow)
		assert.Nil(t, patch.description)
		assert.Nil(t, patch.comments)

		missing, missingErr := (ReplaceDeviceTypeCommand{ID: 41}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "manufacturer", Reason: "required", Description: "This field is required."},
			{Field: "model", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(missingErr))
		require.NotNil(t, missing.uHeight)
	})

	t.Run("PUT and PATCH nulls retain valid airflow intent and sibling intent", func(t *testing.T) {
		t.Parallel()
		command := UpdateDeviceTypeCommand{
			ID: 41, Manufacturer: NullField[shared.ID](), Model: NullField[string](),
			Slug: NullField[string](), PartNumber: NullField[string](),
			UHeight: NullField[string](), ExcludeFromUtilization: NullField[bool](),
			IsFullDepth: NullField[bool](), Airflow: NullField[string](),
			Description: NullField[string](), Comments: FieldValue("retained sibling"),
		}
		patch, err := command.patch()
		assert.Equal(t, []shared.FieldViolation{
			nullDeviceTypeViolation("manufacturer"),
			nullDeviceTypeViolation("model"),
			nullDeviceTypeViolation("slug"),
			nullDeviceTypeViolation("part_number"),
			nullDeviceTypeViolation("u_height"),
			nullDeviceTypeViolation("exclude_from_utilization"),
			nullDeviceTypeViolation("is_full_depth"),
			nullDeviceTypeViolation("description"),
		}, shared.ViolationsOf(err))
		require.NotNil(t, patch.airflow)
		assert.True(t, patch.airflow.IsNull())
		require.NotNil(t, patch.comments)
		assert.Equal(t, "retained sibling", *patch.comments)
	})

	t.Run("PATCH omission preserves every field with a concrete sibling", func(t *testing.T) {
		t.Parallel()
		for _, field := range []string{
			"manufacturer", "model", "slug", "part_number", "u_height",
			"exclude_from_utilization", "is_full_depth", "airflow", "description", "comments",
		} {
			field := field
			t.Run(field, func(t *testing.T) {
				t.Parallel()
				command := UpdateDeviceTypeCommand{ID: 41}
				if field == "comments" {
					command.Description = FieldValue("changed sibling")
				} else {
					command.Comments = FieldValue("changed sibling")
				}
				patch, err := command.patch()
				require.NoError(t, err)
				assertDeviceTypePatchFieldNil(t, patch, field)
			})
		}
	})

	t.Run("manufacturer IDs must be positive", func(t *testing.T) {
		t.Parallel()
		for _, command := range []UpdateDeviceTypeCommand{
			{ID: 41, Manufacturer: FieldValue(shared.ID(0)), Comments: FieldValue("sibling")},
			{ID: 41, Manufacturer: FieldValue(shared.ID(-1)), Comments: FieldValue("sibling")},
		} {
			_, err := command.patch()
			assert.Equal(t, []shared.FieldViolation{{
				Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
			}}, shared.ViolationsOf(err))
		}
	})
}

func nullDeviceTypeViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "null", Description: "This field may not be null.",
	}
}

func assertDeviceTypePatchFieldNil(t *testing.T, patch deviceTypeCommandPatch, field string) {
	t.Helper()
	switch field {
	case "manufacturer":
		assert.Nil(t, patch.manufacturerID)
	case "model":
		assert.Nil(t, patch.model)
	case "slug":
		assert.Nil(t, patch.slug)
	case "part_number":
		assert.Nil(t, patch.partNumber)
	case "u_height":
		assert.Nil(t, patch.uHeight)
	case "exclude_from_utilization":
		assert.Nil(t, patch.excludeFromUtilization)
	case "is_full_depth":
		assert.Nil(t, patch.isFullDepth)
	case "airflow":
		assert.Nil(t, patch.airflow)
	case "description":
		assert.Nil(t, patch.description)
	case "comments":
		assert.Nil(t, patch.comments)
	}
}
