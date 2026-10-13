package ipam

import (
	"sort"

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

func (command ReplaceVRFCommand) patch() (ipamdomain.VRFPatch, error) {
	return buildVRFPatch(
		command.Name,
		command.RD,
		command.EnforceUnique,
		command.Description,
		command.Comments,
		true,
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
		return values, shared.NewValidationError(violations...)
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
	rd, err := ipamdomain.ValidateVRFRouteDistinguisher(value)
	if err != nil {
		*violations = append(*violations, shared.ViolationsOf(err)...)
		return ipamdomain.NullRouteDistinguisher()
	}
	return ipamdomain.NonNullRouteDistinguisher(rd)
}

func (command UpdateVRFCommand) patch() (ipamdomain.VRFPatch, error) {
	return buildVRFPatch(command.Name, command.RD, command.EnforceUnique, command.Description, command.Comments, false)
}

func buildVRFPatch(
	name Field[string],
	rd Field[string],
	enforceUnique Field[bool],
	description Field[string],
	comments Field[string],
	requireName bool,
) (ipamdomain.VRFPatch, error) {
	var violations []shared.FieldViolation
	if requireName && name.State() == FieldOmitted {
		violations = append(violations, requiredViolation("name"))
	}
	patch := ipamdomain.VRFPatch{
		Name:          patchString(&violations, "name", name),
		RD:            patchRouteDistinguisher(&violations, rd),
		EnforceUnique: patchBool(&violations, "enforce_unique", enforceUnique),
		Description:   patchString(&violations, "description", description),
		Comments:      patchString(&violations, "comments", comments),
	}
	if len(violations) > 0 {
		return patch, shared.NewValidationError(violations...)
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
		rd, err := ipamdomain.ValidateVRFRouteDistinguisher(value)
		if err != nil {
			*violations = append(*violations, shared.ViolationsOf(err)...)
			return nil
		}
		nullable := ipamdomain.NonNullRouteDistinguisher(rd)
		return &nullable
	}
}

// Presence errors supersede fallback-name blanks, but never hide independent
// text failures. Deduplication includes the reason so length and NUL both survive.
func mergeVRFMutationErrors(errs ...error) error {
	presenceErrors := make(map[string]bool)
	for _, err := range errs {
		if err != nil && (!shared.HasReason(err, shared.ErrorReasonValidation) || len(shared.ViolationsOf(err)) == 0) {
			return err
		}
		for _, violation := range shared.ViolationsOf(err) {
			if violation.Reason == "required" || violation.Reason == "null" {
				presenceErrors[violation.Field] = true
			}
		}
	}
	seen := make(map[[2]string]bool)
	var violations []shared.FieldViolation
	for _, err := range errs {
		for _, violation := range shared.ViolationsOf(err) {
			key := [2]string{violation.Field, violation.Reason}
			if seen[key] || (violation.Reason == "blank" && presenceErrors[violation.Field]) {
				continue
			}
			seen[key] = true
			violations = append(violations, violation)
		}
	}
	if len(violations) == 0 {
		return nil
	}
	order := map[string]int{"name": 0, "rd": 1, "enforce_unique": 2, "description": 3, "comments": 4}
	sort.SliceStable(violations, func(left, right int) bool {
		leftOrder, leftKnown := order[violations[left].Field]
		rightOrder, rightKnown := order[violations[right].Field]
		if leftKnown != rightKnown {
			return leftKnown
		}
		return leftKnown && leftOrder < rightOrder
	})
	return shared.NewValidationError(violations...)
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
