package dcim

import (
	"sort"

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
		return values, newRackRoleValidationError(violations)
	}
	return values, nil
}

func (command ReplaceRackRoleCommand) patch() (dcimdomain.RackRolePatch, error) {
	return buildRackRolePatch(
		command.Name, command.Slug, command.Color, command.Description, true,
	)
}

func (command UpdateRackRoleCommand) patch() (dcimdomain.RackRolePatch, error) {
	return buildRackRolePatch(
		command.Name, command.Slug, command.Color, command.Description, false,
	)
}

func buildRackRolePatch(
	name Field[string],
	slug Field[string],
	color Field[string],
	description Field[string],
	requireIdentity bool,
) (dcimdomain.RackRolePatch, error) {
	var violations []shared.FieldViolation
	if requireIdentity {
		requireRackRoleField(&violations, "name", name)
		requireRackRoleField(&violations, "slug", slug)
	}
	patch := dcimdomain.RackRolePatch{
		Name:        patchValue(&violations, "name", name),
		Slug:        patchValue(&violations, "slug", slug),
		Color:       patchValue(&violations, "color", color),
		Description: patchValue(&violations, "description", description),
	}
	if len(violations) > 0 {
		return patch, newRackRoleValidationError(violations)
	}
	return patch, nil
}

func requireRackRoleField(
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

var rackRoleValidationFieldOrder = map[string]int{
	"name": 0, "slug": 1, "color": 2, "description": 3, "update_mask": 4,
}

func newRackRoleValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := rackRoleValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := rackRoleValidationFieldOrder[ordered[right].Field]
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

func mergeRackRoleMutationErrors(commandErr, domainErr error) error {
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
		return newRackRoleValidationError(merged)
	}
	if commandErr != nil {
		return commandErr
	}
	return domainErr
}
