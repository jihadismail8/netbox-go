// Package identity contains credential-free identity domain types.
package identity

import "context"

// Principal is the authenticated actor passed explicitly to every use case.
// It contains authorization claims, never credentials or password/token data.
type Principal struct {
	ID          int64
	Username    string
	Email       string
	FirstName   string
	LastName    string
	IsStaff     bool
	IsSuperuser bool
	Permissions map[string]struct{}
	// ObjectVisibility optionally constrains a permission to explicit object
	// IDs. Absence means no additional object constraint.
	ObjectVisibility map[string]map[int64]struct{}
}

func (p Principal) Authenticated() bool { return p.ID > 0 }
func (p Principal) Has(permission string) bool {
	if p.IsSuperuser {
		return true
	}
	_, ok := p.Permissions[permission]
	return ok
}

type principalContextKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok && principal.Authenticated()
}
