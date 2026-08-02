package shared

import "time"

// Timestamp is a UTC domain timestamp. Adapters are responsible for converting
// it to their storage or wire representation.
type Timestamp struct {
	time.Time
}

func NewTimestamp(value time.Time) Timestamp {
	if value.IsZero() {
		return Timestamp{}
	}
	return Timestamp{Time: value.UTC().Round(0)}
}

// Clock makes mutation timestamps deterministic in application tests.
type Clock interface {
	Now() Timestamp
}

// SystemClock is the production wall clock implementation.
type SystemClock struct{}

func (SystemClock) Now() Timestamp {
	return NewTimestamp(time.Now())
}
