package dcim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/shared"
)

func TestInterfaceTemplateScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST required fields and defaults", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateInterfaceTemplateCommand{}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "device_type", Reason: "required", Description: "This field is required."},
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "type", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.Empty(t, values.label)
		assert.True(t, values.enabled)
		assert.False(t, values.mgmtOnly)
		assert.Empty(t, values.description)
	})

	t.Run("POST nulls aggregate in field order", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateInterfaceTemplateCommand{
			DeviceType: NullField[shared.ID](), Name: NullField[string](),
			Label: NullField[string](), Type: NullField[string](),
			Enabled: NullField[bool](), MgmtOnly: NullField[bool](),
			Description: NullField[string](),
		}).values()
		assert.Equal(t, []shared.FieldViolation{
			nullInterfaceTemplateViolation("device_type"),
			nullInterfaceTemplateViolation("name"),
			nullInterfaceTemplateViolation("label"),
			nullInterfaceTemplateViolation("type"),
			nullInterfaceTemplateViolation("enabled"),
			nullInterfaceTemplateViolation("mgmt_only"),
			nullInterfaceTemplateViolation("description"),
		}, shared.ViolationsOf(err))
		assert.True(t, values.enabled, "invalid input must retain the POST default")
	})

	t.Run("POST explicit blank and false values remain concrete", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateInterfaceTemplateCommand{
			DeviceType: FieldValue(shared.ID(73)), Name: FieldValue("  Ethernet1  "),
			Label: FieldValue(""), Type: FieldValue("bridge"),
			Enabled: FieldValue(false), MgmtOnly: FieldValue(false),
			Description: FieldValue(""),
		}).values()
		require.NoError(t, err)
		assert.Equal(t, shared.ID(73), values.deviceTypeID)
		assert.Empty(t, values.label)
		assert.False(t, values.enabled)
		assert.False(t, values.mgmtOnly)
		assert.Empty(t, values.description)
	})

	t.Run("PUT requires the triplet and preserves every optional omission", func(t *testing.T) {
		t.Parallel()
		patch, err := (ReplaceInterfaceTemplateCommand{
			ID: 41,
			CreateInterfaceTemplateCommand: CreateInterfaceTemplateCommand{
				DeviceType: FieldValue(shared.ID(73)), Name: FieldValue("Ethernet1"),
				Type: FieldValue("bridge"),
			},
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.deviceTypeID)
		assert.Equal(t, shared.ID(73), *patch.deviceTypeID)
		require.NotNil(t, patch.name)
		require.NotNil(t, patch.interfaceType)
		assert.Nil(t, patch.label)
		assert.Nil(t, patch.enabled)
		assert.Nil(t, patch.mgmtOnly)
		assert.Nil(t, patch.description)

		_, missingErr := (ReplaceInterfaceTemplateCommand{ID: 41}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "device_type", Reason: "required", Description: "This field is required."},
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "type", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(missingErr))
	})

	t.Run("PUT and PATCH nulls retain concrete sibling intent", func(t *testing.T) {
		t.Parallel()
		command := UpdateInterfaceTemplateCommand{
			ID: 41, DeviceType: NullField[shared.ID](), Name: NullField[string](),
			Label: NullField[string](), Type: NullField[string](),
			Enabled: NullField[bool](), MgmtOnly: NullField[bool](),
			Description: FieldValue("retained sibling"),
		}
		patch, err := command.patch()
		assert.Equal(t, []shared.FieldViolation{
			nullInterfaceTemplateViolation("device_type"),
			nullInterfaceTemplateViolation("name"),
			nullInterfaceTemplateViolation("label"),
			nullInterfaceTemplateViolation("type"),
			nullInterfaceTemplateViolation("enabled"),
			nullInterfaceTemplateViolation("mgmt_only"),
		}, shared.ViolationsOf(err))
		require.NotNil(t, patch.description)
		assert.Equal(t, "retained sibling", *patch.description)
	})

	t.Run("PATCH omission preserves every field with a concrete sibling", func(t *testing.T) {
		t.Parallel()
		for _, field := range []string{
			"device_type", "name", "label", "type", "enabled", "mgmt_only", "description",
		} {
			field := field
			t.Run(field, func(t *testing.T) {
				t.Parallel()
				command := UpdateInterfaceTemplateCommand{ID: 41}
				if field == "description" {
					command.Name = FieldValue("changed sibling")
				} else {
					command.Description = FieldValue("changed sibling")
				}
				patch, err := command.patch()
				require.NoError(t, err)
				assertInterfaceTemplatePatchFieldNil(t, patch, field)
			})
		}
	})

	t.Run("device type IDs must be positive", func(t *testing.T) {
		t.Parallel()
		for _, id := range []shared.ID{0, -1} {
			_, err := (UpdateInterfaceTemplateCommand{
				ID: 41, DeviceType: FieldValue(id), Description: FieldValue("sibling"),
			}).patch()
			assert.Equal(t, []shared.FieldViolation{{
				Field: "device_type", Reason: "invalid_choice", Description: "Select a valid choice.",
			}}, shared.ViolationsOf(err))
		}
	})
}

func nullInterfaceTemplateViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "null", Description: "This field may not be null.",
	}
}

func assertInterfaceTemplatePatchFieldNil(
	t *testing.T,
	patch interfaceTemplateCommandPatch,
	field string,
) {
	t.Helper()
	switch field {
	case "device_type":
		assert.Nil(t, patch.deviceTypeID)
	case "name":
		assert.Nil(t, patch.name)
	case "label":
		assert.Nil(t, patch.label)
	case "type":
		assert.Nil(t, patch.interfaceType)
	case "enabled":
		assert.Nil(t, patch.enabled)
	case "mgmt_only":
		assert.Nil(t, patch.mgmtOnly)
	case "description":
		assert.Nil(t, patch.description)
	}
}
