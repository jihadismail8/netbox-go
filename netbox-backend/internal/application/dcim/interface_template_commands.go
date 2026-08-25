package dcim

import (
	"sort"

	"netbox-go/internal/domain/shared"
)

type CreateInterfaceTemplateCommand struct {
	DeviceType  Field[shared.ID]
	Name        Field[string]
	Label       Field[string]
	Type        Field[string]
	Enabled     Field[bool]
	MgmtOnly    Field[bool]
	Description Field[string]
}

type ReplaceInterfaceTemplateCommand struct {
	ID shared.ID
	CreateInterfaceTemplateCommand
}

type UpdateInterfaceTemplateCommand struct {
	ID          shared.ID
	DeviceType  Field[shared.ID]
	Name        Field[string]
	Label       Field[string]
	Type        Field[string]
	Enabled     Field[bool]
	MgmtOnly    Field[bool]
	Description Field[string]
}

type DeleteInterfaceTemplateCommand struct{ ID shared.ID }

type interfaceTemplateCommandValues struct {
	deviceTypeID  shared.ID
	name          string
	label         string
	interfaceType string
	enabled       bool
	mgmtOnly      bool
	description   string
}

func (command CreateInterfaceTemplateCommand) values() (interfaceTemplateCommandValues, error) {
	var violations []shared.FieldViolation
	values := interfaceTemplateCommandValues{
		deviceTypeID: fullFieldValue(
			&violations, "device_type", command.DeviceType, shared.ID(0), true,
		),
		name:          valueForFullMutation(&violations, "name", command.Name, "", true),
		label:         valueForFullMutation(&violations, "label", command.Label, "", false),
		interfaceType: valueForFullMutation(&violations, "type", command.Type, "", true),
		enabled:       fullFieldValue(&violations, "enabled", command.Enabled, true, false),
		mgmtOnly:      fullFieldValue(&violations, "mgmt_only", command.MgmtOnly, false, false),
		description: valueForFullMutation(
			&violations, "description", command.Description, "", false,
		),
	}
	if values.deviceTypeID <= 0 && command.DeviceType.State() == FieldPresent {
		violations = append(violations, shared.FieldViolation{
			Field: "device_type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return values, newInterfaceTemplateValidationError(violations)
	}
	return values, nil
}

type interfaceTemplateCommandPatch struct {
	deviceTypeID  *shared.ID
	name          *string
	label         *string
	interfaceType *string
	enabled       *bool
	mgmtOnly      *bool
	description   *string
}

func (patch interfaceTemplateCommandPatch) empty() bool {
	return patch.deviceTypeID == nil && patch.name == nil && patch.label == nil &&
		patch.interfaceType == nil && patch.enabled == nil && patch.mgmtOnly == nil &&
		patch.description == nil
}

func (command UpdateInterfaceTemplateCommand) patch() (interfaceTemplateCommandPatch, error) {
	return buildInterfaceTemplatePatch(
		command.DeviceType, command.Name, command.Label, command.Type,
		command.Enabled, command.MgmtOnly, command.Description, false,
	)
}

func (command ReplaceInterfaceTemplateCommand) patch() (interfaceTemplateCommandPatch, error) {
	return buildInterfaceTemplatePatch(
		command.DeviceType, command.Name, command.Label, command.Type,
		command.Enabled, command.MgmtOnly, command.Description, true,
	)
}

func buildInterfaceTemplatePatch(
	deviceType Field[shared.ID],
	name Field[string],
	label Field[string],
	interfaceType Field[string],
	enabled Field[bool],
	mgmtOnly Field[bool],
	description Field[string],
	replace bool,
) (interfaceTemplateCommandPatch, error) {
	var violations []shared.FieldViolation
	if replace {
		requireInterfaceTemplateField(&violations, "device_type", deviceType)
		requireInterfaceTemplateField(&violations, "name", name)
		requireInterfaceTemplateField(&violations, "type", interfaceType)
	}
	patch := interfaceTemplateCommandPatch{
		deviceTypeID:  patchFieldValue(&violations, "device_type", deviceType),
		name:          patchValue(&violations, "name", name),
		label:         patchValue(&violations, "label", label),
		interfaceType: patchValue(&violations, "type", interfaceType),
		enabled:       patchFieldValue(&violations, "enabled", enabled),
		mgmtOnly:      patchFieldValue(&violations, "mgmt_only", mgmtOnly),
		description:   patchValue(&violations, "description", description),
	}
	if patch.deviceTypeID != nil && !patch.deviceTypeID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "device_type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return patch, newInterfaceTemplateValidationError(violations)
	}
	if patch.empty() {
		return patch, newInterfaceTemplateValidationError([]shared.FieldViolation{{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		}})
	}
	return patch, nil
}

func requireInterfaceTemplateField[T any](
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

var interfaceTemplateValidationFieldOrder = map[string]int{
	"device_type": 0, "name": 1, "label": 2, "type": 3,
	"enabled": 4, "mgmt_only": 5, "description": 6, "update_mask": 7,
}

func newInterfaceTemplateValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := interfaceTemplateValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := interfaceTemplateValidationFieldOrder[ordered[right].Field]
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

func mergeInterfaceTemplateMutationErrors(errs ...error) error {
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
		return newInterfaceTemplateValidationError(merged)
	}
	return nil
}
