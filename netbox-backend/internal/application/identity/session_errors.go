package identity

import (
	"errors"

	"netbox-go/internal/domain/shared"
)

// SessionCredentialFailureKind identifies an expected browser-session
// rejection without coupling the application service to a transport response.
type SessionCredentialFailureKind uint8

const (
	SessionCredentialFailureMissing SessionCredentialFailureKind = iota + 1
	SessionCredentialFailureUnknown
	SessionCredentialFailureExpired
	SessionCredentialFailureInactiveOwner
)

// SessionCredentialFailure carries only the transport-neutral classification
// of a rejected session. It deliberately contains no credential material.
type SessionCredentialFailure struct {
	Kind SessionCredentialFailureKind
}

func (*SessionCredentialFailure) Error() string {
	return "session credential rejected"
}

func sessionCredentialError(kind SessionCredentialFailureKind) error {
	return shared.WrapError(
		shared.ErrorReasonUnauthenticated,
		unauthenticatedMessage,
		&SessionCredentialFailure{Kind: kind},
	)
}

// SessionCredentialAllowsTokenFallback reports whether a rejected browser
// session is one of the expected credential states for which an adapter may
// try another credential. A broad unauthenticated reason alone is insufficient.
func SessionCredentialAllowsTokenFallback(err error) bool {
	if shared.ReasonOf(err) != shared.ErrorReasonUnauthenticated {
		return false
	}
	var failure *SessionCredentialFailure
	if !errors.As(err, &failure) {
		return false
	}
	switch failure.Kind {
	case SessionCredentialFailureMissing,
		SessionCredentialFailureUnknown,
		SessionCredentialFailureExpired,
		SessionCredentialFailureInactiveOwner:
		return true
	default:
		return false
	}
}
