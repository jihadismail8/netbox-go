package dcim

import (
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
		return interfaceTemplateCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command ReplaceInterfaceTemplateCommand) values() (interfaceTemplateCommandValues, error) {
	return command.CreateInterfaceTemplateCommand.values()
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
	var violations []shared.FieldViolation
	patch := interfaceTemplateCommandPatch{
		deviceTypeID:  patchFieldValue(&violations, "device_type", command.DeviceType),
		name:          patchValue(&violations, "name", command.Name),
		label:         patchValue(&violations, "label", command.Label),
		interfaceType: patchValue(&violations, "type", command.Type),
		enabled:       patchFieldValue(&violations, "enabled", command.Enabled),
		mgmtOnly:      patchFieldValue(&violations, "mgmt_only", command.MgmtOnly),
		description:   patchValue(&violations, "description", command.Description),
	}
	if patch.deviceTypeID != nil && !patch.deviceTypeID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "device_type", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return interfaceTemplateCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return interfaceTemplateCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}
