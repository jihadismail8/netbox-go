package dcim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST required fields and defaults", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceCommand{}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "device_type", Reason: "required", Description: "This field is required."},
			{Field: "role", Reason: "required", Description: "This field is required."},
			{Field: "site", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.True(t, values.name.IsNull())
		assert.True(t, values.rackID.IsNull())
		assert.True(t, values.position.IsNull())
		assert.Empty(t, values.face)
		assert.Equal(t, dcimdomain.DeviceStatusActive.String(), values.status)
		assert.Empty(t, values.serial)
		assert.True(t, values.assetTag.IsNull())
		assert.True(t, values.airflow.IsNull())
		assert.Empty(t, values.description)
		assert.Empty(t, values.comments)
	})

	t.Run("POST nullable fields clear and airflow remains inheritable", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceCommand{
			DeviceType: FieldValue(shared.ID(10)), Role: FieldValue(shared.ID(11)),
			Site: FieldValue(shared.ID(12)), Name: NullField[string](),
			Rack: NullField[shared.ID](), Position: NullField[string](),
			Face: NullField[string](), AssetTag: NullField[string](),
			Airflow: NullField[string](),
		}).values()
		require.NoError(t, err)
		assert.True(t, values.name.IsNull())
		assert.True(t, values.rackID.IsNull())
		assert.True(t, values.position.IsNull())
		assert.Empty(t, values.face)
		assert.True(t, values.assetTag.IsNull())
		assert.True(t, values.airflow.IsNull())
	})

	t.Run("POST blank zero and concrete intent remains present", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateDeviceCommand{
			DeviceType: FieldValue(shared.ID(10)), Role: FieldValue(shared.ID(11)),
			Name: FieldValue(""), Site: FieldValue(shared.ID(12)),
			Rack: FieldValue(shared.ID(13)), Position: FieldValue("0"),
			Face: FieldValue(""), Status: FieldValue("planned"), Serial: FieldValue(""),
			AssetTag: FieldValue(""), Airflow: FieldValue(""),
			Description: FieldValue(""), Comments: FieldValue(""),
		}).values()
		require.NoError(t, err)
		assert.Equal(t, "", deviceCommandNullableValue(t, values.name))
		assert.Equal(t, shared.ID(13), deviceCommandNullableValue(t, values.rackID))
		assert.Equal(t, "0", deviceCommandNullableValue(t, values.position))
		assert.Equal(t, "", deviceCommandNullableValue(t, values.assetTag))
		airflow, present := values.airflow.Get()
		require.True(t, present)
		assert.Empty(t, airflow)
	})

	t.Run("PUT resets only omitted rack position and face", func(t *testing.T) {
		t.Parallel()
		patch, err := (ReplaceDeviceCommand{
			ID: 41,
			CreateDeviceCommand: CreateDeviceCommand{
				DeviceType: FieldValue(shared.ID(10)), Role: FieldValue(shared.ID(11)),
				Site: FieldValue(shared.ID(12)),
			},
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.deviceTypeID)
		require.NotNil(t, patch.roleID)
		require.NotNil(t, patch.siteID)
		require.NotNil(t, patch.rackID)
		assert.True(t, patch.rackID.IsNull())
		require.NotNil(t, patch.position)
		assert.True(t, patch.position.IsNull())
		require.NotNil(t, patch.face)
		assert.Empty(t, *patch.face)
		assert.Nil(t, patch.name)
		assert.Nil(t, patch.status)
		assert.Nil(t, patch.serial)
		assert.Nil(t, patch.assetTag)
		assert.Nil(t, patch.airflow)
		assert.Nil(t, patch.description)
		assert.Nil(t, patch.comments)

		_, missingErr := (ReplaceDeviceCommand{ID: 41}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "device_type", Reason: "required", Description: "This field is required."},
			{Field: "role", Reason: "required", Description: "This field is required."},
			{Field: "site", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(missingErr))
	})

	t.Run("PATCH nullable clears and choice nulls normalize to blank", func(t *testing.T) {
		t.Parallel()
		patch, err := (UpdateDeviceCommand{
			ID: 41, Name: NullField[string](), Rack: NullField[shared.ID](),
			Position: NullField[string](), Face: NullField[string](),
			AssetTag: NullField[string](), Airflow: NullField[string](),
			Comments: FieldValue("sibling"),
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.name)
		assert.True(t, patch.name.IsNull())
		require.NotNil(t, patch.rackID)
		assert.True(t, patch.rackID.IsNull())
		require.NotNil(t, patch.position)
		assert.True(t, patch.position.IsNull())
		require.NotNil(t, patch.face)
		assert.Empty(t, *patch.face)
		require.NotNil(t, patch.assetTag)
		assert.True(t, patch.assetTag.IsNull())
		require.NotNil(t, patch.airflow)
		airflow, present := patch.airflow.Get()
		require.True(t, present)
		assert.Empty(t, airflow)
		require.NotNil(t, patch.comments)
		assert.Equal(t, "sibling", *patch.comments)
	})

	t.Run("PATCH non-null fields report ordered null violations", func(t *testing.T) {
		t.Parallel()
		_, err := (UpdateDeviceCommand{
			ID: 41, DeviceType: NullField[shared.ID](), Role: NullField[shared.ID](),
			Site: NullField[shared.ID](), Status: NullField[string](),
			Serial: NullField[string](), Description: NullField[string](),
			Comments: NullField[string](),
		}).patch()
		assert.Equal(t, []string{
			"device_type", "role", "site", "status", "serial", "description", "comments",
		}, deviceCommandViolationFields(err))
	})

	t.Run("PATCH omission preserves every field with a concrete sibling", func(t *testing.T) {
		t.Parallel()
		for _, field := range []string{
			"device_type", "role", "name", "site", "rack", "position", "face",
			"status", "serial", "asset_tag", "airflow", "description", "comments",
		} {
			field := field
			t.Run(field, func(t *testing.T) {
				command := UpdateDeviceCommand{ID: 41, Comments: FieldValue("sibling")}
				if field == "comments" {
					command.Description = FieldValue("sibling")
					command.Comments = OmittedField[string]()
				}
				patch, err := command.patch()
				require.NoError(t, err)
				assertDeviceCommandPatchFieldNil(t, patch, field)
			})
		}
	})

	t.Run("relationship IDs must be positive", func(t *testing.T) {
		t.Parallel()
		_, err := (UpdateDeviceCommand{
			ID: 41, DeviceType: FieldValue(shared.ID(0)), Role: FieldValue(shared.ID(-1)),
			Site: FieldValue(shared.ID(0)), Rack: FieldValue(shared.ID(-1)),
			Description: FieldValue("sibling"),
		}).patch()
		assert.Equal(t, []string{"device_type", "role", "site", "rack"}, deviceCommandViolationFields(err))
	})
}

func deviceCommandViolationFields(err error) []string {
	violations := shared.ViolationsOf(err)
	fields := make([]string, len(violations))
	for index, violation := range violations {
		fields[index] = violation.Field
	}
	return fields
}

func deviceCommandNullableValue[T any](t *testing.T, value dcimdomain.DeviceNullable[T]) T {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}

func assertDeviceCommandPatchFieldNil(t *testing.T, patch deviceCommandPatch, field string) {
	t.Helper()
	switch field {
	case "device_type":
		assert.Nil(t, patch.deviceTypeID)
	case "role":
		assert.Nil(t, patch.roleID)
	case "name":
		assert.Nil(t, patch.name)
	case "site":
		assert.Nil(t, patch.siteID)
	case "rack":
		assert.Nil(t, patch.rackID)
	case "position":
		assert.Nil(t, patch.position)
	case "face":
		assert.Nil(t, patch.face)
	case "status":
		assert.Nil(t, patch.status)
	case "serial":
		assert.Nil(t, patch.serial)
	case "asset_tag":
		assert.Nil(t, patch.assetTag)
	case "airflow":
		assert.Nil(t, patch.airflow)
	case "description":
		assert.Nil(t, patch.description)
	case "comments":
		assert.Nil(t, patch.comments)
	}
}
