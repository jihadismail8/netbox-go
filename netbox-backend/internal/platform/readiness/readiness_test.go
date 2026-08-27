package readiness

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

type scriptedPinger struct {
	mu       sync.Mutex
	results  []error
	contexts []context.Context
}

func (p *scriptedPinger) PingContext(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.contexts = append(p.contexts, ctx)
	index := len(p.contexts) - 1
	if index >= len(p.results) {
		return errors.New("unexpected readiness ping")
	}
	return p.results[index]
}

func (p *scriptedPinger) snapshot() []context.Context {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]context.Context(nil), p.contexts...)
}

type pingerFunc func(context.Context) error

func (f pingerFunc) PingContext(ctx context.Context) error {
	return f(ctx)
}

func TestPostgreSQLReadinessTracksSuccessLossAndRecovery(t *testing.T) {
	dependencyCause := errors.New("private PostgreSQL endpoint rejected the connection")
	pinger := &scriptedPinger{results: []error{nil, dependencyCause, nil}}
	checker := NewPostgreSQL(pinger, time.Second)

	type contextKey struct{}
	marker := &struct{}{}
	ctx := context.WithValue(t.Context(), contextKey{}, marker)

	want := []error{nil, ErrUnavailable, nil}
	for index, wantErr := range want {
		err := checker.Check(ctx)
		if err != wantErr {
			t.Fatalf("Check() call %d error = %v, want %v", index+1, err, wantErr)
		}
		if err != nil && strings.Contains(err.Error(), dependencyCause.Error()) {
			t.Fatalf("Check() call %d disclosed dependency cause %q", index+1, err)
		}
	}

	contexts := pinger.snapshot()
	if len(contexts) != len(want) {
		t.Fatalf("PingContext() calls = %d, want %d", len(contexts), len(want))
	}
	for index, pingCtx := range contexts {
		if got := pingCtx.Value(contextKey{}); got != marker {
			t.Fatalf("PingContext() call %d lost caller context value", index+1)
		}
	}

	if err := NewPostgreSQL(nil, time.Second).Check(ctx); err != ErrUnavailable {
		t.Fatalf("nil pinger Check() error = %v, want %v", err, ErrUnavailable)
	}
	var typedNilPinger *scriptedPinger
	if err := NewPostgreSQL(typedNilPinger, time.Second).Check(ctx); err != ErrUnavailable {
		t.Fatalf("typed nil pinger Check() error = %v, want %v", err, ErrUnavailable)
	}
	invalidTimeoutPinger := &scriptedPinger{results: []error{nil}}
	if err := NewPostgreSQL(invalidTimeoutPinger, 0).Check(ctx); err != ErrUnavailable {
		t.Fatalf("invalid timeout Check() error = %v, want %v", err, ErrUnavailable)
	}
	if calls := len(invalidTimeoutPinger.snapshot()); calls != 0 {
		t.Fatalf("invalid timeout PingContext() calls = %d, want 0", calls)
	}
	var nilChecker *PostgreSQL
	if err := nilChecker.Check(ctx); err != ErrUnavailable {
		t.Fatalf("nil checker Check() error = %v, want %v", err, ErrUnavailable)
	}
}

func TestPostgreSQLReadinessHonorsCallerCancellationAndOneSecondCap(t *testing.T) {
	t.Run("earlier caller deadline wins", func(t *testing.T) {
		callerCtx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
		defer cancel()
		callerDeadline, ok := callerCtx.Deadline()
		if !ok {
			t.Fatal("caller context has no deadline")
		}

		var observedDeadline time.Time
		checker := NewPostgreSQL(pingerFunc(func(ctx context.Context) error {
			var deadlineOK bool
			observedDeadline, deadlineOK = ctx.Deadline()
			if !deadlineOK {
				return errors.New("bounded context has no deadline")
			}
			return nil
		}), time.Hour)

		if err := checker.Check(callerCtx); err != nil {
			t.Fatalf("Check() error = %v", err)
		}
		if !observedDeadline.Equal(callerDeadline) {
			t.Fatalf("PingContext() deadline = %v, want caller deadline %v", observedDeadline, callerDeadline)
		}
	})

	t.Run("checker caps a longer caller", func(t *testing.T) {
		const maxTimeout = time.Second
		var observedDeadline time.Time
		checker := NewPostgreSQL(pingerFunc(func(ctx context.Context) error {
			var ok bool
			observedDeadline, ok = ctx.Deadline()
			if !ok {
				return errors.New("bounded context has no deadline")
			}
			return nil
		}), maxTimeout)

		before := time.Now()
		if err := checker.Check(t.Context()); err != nil {
			t.Fatalf("Check() error = %v", err)
		}
		after := time.Now()
		if observedDeadline.Before(before.Add(maxTimeout)) || observedDeadline.After(after.Add(maxTimeout)) {
			t.Fatalf(
				"PingContext() deadline = %v, want between %v and %v",
				observedDeadline,
				before.Add(maxTimeout),
				after.Add(maxTimeout),
			)
		}
	})

	t.Run("caller cancellation reaches one ping and remains non-disclosing", func(t *testing.T) {
		callerCtx, cancel := context.WithCancel(t.Context())
		cancel()

		calls := 0
		checker := NewPostgreSQL(pingerFunc(func(ctx context.Context) error {
			calls++
			if err := ctx.Err(); err != context.Canceled {
				return errors.New("caller cancellation was not propagated")
			}
			return context.Canceled
		}), time.Second)

		err := checker.Check(callerCtx)
		if err != ErrUnavailable {
			t.Fatalf("Check() error = %v, want %v", err, ErrUnavailable)
		}
		if calls != 1 {
			t.Fatalf("PingContext() calls = %d, want 1", calls)
		}
		if strings.Contains(err.Error(), context.Canceled.Error()) {
			t.Fatalf("Check() disclosed cancellation cause %q", err)
		}
	})
}
