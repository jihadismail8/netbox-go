package dcim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestRackScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST required fields and defaults", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateRackCommand{}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "site", Reason: "required", Description: "This field is required."},
			{Field: "name", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.True(t, values.facilityID.IsNull())
		assert.True(t, values.rackTypeID.IsNull())
		assert.Equal(t, dcimdomain.RackStatusActive.String(), values.status)
		assert.True(t, values.roleID.IsNull())
		assert.Empty(t, values.serial)
		assert.True(t, values.assetTag.IsNull())
		assert.True(t, values.formFactor.IsNull())
		assert.Equal(t, dcimdomain.RackDefaultWidth, values.width)
		assert.Equal(t, dcimdomain.RackDefaultUHeight, values.uHeight)
		assert.Equal(t, dcimdomain.RackDefaultStartingUnit, values.startingUnit)
		assert.False(t, values.descUnits)
		assert.True(t, values.airflow.IsNull())
		assert.Empty(t, values.description)
		assert.Empty(t, values.comments)
	})

	t.Run("POST nulls aggregate while nullable fields clear", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateRackCommand{
			Site: NullField[shared.ID](), Name: NullField[string](),
			FacilityID: NullField[string](), RackType: NullField[shared.ID](),
			Status: NullField[string](), Role: NullField[shared.ID](),
			Serial: NullField[string](), AssetTag: NullField[string](),
			FormFactor: NullField[string](), Width: NullField[uint32](),
			UHeight: NullField[uint32](), StartingUnit: NullField[uint32](),
			DescUnits: NullField[bool](), Airflow: NullField[string](),
			Description: NullField[string](), Comments: NullField[string](),
		}).values()
		assert.Equal(t, []string{
			"site", "name", "status", "serial", "width", "u_height",
			"starting_unit", "desc_units", "airflow", "description", "comments",
		}, rackViolationFields(err))
		assert.True(t, values.facilityID.IsNull())
		assert.True(t, values.rackTypeID.IsNull())
		assert.True(t, values.roleID.IsNull())
		assert.True(t, values.assetTag.IsNull())
		assert.True(t, values.formFactor.IsNull())
		assert.True(t, values.airflow.IsNull())
	})

	t.Run("POST concrete blank zero and false intent remains present", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateRackCommand{
			Site: FieldValue(shared.ID(3)), Name: FieldValue("  A01  "),
			FacilityID: FieldValue(""), RackType: FieldValue(shared.ID(8)),
			Status: FieldValue("planned"), Role: FieldValue(shared.ID(9)),
			Serial: FieldValue(""), AssetTag: FieldValue(""),
			FormFactor: FieldValue(""), Width: FieldValue(uint32(23)),
			UHeight:      FieldValue(uint32(100)),
			StartingUnit: FieldValue(dcimdomain.RackTypeMaximumStartingUnit),
			DescUnits:    FieldValue(false), Airflow: FieldValue(""),
			Description: FieldValue(""), Comments: FieldValue(""),
		}).values()
		require.NoError(t, err)
		assert.Equal(t, shared.ID(3), values.siteID)
		assert.Equal(t, "  A01  ", values.name)
		assert.Equal(t, "", rackCommandNullableValue(t, values.facilityID))
		assert.Equal(t, shared.ID(8), rackCommandNullableValue(t, values.rackTypeID))
		assert.Equal(t, shared.ID(9), rackCommandNullableValue(t, values.roleID))
		assert.Equal(t, "", rackCommandNullableValue(t, values.assetTag))
		assert.Equal(t, "", rackCommandNullableValue(t, values.formFactor))
		assert.Equal(t, "", rackCommandNullableValue(t, values.airflow))
		assert.False(t, values.descUnits)
	})

	t.Run("PUT resets only omitted facility and rack type", func(t *testing.T) {
		t.Parallel()
		patch, err := (ReplaceRackCommand{
			ID: 41,
			CreateRackCommand: CreateRackCommand{
				Site: FieldValue(shared.ID(3)), Name: FieldValue("A01"),
			},
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.siteID)
		require.NotNil(t, patch.name)
		require.NotNil(t, patch.facilityID)
		assert.True(t, patch.facilityID.IsNull())
		require.NotNil(t, patch.rackTypeID)
		assert.True(t, patch.rackTypeID.IsNull())
		assert.Nil(t, patch.status)
		assert.Nil(t, patch.roleID)
		assert.Nil(t, patch.serial)
		assert.Nil(t, patch.assetTag)
		assert.Nil(t, patch.formFactor)
		assert.Nil(t, patch.width)
		assert.Nil(t, patch.uHeight)
		assert.Nil(t, patch.startingUnit)
		assert.Nil(t, patch.descUnits)
		assert.Nil(t, patch.airflow)
		assert.Nil(t, patch.description)
		assert.Nil(t, patch.comments)

		_, missingErr := (ReplaceRackCommand{ID: 41}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "site", Reason: "required", Description: "This field is required."},
			{Field: "name", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(missingErr))
	})

	t.Run("PATCH nullable clears and airflow null rejection retain sibling intent", func(t *testing.T) {
		t.Parallel()
		patch, err := (UpdateRackCommand{
			ID: 41, FacilityID: NullField[string](), RackType: NullField[shared.ID](),
			Role: NullField[shared.ID](), AssetTag: NullField[string](),
			FormFactor: NullField[string](), Airflow: NullField[string](),
			Description: FieldValue("retained sibling"),
		}).patch()
		assert.Equal(t, []shared.FieldViolation{rackNullViolation("airflow")}, shared.ViolationsOf(err))
		require.NotNil(t, patch.facilityID)
		assert.True(t, patch.facilityID.IsNull())
		require.NotNil(t, patch.rackTypeID)
		assert.True(t, patch.rackTypeID.IsNull())
		require.NotNil(t, patch.roleID)
		assert.True(t, patch.roleID.IsNull())
		require.NotNil(t, patch.assetTag)
		assert.True(t, patch.assetTag.IsNull())
		require.NotNil(t, patch.formFactor)
		assert.True(t, patch.formFactor.IsNull())
		assert.Nil(t, patch.airflow)
		require.NotNil(t, patch.description)
		assert.Equal(t, "retained sibling", *patch.description)
	})

	t.Run("PATCH non-null fields report field-specific null violations", func(t *testing.T) {
		t.Parallel()
		_, err := (UpdateRackCommand{
			ID: 41, Site: NullField[shared.ID](), Name: NullField[string](),
			Status: NullField[string](), Serial: NullField[string](),
			Width: NullField[uint32](), UHeight: NullField[uint32](),
			StartingUnit: NullField[uint32](), DescUnits: NullField[bool](),
			Airflow: NullField[string](), Description: NullField[string](),
			Comments: NullField[string](),
		}).patch()
		assert.Equal(t, []string{
			"site", "name", "status", "serial", "width", "u_height",
			"starting_unit", "desc_units", "airflow", "description", "comments",
		}, rackViolationFields(err))

		patch, blankErr := (UpdateRackCommand{
			ID: 41, Airflow: FieldValue(""),
		}).patch()
		require.NoError(t, blankErr)
		require.NotNil(t, patch.airflow)
		assert.Equal(t, "", rackCommandNullableValue(t, *patch.airflow))
	})

	t.Run("PATCH omission preserves every field with a concrete sibling", func(t *testing.T) {
		t.Parallel()
		for _, field := range []string{
			"site", "name", "facility_id", "rack_type", "status", "role", "serial",
			"asset_tag", "form_factor", "width", "u_height", "starting_unit",
			"desc_units", "airflow", "description", "comments",
		} {
			field := field
			t.Run(field, func(t *testing.T) {
				command := UpdateRackCommand{ID: 41, Comments: FieldValue("sibling")}
				if field == "comments" {
					command.Description = FieldValue("sibling")
					command.Comments = OmittedField[string]()
				}
				patch, err := command.patch()
				require.NoError(t, err)
				assertRackCommandPatchFieldNil(t, patch, field)
			})
		}
	})

	t.Run("relationship IDs must be positive", func(t *testing.T) {
		t.Parallel()
		_, err := (UpdateRackCommand{
			ID: 41, Site: FieldValue(shared.ID(0)), RackType: FieldValue(shared.ID(-1)),
			Role: FieldValue(shared.ID(0)), Description: FieldValue("sibling"),
		}).patch()
		assert.Equal(t, []string{"site", "rack_type", "role"}, rackViolationFields(err))
	})
}

