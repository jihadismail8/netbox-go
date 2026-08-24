package dcim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var rackRoleCommandTestTime = shared.NewTimestamp(
	time.Date(2026, time.August, 24, 16, 30, 0, 0, time.UTC),
)

type rackRoleScalarCommandField struct {
	name     string
	concrete string
}

var rackRoleScalarCommandFields = []rackRoleScalarCommandField{
	{name: "name", concrete: "Concrete Rack Role"},
	{name: "slug", concrete: "concrete-rack-role"},
	{name: "color", concrete: "00ff00"},
	{name: "description", concrete: "Distribution racks"},
}

func TestRackRoleScalarCommandPresenceMatrix(t *testing.T) {
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
		for _, field := range rackRoleScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineCreateRackRoleCommand()
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setCreateRackRoleCommandField(&command, field.name, value)

					values, commandErr := command.values()
					assertRackRolePresenceViolation(
						t, commandErr, expectedRackRolePresenceViolation("POST", field.name, state.name),
					)
					assert.Equal(t, expectedCreateRackRoleValues(field, state.name), values)
					if commandErr != nil {
						return
					}
					_, domainErr := dcimdomain.NewRackRole(values, rackRoleCommandTestTime)
					assertRackRoleDomainOutcome(t, domainErr, field.name, state.name)
				})
			}
		}
	})

	t.Run("PUT", func(t *testing.T) {
		for _, field := range rackRoleScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := baselineReplaceRackRoleCommand()
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setReplaceRackRoleCommandField(&command, field.name, value)

					patch, commandErr := command.patch()
					assertRackRolePresenceViolation(
						t, commandErr, expectedRackRolePresenceViolation("PUT", field.name, state.name),
					)
					assert.Equal(t, expectedReplaceRackRolePatch(field, state.name), patch)
					if commandErr != nil {
						return
					}
					candidate := newRackRoleCommandTestAggregate(t)
					domainErr := candidate.ApplyPatch(patch, rackRoleCommandTestTimeAfter())
					assertRackRoleDomainOutcome(t, domainErr, field.name, state.name)
				})
			}
		}

		patch, err := (ReplaceRackRoleCommand{ID: 1}).patch()
		assert.Equal(t, dcimdomain.RackRolePatch{}, patch)
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
	})

	t.Run("PATCH", func(t *testing.T) {
		for _, field := range rackRoleScalarCommandFields {
			field := field
			for _, state := range states {
				state := state
				t.Run(field.name+"/"+state.name, func(t *testing.T) {
					command := UpdateRackRoleCommand{ID: 1}
					expectedPatch := expectedUpdateRackRolePatch(field, state.name)
					if state.name == "omitted" {
						addRackRolePatchSibling(&command, &expectedPatch, field.name)
					}
					value := state.value
					if state.name == "concrete" {
						value = FieldValue(field.concrete)
					}
					setUpdateRackRoleCommandField(&command, field.name, value)

					patch, commandErr := command.patch()
					assertRackRolePresenceViolation(
						t, commandErr, expectedRackRolePresenceViolation("PATCH", field.name, state.name),
					)
					assert.Equal(t, expectedPatch, patch)
					if commandErr != nil {
						return
					}
					candidate := newRackRoleCommandTestAggregate(t)
					before := candidate.Values()
					domainErr := candidate.ApplyPatch(patch, rackRoleCommandTestTimeAfter())
					assertRackRoleDomainOutcome(t, domainErr, field.name, state.name)
					if state.name == "omitted" {
						assertRackRoleOmittedScalarPreserved(t, candidate.Values(), before, field.name)
					}
				})
			}
		}

		patch, err := (UpdateRackRoleCommand{
			ID: 1, Name: NullField[string](), Color: FieldValue("00ff00"),
		}).patch()
		assert.Equal(t, []shared.FieldViolation{{
			Field: "name", Reason: "null", Description: "This field may not be null.",
		}}, shared.ViolationsOf(err))
		assert.Nil(t, patch.Name)
		require.NotNil(t, patch.Color)
		assert.Equal(t, "00ff00", *patch.Color)
	})
}

func baselineCreateRackRoleCommand() CreateRackRoleCommand {
	return CreateRackRoleCommand{Name: FieldValue("Rack Role"), Slug: FieldValue("rack-role")}
}

