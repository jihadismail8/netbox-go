package identity

import (
	"strings"
	"time"
)

type User struct {
	ID          int64
	Username    string
	Email       string
	FirstName   string
	LastName    string
	IsStaff     bool
	IsSuperuser bool
	IsActive    bool
	Permissions []string
	// ObjectVisibility contains the object IDs granted for a permission when
	// that permission is object-scoped. A missing permission key means the
	// effective grant is global.
	ObjectVisibility map[string]map[int64]struct{}
	Created          time.Time
	Updated          time.Time
}

func (u User) Principal() Principal {
	permissions := make(map[string]struct{}, len(u.Permissions))
	for _, permission := range u.Permissions {
		permissions[permission] = struct{}{}
	}
	visibility := make(map[string]map[int64]struct{}, len(u.ObjectVisibility))
	// Superusers are never constrained by object grants. Keeping the map empty
	// also preserves that invariant if an administrator happens to be assigned
	// a narrower grant in addition to superuser status.
	if !u.IsSuperuser {
		for permission, ids := range u.ObjectVisibility {
			copied := make(map[int64]struct{}, len(ids))
			for id := range ids {
				copied[id] = struct{}{}
			}
			visibility[permission] = copied
		}
	}
	return Principal{ID: u.ID, Username: u.Username, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName, IsStaff: u.IsStaff, IsSuperuser: u.IsSuperuser, Permissions: permissions, ObjectVisibility: visibility}
}

// Group is a Go-owned RBAC group. Membership and grants are intentionally
// managed behind the identity application service rather than public routes.
type Group struct {
	ID   int64
	Name string
}

// PermissionGrant is one NetBox/Django-style model permission. ObjectID nil
// represents a global model grant; a non-nil value scopes it to that object.
type PermissionGrant struct {
	ID       int64
	Name     string
	AppLabel string
	Action   string
	Model    string
	ObjectID *int64
}

func (p PermissionGrant) Codename() string {
	return strings.ToLower(strings.TrimSpace(p.AppLabel)) + "." +
		strings.ToLower(strings.TrimSpace(p.Action)) + "_" +
		strings.ToLower(strings.ReplaceAll(strings.TrimSpace(p.Model), "_", ""))
}

type APIToken struct {
	ID           int64
	UserID       int64
	Display      string
	Description  string
	WriteEnabled bool
	AllowedIPs   []string
	Created      time.Time
	Expires      *time.Time
	LastUsed     *time.Time
}

type BrowserSession struct {
	User      User
	Secret    string
	CSRFToken string
	Expires   time.Time
}
