package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateRackRoleCommand struct {
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	Description Field[string]
}

type ReplaceRackRoleCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	Description Field[string]
}

type UpdateRackRoleCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Color       Field[string]
	Description Field[string]
}

type DeleteRackRoleCommand struct{ ID shared.ID }

func (command CreateRackRoleCommand) values() (dcimdomain.RackRoleValues, error) {
	return fullRackRoleValues(command.Name, command.Slug, command.Color, command.Description)
}

func (command ReplaceRackRoleCommand) values() (dcimdomain.RackRoleValues, error) {
	return fullRackRoleValues(command.Name, command.Slug, command.Color, command.Description)
}

func fullRackRoleValues(
	name Field[string],
	slug Field[string],
	color Field[string],
	description Field[string],
) (dcimdomain.RackRoleValues, error) {
	var violations []shared.FieldViolation
	values := dcimdomain.RackRoleValues{
		Name: valueForFullMutation(&violations, "name", name, "", true),
		Slug: valueForFullMutation(&violations, "slug", slug, "", true),
		Color: valueForFullMutation(
			&violations, "color", color, dcimdomain.RackRoleDefaultColor, false,
		),
		Description: valueForFullMutation(&violations, "description", description, "", false),
	}
	if len(violations) > 0 {
		return dcimdomain.RackRoleValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command UpdateRackRoleCommand) patch() (dcimdomain.RackRolePatch, error) {
	var violations []shared.FieldViolation
	patch := dcimdomain.RackRolePatch{
		Name:        patchValue(&violations, "name", command.Name),
		Slug:        patchValue(&violations, "slug", command.Slug),
		Color:       patchValue(&violations, "color", command.Color),
		Description: patchValue(&violations, "description", command.Description),
	}
	if len(violations) > 0 {
		return dcimdomain.RackRolePatch{}, shared.NewValidationError(violations...)
	}
	return patch, nil
}
