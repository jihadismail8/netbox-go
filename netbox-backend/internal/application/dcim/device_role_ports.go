package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type DeviceRoleListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []DeviceRoleSort
	DefaultTreeOrder      bool
	Names                 []string
	Slugs                 []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type DeviceRolePage struct {
	Count   uint64
	Results []*dcimdomain.DeviceRole
}

type DeviceRoleDependent struct {
	ID      shared.ID
	Display string
}

type DeviceRoleRepository interface {
	List(context.Context, DeviceRoleListCriteria) (DeviceRolePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.DeviceRole, error)
	ListHierarchyForUpdate(context.Context) ([]*dcimdomain.DeviceRole, error)
	Create(context.Context, *dcimdomain.DeviceRole) error
	Update(context.Context, *dcimdomain.DeviceRole) error
	FindDeviceUsingRoles(context.Context, []shared.ID) (*DeviceRoleDependent, error)
	Delete(context.Context, *dcimdomain.DeviceRole) error
}
