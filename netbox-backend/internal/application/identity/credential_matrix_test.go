package identity_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type credentialTouch struct {
	id int64
	at time.Time
}

type credentialSpyStore struct {
	application.Store
	record      application.TokenRecord
	user        domain.User
	lookupErr   error
	touchErr    error
	lookupCalls int
	touches     []credentialTouch
}

func (store *credentialSpyStore) TokenByHash(context.Context, []byte) (application.TokenRecord, domain.User, error) {
	store.lookupCalls++
	return store.record, store.user, store.lookupErr
}

func (store *credentialSpyStore) TouchToken(_ context.Context, id int64, at time.Time) error {
	store.touches = append(store.touches, credentialTouch{id: id, at: at})
	return store.touchErr
}

func TestTokenCredentialMatrixLookupClassification(t *testing.T) {
	now := time.Date(2026, 8, 17, 6, 0, 0, 0, time.UTC)
	infrastructureFailure := errors.New("credential lookup unavailable")

	tests := []struct {
		name        string
		secret      string
		lookupErr   error
		wantReason  shared.ErrorReason
		wantCause   error
		wantLookups int
	}{
		{
			name:        "missing",
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantLookups: 0,
		},
		{
			name:        "unknown",
			secret:      "present",
			lookupErr:   application.ErrNotFound,
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantLookups: 1,
		},
		{
			name:        "infrastructure failure",
			secret:      "present",
			lookupErr:   infrastructureFailure,
			wantReason:  shared.ErrorReasonInternal,
			wantCause:   infrastructureFailure,
			wantLookups: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &credentialSpyStore{lookupErr: test.lookupErr}
			service := application.NewService(store, &testClock{now: now})

			user, err := service.AuthenticateToken(t.Context(), test.secret, "192.0.2.1", false)

			require.Equal(t, domain.User{}, user)
			require.Equal(t, test.wantReason, shared.ReasonOf(err))
			if test.wantCause != nil {
				require.ErrorIs(t, err, test.wantCause)
			}
			require.Equal(t, test.wantLookups, store.lookupCalls)
			require.Empty(t, store.touches)
		})
	}
}

