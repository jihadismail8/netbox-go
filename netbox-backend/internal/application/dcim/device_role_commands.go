package dcim

import (
	"sort"

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
		return values, newDeviceRoleValidationError(violations)
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

func (command ReplaceDeviceRoleCommand) patch() (dcimdomain.DeviceRolePatch, error) {
	return buildDeviceRolePatch(
		command.Parent, command.Name, command.Slug, command.Color, command.VMRole,
		command.Description, command.Comments, true,
	)
}

func (command UpdateDeviceRoleCommand) patch() (dcimdomain.DeviceRolePatch, error) {
	return buildDeviceRolePatch(
		command.Parent, command.Name, command.Slug, command.Color, command.VMRole,
		command.Description, command.Comments, false,
	)
}

func buildDeviceRolePatch(
	parent Field[shared.ID],
	name Field[string],
	slug Field[string],
	color Field[string],
	vmRole Field[bool],
	description Field[string],
	comments Field[string],
	requireIdentity bool,
) (dcimdomain.DeviceRolePatch, error) {
	var violations []shared.FieldViolation
	if requireIdentity {
		requireDeviceRoleField(&violations, "name", name)
		requireDeviceRoleField(&violations, "slug", slug)
	}
	patch := dcimdomain.DeviceRolePatch{
		Parent:      deviceRoleParentPatch(&violations, parent, requireIdentity),
		Name:        patchValue(&violations, "name", name),
		Slug:        patchValue(&violations, "slug", slug),
		Color:       patchValue(&violations, "color", color),
		VMRole:      boolPatchValue(&violations, "vm_role", vmRole),
		Description: patchValue(&violations, "description", description),
		Comments:    patchValue(&violations, "comments", comments),
	}
	if len(violations) > 0 {
		return patch, newDeviceRoleValidationError(violations)
	}
	return patch, nil
}

func requireDeviceRoleField[T any](
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

func deviceRoleParentPatch(
	violations *[]shared.FieldViolation,
	field Field[shared.ID],
	resetOmitted bool,
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
		if resetOmitted {
			value := dcimdomain.RootDeviceRoleParent()
			return &value
		}
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

var deviceRoleValidationFieldOrder = map[string]int{
	"parent": 0, "name": 1, "slug": 2, "color": 3, "vm_role": 4,
	"description": 5, "comments": 6, "update_mask": 7,
}

func newDeviceRoleValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := deviceRoleValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := deviceRoleValidationFieldOrder[ordered[right].Field]
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

func mergeDeviceRoleMutationErrors(errs ...error) error {
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
		return newDeviceRoleValidationError(merged)
	}
	return nil
}
