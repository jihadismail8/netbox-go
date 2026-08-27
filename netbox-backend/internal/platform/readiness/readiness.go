// Package readiness owns process-level dependency readiness checks.
package readiness

import (
	"context"
	"errors"
	"reflect"
	"time"
)

// ErrUnavailable is the only error exposed when a required dependency is not
// ready. Dependency causes are intentionally not retained or wrapped.
var ErrUnavailable = errors.New("readiness unavailable")

// Checker reports whether the process's required dependencies are ready.
type Checker interface {
	Check(context.Context) error
}

// Pinger is the narrow PostgreSQL connectivity contract used by readiness.
type Pinger interface {
	PingContext(context.Context) error
}

// PostgreSQL checks one PostgreSQL connection attempt per call.
type PostgreSQL struct {
	pinger     Pinger
	maxTimeout time.Duration
}

// NewPostgreSQL constructs a PostgreSQL readiness checker. Invalid or missing
// dependencies remain representable so production composition fails closed at
// check time rather than installing an always-ready fallback.
func NewPostgreSQL(pinger Pinger, maxTimeout time.Duration) *PostgreSQL {
	return &PostgreSQL{pinger: pinger, maxTimeout: maxTimeout}
}

// Check performs one bounded connectivity check. It deliberately collapses
// every failure to ErrUnavailable so transport adapters cannot expose a
// database cause or configuration detail.
func (c *PostgreSQL) Check(ctx context.Context) error {
	if c == nil || nilPinger(c.pinger) || c.maxTimeout <= 0 || ctx == nil {
		return ErrUnavailable
	}

	checkCtx, cancel := context.WithTimeout(ctx, c.maxTimeout)
	defer cancel()

	if err := c.pinger.PingContext(checkCtx); err != nil {
		return ErrUnavailable
	}
	return nil
}

func nilPinger(pinger Pinger) bool {
	if pinger == nil {
		return true
	}
	value := reflect.ValueOf(pinger)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
