package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type RackTypeListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []RackTypeSort
	ManufacturerIDs       []int64
	ManufacturerSlugs     []string
	Models                []string
	Slugs                 []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type RackTypePage struct {
	Count   uint64
	Results []*dcimdomain.RackType
}

type RackPropagationChange struct {
	ID             shared.ID
	Representation string
	Before         dcimdomain.RackSnapshot
	After          dcimdomain.RackSnapshot
}

type RackTypeRepository interface {
	List(context.Context, RackTypeListCriteria) (RackTypePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.RackType, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.RackType, error)
	Create(context.Context, *dcimdomain.RackType) error
	Update(context.Context, *dcimdomain.RackType) error
	Delete(context.Context, *dcimdomain.RackType) error
	PropagateToRacks(
		context.Context,
		shared.ID,
		dcimdomain.RackPhysicalAttributes,
		shared.Timestamp,
	) ([]RackPropagationChange, error)
}

type RackTypeManufacturerReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.Manufacturer, error)
}
