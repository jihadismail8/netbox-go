package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateManufacturerCommand struct {
	Name        Field[string]
	Slug        Field[string]
	Description Field[string]
}

type ReplaceManufacturerCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Description Field[string]
}

type UpdateManufacturerCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Description Field[string]
}

type DeleteManufacturerCommand struct{ ID shared.ID }

func (command CreateManufacturerCommand) values() (dcimdomain.ManufacturerValues, error) {
	return fullManufacturerValues(command.Name, command.Slug, command.Description)
}

func (command ReplaceManufacturerCommand) values() (dcimdomain.ManufacturerValues, error) {
	return fullManufacturerValues(command.Name, command.Slug, command.Description)
}

func fullManufacturerValues(
	name Field[string],
	slug Field[string],
	description Field[string],
) (dcimdomain.ManufacturerValues, error) {
	var violations []shared.FieldViolation
	values := dcimdomain.ManufacturerValues{
		Name:        valueForFullMutation(&violations, "name", name, "", true),
		Slug:        valueForFullMutation(&violations, "slug", slug, "", true),
		Description: valueForFullMutation(&violations, "description", description, "", false),
	}
	if len(violations) > 0 {
		return dcimdomain.ManufacturerValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command UpdateManufacturerCommand) patch() (dcimdomain.ManufacturerPatch, error) {
	var violations []shared.FieldViolation
	patch := dcimdomain.ManufacturerPatch{
		Name:        patchValue(&violations, "name", command.Name),
		Slug:        patchValue(&violations, "slug", command.Slug),
		Description: patchValue(&violations, "description", command.Description),
	}
	if len(violations) > 0 {
		return dcimdomain.ManufacturerPatch{}, shared.NewValidationError(violations...)
	}
	return patch, nil
}
