package dcim

import (
	"sort"

	"netbox-go/internal/application/presence"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type FieldState = presence.FieldState

const (
	FieldOmitted = presence.Omitted
	FieldNull    = presence.Null
	FieldPresent = presence.Present
)

// Field retains the omitted/null/value distinction shared by REST decoding
// and protobuf presence plus FieldMask.
type Field[T any] = presence.Field[T]

func OmittedField[T any]() Field[T] {
	return presence.OmittedField[T]()
}

func NullField[T any]() Field[T] {
	return presence.NullField[T]()
}

func FieldValue[T any](value T) Field[T] {
	return presence.Value(value)
}

// CreateSiteCommand retains presence for required and optional fields. The
// zero Field value is omitted.
type CreateSiteCommand struct {
	Name        Field[string]
	Slug        Field[string]
	Status      Field[string]
	Facility    Field[string]
	Description Field[string]
	Comments    Field[string]
}

// ReplaceSiteCommand requires the baseline PUT identity fields while retaining
// presence for optional fields. Omitted optional fields preserve locked state.
type ReplaceSiteCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Status      Field[string]
	Facility    Field[string]
	Description Field[string]
	Comments    Field[string]
}

// UpdateSiteCommand preserves field presence for REST PATCH and protobuf
// FieldMask requests. Empty and explicit null are different states.
type UpdateSiteCommand struct {
	ID          shared.ID
	Name        Field[string]
	Slug        Field[string]
	Status      Field[string]
	Facility    Field[string]
	Description Field[string]
	Comments    Field[string]
}

type DeleteSiteCommand struct {
	ID shared.ID
}

func (command CreateSiteCommand) values() (dcimdomain.SiteValues, error) {
	return fullSiteValues(
		command.Name,
		command.Slug,
		command.Status,
		command.Facility,
		command.Description,
		command.Comments,
	)
}

func fullSiteValues(
	name Field[string],
	slug Field[string],
	status Field[string],
	facility Field[string],
	description Field[string],
	comments Field[string],
) (dcimdomain.SiteValues, error) {
	var violations []shared.FieldViolation
	values := dcimdomain.SiteValues{
		Name:     valueForFullMutation(&violations, "name", name, "", true),
		Slug:     valueForFullMutation(&violations, "slug", slug, "", true),
		Status:   fullSiteStatus(&violations, status),
		Facility: valueForFullMutation(&violations, "facility", facility, "", false),
		Description: valueForFullMutation(
			&violations,
			"description",
			description,
			"",
			false,
		),
		Comments: valueForFullMutation(&violations, "comments", comments, "", false),
	}
	if len(violations) > 0 {
		return values, newSiteValidationError(violations)
	}
	return values, nil
}

func fullSiteStatus(
	violations *[]shared.FieldViolation,
	field Field[string],
) string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		if value == "" {
			*violations = append(*violations, blankSiteStatusViolation())
		}
		return value
	case FieldNull:
		*violations = append(*violations, blankSiteStatusViolation())
	}
	return dcimdomain.SiteStatusActive.String()
}

func blankSiteStatusViolation() shared.FieldViolation {
	return shared.FieldViolation{
		Field:       "status",
		Reason:      "blank",
		Description: "This field may not be blank.",
	}
}

func valueForFullMutation(
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
		*violations = append(*violations, shared.FieldViolation{
			Field:       name,
			Reason:      "null",
			Description: "This field may not be null.",
		})
	case FieldOmitted:
		if required {
			*violations = append(*violations, shared.FieldViolation{
				Field:       name,
				Reason:      "required",
				Description: "This field is required.",
			})
		}
	}
	return fallback
}

func (command ReplaceSiteCommand) patch() (dcimdomain.SitePatch, error) {
	return buildSitePatch(
		command.Name,
		command.Slug,
		command.Status,
		command.Facility,
		command.Description,
		command.Comments,
		true,
	)
}

func (command UpdateSiteCommand) patch() (dcimdomain.SitePatch, error) {
	return buildSitePatch(
		command.Name,
		command.Slug,
		command.Status,
		command.Facility,
		command.Description,
		command.Comments,
		false,
	)
}

func buildSitePatch(
	name Field[string],
	slug Field[string],
	status Field[string],
	facility Field[string],
	description Field[string],
	comments Field[string],
	requireIdentity bool,
) (dcimdomain.SitePatch, error) {
	var violations []shared.FieldViolation
	if requireIdentity {
		requireSiteField(&violations, "name", name)
		requireSiteField(&violations, "slug", slug)
	}
	patch := dcimdomain.SitePatch{
		Name:        patchValue(&violations, "name", name),
		Slug:        patchValue(&violations, "slug", slug),
		Status:      patchSiteStatus(&violations, status),
		Facility:    patchValue(&violations, "facility", facility),
		Description: patchValue(&violations, "description", description),
		Comments:    patchValue(&violations, "comments", comments),
	}
	if len(violations) > 0 {
		return patch, newSiteValidationError(violations)
	}
	return patch, nil
}

func requireSiteField(
	violations *[]shared.FieldViolation,
	name string,
	field Field[string],
) {
	if field.State() != FieldOmitted {
		return
	}
	*violations = append(*violations, shared.FieldViolation{
		Field:       name,
		Reason:      "required",
		Description: "This field is required.",
	})
}

func patchSiteStatus(
	violations *[]shared.FieldViolation,
	field Field[string],
) *string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		if value == "" {
			*violations = append(*violations, blankSiteStatusViolation())
		}
		return &value
	case FieldNull:
		*violations = append(*violations, blankSiteStatusViolation())
	}
	return nil
}

func patchValue(
	violations *[]shared.FieldViolation,
	name string,
	field Field[string],
) *string {
	switch field.State() {
	case FieldPresent:
		value, _ := field.Get()
		return &value
	case FieldNull:
		*violations = append(*violations, shared.FieldViolation{
			Field:       name,
			Reason:      "null",
			Description: "This field may not be null.",
		})
	}
	return nil
}

var siteValidationFieldOrder = map[string]int{
	"name":        0,
	"slug":        1,
	"status":      2,
	"facility":    3,
	"description": 4,
	"comments":    5,
	"update_mask": 6,
}

func newSiteValidationError(violations []shared.FieldViolation) error {
	ordered := append([]shared.FieldViolation(nil), violations...)
	sort.SliceStable(ordered, func(left, right int) bool {
		leftOrder, leftKnown := siteValidationFieldOrder[ordered[left].Field]
		rightOrder, rightKnown := siteValidationFieldOrder[ordered[right].Field]
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

func mergeSiteMutationErrors(commandErr, domainErr error) error {
	if domainErr != nil && !shared.HasReason(domainErr, shared.ErrorReasonValidation) {
		return domainErr
	}
	if commandErr != nil && !shared.HasReason(commandErr, shared.ErrorReasonValidation) {
		return commandErr
	}

	commandViolations := shared.ViolationsOf(commandErr)
	commandFields := make(map[string]struct{}, len(commandViolations))
	merged := make([]shared.FieldViolation, 0, len(commandViolations)+len(shared.ViolationsOf(domainErr)))
	for _, violation := range commandViolations {
		commandFields[violation.Field] = struct{}{}
		merged = append(merged, violation)
	}
	for _, violation := range shared.ViolationsOf(domainErr) {
		if _, duplicate := commandFields[violation.Field]; duplicate {
			continue
		}
		merged = append(merged, violation)
	}
	if len(merged) > 0 {
		return newSiteValidationError(merged)
	}
	if commandErr != nil {
		return commandErr
	}
	return domainErr
}
