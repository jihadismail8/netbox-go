package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type DeviceListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []DeviceSort
	SiteIDs               []int64
	SiteSlugs             []string
	RackIDs               []int64
	DeviceTypeIDs         []int64
	DeviceTypeSlugs       []string
	RoleIDs               []int64
	RoleSlugs             []string
	Names                 []string
	Statuses              []dcimdomain.DeviceStatus
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type DevicePage struct {
	Count   uint64
	Results []*dcimdomain.Device
}

type DeviceRackOccupant struct {
	ID                shared.ID
	PositionHalfUnits uint16
	HeightHalfUnits   uint16
	Face              dcimdomain.DeviceFace
	FullDepth         bool
}

type DeviceRepository interface {
	List(context.Context, DeviceListCriteria) (DevicePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.Device, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Device, error)
	GetDeviceReference(context.Context, shared.ID) (dcimdomain.DeviceReference, error)
	Create(context.Context, *dcimdomain.Device) error
	Update(context.Context, *dcimdomain.Device) error
	Delete(context.Context, *dcimdomain.Device) error
	ListRackOccupantsForUpdate(
		context.Context,
		shared.ID,
		shared.ID,
	) ([]DeviceRackOccupant, error)
}

type DeviceTypeReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.DeviceType, error)
}

type DeviceRoleReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.DeviceRole, error)
}

type DeviceSiteReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.Site, error)
}

type DeviceRackReader interface {
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Rack, error)
}

type DeviceInterfaceTemplateReader interface {
	List(context.Context, InterfaceTemplateListCriteria) (InterfaceTemplatePage, error)
}

type DeviceInterfaceCreator interface {
	Create(context.Context, *dcimdomain.Interface) error
}
