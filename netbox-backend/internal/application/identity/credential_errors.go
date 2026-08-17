package identity

import "netbox-go/internal/domain/shared"

const unauthenticatedMessage = "Authentication credentials were not provided."

// TokenCredentialFailureKind identifies why a presented token credential was
// rejected without coupling the application service to a transport response.
type TokenCredentialFailureKind uint8

const (
	TokenCredentialFailureMissing TokenCredentialFailureKind = iota + 1
	TokenCredentialFailureUnknown
	TokenCredentialFailureRevoked
	TokenCredentialFailureExpired
	TokenCredentialFailureInactiveOwner
	TokenCredentialFailureSourceUnavailable
	TokenCredentialFailureSourceDenied
)

// TokenCredentialFailure carries the transport-neutral classification for a
// rejected token. SourceIP is populated only for a denied, parseable peer.
type TokenCredentialFailure struct {
	Kind     TokenCredentialFailureKind
	SourceIP string
}

func (*TokenCredentialFailure) Error() string {
	return "token credential rejected"
}

func tokenCredentialError(kind TokenCredentialFailureKind, sourceIP string) error {
	if kind != TokenCredentialFailureSourceDenied {
		sourceIP = ""
	}
	return shared.WrapError(
		shared.ErrorReasonUnauthenticated,
		unauthenticatedMessage,
		&TokenCredentialFailure{Kind: kind, SourceIP: sourceIP},
	)
}
