package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type RackListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []RackSort
	SiteIDs               []int64
	SiteSlugs             []string
	Names                 []string
	Statuses              []dcimdomain.RackStatus
	RoleIDs               []int64
	RoleSlugs             []string
	RackTypeIDs           []int64
	RackTypeSlugs         []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type RackPage struct {
	Count   uint64
	Results []*dcimdomain.Rack
}

type RackDevicePlacement struct {
	ID                shared.ID
	PositionHalfUnits int32
	HeightHalfUnits   uint16
}

type RackSitePropagationChange struct {
	ID             shared.ID
	Representation string
	Before         shared.ObjectSnapshot
	After          shared.ObjectSnapshot
}

type RackRepository interface {
	List(context.Context, RackListCriteria) (RackPage, error)
	Get(context.Context, shared.ID) (*dcimdomain.Rack, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Rack, error)
	Create(context.Context, *dcimdomain.Rack) error
	Update(context.Context, *dcimdomain.Rack) error
	Delete(context.Context, *dcimdomain.Rack) error
	MountedDevices(context.Context, shared.ID) ([]RackDevicePlacement, error)
	PropagateSiteToDevices(
		context.Context,
		shared.ID,
		shared.ID,
		shared.Timestamp,
	) ([]RackSitePropagationChange, error)
}

type RackSiteReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.Site, error)
}

type RackTypeReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.RackType, error)
}

type RackRoleReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.RackRole, error)
}
