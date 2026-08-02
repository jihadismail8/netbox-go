package dcim

import (
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

// ReplaceSiteCommand is a full replacement: omitted optional fields reset to
// contract defaults while name and slug remain required.
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

func (command ReplaceSiteCommand) values() (dcimdomain.SiteValues, error) {
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
		Name: valueForFullMutation(&violations, "name", name, "", true),
		Slug: valueForFullMutation(&violations, "slug", slug, "", true),
		Status: valueForFullMutation(
			&violations,
			"status",
			status,
			dcimdomain.SiteStatusActive.String(),
			false,
		),
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
		return dcimdomain.SiteValues{}, shared.NewValidationError(violations...)
	}
	return values, nil
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

func (command UpdateSiteCommand) patch() (dcimdomain.SitePatch, error) {
	var violations []shared.FieldViolation
	patch := dcimdomain.SitePatch{
		Name:        patchValue(&violations, "name", command.Name),
		Slug:        patchValue(&violations, "slug", command.Slug),
		Status:      patchValue(&violations, "status", command.Status),
		Facility:    patchValue(&violations, "facility", command.Facility),
		Description: patchValue(&violations, "description", command.Description),
		Comments:    patchValue(&violations, "comments", command.Comments),
	}
	if len(violations) > 0 {
		return dcimdomain.SitePatch{}, shared.NewValidationError(violations...)
	}
	return patch, nil
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
