package dcim

import (
	"sort"

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
		return values, newManufacturerValidationError(violations)
	}
	return values, nil
}

func (command ReplaceManufacturerCommand) patch() (dcimdomain.ManufacturerPatch, error) {
	return buildManufacturerPatch(command.Name, command.Slug, command.Description, true)
}

func (command UpdateManufacturerCommand) patch() (dcimdomain.ManufacturerPatch, error) {
	return buildManufacturerPatch(command.Name, command.Slug, command.Description, false)
}

func buildManufacturerPatch(
	name Field[string],
	slug Field[string],
	description Field[string],
	requireIdentity bool,
) (dcimdomain.ManufacturerPatch, error) {
	var violations []shared.FieldViolation
	if requireIdentity {
		requireManufacturerField(&violations, "name", name)
		requireManufacturerField(&violations, "slug", slug)
	}
	patch := dcimdomain.ManufacturerPatch{
		Name:        patchValue(&violations, "name", name),
		Slug:        patchValue(&violations, "slug", slug),
		Description: patchValue(&violations, "description", description),
	}
	if len(violations) > 0 {
		return patch, newManufacturerValidationError(violations)
	}
	return patch, nil
}

func requireManufacturerField(
	violations *[]shared.FieldViolation,
	name string,
	field Field[string],
) {
	if field.State() != FieldOmitted {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field: name, Reason: "required", Description: "This field is required.",
	})
}

var manufacturerValidationFieldOrder = map[string]int{
	"name": 0, "slug": 1, "description": 2, "update_mask": 3,
}

func newManufacturerValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := manufacturerValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := manufacturerValidationFieldOrder[ordered[right].Field]
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

func mergeManufacturerMutationErrors(commandErr, domainErr error) error {
	if domainErr != nil && !shared.HasReason(domainErr, shared.ErrorReasonValidation) {
		return domainErr
	}
	if commandErr != nil && !shared.HasReason(commandErr, shared.ErrorReasonValidation) {
		return commandErr
	}

	commandViolations := shared.ViolationsOf(commandErr)
	commandFields := make(map[string]struct{}, len(commandViolations))
	merged := make([]shared.FieldViolation, 0, len(commandViolations)+len(shared.ViolationsOf(domainErr)))
	for _, violation := range commandViolations {
		commandFields[violation.Field] = struct{}{}
		merged = append(merged, violation)
	}
	for _, violation := range shared.ViolationsOf(domainErr) {
		if _, duplicate := commandFields[violation.Field]; duplicate {
			continue
		}
		merged = append(merged, violation)
	}
	if len(merged) > 0 {
		return newManufacturerValidationError(merged)
	}
	if commandErr != nil {
		return commandErr
	}
	return domainErr
}
