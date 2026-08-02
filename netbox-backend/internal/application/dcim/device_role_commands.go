package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateDeviceRoleCommand struct {
	Parent      Field[shared.ID]
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	VMRole      Field[bool]
	Description Field[string]
	Comments    Field[string]
}

type ReplaceDeviceRoleCommand struct {
	ID          shared.ID
	Parent      Field[shared.ID]
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	VMRole      Field[bool]
	Description Field[string]
	Comments    Field[string]
}

type UpdateDeviceRoleCommand struct {
	ID          shared.ID
	Parent      Field[shared.ID]
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	VMRole      Field[bool]
	Description Field[string]
	Comments    Field[string]
}

type DeleteDeviceRoleCommand struct{ ID shared.ID }

func (command CreateDeviceRoleCommand) values() (dcimdomain.DeviceRoleValues, error) {
	return fullDeviceRoleValues(
		command.Parent, command.Name, command.Slug, command.Color, command.VMRole,
		command.Description, command.Comments,
	)
}

func (command ReplaceDeviceRoleCommand) values() (dcimdomain.DeviceRoleValues, error) {
	return fullDeviceRoleValues(
		command.Parent, command.Name, command.Slug, command.Color, command.VMRole,
		command.Description, command.Comments,
	)
}

func fullDeviceRoleValues(
	parent Field[shared.ID],
	name Field[string],
	slug Field[string],
	color Field[string],
	vmRole Field[bool],
	description Field[string],
	comments Field[string],
) (dcimdomain.DeviceRoleValues, error) {
	var violations []shared.FieldViolation
	values := dcimdomain.DeviceRoleValues{
		Parent: deviceRoleParentForFullMutation(&violations, parent),
		Name:   valueForFullMutation(&violations, "name", name, "", true),
		Slug:   valueForFullMutation(&violations, "slug", slug, "", true),
		Color: valueForFullMutation(
			&violations, "color", color, dcimdomain.DeviceRoleDefaultColor, false,
		),
		VMRole:      boolForFullMutation(&violations, "vm_role", vmRole, true),
		Description: valueForFullMutation(&violations, "description", description, "", false),
		Comments:    valueForFullMutation(&violations, "comments", comments, "", false),
	}
	if len(violations) > 0 {
		return dcimdomain.DeviceRoleValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func deviceRoleParentForFullMutation(
	violations *[]shared.FieldViolation,
	field Field[shared.ID],
) dcimdomain.DeviceRoleParent {
	if field.State() != FieldPresent {
		return dcimdomain.RootDeviceRoleParent()
	}
	id, _ := field.Get()
	if !id.IsValid() {
		*violations = append(*violations, shared.FieldViolation{
			Field: "parent", Reason: "invalid", Description: "A valid object ID is required.",
		})
	}
	return dcimdomain.NonRootDeviceRoleParent(id)
}

func boolForFullMutation(
	violations *[]shared.FieldViolation,
	name string,
	field Field[bool],
	fallback bool,
) bool {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return value
	case FieldNull:
		*violations = append(*violations, shared.FieldViolation{
			Field: name, Reason: "null", Description: "This field may not be null.",
		})
	}
	return fallback
}

func (command UpdateDeviceRoleCommand) patch() (dcimdomain.DeviceRolePatch, error) {
	var violations []shared.FieldViolation
	patch := dcimdomain.DeviceRolePatch{
		Parent:      deviceRoleParentPatch(&violations, command.Parent),
		Name:        patchValue(&violations, "name", command.Name),
		Slug:        patchValue(&violations, "slug", command.Slug),
		Color:       patchValue(&violations, "color", command.Color),
		VMRole:      boolPatchValue(&violations, "vm_role", command.VMRole),
		Description: patchValue(&violations, "description", command.Description),
		Comments:    patchValue(&violations, "comments", command.Comments),
	}
	if len(violations) > 0 {
		return dcimdomain.DeviceRolePatch{}, shared.NewValidationError(violations...)
	}
	return patch, nil
}

func deviceRoleParentPatch(
	violations *[]shared.FieldViolation,
	field Field[shared.ID],
) *dcimdomain.DeviceRoleParent {
	switch field.State() {
	case FieldNull:
		value := dcimdomain.RootDeviceRoleParent()
		return &value
	case FieldPresent:
		id, _ := field.Get()
		if !id.IsValid() {
			*violations = append(*violations, shared.FieldViolation{
				Field: "parent", Reason: "invalid", Description: "A valid object ID is required.",
			})
		}
		value := dcimdomain.NonRootDeviceRoleParent(id)
		return &value
	default:
		return nil
	}
}

func boolPatchValue(
	violations *[]shared.FieldViolation,
	name string,
	field Field[bool],
) *bool {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return &value
	case FieldNull:
		*violations = append(*violations, shared.FieldViolation{
			Field: name, Reason: "null", Description: "This field may not be null.",
		})
	}
	return nil
}
