package dcim

import (
	"sort"

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
		formFactor:     valueForFullMutation(&violations, "form_factor", command.FormFactor, "", false),
		width:          fullRackTypeWidth(&violations, command.Width),
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
		return values, newRackTypeValidationError(violations)
	}
	return values, nil
}

func fullRackTypeWidth(violations *[]shared.FieldViolation, field Field[uint32]) uint32 {
	if field.State() == FieldNull {
		*violations = append(*violations, blankRackTypeWidthViolation())
		return dcimdomain.RackTypeDefaultWidth
	}
	return fullFieldValue(
		violations, "width", field, dcimdomain.RackTypeDefaultWidth, false,
	)
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

func (command ReplaceRackTypeCommand) patch() (rackTypeCommandPatch, error) {
	return buildRackTypePatch(
		command.Manufacturer, command.Model, command.Slug, command.FormFactor,
		command.Width, command.UHeight, command.StartingUnit, command.DescUnits,
		command.Description, command.Comments, true,
	)
}

func (command UpdateRackTypeCommand) patch() (rackTypeCommandPatch, error) {
	return buildRackTypePatch(
		command.Manufacturer, command.Model, command.Slug, command.FormFactor,
		command.Width, command.UHeight, command.StartingUnit, command.DescUnits,
		command.Description, command.Comments, false,
	)
}

func buildRackTypePatch(
	manufacturer Field[shared.ID],
	model Field[string],
	slug Field[string],
	formFactor Field[string],
	width Field[uint32],
	uHeight Field[uint32],
	startingUnit Field[uint32],
	descUnits Field[bool],
	description Field[string],
	comments Field[string],
	requireIdentity bool,
) (rackTypeCommandPatch, error) {
	var violations []shared.FieldViolation
	if requireIdentity {
		requireRackTypeField(&violations, "manufacturer", manufacturer)
		requireRackTypeField(&violations, "model", model)
		requireRackTypeField(&violations, "slug", slug)
	}
	patch := rackTypeCommandPatch{
		manufacturerID: patchFieldValue(&violations, "manufacturer", manufacturer),
		model:          patchValue(&violations, "model", model),
		slug:           patchValue(&violations, "slug", slug),
		formFactor:     patchValue(&violations, "form_factor", formFactor),
		width:          patchRackTypeWidth(&violations, width),
		uHeight:        patchFieldValue(&violations, "u_height", uHeight),
		startingUnit:   patchFieldValue(&violations, "starting_unit", startingUnit),
		descUnits:      patchFieldValue(&violations, "desc_units", descUnits),
		description:    patchValue(&violations, "description", description),
		comments:       patchValue(&violations, "comments", comments),
	}
	if patch.manufacturerID != nil && !patch.manufacturerID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return patch, newRackTypeValidationError(violations)
	}
	if patch.empty() {
		return patch, newRackTypeValidationError([]shared.FieldViolation{{
			Field: "update_mask", Reason: "required", Description: "At least one writable field must be supplied.",
		}})
	}
	return patch, nil
}

func requireRackTypeField[T any](
	violations *[]shared.FieldViolation,
	name string,
	field Field[T],
) {
	if field.State() != FieldOmitted {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field: name, Reason: "required", Description: "This field is required.",
	})
}

func patchRackTypeWidth(
	violations *[]shared.FieldViolation,
	field Field[uint32],
) *uint32 {
	if field.State() == FieldNull {
		*violations = append(*violations, blankRackTypeWidthViolation())
		return nil
	}
	return patchFieldValue(violations, "width", field)
}

func blankRackTypeWidthViolation() shared.FieldViolation {
	return shared.FieldViolation{
		Field: "width", Reason: "blank", Description: "This field may not be blank.",
	}
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

var rackTypeValidationFieldOrder = map[string]int{
	"manufacturer":  0,
	"model":         1,
	"slug":          2,
	"form_factor":   3,
	"width":         4,
	"u_height":      5,
	"starting_unit": 6,
	"desc_units":    7,
	"description":   8,
	"comments":      9,
	"update_mask":   10,
}

func newRackTypeValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := rackTypeValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := rackTypeValidationFieldOrder[ordered[right].Field]
		switch {
		case leftKnown && rightKnown:
			return leftOrder < rightOrder
		case leftKnown:
			return true
		case rightKnown:
			return false
		default:
			return false
		}
	})
	return shared.NewValidationError(ordered...)
}

func mergeRackTypeMutationErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil && !shared.HasReason(err, shared.ErrorReasonValidation) {
			return err
		}
	}

	seenFields := make(map[string]struct{})
	var merged []shared.FieldViolation
	for _, err := range errs {
		for _, violation := range shared.ViolationsOf(err) {
			if _, duplicate := seenFields[violation.Field]; duplicate {
				continue
			}
			seenFields[violation.Field] = struct{}{}
			merged = append(merged, violation)
		}
	}
	if len(merged) > 0 {
		return newRackTypeValidationError(merged)
	}
	return nil
}
