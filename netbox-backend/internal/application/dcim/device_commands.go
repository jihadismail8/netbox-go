package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateDeviceCommand struct {
	DeviceType  Field[shared.ID]
	Role        Field[shared.ID]
	Name        Field[string]
	Site        Field[shared.ID]
	Rack        Field[shared.ID]
	Position    Field[string]
	Face        Field[string]
	Status      Field[string]
	Serial      Field[string]
	AssetTag    Field[string]
	Airflow     Field[string]
	Description Field[string]
	Comments    Field[string]
}

type ReplaceDeviceCommand struct {
	ID shared.ID
	CreateDeviceCommand
}

type UpdateDeviceCommand struct {
	ID          shared.ID
	DeviceType  Field[shared.ID]
	Role        Field[shared.ID]
	Name        Field[string]
	Site        Field[shared.ID]
	Rack        Field[shared.ID]
	Position    Field[string]
	Face        Field[string]
	Status      Field[string]
	Serial      Field[string]
	AssetTag    Field[string]
	Airflow     Field[string]
	Description Field[string]
	Comments    Field[string]
}

type DeleteDeviceCommand struct{ ID shared.ID }

type deviceCommandValues struct {
	deviceTypeID shared.ID
	roleID       shared.ID
	name         dcimdomain.DeviceNullable[string]
	siteID       shared.ID
	rackID       dcimdomain.DeviceNullable[shared.ID]
	position     dcimdomain.DeviceNullable[string]
	face         string
	status       string
	serial       string
	assetTag     dcimdomain.DeviceNullable[string]
	airflow      dcimdomain.NullableDeviceAirflow
	description  string
	comments     string
}

func (command CreateDeviceCommand) values() (deviceCommandValues, error) {
	var violations []shared.FieldViolation
	values := deviceCommandValues{
		deviceTypeID: fullFieldValue(&violations, "device_type", command.DeviceType, shared.ID(0), true),
		roleID:       fullFieldValue(&violations, "role", command.Role, shared.ID(0), true),
		name:         fullDeviceNullable(command.Name),
		siteID:       fullFieldValue(&violations, "site", command.Site, shared.ID(0), true),
		rackID:       fullDeviceNullable(command.Rack),
		position:     fullDeviceNullable(command.Position),
		face:         fullBlankableDeviceChoice(command.Face),
		status:       valueForFullMutation(&violations, "status", command.Status, dcimdomain.DeviceStatusActive.String(), false),
		serial:       valueForFullMutation(&violations, "serial", command.Serial, "", false),
		assetTag:     fullDeviceNullable(command.AssetTag),
		airflow:      fullDeviceAirflow(command.Airflow),
		description:  valueForFullMutation(&violations, "description", command.Description, "", false),
		comments:     valueForFullMutation(&violations, "comments", command.Comments, "", false),
	}
	validateRequiredDeviceID(&violations, "device_type", command.DeviceType, values.deviceTypeID)
	validateRequiredDeviceID(&violations, "role", command.Role, values.roleID)
	validateRequiredDeviceID(&violations, "site", command.Site, values.siteID)
	validateNullableDeviceID(&violations, "rack", values.rackID)
	if len(violations) > 0 {
		return deviceCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command ReplaceDeviceCommand) values() (deviceCommandValues, error) {
	return command.CreateDeviceCommand.values()
}

type deviceCommandPatch struct {
	deviceTypeID *shared.ID
	roleID       *shared.ID
	name         *dcimdomain.DeviceNullable[string]
	siteID       *shared.ID
	rackID       *dcimdomain.DeviceNullable[shared.ID]
	position     *dcimdomain.DeviceNullable[string]
	face         *string
	status       *string
	serial       *string
	assetTag     *dcimdomain.DeviceNullable[string]
	airflow      *dcimdomain.NullableDeviceAirflow
	description  *string
	comments     *string
}

func (patch deviceCommandPatch) empty() bool {
	return patch.deviceTypeID == nil && patch.roleID == nil && patch.name == nil &&
		patch.siteID == nil && patch.rackID == nil && patch.position == nil &&
		patch.face == nil && patch.status == nil && patch.serial == nil &&
		patch.assetTag == nil && patch.airflow == nil &&
		patch.description == nil && patch.comments == nil
}

func (command UpdateDeviceCommand) patch() (deviceCommandPatch, error) {
	var violations []shared.FieldViolation
	patch := deviceCommandPatch{
		deviceTypeID: patchFieldValue(&violations, "device_type", command.DeviceType),
		roleID:       patchFieldValue(&violations, "role", command.Role),
		name:         patchDeviceNullable(command.Name),
		siteID:       patchFieldValue(&violations, "site", command.Site),
		rackID:       patchDeviceNullable(command.Rack),
		position:     patchDeviceNullable(command.Position),
		face:         patchBlankableDeviceChoice(command.Face),
		status:       patchValue(&violations, "status", command.Status),
		serial:       patchValue(&violations, "serial", command.Serial),
		assetTag:     patchDeviceNullable(command.AssetTag),
		airflow:      patchDeviceAirflow(command.Airflow),
		description:  patchValue(&violations, "description", command.Description),
		comments:     patchValue(&violations, "comments", command.Comments),
	}
	if patch.deviceTypeID != nil && !patch.deviceTypeID.IsValid() {
		violations = append(violations, deviceChoiceViolation("device_type"))
	}
	if patch.roleID != nil && !patch.roleID.IsValid() {
		violations = append(violations, deviceChoiceViolation("role"))
	}
	if patch.siteID != nil && !patch.siteID.IsValid() {
		violations = append(violations, deviceChoiceViolation("site"))
	}
	if patch.rackID != nil {
		validateNullableDeviceID(&violations, "rack", *patch.rackID)
	}
	if len(violations) > 0 {
		return deviceCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return deviceCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func fullDeviceNullable[T any](field Field[T]) dcimdomain.DeviceNullable[T] {
	if value, present := field.Get(); present {
		return dcimdomain.NonNullDeviceValue(value)
	}
	return dcimdomain.NullDeviceValue[T]()
}

func patchDeviceNullable[T any](field Field[T]) *dcimdomain.DeviceNullable[T] {
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

func fullBlankableDeviceChoice(field Field[string]) string {
	value, present := field.Get()
	if present {
		return value
	}
	return ""
}

func patchBlankableDeviceChoice(field Field[string]) *string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return &value
	case FieldNull:
		value := ""
		return &value
	default:
		return nil
	}
}

func fullDeviceAirflow(field Field[string]) dcimdomain.NullableDeviceAirflow {
	value, present := field.Get()
	if !present {
		return dcimdomain.NullDeviceAirflow()
	}
	return dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflow(value))
}

func patchDeviceAirflow(field Field[string]) *dcimdomain.NullableDeviceAirflow {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		airflow := dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflow(value))
		return &airflow
	case FieldNull:
		airflow := dcimdomain.NullDeviceAirflow()
		return &airflow
	default:
		return nil
	}
}

func validateRequiredDeviceID(
	violations *[]shared.FieldViolation,
	field string,
	input Field[shared.ID],
	id shared.ID,
) {
	if input.State() == FieldPresent && !id.IsValid() {
		*violations = append(*violations, deviceChoiceViolation(field))
	}
}

func validateNullableDeviceID(
	violations *[]shared.FieldViolation,
	field string,
	value dcimdomain.DeviceNullable[shared.ID],
) {
	id, present := value.Get()
	if present && !id.IsValid() {
		*violations = append(*violations, deviceChoiceViolation(field))
	}
}

func deviceChoiceViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field: field, Reason: "invalid_choice", Description: "Select a valid choice.",
	}
}
