package dcim

import (
	"sort"

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
	MTU         Field[int64]
	Speed       Field[int64]
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
	MTU         Field[int64]
	Speed       Field[int64]
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
	mtu           dcimdomain.DeviceNullable[int64]
	speed         dcimdomain.DeviceNullable[int64]
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
			&violations, "type", interfaceTypeField(command.Type), "", true,
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
		return values, newInterfaceValidationError(violations)
	}
	return values, nil
}

type interfaceCommandPatch struct {
	deviceID      *shared.ID
	name          *string
	label         *string
	interfaceType *string
	enabled       *bool
	mgmtOnly      *bool
	mtu           *dcimdomain.DeviceNullable[int64]
	speed         *dcimdomain.DeviceNullable[int64]
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
	return buildInterfacePatch(
		command.Device, command.Name, command.Label, command.Type,
		command.Enabled, command.MgmtOnly, command.MTU, command.Speed,
		command.Duplex, command.Description, false,
	)
}

func (command ReplaceInterfaceCommand) patch() (interfaceCommandPatch, error) {
	return buildInterfacePatch(
		command.Device, command.Name, command.Label, command.Type,
		command.Enabled, command.MgmtOnly, command.MTU, command.Speed,
		command.Duplex, command.Description, true,
	)
}

func buildInterfacePatch(
	device Field[shared.ID], name Field[string], label Field[string], interfaceType Field[string],
	enabled Field[bool], mgmtOnly Field[bool], mtu Field[int64], speed Field[int64],
	duplex Field[string], description Field[string], replace bool,
) (interfaceCommandPatch, error) {
	var violations []shared.FieldViolation
	if replace {
		requireInterfaceField(&violations, "device", device)
		requireInterfaceField(&violations, "name", name)
		requireInterfaceField(&violations, "type", interfaceType)
	}
	patch := interfaceCommandPatch{
		deviceID:      patchFieldValue(&violations, "device", device),
		name:          patchValue(&violations, "name", name),
		label:         patchValue(&violations, "label", label),
		interfaceType: patchValue(&violations, "type", interfaceTypeField(interfaceType)),
		enabled:       patchFieldValue(&violations, "enabled", enabled),
		mgmtOnly:      patchFieldValue(&violations, "mgmt_only", mgmtOnly),
		mtu:           nullablePatchValue(mtu),
		speed:         nullablePatchValue(speed),
		duplex:        nullablePatchValue(duplex),
		description:   patchValue(&violations, "description", description),
	}
	if patch.deviceID != nil && !patch.deviceID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "device", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return patch, newInterfaceValidationError(violations)
	}
	if patch.empty() {
		return interfaceCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func interfaceTypeField(field Field[string]) Field[string] {
	if field.State() == FieldNull {
		return FieldValue("")
	}
	return field
}

func requireInterfaceField[T any](violations *[]shared.FieldViolation, name string, field Field[T]) {
	if field.State() == FieldOmitted {
		*violations = append(*violations, shared.FieldViolation{
			Field: name, Reason: "required", Description: "This field is required.",
		})
	}
}

func newInterfaceValidationError(violations []shared.FieldViolation) error {
	order := map[string]int{
		"device": 0, "name": 1, "label": 2, "type": 3, "enabled": 4,
		"mgmt_only": 5, "mtu": 6, "speed": 7, "duplex": 8, "description": 9, "update_mask": 10,
	}
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := order[ordered[left].Field]
		rightOrder, rightKnown := order[ordered[right].Field]
		if leftKnown != rightKnown {
			return leftKnown
		}
		return leftKnown && leftOrder < rightOrder
	})
	return shared.NewValidationError(ordered...)
}

func mergeInterfaceMutationErrors(errs ...error) error {
	for _, err := range errs {
		if err != nil && !shared.HasReason(err, shared.ErrorReasonValidation) {
			return err
		}
	}
	seen := make(map[string]struct{})
	var merged []shared.FieldViolation
	for _, err := range errs {
		violations := shared.ViolationsOf(err)
		if err != nil && len(violations) == 0 {
			return err
		}
		for _, violation := range violations {
			if _, duplicate := seen[violation.Field]; duplicate {
				continue
			}
			seen[violation.Field] = struct{}{}
			merged = append(merged, violation)
		}
	}
	if len(merged) > 0 {
		return newInterfaceValidationError(merged)
	}
	return nil
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
