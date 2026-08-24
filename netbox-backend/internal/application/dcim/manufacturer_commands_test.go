package dcim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var manufacturerCommandTestTime = shared.NewTimestamp(
	time.Date(2026, time.August, 24, 7, 0, 0, 0, time.UTC),
)

type manufacturerScalarCommandField struct {
	name     string
	concrete string
}

var manufacturerScalarCommandFields = []manufacturerScalarCommandField{
	{name: "name", concrete: "Concrete Manufacturer"},
	{name: "slug", concrete: "concrete-manufacturer"},
	{name: "description", concrete: "Hardware vendor"},
}

func TestManufacturerScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	states := []struct {
		name  string
		value Field[string]
	}{
		{name: "omitted", value: OmittedField[string]()},
		{name: "null", value: NullField[string]()},
		{name: "blank", value: FieldValue("")},
		{name: "concrete", value: FieldValue("placeholder")},
	}

	t.Run("POST", func(t *testing.T) {
		for _, field := range manufacturerScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineCreateManufacturerCommand()
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setCreateManufacturerCommandField(&command, field.name, value)

					values, commandErr := command.values()
					assertManufacturerPresenceViolation(
						t, commandErr, expectedManufacturerPresenceViolation("POST", field.name, state.name),
					)
					assert.Equal(t, expectedCreateManufacturerValues(field, state.name), values)
					if commandErr != nil {
						return
					}
					_, domainErr := dcimdomain.NewManufacturer(values, manufacturerCommandTestTime)
					assertManufacturerDomainOutcome(t, domainErr, field.name, state.name)
				})
			}
		}
	})

	t.Run("PUT", func(t *testing.T) {
		for _, field := range manufacturerScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineReplaceManufacturerCommand()
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setReplaceManufacturerCommandField(&command, field.name, value)

					patch, commandErr := command.patch()
					assertManufacturerPresenceViolation(
						t, commandErr, expectedManufacturerPresenceViolation("PUT", field.name, state.name),
					)
					assert.Equal(t, expectedReplaceManufacturerPatch(field, state.name), patch)
					if commandErr != nil {
						return
					}
					candidate := newManufacturerCommandTestAggregate(t)
					domainErr := candidate.ApplyPatch(patch, manufacturerCommandTestTimeAfter())
					assertManufacturerDomainOutcome(t, domainErr, field.name, state.name)
				})
			}
		}

		patch, err := (ReplaceManufacturerCommand{ID: 1}).patch()
		assert.Equal(t, dcimdomain.ManufacturerPatch{}, patch)
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
	})

	t.Run("PATCH", func(t *testing.T) {
		for _, field := range manufacturerScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := UpdateManufacturerCommand{ID: 1}
					expectedPatch := expectedUpdateManufacturerPatch(field, state.name)
					if state.name == "omitted" {
						addManufacturerPatchSibling(&command, &expectedPatch, field.name)
					}
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setUpdateManufacturerCommandField(&command, field.name, value)

					patch, commandErr := command.patch()
					assertManufacturerPresenceViolation(
						t, commandErr, expectedManufacturerPresenceViolation("PATCH", field.name, state.name),
					)
					assert.Equal(t, expectedPatch, patch)
					if commandErr != nil {
						return
					}
					candidate := newManufacturerCommandTestAggregate(t)
					before := candidate.Values()
					domainErr := candidate.ApplyPatch(patch, manufacturerCommandTestTimeAfter())
					assertManufacturerDomainOutcome(t, domainErr, field.name, state.name)
					if state.name == "omitted" {
						assertManufacturerOmittedScalarPreserved(t, candidate.Values(), before, field.name)
					}
				})
			}
		}

		patch, err := (UpdateManufacturerCommand{
			ID: 1, Name: NullField[string](), Description: FieldValue("retained sibling"),
		}).patch()
		assert.Equal(t, []shared.FieldViolation{{
			Field: "name", Reason: "null", Description: "This field may not be null.",
		}}, shared.ViolationsOf(err))
		assert.Nil(t, patch.Name)
		require.NotNil(t, patch.Description)
		assert.Equal(t, "retained sibling", *patch.Description)
	})
}

func baselineCreateManufacturerCommand() CreateManufacturerCommand {
	return CreateManufacturerCommand{Name: FieldValue("Manufacturer"), Slug: FieldValue("manufacturer")}
}

func baselineReplaceManufacturerCommand() ReplaceManufacturerCommand {
	return ReplaceManufacturerCommand{
		ID: 1, Name: FieldValue("Replacement Manufacturer"), Slug: FieldValue("replacement-manufacturer"),
	}
}