func baselineReplaceRackRoleCommand() ReplaceRackRoleCommand {
	return ReplaceRackRoleCommand{
		ID: 1, Name: FieldValue("Replacement Rack Role"), Slug: FieldValue("replacement-rack-role"),
	}
}

func setCreateRackRoleCommandField(command *CreateRackRoleCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "color":
		command.Color = value
	case "description":
		command.Description = value
	}
}

func setReplaceRackRoleCommandField(command *ReplaceRackRoleCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "color":
		command.Color = value
	case "description":
		command.Description = value
	}
}

func setUpdateRackRoleCommandField(command *UpdateRackRoleCommand, field string, value Field[string]) {
	switch field {
	case "name":
		command.Name = value
	case "slug":
		command.Slug = value
	case "color":
		command.Color = value
	case "description":
		command.Description = value
	}
}

func addRackRolePatchSibling(
	command *UpdateRackRoleCommand,
	expected *dcimdomain.RackRolePatch,
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

func expectedRackRolePresenceViolation(operation, field, state string) *shared.FieldViolation {
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

func expectedCreateRackRoleValues(
	field rackRoleScalarCommandField,
	state string,
) dcimdomain.RackRoleValues {
	values := dcimdomain.RackRoleValues{
		Name: "Rack Role", Slug: "rack-role", Color: dcimdomain.RackRoleDefaultColor,
	}
	value := ""
	if state == "concrete" {
		value = field.concrete
	} else if field.name == "color" && (state == "omitted" || state == "null") {
		value = dcimdomain.RackRoleDefaultColor
	}
	setRackRoleValuesField(&values, field.name, value)
	return values
}

func expectedReplaceRackRolePatch(
	field rackRoleScalarCommandField,
	state string,
) dcimdomain.RackRolePatch {
	name := "Replacement Rack Role"
	slug := "replacement-rack-role"
	patch := dcimdomain.RackRolePatch{Name: &name, Slug: &slug}
	setExpectedRackRolePatchField(&patch, field, state)
	return patch
}

func expectedUpdateRackRolePatch(
	field rackRoleScalarCommandField,
	state string,
) dcimdomain.RackRolePatch {
	patch := dcimdomain.RackRolePatch{}
	setExpectedRackRolePatchField(&patch, field, state)
	return patch
}

func setExpectedRackRolePatchField(
	patch *dcimdomain.RackRolePatch,
	field rackRoleScalarCommandField,
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
	case "color":
		patch.Color = value
	case "description":
		patch.Description = value
	}
}

func setRackRoleValuesField(values *dcimdomain.RackRoleValues, field, value string) {
	switch field {
	case "name":
		values.Name = value
	case "slug":
		values.Slug = value
	case "color":
		values.Color = value
	case "description":
		values.Description = value
	}
}

func newRackRoleCommandTestAggregate(t *testing.T) *dcimdomain.RackRole {
	t.Helper()
	role, err := dcimdomain.NewRackRole(dcimdomain.RackRoleValues{
		Name: "Original Rack Role", Slug: "original-rack-role", Color: "123abc",
		Description: "Original description",
	}, rackRoleCommandTestTime)
	require.NoError(t, err)
	return role
}

func rackRoleCommandTestTimeAfter() shared.Timestamp {
	return shared.NewTimestamp(rackRoleCommandTestTime.Add(time.Minute))
}

func assertRackRolePresenceViolation(
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

func assertRackRoleDomainOutcome(t *testing.T, err error, field, state string) {
	t.Helper()
	if state == "blank" && (field == "name" || field == "slug" || field == "color") {
		require.Error(t, err)
		assert.Equal(t, []shared.FieldViolation{{
			Field: field, Reason: "required", Description: "This field may not be blank.",
		}}, shared.ViolationsOf(err))
		return
	}
	require.NoError(t, err)
}

func assertRackRoleOmittedScalarPreserved(
	t *testing.T,
	after dcimdomain.RackRoleValues,
	before dcimdomain.RackRoleValues,
	field string,
) {
	t.Helper()
	switch field {
	case "name":
		assert.Equal(t, before.Name, after.Name)
	case "slug":
		assert.Equal(t, before.Slug, after.Slug)
	case "color":
		assert.Equal(t, before.Color, after.Color)
	case "description":
		assert.Equal(t, before.Description, after.Description)
	}
}
