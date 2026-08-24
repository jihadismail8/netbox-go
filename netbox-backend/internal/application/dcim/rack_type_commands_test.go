package dcim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var rackTypeCommandTestTime = shared.NewTimestamp(
	time.Date(2026, time.August, 24, 20, 30, 0, 0, time.UTC),
)

func TestRackTypeScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST omission and defaults", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateRackTypeCommand{
			Manufacturer: FieldValue(shared.ID(9)),
			Model:        FieldValue("Rack 42"), Slug: FieldValue("rack-42"),
			FormFactor: FieldValue("4-post-cabinet"),
		}).values()
		require.NoError(t, err)
		assert.Equal(t, rackTypeCommandValues{
			manufacturerID: 9, model: "Rack 42", slug: "rack-42",
			formFactor: "4-post-cabinet", width: dcimdomain.RackTypeDefaultWidth,
			uHeight:      dcimdomain.RackTypeDefaultUHeight,
			startingUnit: dcimdomain.RackTypeDefaultStartingUnit,
		}, values)

		omitted, commandErr := (CreateRackTypeCommand{}).values()
		assert.Equal(t, rackTypeCommandValues{
			width:        dcimdomain.RackTypeDefaultWidth,
			uHeight:      dcimdomain.RackTypeDefaultUHeight,
			startingUnit: dcimdomain.RackTypeDefaultStartingUnit,
		}, omitted)
		assert.Equal(t, []shared.FieldViolation{
			{Field: "manufacturer", Reason: "required", Description: "This field is required."},
			{Field: "model", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(commandErr))
		_, domainErr := dcimdomain.NewRackType(rackTypeDomainValues(omitted), rackTypeCommandTestTime)
		assert.Equal(t, []string{"manufacturer", "model", "slug", "form_factor"},
			violationFields(mergeRackTypeMutationErrors(commandErr, domainErr)))
	})

	t.Run("POST null states retain defaults and aggregate in field order", func(t *testing.T) {
		t.Parallel()
		values, err := (CreateRackTypeCommand{
			Manufacturer: NullField[shared.ID](), Model: NullField[string](),
			Slug: NullField[string](), FormFactor: NullField[string](),
			Width: NullField[uint32](), UHeight: NullField[uint32](),
			StartingUnit: NullField[uint32](), DescUnits: NullField[bool](),
			Description: NullField[string](), Comments: NullField[string](),
		}).values()
		assert.Equal(t, dcimdomain.RackTypeDefaultWidth, values.width)
		assert.Equal(t, dcimdomain.RackTypeDefaultUHeight, values.uHeight)
		assert.Equal(t, dcimdomain.RackTypeDefaultStartingUnit, values.startingUnit)
		assert.Equal(t, []shared.FieldViolation{
			nullRackTypeViolation("manufacturer"),
			nullRackTypeViolation("model"),
			nullRackTypeViolation("slug"),
			nullRackTypeViolation("form_factor"),
			blankRackTypeWidthViolation(),
			nullRackTypeViolation("u_height"),
			nullRackTypeViolation("starting_unit"),
			nullRackTypeViolation("desc_units"),
			nullRackTypeViolation("description"),
			nullRackTypeViolation("comments"),
		}, shared.ViolationsOf(err))
	})

	t.Run("PUT requires identity and preserves every omitted optional", func(t *testing.T) {
		t.Parallel()
		patch, err := (ReplaceRackTypeCommand{
			ID: 1,
			CreateRackTypeCommand: CreateRackTypeCommand{
				Manufacturer: FieldValue(shared.ID(9)),
				Model:        FieldValue("Replacement"), Slug: FieldValue("replacement"),
			},
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.manufacturerID)
		assert.Equal(t, shared.ID(9), *patch.manufacturerID)
		require.NotNil(t, patch.model)
		assert.Equal(t, "Replacement", *patch.model)
		require.NotNil(t, patch.slug)
		assert.Equal(t, "replacement", *patch.slug)
		assert.Nil(t, patch.formFactor)
		assert.Nil(t, patch.width)
		assert.Nil(t, patch.uHeight)
		assert.Nil(t, patch.startingUnit)
		assert.Nil(t, patch.descUnits)
		assert.Nil(t, patch.description)
		assert.Nil(t, patch.comments)

		emptyPatch, err := (ReplaceRackTypeCommand{ID: 1}).patch()
		assert.True(t, emptyPatch.empty())
		assert.Equal(t, []shared.FieldViolation{
			{Field: "manufacturer", Reason: "required", Description: "This field is required."},
			{Field: "model", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
	})

	t.Run("PUT and PATCH null states aggregate without discarding siblings", func(t *testing.T) {
		t.Parallel()
		for _, test := range []struct {
			name  string
			patch func() (rackTypeCommandPatch, error)
		}{
			{
				name: "PUT",
				patch: func() (rackTypeCommandPatch, error) {
					return (ReplaceRackTypeCommand{ID: 1, CreateRackTypeCommand: allNullRackTypeCommand()}).patch()
				},
			},
			{
				name: "PATCH",
				patch: func() (rackTypeCommandPatch, error) {
					command := allNullRackTypeCommand()
					return (UpdateRackTypeCommand{
						ID:           commandID,
						Manufacturer: command.Manufacturer, Model: command.Model, Slug: command.Slug,
						FormFactor: command.FormFactor, Width: command.Width, UHeight: command.UHeight,
						StartingUnit: command.StartingUnit, DescUnits: command.DescUnits,
						Description: command.Description, Comments: command.Comments,
					}).patch()
				},
			},
		} {
			test := test
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				patch, err := test.patch()
				assert.True(t, patch.empty())
				assert.Equal(t, []shared.FieldViolation{
					nullRackTypeViolation("manufacturer"),
					nullRackTypeViolation("model"),
					nullRackTypeViolation("slug"),
					nullRackTypeViolation("form_factor"),
					blankRackTypeWidthViolation(),
					nullRackTypeViolation("u_height"),
					nullRackTypeViolation("starting_unit"),
					nullRackTypeViolation("desc_units"),
					nullRackTypeViolation("description"),
					nullRackTypeViolation("comments"),
				}, shared.ViolationsOf(err))
			})
		}
	})

	t.Run("PATCH omissions preserve their field while a concrete sibling remains", func(t *testing.T) {
		t.Parallel()
		for _, field := range []string{
			"manufacturer", "model", "slug", "form_factor", "width", "u_height",
			"starting_unit", "desc_units", "description", "comments",
		} {
			field := field
			t.Run(field, func(t *testing.T) {
				t.Parallel()
				command := UpdateRackTypeCommand{ID: commandID}
				if field == "comments" {
					command.Description = FieldValue("changed sibling")
				} else {
					command.Comments = FieldValue("changed sibling")
				}
				patch, err := command.patch()
				require.NoError(t, err)
				assertRackTypePatchFieldNil(t, patch, field)
			})
		}
	})

	t.Run("PATCH retains zero false blank and concrete values", func(t *testing.T) {
		t.Parallel()
		patch, err := (UpdateRackTypeCommand{
			ID: commandID, Manufacturer: FieldValue(shared.ID(9)),
			Model: FieldValue(""), Slug: FieldValue(""), FormFactor: FieldValue(""),
			Width: FieldValue(uint32(0)), UHeight: FieldValue(uint32(0)),
			StartingUnit: FieldValue(uint32(0)), DescUnits: FieldValue(false),
			Description: FieldValue(""), Comments: FieldValue(""),
		}).patch()
		require.NoError(t, err)
		assert.Equal(t, shared.ID(9), *patch.manufacturerID)
		assert.Equal(t, "", *patch.model)
		assert.Equal(t, "", *patch.slug)
		assert.Equal(t, "", *patch.formFactor)
		assert.Equal(t, uint32(0), *patch.width)
		assert.Equal(t, uint32(0), *patch.uHeight)
		assert.Equal(t, uint32(0), *patch.startingUnit)
		assert.False(t, *patch.descUnits)
		assert.Equal(t, "", *patch.description)
		assert.Equal(t, "", *patch.comments)
	})
}

const commandID shared.ID = 1

func allNullRackTypeCommand() CreateRackTypeCommand {
	return CreateRackTypeCommand{
		Manufacturer: NullField[shared.ID](), Model: NullField[string](),
		Slug: NullField[string](), FormFactor: NullField[string](),
		Width: NullField[uint32](), UHeight: NullField[uint32](),
		StartingUnit: NullField[uint32](), DescUnits: NullField[bool](),
		Description: NullField[string](), Comments: NullField[string](),
	}
}

func nullRackTypeViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "null", Description: "This field may not be null.",
	}
}

func rackTypeDomainValues(values rackTypeCommandValues) dcimdomain.RackTypeValues {
	return dcimdomain.RackTypeValues{
		Model: values.model, Slug: values.slug, FormFactor: values.formFactor,
		Width: values.width, UHeight: values.uHeight, StartingUnit: values.startingUnit,
		DescUnits: values.descUnits, Description: values.description, Comments: values.comments,
	}
}

func violationFields(err error) []string {
	violations := shared.ViolationsOf(err)
	fields := make([]string, len(violations))
	for index, violation := range violations {
		fields[index] = violation.Field
	}
	return fields
}

func assertRackTypePatchFieldNil(t *testing.T, patch rackTypeCommandPatch, field string) {
	t.Helper()
	switch field {
	case "manufacturer":
		assert.Nil(t, patch.manufacturerID)
	case "model":
		assert.Nil(t, patch.model)
	case "slug":
		assert.Nil(t, patch.slug)
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
	case "description":
		assert.Nil(t, patch.description)
	case "comments":
		assert.Nil(t, patch.comments)
	}
}