func setCreateManufacturerCommandField(command *CreateManufacturerCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "description":
		command.Description = value
	}
}

func setReplaceManufacturerCommandField(command *ReplaceManufacturerCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "description":
		command.Description = value
	}
}

func setUpdateManufacturerCommandField(command *UpdateManufacturerCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "description":
		command.Description = value
	}
}

func addManufacturerPatchSibling(
	command *UpdateManufacturerCommand,
	expected *dcimdomain.ManufacturerPatch,
	omittedField string,
) {
	if omittedField == "description" {
		value := "Retained sibling"
		command.Name = FieldValue(value)
		expected.Name = &value
		return
	}
	value := "retained sibling"
	command.Description = FieldValue(value)
	expected.Description = &value
}

func expectedManufacturerPresenceViolation(operation, field, state string) *shared.FieldViolation {
	if state == "omitted" && operation != "PATCH" && (field == "name" || field == "slug") {
		return &shared.FieldViolation{
			Field: field, Reason: "required", Description: "This field is required.",
		}
	}
	if state == "null" {
		return &shared.FieldViolation{
			Field: field, Reason: "null", Description: "This field may not be null.",
		}
	}
	return nil
}

func expectedCreateManufacturerValues(
	field manufacturerScalarCommandField,
	state string,
) dcimdomain.ManufacturerValues {
	values := dcimdomain.ManufacturerValues{Name: "Manufacturer", Slug: "manufacturer"}
	value := ""
	if state == "concrete" {
		value = field.concrete
	}
	setManufacturerValuesField(&values, field.name, value)
	return values
}

func expectedReplaceManufacturerPatch(
	field manufacturerScalarCommandField,
	state string,
) dcimdomain.ManufacturerPatch {
	name := "Replacement Manufacturer"
	slug := "replacement-manufacturer"
	patch := dcimdomain.ManufacturerPatch{Name: &name, Slug: &slug}
	setExpectedManufacturerPatchField(&patch, field, state)
	return patch
}

func expectedUpdateManufacturerPatch(
	field manufacturerScalarCommandField,
	state string,
) dcimdomain.ManufacturerPatch {
	patch := dcimdomain.ManufacturerPatch{}
	setExpectedManufacturerPatchField(&patch, field, state)
	return patch
}

func setExpectedManufacturerPatchField(
	patch *dcimdomain.ManufacturerPatch,
	field manufacturerScalarCommandField,
	state string,
) {
	var value *string
	switch state {
	case "blank":
		blank := ""
		value = &blank
	case "concrete":
		concrete := field.concrete
		value = &concrete
	}
	switch field.name {
	case "name":
		patch.Name = value
	case "slug":
		patch.Slug = value
	case "description":
		patch.Description = value
	}
}

func setManufacturerValuesField(values *dcimdomain.ManufacturerValues, field, value string) {
	switch field {
	case "name":
		values.Name = value
	case "slug":
		values.Slug = value
	case "description":
		values.Description = value
	}
}

func newManufacturerCommandTestAggregate(t *testing.T) *dcimdomain.Manufacturer {
	t.Helper()
	manufacturer, err := dcimdomain.NewManufacturer(dcimdomain.ManufacturerValues{
		Name: "Original Manufacturer", Slug: "original-manufacturer",
		Description: "Original description",
	}, manufacturerCommandTestTime)
	require.NoError(t, err)
	return manufacturer
}

func manufacturerCommandTestTimeAfter() shared.Timestamp {
	return shared.NewTimestamp(manufacturerCommandTestTime.Add(time.Minute))
}

func assertManufacturerPresenceViolation(
	t *testing.T,
	err error,
	expected *shared.FieldViolation,
) {
	t.Helper()
	if expected == nil {
		require.NoError(t, err)
		return
	}
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{*expected}, shared.ViolationsOf(err))
}

func assertManufacturerDomainOutcome(t *testing.T, err error, field, state string) {
	t.Helper()
	if state == "blank" && (field == "name" || field == "slug") {
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{{
			Field: field, Reason: "required", Description: "This field may not be blank.",
		}}, shared.ViolationsOf(err))
		return
	}
	require.NoError(t, err)
}

func assertManufacturerOmittedScalarPreserved(
	t *testing.T,
	after dcimdomain.ManufacturerValues,
	before dcimdomain.ManufacturerValues,
	field string,
) {
	t.Helper()
	switch field {
	case "name":
		assert.Equal(t, before.Name, after.Name)
	case "slug":
		assert.Equal(t, before.Slug, after.Slug)
	case "description":
		assert.Equal(t, before.Description, after.Description)
	}
}