func TestTokenCredentialOutcomeClassification(t *testing.T) {
	now := time.Date(2026, 8, 17, 6, 30, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	expired := now
	revoked := now.Add(-time.Hour)
	lookupFailure := errors.New("credential lookup unavailable")
	touchFailure := errors.New("credential touch unavailable")
	activeUser := activeCredentialUser()

	restrictedToken := func(networks ...string) domain.APIToken {
		token := credentialToken(activeUser.ID, &stale)
		token.AllowedIPs = append([]string(nil), networks...)
		return token
	}

	tests := []struct {
		name         string
		secret       string
		record       application.TokenRecord
		user         domain.User
		lookupErr    error
		touchErr     error
		remote       string
		write        bool
		wantReason   shared.ErrorReason
		wantMessage  string
		wantKind     application.TokenCredentialFailureKind
		wantSourceIP string
		wantTyped    bool
		wantLookups  int
		wantTouch    bool
		wantUser     bool
		wantCause    error
	}{
		{
			name:        "missing",
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureMissing,
			wantTyped:   true,
			wantLookups: 0,
		},
		{
			name:        "unknown",
			secret:      "present",
			lookupErr:   application.ErrNotFound,
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureUnknown,
			wantTyped:   true,
			wantLookups: 1,
		},
		{
			name:   "revoked",
			secret: "present",
			record: application.TokenRecord{
				Token:     credentialToken(activeUser.ID, &stale),
				RevokedAt: &revoked,
			},
			user:        activeUser,
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureRevoked,
			wantTyped:   true,
			wantLookups: 1,
		},
		{
			name:   "expired",
			secret: "present",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(activeUser.ID, &stale)
				token.Expires = &expired
				return token
			}()},
			user:        activeUser,
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureExpired,
			wantTyped:   true,
			wantLookups: 1,
			wantTouch:   true,
		},
		{
			name:        "inactive owner",
			secret:      "present",
			record:      application.TokenRecord{Token: credentialToken(activeUser.ID, &stale)},
			user:        domain.User{ID: activeUser.ID, Username: "inactive"},
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureInactiveOwner,
			wantTyped:   true,
			wantLookups: 1,
			wantTouch:   true,
		},
		{
			name:        "restricted source unavailable when absent",
			secret:      "present",
			record:      application.TokenRecord{Token: restrictedToken("192.0.2.0/24")},
			user:        activeUser,
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureSourceUnavailable,
			wantTyped:   true,
			wantLookups: 1,
			wantTouch:   true,
		},
		{
			name:        "restricted source unavailable when malformed",
			secret:      "present",
			record:      application.TokenRecord{Token: restrictedToken("192.0.2.0/24")},
			user:        activeUser,
			remote:      "not-an-address",
			wantReason:  shared.ErrorReasonUnauthenticated,
			wantMessage: "Authentication credentials were not provided.",
			wantKind:    application.TokenCredentialFailureSourceUnavailable,
			wantTyped:   true,
			wantLookups: 1,
			wantTouch:   true,
		},
		{
			name:         "restricted IPv4 source denied and canonicalized",
			secret:       "present",
			record:       application.TokenRecord{Token: restrictedToken("192.0.2.0/24")},
			user:         activeUser,
			remote:       "198.51.100.1:443",
			wantReason:   shared.ErrorReasonUnauthenticated,
			wantMessage:  "Authentication credentials were not provided.",
			wantKind:     application.TokenCredentialFailureSourceDenied,
			wantSourceIP: "198.51.100.1",
			wantTyped:    true,
			wantLookups:  1,
			wantTouch:    true,
		},
		{
			name:         "restricted IPv6 source denied and canonicalized",
			secret:       "present",
			record:       application.TokenRecord{Token: restrictedToken("2001:db8::/32")},
			user:         activeUser,
			remote:       "[2001:0db9:0:0::1]:443",
			wantReason:   shared.ErrorReasonUnauthenticated,
			wantMessage:  "Authentication credentials were not provided.",
			wantKind:     application.TokenCredentialFailureSourceDenied,
			wantSourceIP: "2001:db9::1",
			wantTyped:    true,
			wantLookups:  1,
			wantTouch:    true,
		},
		{
			name:         "invalid persisted prefix fails closed",
			secret:       "present",
			record:       application.TokenRecord{Token: restrictedToken("invalid-prefix")},
			user:         activeUser,
			remote:       "198.51.100.2",
			wantReason:   shared.ErrorReasonUnauthenticated,
			wantMessage:  "Authentication credentials were not provided.",
			wantKind:     application.TokenCredentialFailureSourceDenied,
			wantSourceIP: "198.51.100.2",
			wantTyped:    true,
			wantLookups:  1,
			wantTouch:    true,
		},
		{
			name:   "source denial precedes write denial",
			secret: "present",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := restrictedToken("192.0.2.0/24")
				token.WriteEnabled = false
				return token
			}()},
			user:         activeUser,
			remote:       "198.51.100.3",
			write:        true,
			wantReason:   shared.ErrorReasonUnauthenticated,
			wantMessage:  "Authentication credentials were not provided.",
			wantKind:     application.TokenCredentialFailureSourceDenied,
			wantSourceIP: "198.51.100.3",
			wantTyped:    true,
			wantLookups:  1,
			wantTouch:    true,
		},
		{
			name:        "unrestricted token ignores malformed source",
			secret:      "present",
			record:      application.TokenRecord{Token: credentialToken(activeUser.ID, &stale)},
			user:        activeUser,
			remote:      "not-an-address",
			wantLookups: 1,
			wantTouch:   true,
			wantUser:    true,
		},
		{
			name:        "restricted bare IPv4 source allowed",
			secret:      "present",
			record:      application.TokenRecord{Token: restrictedToken("192.0.2.0/24")},
			user:        activeUser,
			remote:      "192.0.2.7",
			wantLookups: 1,
			wantTouch:   true,
			wantUser:    true,
		},
		{
			name:        "restricted bare IPv6 source allowed",
			secret:      "present",
			record:      application.TokenRecord{Token: restrictedToken("2001:db8::/32")},
			user:        activeUser,
			remote:      "2001:db8::7",
			wantLookups: 1,
			wantTouch:   true,
			wantUser:    true,
		},
		{
			name:   "write disabled",
			secret: "present",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(activeUser.ID, &stale)
				token.WriteEnabled = false
				return token
			}()},
			user:        activeUser,
			remote:      "192.0.2.1",
			write:       true,
			wantReason:  shared.ErrorReasonForbidden,
			wantMessage: "You do not have permission to perform this action.",
			wantLookups: 1,
			wantTouch:   true,
		},
		{
			name:        "lookup infrastructure failure",
			secret:      "present",
			lookupErr:   lookupFailure,
			wantReason:  shared.ErrorReasonInternal,
			wantMessage: "An internal error occurred.",
			wantLookups: 1,
			wantCause:   lookupFailure,
		},
		{
			name:        "touch infrastructure failure",
			secret:      "present",
			record:      application.TokenRecord{Token: credentialToken(activeUser.ID, &stale)},
			user:        activeUser,
			touchErr:    touchFailure,
			wantReason:  shared.ErrorReasonInternal,
			wantMessage: "An internal error occurred.",
			wantLookups: 1,
			wantTouch:   true,
			wantCause:   touchFailure,
		},
		{
			name:        "valid",
			secret:      "present",
			record:      application.TokenRecord{Token: credentialToken(activeUser.ID, &stale)},
			user:        activeUser,
			remote:      "192.0.2.1",
			wantLookups: 1,
			wantTouch:   true,
			wantUser:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &credentialSpyStore{
				record:    test.record,
				user:      test.user,
				lookupErr: test.lookupErr,
				touchErr:  test.touchErr,
			}
			service := application.NewService(store, &testClock{now: now})

			user, err := service.AuthenticateToken(t.Context(), test.secret, test.remote, test.write)

			var credentialFailure *application.TokenCredentialFailure
			if test.wantUser {
				require.NoError(t, err)
				require.Equal(t, test.user, user)
				require.False(t, errors.As(err, &credentialFailure))
			} else {
				require.Equal(t, domain.User{}, user)
				require.Equal(t, test.wantReason, shared.ReasonOf(err))
				var appErr *shared.Error
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, test.wantMessage, appErr.Message)
				if test.wantCause != nil {
					require.ErrorIs(t, err, test.wantCause)
				}
				if test.wantTyped {
					require.True(t, errors.As(err, &credentialFailure), "expected typed credential failure")
					require.Equal(t, test.wantKind, credentialFailure.Kind)
					require.Equal(t, test.wantSourceIP, credentialFailure.SourceIP)
				} else {
					require.False(t, errors.As(err, &credentialFailure))
				}
			}
			require.Equal(t, test.wantLookups, store.lookupCalls)
			if test.wantTouch {
				require.Equal(t, []credentialTouch{{id: 17, at: now}}, store.touches)
			} else {
				require.Empty(t, store.touches)
			}
		})
	}
}

