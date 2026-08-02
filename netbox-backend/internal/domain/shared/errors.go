package shared

import (
	"errors"
	"fmt"
)

// ErrorReason is the canonical, transport-neutral error taxonomy. REST and
// gRPC adapters map these reasons without inventing their own business errors.
type ErrorReason string

const (
	ErrorReasonValidation      ErrorReason = "validation"
	ErrorReasonUnauthenticated ErrorReason = "unauthenticated"
	ErrorReasonForbidden       ErrorReason = "forbidden"
	ErrorReasonNotFound        ErrorReason = "not_found"
	ErrorReasonConflict        ErrorReason = "conflict"
	ErrorReasonProtected       ErrorReason = "protected"
	ErrorReasonRateLimited     ErrorReason = "rate_limited"
	ErrorReasonInternal        ErrorReason = "internal"
)

// FieldViolation describes one stable validation failure.
type FieldViolation struct {
	Field       string
	Reason      string
	Description string
}

// Error carries a canonical reason and optional field violations while still
// preserving the underlying infrastructure error for logs.
type Error struct {
	Reason          ErrorReason
	Message         string
	FieldViolations []FieldViolation
	Cause           error
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}
	if err.Message != "" {
		return err.Message
	}
	return string(err.Reason)
}

func (err *Error) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Cause
}

// NewError creates a canonical error without field violations.
func NewError(reason ErrorReason, message string) *Error {
	return &Error{Reason: reason, Message: message}
}

// WrapError retains a technical cause behind a safe application message.
func WrapError(reason ErrorReason, message string, cause error) *Error {
	return &Error{Reason: reason, Message: message, Cause: cause}
}

// NewValidationError creates one validation error containing all violations.
func NewValidationError(violations ...FieldViolation) *Error {
	cloned := append([]FieldViolation(nil), violations...)
	return &Error{
		Reason:          ErrorReasonValidation,
		Message:         "Invalid input.",
		FieldViolations: cloned,
	}
}

// Invalid creates the canonical single-field validation shape for malformed
// typed input.
func Invalid(field, description string) *Error {
	return NewValidationError(FieldViolation{
		Field:       field,
		Reason:      "invalid",
		Description: description,
	})
}

// Unauthenticated returns the canonical missing-credentials error shared by
// REST and gRPC adapters.
func Unauthenticated() *Error {
	return NewError(
		ErrorReasonUnauthenticated,
		"Authentication credentials were not provided.",
	)
}

// NotFound hides storage details behind a stable domain error.
func NotFound(resource string, id ID) *Error {
	return NewError(
		ErrorReasonNotFound,
		fmt.Sprintf("%s with ID %s was not found.", resource, id),
	)
}

// Conflict reports a uniqueness or protected-relationship failure.
func Conflict(message string, cause error) *Error {
	return WrapError(ErrorReasonConflict, message, cause)
}

// ConflictWithViolations preserves conflict semantics for non-HTTP
// transports while allowing NetBox-compatible REST adapters to render a
// database-enforced uniqueness race against the field which caused it.
func ConflictWithViolations(
	message string,
	cause error,
	violations ...FieldViolation,
) *Error {
	cloned := append([]FieldViolation(nil), violations...)
	return &Error{
		Reason:          ErrorReasonConflict,
		Message:         message,
		FieldViolations: cloned,
		Cause:           cause,
	}
}

// ReasonOf extracts a canonical reason through wrapped errors.
func ReasonOf(err error) ErrorReason {
	var domainError *Error
	if errors.As(err, &domainError) {
		return domainError.Reason
	}
	return ErrorReasonInternal
}

// ViolationsOf returns a defensive copy of validation details.
func ViolationsOf(err error) []FieldViolation {
	var domainError *Error
	if !errors.As(err, &domainError) {
		return nil
	}
	return append([]FieldViolation(nil), domainError.FieldViolations...)
}

// HasReason reports whether err contains the requested canonical reason.
func HasReason(err error, reason ErrorReason) bool {
	var domainError *Error
	return errors.As(err, &domainError) && domainError.Reason == reason
}
