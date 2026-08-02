package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateInterfaceCommand struct {
	Device      Field[shared.ID]
	Name        Field[string]
	Label       Field[string]
	Type        Field[string]
	Enabled     Field[bool]
	MgmtOnly    Field[bool]
	MTU         Field[uint32]
	Speed       Field[uint64]
	Duplex      Field[string]
	Description Field[string]
}

type ReplaceInterfaceCommand struct {
	ID shared.ID
	CreateInterfaceCommand
}

type UpdateInterfaceCommand struct {
	ID          shared.ID
	Device      Field[shared.ID]
	Name        Field[string]
	Label       Field[string]
	Type        Field[string]
	Enabled     Field[bool]
	MgmtOnly    Field[bool]
	MTU         Field[uint32]
	Speed       Field[uint64]
	Duplex      Field[string]
	Description Field[string]
}

type DeleteInterfaceCommand struct{ ID shared.ID }

type interfaceCommandValues struct {
	deviceID      shared.ID
	name          string
	label         string
	interfaceType string
	enabled       bool
	mgmtOnly      bool
	mtu           dcimdomain.DeviceNullable[uint32]
	speed         dcimdomain.DeviceNullable[uint64]
	duplex        dcimdomain.DeviceNullable[string]
	description   string
}

func (command CreateInterfaceCommand) values() (interfaceCommandValues, error) {
	var violations []shared.FieldViolation
	values := interfaceCommandValues{
		deviceID: fullFieldValue(
			&violations, "device", command.Device, shared.ID(0), true,
		),
		name:  valueForFullMutation(&violations, "name", command.Name, "", true),
		label: valueForFullMutation(&violations, "label", command.Label, "", false),
		interfaceType: valueForFullMutation(
			&violations, "type", command.Type, "", true,
		),
		enabled:  fullFieldValue(&violations, "enabled", command.Enabled, true, false),
		mgmtOnly: fullFieldValue(&violations, "mgmt_only", command.MgmtOnly, false, false),
		mtu:      nullableFullValue(command.MTU),
		speed:    nullableFullValue(command.Speed),
		duplex:   nullableFullValue(command.Duplex),
		description: valueForFullMutation(
			&violations, "description", command.Description, "", false,
		),
	}
	if values.deviceID <= 0 && command.Device.State() == FieldPresent {
		violations = append(violations, shared.FieldViolation{
			Field: "device", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return interfaceCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command ReplaceInterfaceCommand) values() (interfaceCommandValues, error) {
	return command.CreateInterfaceCommand.values()
}

type interfaceCommandPatch struct {
	deviceID      *shared.ID
	name          *string
	label         *string
	interfaceType *string
	enabled       *bool
	mgmtOnly      *bool
	mtu           *dcimdomain.DeviceNullable[uint32]
	speed         *dcimdomain.DeviceNullable[uint64]
	duplex        *dcimdomain.DeviceNullable[string]
	description   *string
}

func (patch interfaceCommandPatch) empty() bool {
	return patch.deviceID == nil && patch.name == nil && patch.label == nil &&
		patch.interfaceType == nil && patch.enabled == nil && patch.mgmtOnly == nil &&
		patch.mtu == nil && patch.speed == nil && patch.duplex == nil &&
		patch.description == nil
}

func (command UpdateInterfaceCommand) patch() (interfaceCommandPatch, error) {
	var violations []shared.FieldViolation
	patch := interfaceCommandPatch{
		deviceID:      patchFieldValue(&violations, "device", command.Device),
		name:          patchValue(&violations, "name", command.Name),
		label:         patchValue(&violations, "label", command.Label),
		interfaceType: patchValue(&violations, "type", command.Type),
		enabled:       patchFieldValue(&violations, "enabled", command.Enabled),
		mgmtOnly:      patchFieldValue(&violations, "mgmt_only", command.MgmtOnly),
		mtu:           nullablePatchValue(command.MTU),
		speed:         nullablePatchValue(command.Speed),
		duplex:        nullablePatchValue(command.Duplex),
		description:   patchValue(&violations, "description", command.Description),
	}
	if patch.deviceID != nil && !patch.deviceID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "device", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return interfaceCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return interfaceCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func nullableFullValue[T any](field Field[T]) dcimdomain.DeviceNullable[T] {
	if value, present := field.Get(); present {
		return dcimdomain.NonNullDeviceValue(value)
	}
	return dcimdomain.NullDeviceValue[T]()
}

func nullablePatchValue[T any](
	field Field[T],
) *dcimdomain.DeviceNullable[T] {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		nullable := dcimdomain.NonNullDeviceValue(value)
		return &nullable
	case FieldNull:
		nullable := dcimdomain.NullDeviceValue[T]()
		return &nullable
	default:
		return nil
	}
}
