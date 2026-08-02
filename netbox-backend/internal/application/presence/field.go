// Package presence models transport-neutral command-field presence.
package presence

// FieldState distinguishes an omitted field, an explicit null, and a concrete
// value. The zero state is deliberately Omitted so zero-value commands are
// safe partial updates.
type FieldState uint8

const (
	Omitted FieldState = iota
	Null
	Present
)

// Field preserves omitted, explicit-null, and concrete-value intent without
// exposing a REST, protobuf, or persistence representation to application
// use cases. Its state is immutable after construction.
type Field[T any] struct {
	state FieldState
	value T
}

func OmittedField[T any]() Field[T] { return Field[T]{state: Omitted} }

func NullField[T any]() Field[T] { return Field[T]{state: Null} }

func Value[T any](value T) Field[T] {
	return Field[T]{state: Present, value: value}
}

func (field Field[T]) State() FieldState { return field.state }

// Get returns a value only for Present. State retains the distinction between
// omitted and explicit null.
func (field Field[T]) Get() (T, bool) {
	return field.value, field.state == Present
}
