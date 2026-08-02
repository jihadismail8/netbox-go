// Package authz owns shared authorization decisions used by all transports.
package authz

type Action string

const (
	View   Action = "view"
	Add    Action = "add"
	Change Action = "change"
	Delete Action = "delete"
)

// ListScope is the complete object-visibility constraint for one authorized
// list operation. When Constrained is false every object of the kind is
// visible; when it is true only ObjectIDs are visible (including an empty
// slice, which means no objects are visible).
type ListScope struct {
	ObjectIDs   []int64
	Constrained bool
}

// PermissionAuthorizer implements the profile's fail-closed permission model,
// including persisted object visibility through the typed resource port.
type PermissionAuthorizer struct{}

// AllowAll is only for deterministic application tests and never runtime wiring.
type AllowAll struct{}
