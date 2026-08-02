package workflow

import (
	"context"

	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func principal(ctx context.Context) (identity.Principal, error) {
	value, ok := identity.PrincipalFromContext(ctx)
	if !ok {
		return value, shared.NewError(
			shared.ErrorReasonUnauthenticated,
			"Authentication credentials were not provided.",
		)
	}
	return value, nil
}
