package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type ManufacturerListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []ManufacturerSort
	Names                 []string
	Slugs                 []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type ManufacturerPage struct {
	Count   uint64
	Results []*dcimdomain.Manufacturer
}

type ManufacturerRepository interface {
	List(context.Context, ManufacturerListCriteria) (ManufacturerPage, error)
	Get(context.Context, shared.ID) (*dcimdomain.Manufacturer, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Manufacturer, error)
	Create(context.Context, *dcimdomain.Manufacturer) error
	Update(context.Context, *dcimdomain.Manufacturer) error
	Delete(context.Context, *dcimdomain.Manufacturer) error
}

type RackRoleListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []RackRoleSort
	Names                 []string
	Slugs                 []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type RackRolePage struct {
	Count   uint64
	Results []*dcimdomain.RackRole
}

type RackRoleRepository interface {
	List(context.Context, RackRoleListCriteria) (RackRolePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.RackRole, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.RackRole, error)
	Create(context.Context, *dcimdomain.RackRole) error
	Update(context.Context, *dcimdomain.RackRole) error
	Delete(context.Context, *dcimdomain.RackRole) error
}
