package ipam

import (
	"netbox-go/internal/application/presence"
	ipamdomain "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type FieldState = presence.FieldState

const (
	FieldOmitted = presence.Omitted
	FieldNull    = presence.Null
	FieldPresent = presence.Present
)

// Field retains omitted/null/value presence across REST and protobuf adapters.
type Field[T any] = presence.Field[T]

func OmittedField[T any]() Field[T] { return presence.OmittedField[T]() }

func NullField[T any]() Field[T] { return presence.NullField[T]() }

func FieldValue[T any](value T) Field[T] {
	return presence.Value(value)
}

type CreateVRFCommand struct {
	Name          Field[string]
	RD            Field[string]
	EnforceUnique Field[bool]
	Description   Field[string]
	Comments      Field[string]
}

type ReplaceVRFCommand struct {
	ID            shared.ID
	Name          Field[string]
	RD            Field[string]
	EnforceUnique Field[bool]
	Description   Field[string]
	Comments      Field[string]
}

type UpdateVRFCommand struct {
	ID            shared.ID
	Name          Field[string]
	RD            Field[string]
	EnforceUnique Field[bool]
	Description   Field[string]
	Comments      Field[string]
}

type DeleteVRFCommand struct {
	ID shared.ID
}

func (command CreateVRFCommand) values() (ipamdomain.VRFValues, error) {
	return fullVRFValues(
		command.Name,
		command.RD,
		command.EnforceUnique,
		command.Description,
		command.Comments,
	)
}

func (command ReplaceVRFCommand) values() (ipamdomain.VRFValues, error) {
	return fullVRFValues(
		command.Name,
		command.RD,
		command.EnforceUnique,
		command.Description,
		command.Comments,
	)
}

func fullVRFValues(
	name Field[string],
	rd Field[string],
	enforceUnique Field[bool],
	description Field[string],
	comments Field[string],
) (ipamdomain.VRFValues, error) {
	var violations []shared.FieldViolation
	values := ipamdomain.VRFValues{
		Name:          fullString(&violations, "name", name, "", true),
		RD:            fullRouteDistinguisher(&violations, rd),
		EnforceUnique: fullBool(&violations, "enforce_unique", enforceUnique, true),
		Description:   fullString(&violations, "description", description, "", false),
		Comments:      fullString(&violations, "comments", comments, "", false),
	}
	if len(violations) > 0 {
		return ipamdomain.VRFValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
}

func fullString(
	violations *[]shared.FieldViolation,
	name string,
	field Field[string],
	fallback string,
	required bool,
) string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return value
	case FieldNull:
		*violations = append(*violations, nullViolation(name))
	case FieldOmitted:
		if required {
			*violations = append(*violations, requiredViolation(name))
		}
	}
	return fallback
}

func fullBool(
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
		*violations = append(*violations, nullViolation(name))
	}
	return fallback
}

func fullRouteDistinguisher(
	violations *[]shared.FieldViolation,
	field Field[string],
) ipamdomain.NullableRouteDistinguisher {
	if field.State() != FieldPresent {
		return ipamdomain.NullRouteDistinguisher()
	}
	value, _ := field.Get()
	rd, err := ipamdomain.ParseRouteDistinguisher(value)
	if err != nil {
		*violations = append(*violations, shared.ViolationsOf(err)...)
		return ipamdomain.NullRouteDistinguisher()
	}
	return ipamdomain.NonNullRouteDistinguisher(rd)
}

func (command UpdateVRFCommand) patch() (ipamdomain.VRFPatch, error) {
	var violations []shared.FieldViolation
	patch := ipamdomain.VRFPatch{
		Name:          patchString(&violations, "name", command.Name),
		RD:            patchRouteDistinguisher(&violations, command.RD),
		EnforceUnique: patchBool(&violations, "enforce_unique", command.EnforceUnique),
		Description:   patchString(&violations, "description", command.Description),
		Comments:      patchString(&violations, "comments", command.Comments),
	}
	if len(violations) > 0 {
		return ipamdomain.VRFPatch{}, shared.NewValidationError(violations...)
	}
	return patch, nil
}

func patchString(
	violations *[]shared.FieldViolation,
	name string,
	field Field[string],
) *string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return &value
	case FieldNull:
		*violations = append(*violations, nullViolation(name))
	}
	return nil
}

func patchBool(
	violations *[]shared.FieldViolation,
	name string,
	field Field[bool],
) *bool {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return &value
	case FieldNull:
		*violations = append(*violations, nullViolation(name))
	}
	return nil
}

func patchRouteDistinguisher(
	violations *[]shared.FieldViolation,
	field Field[string],
) *ipamdomain.NullableRouteDistinguisher {
	switch field.State() {
	case FieldOmitted:
		return nil
	case FieldNull:
		nullable := ipamdomain.NullRouteDistinguisher()
		return &nullable
	default:
		value, _ := field.Get()
		rd, err := ipamdomain.ParseRouteDistinguisher(value)
		if err != nil {
			*violations = append(*violations, shared.ViolationsOf(err)...)
			return nil
		}
		nullable := ipamdomain.NonNullRouteDistinguisher(rd)
		return &nullable
	}
}

func nullViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field:       field,
		Reason:      "null",
		Description: "This field may not be null.",
	}
}

func requiredViolation(field string) shared.FieldViolation {
	return shared.FieldViolation{
		Field:       field,
		Reason:      "required",
		Description: "This field is required.",
	}
}