func TestTokenCredentialMatrixLastUsedStrictBoundary(t *testing.T) {
	now := time.Date(2026, 8, 17, 6, 0, 0, 0, time.UTC)
	user := activeCredentialUser()

	tests := []struct {
		name      string
		lastUsed  *time.Time
		wantTouch bool
	}{
		{name: "absent", wantTouch: true},
		{name: "below one minute", lastUsed: timePointer(now.Add(-time.Minute + time.Nanosecond))},
		{name: "exactly one minute", lastUsed: timePointer(now.Add(-time.Minute))},
		{name: "over one minute", lastUsed: timePointer(now.Add(-time.Minute - time.Nanosecond)), wantTouch: true},
		{name: "future timestamp", lastUsed: timePointer(now.Add(time.Minute))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &credentialSpyStore{
				record: application.TokenRecord{Token: credentialToken(user.ID, test.lastUsed)},
				user:   user,
			}
			service := application.NewService(store, &testClock{now: now})

			authenticated, err := service.AuthenticateToken(t.Context(), "present", "192.0.2.1", false)

			require.NoError(t, err)
			require.Equal(t, user, authenticated)
			require.Equal(t, 1, store.lookupCalls)
			if !test.wantTouch {
				require.Empty(t, store.touches)
				return
			}
			require.Equal(t, []credentialTouch{{id: 17, at: now}}, store.touches)
		})
	}
}

