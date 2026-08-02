package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateRackTypeCommand struct {
	Manufacturer Field[shared.ID]
	Model        Field[string]
	Slug         Field[string]
	FormFactor   Field[string]
	Width        Field[uint32]
	UHeight      Field[uint32]
	StartingUnit Field[uint32]
	DescUnits    Field[bool]
	Description  Field[string]
	Comments     Field[string]
}

type ReplaceRackTypeCommand struct {
	ID shared.ID
	CreateRackTypeCommand
}

type UpdateRackTypeCommand struct {
	ID           shared.ID
	Manufacturer Field[shared.ID]
	Model        Field[string]
	Slug         Field[string]
	FormFactor   Field[string]
	Width        Field[uint32]
	UHeight      Field[uint32]
	StartingUnit Field[uint32]
	DescUnits    Field[bool]
	Description  Field[string]
	Comments     Field[string]
}

type DeleteRackTypeCommand struct{ ID shared.ID }

type rackTypeCommandValues struct {
	manufacturerID shared.ID
	model          string
	slug           string
	formFactor     string
	width          uint32
	uHeight        uint32
	startingUnit   uint32
	descUnits      bool
	description    string
	comments       string
}

func (command CreateRackTypeCommand) values() (rackTypeCommandValues, error) {
	var violations []shared.FieldViolation
	values := rackTypeCommandValues{
		manufacturerID: fullFieldValue(&violations, "manufacturer", command.Manufacturer, shared.ID(0), true),
		model:          valueForFullMutation(&violations, "model", command.Model, "", true),
		slug:           valueForFullMutation(&violations, "slug", command.Slug, "", true),
		formFactor:     valueForFullMutation(&violations, "form_factor", command.FormFactor, "", true),
		width:          fullFieldValue(&violations, "width", command.Width, dcimdomain.RackTypeDefaultWidth, false),
		uHeight:        fullFieldValue(&violations, "u_height", command.UHeight, dcimdomain.RackTypeDefaultUHeight, false),
		startingUnit:   fullFieldValue(&violations, "starting_unit", command.StartingUnit, dcimdomain.RackTypeDefaultStartingUnit, false),
		descUnits:      fullFieldValue(&violations, "desc_units", command.DescUnits, false, false),
		description:    valueForFullMutation(&violations, "description", command.Description, "", false),
		comments:       valueForFullMutation(&violations, "comments", command.Comments, "", false),
	}
	if values.manufacturerID <= 0 && command.Manufacturer.State() == FieldPresent {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return rackTypeCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command ReplaceRackTypeCommand) values() (rackTypeCommandValues, error) {
	return command.CreateRackTypeCommand.values()
}

type rackTypeCommandPatch struct {
	manufacturerID *shared.ID
	model          *string
	slug           *string
	formFactor     *string
	width          *uint32
	uHeight        *uint32
	startingUnit   *uint32
	descUnits      *bool
	description    *string
	comments       *string
}

func (patch rackTypeCommandPatch) empty() bool {
	return patch.manufacturerID == nil && patch.model == nil && patch.slug == nil &&
		patch.formFactor == nil && patch.width == nil && patch.uHeight == nil &&
		patch.startingUnit == nil && patch.descUnits == nil && patch.description == nil &&
		patch.comments == nil
}

func (command UpdateRackTypeCommand) patch() (rackTypeCommandPatch, error) {
	var violations []shared.FieldViolation
	patch := rackTypeCommandPatch{
		manufacturerID: patchFieldValue(&violations, "manufacturer", command.Manufacturer),
		model:          patchValue(&violations, "model", command.Model),
		slug:           patchValue(&violations, "slug", command.Slug),
		formFactor:     patchValue(&violations, "form_factor", command.FormFactor),
		width:          patchFieldValue(&violations, "width", command.Width),
		uHeight:        patchFieldValue(&violations, "u_height", command.UHeight),
		startingUnit:   patchFieldValue(&violations, "starting_unit", command.StartingUnit),
		descUnits:      patchFieldValue(&violations, "desc_units", command.DescUnits),
		description:    patchValue(&violations, "description", command.Description),
		comments:       patchValue(&violations, "comments", command.Comments),
	}
	if patch.manufacturerID != nil && !patch.manufacturerID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return rackTypeCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return rackTypeCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func fullFieldValue[T any](
	violations *[]shared.FieldViolation,
	name string,
	field Field[T],
	fallback T,
	required bool,
) T {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return value
	case FieldNull:
		*violations = append(*violations, shared.FieldViolation{
			Field: name, Reason: "null", Description: "This field may not be null.",
		})
	case FieldOmitted:
		if required {
			*violations = append(*violations, shared.FieldViolation{
				Field: name, Reason: "required", Description: "This field is required.",
			})
		}
	}
	return fallback
}

func patchFieldValue[T any](
	violations *[]shared.FieldViolation,
	name string,
	field Field[T],
) *T {
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
