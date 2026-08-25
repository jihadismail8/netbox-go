package dcim

import (
	"sort"

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
		return values, newDeviceTypeValidationError(violations)
	}
	return values, nil
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
	return buildDeviceTypePatch(
		command.Manufacturer, command.Model, command.Slug, command.PartNumber,
		command.UHeight, command.ExcludeFromUtilization, command.IsFullDepth,
		command.Airflow, command.Description, command.Comments, false,
	)
}

func (command ReplaceDeviceTypeCommand) patch() (deviceTypeCommandPatch, error) {
	return buildDeviceTypePatch(
		command.Manufacturer, command.Model, command.Slug, command.PartNumber,
		command.UHeight, command.ExcludeFromUtilization, command.IsFullDepth,
		command.Airflow, command.Description, command.Comments, true,
	)
}

func buildDeviceTypePatch(
	manufacturer Field[shared.ID],
	model Field[string],
	slug Field[string],
	partNumber Field[string],
	uHeight Field[string],
	excludeFromUtilization Field[bool],
	isFullDepth Field[bool],
	airflow Field[string],
	description Field[string],
	comments Field[string],
	replace bool,
) (deviceTypeCommandPatch, error) {
	var violations []shared.FieldViolation
	if replace {
		requireDeviceTypeField(&violations, "manufacturer", manufacturer)
		requireDeviceTypeField(&violations, "model", model)
		requireDeviceTypeField(&violations, "slug", slug)
	}
	patch := deviceTypeCommandPatch{
		manufacturerID: patchFieldValue(&violations, "manufacturer", manufacturer),
		model:          patchValue(&violations, "model", model),
		slug:           patchValue(&violations, "slug", slug),
		partNumber:     patchValue(&violations, "part_number", partNumber),
		uHeight:        patchValue(&violations, "u_height", uHeight),
		excludeFromUtilization: patchFieldValue(
			&violations, "exclude_from_utilization", excludeFromUtilization,
		),
		isFullDepth: patchFieldValue(&violations, "is_full_depth", isFullDepth),
		airflow:     patchDeviceTypeAirflow(airflow),
		description: patchValue(&violations, "description", description),
		comments:    patchValue(&violations, "comments", comments),
	}
	if replace && uHeight.State() == FieldOmitted {
		defaultHeight := dcimdomain.DeviceTypeDefaultHeight
		patch.uHeight = &defaultHeight
	}
	if patch.manufacturerID != nil && !patch.manufacturerID.IsValid() {
		violations = append(violations, shared.FieldViolation{
			Field: "manufacturer", Reason: "invalid_choice", Description: "Select a valid choice.",
		})
	}
	if len(violations) > 0 {
		return patch, newDeviceTypeValidationError(violations)
	}
	if patch.empty() {
		return patch, newDeviceTypeValidationError([]shared.FieldViolation{{
			Field: "update_mask", Reason: "required",
			Description: "At least one writable field must be supplied.",
		}})
	}
	return patch, nil
}

func requireDeviceTypeField[T any](
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

var deviceTypeValidationFieldOrder = map[string]int{
	"manufacturer": 0, "model": 1, "slug": 2, "part_number": 3,
	"u_height": 4, "exclude_from_utilization": 5, "is_full_depth": 6,
	"airflow": 7, "description": 8, "comments": 9, "update_mask": 10,
}

func newDeviceTypeValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := deviceTypeValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := deviceTypeValidationFieldOrder[ordered[right].Field]
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

func mergeDeviceTypeMutationErrors(errs ...error) error {
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
		return newDeviceTypeValidationError(merged)
	}
	return nil
}