func TestTokenCredentialMatrixRestrictionOrdering(t *testing.T) {
	now := time.Date(2026, 8, 17, 6, 0, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	expiredBefore := now.Add(-time.Nanosecond)
	expiredAt := now
	expiresAfter := now.Add(time.Nanosecond)
	revoked := now.Add(-time.Hour)

	tests := []struct {
		name       string
		record     application.TokenRecord
		user       domain.User
		remote     string
		write      bool
		wantReason shared.ErrorReason
		wantTouch  bool
		wantUser   bool
	}{
		{
			name: "revoked",
			record: application.TokenRecord{
				Token:     credentialToken(41, nil),
				RevokedAt: &revoked,
			},
			user:       activeCredentialUser(),
			remote:     "192.0.2.1",
			wantReason: shared.ErrorReasonUnauthenticated,
		},
		{
			name: "expiry before boundary",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.Expires = &expiredBefore
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "192.0.2.1",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "expiry at boundary",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.Expires = &expiredAt
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "192.0.2.1",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "expiry after boundary",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.Expires = &expiresAfter
				return token
			}()},
			user:      activeCredentialUser(),
			remote:    "192.0.2.1",
			wantTouch: true,
			wantUser:  true,
		},
		{
			name:   "inactive owner",
			record: application.TokenRecord{Token: credentialToken(41, &stale)},
			user: domain.User{
				ID: 41, Username: "inactive", IsActive: false,
			},
			remote:     "192.0.2.1",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "source IPv4 allowed",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user:      activeCredentialUser(),
			remote:    "192.0.2.7:443",
			wantTouch: true,
			wantUser:  true,
		},
		{
			name: "source IPv6 allowed",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"2001:db8::/32"}
				return token
			}()},
			user:      activeCredentialUser(),
			remote:    "[2001:db8::7]:443",
			wantTouch: true,
			wantUser:  true,
		},
		{
			name: "source IPv6 denied",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"2001:db8::/32"}
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "[2001:db9::7]:443",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "source address missing",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user:       activeCredentialUser(),
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "source address malformed",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "not-an-address",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "one of multiple prefixes matches",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24", "2001:db8::/32"}
				return token
			}()},
			user:      activeCredentialUser(),
			remote:    "[2001:db8::9]:443",
			wantTouch: true,
			wantUser:  true,
		},
		{
			name: "source IP denied",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "198.51.100.1",
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "source IP denial precedes write denial",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.WriteEnabled = false
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "198.51.100.1",
			write:      true,
			wantReason: shared.ErrorReasonUnauthenticated,
			wantTouch:  true,
		},
		{
			name: "write disabled safe operation",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.WriteEnabled = false
				return token
			}()},
			user:      activeCredentialUser(),
			remote:    "192.0.2.1",
			wantTouch: true,
			wantUser:  true,
		},
		{
			name: "write disabled unsafe operation",
			record: application.TokenRecord{Token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.WriteEnabled = false
				return token
			}()},
			user:       activeCredentialUser(),
			remote:     "192.0.2.1",
			write:      true,
			wantReason: shared.ErrorReasonForbidden,
			wantTouch:  true,
		},
		{
			name:      "write enabled unsafe operation",
			record:    application.TokenRecord{Token: credentialToken(41, &stale)},
			user:      activeCredentialUser(),
			remote:    "192.0.2.1",
			write:     true,
			wantTouch: true,
			wantUser:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &credentialSpyStore{record: test.record, user: test.user}
			service := application.NewService(store, &testClock{now: now})

			user, err := service.AuthenticateToken(t.Context(), "present", test.remote, test.write)

			if test.wantUser {
				require.NoError(t, err)
				require.Equal(t, test.user, user)
			} else {
				require.Equal(t, domain.User{}, user)
				require.Equal(t, test.wantReason, shared.ReasonOf(err))
			}
			if test.wantTouch {
				require.Equal(t, []credentialTouch{{id: 17, at: now}}, store.touches)
			} else {
				require.Empty(t, store.touches)
			}
		})
	}
}

func TestTokenCredentialMatrixTouchFailurePrecedesRestrictions(t *testing.T) {
	now := time.Date(2026, 8, 17, 6, 0, 0, 0, time.UTC)
	stale := now.Add(-2 * time.Minute)
	expired := now
	touchFailure := errors.New("credential touch unavailable")

	tests := []struct {
		name   string
		token  domain.APIToken
		user   domain.User
		remote string
		write  bool
	}{
		{
			name:  "otherwise valid",
			token: credentialToken(41, &stale),
			user:  activeCredentialUser(), remote: "192.0.2.1",
		},
		{
			name: "expired",
			token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.Expires = &expired
				return token
			}(),
			user: activeCredentialUser(), remote: "192.0.2.1",
		},
		{
			name:  "inactive owner",
			token: credentialToken(41, &stale),
			user:  domain.User{ID: 41, Username: "inactive"}, remote: "192.0.2.1",
		},
		{
			name: "source IP denied",
			token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.AllowedIPs = []string{"192.0.2.0/24"}
				return token
			}(),
			user: activeCredentialUser(), remote: "198.51.100.1",
		},
		{
			name: "write disabled",
			token: func() domain.APIToken {
				token := credentialToken(41, &stale)
				token.WriteEnabled = false
				return token
			}(),
			user: activeCredentialUser(), remote: "192.0.2.1", write: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &credentialSpyStore{
				record:   application.TokenRecord{Token: test.token},
				user:     test.user,
				touchErr: touchFailure,
			}
			service := application.NewService(store, &testClock{now: now})

			user, err := service.AuthenticateToken(t.Context(), "present", test.remote, test.write)

			require.Equal(t, domain.User{}, user)
			require.Equal(t, shared.ErrorReasonInternal, shared.ReasonOf(err))
			require.ErrorIs(t, err, touchFailure)
			require.Equal(t, []credentialTouch{{id: 17, at: now}}, store.touches)
		})
	}
}

func activeCredentialUser() domain.User {
	return domain.User{ID: 41, Username: "credential-user", IsActive: true}
}

func credentialToken(userID int64, lastUsed *time.Time) domain.APIToken {
	return domain.APIToken{
		ID: 17, UserID: userID, WriteEnabled: true, LastUsed: lastUsed,
	}
}

func timePointer(value time.Time) *time.Time { return &value }
