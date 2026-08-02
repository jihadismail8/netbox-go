package dcim

import (
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type CreateDeviceTypeCommand struct {
	Manufacturer           Field[shared.ID]
	Model                  Field[string]
	Slug                   Field[string]
	PartNumber             Field[string]
	UHeight                Field[string]
	ExcludeFromUtilization Field[bool]
	IsFullDepth            Field[bool]
	Airflow                Field[string]
	Description            Field[string]
	Comments               Field[string]
}

type ReplaceDeviceTypeCommand struct {
	ID shared.ID
	CreateDeviceTypeCommand
}

type UpdateDeviceTypeCommand struct {
	ID                     shared.ID
	Manufacturer           Field[shared.ID]
	Model                  Field[string]
	Slug                   Field[string]
	PartNumber             Field[string]
	UHeight                Field[string]
	ExcludeFromUtilization Field[bool]
	IsFullDepth            Field[bool]
	Airflow                Field[string]
	Description            Field[string]
	Comments               Field[string]
}

type DeleteDeviceTypeCommand struct{ ID shared.ID }

type deviceTypeCommandValues struct {
	manufacturerID         shared.ID
	model                  string
	slug                   string
	partNumber             string
	uHeight                string
	excludeFromUtilization bool
	isFullDepth            bool
	airflow                dcimdomain.NullableDeviceAirflow
	description            string
	comments               string
}

func (command CreateDeviceTypeCommand) values() (deviceTypeCommandValues, error) {
	var violations []shared.FieldViolation
	values := deviceTypeCommandValues{
		manufacturerID: fullFieldValue(
			&violations, "manufacturer", command.Manufacturer, shared.ID(0), true,
		),
		model:      valueForFullMutation(&violations, "model", command.Model, "", true),
		slug:       valueForFullMutation(&violations, "slug", command.Slug, "", true),
		partNumber: valueForFullMutation(&violations, "part_number", command.PartNumber, "", false),
		uHeight: valueForFullMutation(
			&violations, "u_height", command.UHeight, dcimdomain.DeviceTypeDefaultHeight, false,
		),
		excludeFromUtilization: fullFieldValue(
			&violations, "exclude_from_utilization", command.ExcludeFromUtilization, false, false,
		),
		isFullDepth: fullFieldValue(
			&violations, "is_full_depth", command.IsFullDepth, true, false,
		),
		airflow: fullDeviceTypeAirflow(&violations, command.Airflow),
		description: valueForFullMutation(
			&violations, "description", command.Description, "", false,
		),
		comments: valueForFullMutation(&violations, "comments", command.Comments, "", false),
	}
	if values.manufacturerID <= 0 && command.Manufacturer.State() == FieldPresent {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return deviceTypeCommandValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func (command ReplaceDeviceTypeCommand) values() (deviceTypeCommandValues, error) {
	return command.CreateDeviceTypeCommand.values()
}

func fullDeviceTypeAirflow(
	violations *[]shared.FieldViolation,
	field Field[string],
) dcimdomain.NullableDeviceAirflow {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflow(value))
	case FieldNull, FieldOmitted:
		return dcimdomain.NullDeviceAirflow()
	default:
		*violations = append(*violations, shared.FieldViolation{
			Field: "airflow", Reason: "invalid", Description: "Invalid airflow value.",
		})
		return dcimdomain.NullDeviceAirflow()
	}
}

type deviceTypeCommandPatch struct {
	manufacturerID         *shared.ID
	model                  *string
	slug                   *string
	partNumber             *string
	uHeight                *string
	excludeFromUtilization *bool
	isFullDepth            *bool
	airflow                *dcimdomain.NullableDeviceAirflow
	description            *string
	comments               *string
}

func (patch deviceTypeCommandPatch) empty() bool {
	return patch.manufacturerID == nil && patch.model == nil && patch.slug == nil &&
		patch.partNumber == nil && patch.uHeight == nil &&
		patch.excludeFromUtilization == nil && patch.isFullDepth == nil &&
		patch.airflow == nil && patch.description == nil && patch.comments == nil
}

func (command UpdateDeviceTypeCommand) patch() (deviceTypeCommandPatch, error) {
	var violations []shared.FieldViolation
	patch := deviceTypeCommandPatch{
		manufacturerID: patchFieldValue(&violations, "manufacturer", command.Manufacturer),
		model:          patchValue(&violations, "model", command.Model),
		slug:           patchValue(&violations, "slug", command.Slug),
		partNumber:     patchValue(&violations, "part_number", command.PartNumber),
		uHeight:        patchValue(&violations, "u_height", command.UHeight),
		excludeFromUtilization: patchFieldValue(
			&violations, "exclude_from_utilization", command.ExcludeFromUtilization,
		),
		isFullDepth: patchFieldValue(&violations, "is_full_depth", command.IsFullDepth),
		airflow:     patchDeviceTypeAirflow(command.Airflow),
		description: patchValue(&violations, "description", command.Description),
		comments:    patchValue(&violations, "comments", command.Comments),
	}
	if patch.manufacturerID != nil && !patch.manufacturerID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return deviceTypeCommandPatch{}, shared.NewValidationError(violations...)
	}
	if patch.empty() {
		return deviceTypeCommandPatch{}, shared.NewValidationError(shared.FieldViolation{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		})
	}
	return patch, nil
}

func patchDeviceTypeAirflow(
	field Field[string],
) *dcimdomain.NullableDeviceAirflow {
	var value dcimdomain.NullableDeviceAirflow
	switch field.State() {
	case FieldPresent:
		raw, _ := field.Get()
		value = dcimdomain.NonNullDeviceAirflow(dcimdomain.DeviceAirflow(raw))
	case FieldNull:
		value = dcimdomain.NullDeviceAirflow()
	default:
		return nil
	}
	return &value
}
