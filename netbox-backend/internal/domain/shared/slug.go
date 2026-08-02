package shared

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Slug is a validated ASCII slug compatible with Django's default SlugField.
type Slug string

// ParseSlug trims transport whitespace and validates the value and length.
func ParseSlug(value string, maxLength int) (Slug, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", NewValidationError(FieldViolation{
			Field:       "slug",
			Reason:      "required",
			Description: "This field may not be blank.",
		})
	}
	if utf8.RuneCountInString(value) > maxLength {
		return "", NewValidationError(FieldViolation{
			Field:  "slug",
			Reason: "max_length",
			Description: fmt.Sprintf(
				"Ensure this field has no more than %d characters.",
				maxLength,
			),
		})
	}
	if !slugPattern.MatchString(value) {
		return "", NewValidationError(FieldViolation{
			Field:       "slug",
			Reason:      "invalid",
			Description: "Enter a valid slug consisting of letters, numbers, underscores, or hyphens.",
		})
	}

	return Slug(value), nil
}

func (slug Slug) String() string {
	return string(slug)
}