func rackViolationFields(err error) []string {
	violations := shared.ViolationsOf(err)
	fields := make([]string, len(violations))
	for index, violation := range violations {
		fields[index] = violation.Field
	}
	return fields
}

func rackCommandNullableValue[T any](t *testing.T, value dcimdomain.RackNullable[T]) T {
	t.Helper()
	actual, present := value.Get()
	require.True(t, present)
	return actual
}

func assertRackCommandPatchFieldNil(t *testing.T, patch rackCommandPatch, field string) {
	t.Helper()
	switch field {
	case "site":
		assert.Nil(t, patch.siteID)
	case "name":
		assert.Nil(t, patch.name)
	case "facility_id":
		assert.Nil(t, patch.facilityID)
	case "rack_type":
		assert.Nil(t, patch.rackTypeID)
	case "status":
		assert.Nil(t, patch.status)
	case "role":
		assert.Nil(t, patch.roleID)
	case "serial":
		assert.Nil(t, patch.serial)
	case "asset_tag":
		assert.Nil(t, patch.assetTag)
	case "form_factor":
		assert.Nil(t, patch.formFactor)
	case "width":
		assert.Nil(t, patch.width)
	case "u_height":
		assert.Nil(t, patch.uHeight)
	case "starting_unit":
		assert.Nil(t, patch.startingUnit)
	case "desc_units":
		assert.Nil(t, patch.descUnits)
	case "airflow":
		assert.Nil(t, patch.airflow)
	case "description":
		assert.Nil(t, patch.description)
	case "comments":
		assert.Nil(t, patch.comments)
	}
}
