package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type DeviceTypeListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []DeviceTypeSort
	ManufacturerIDs       []int64
	ManufacturerSlugs     []string
	Models                []string
	Slugs                 []string
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type DeviceTypePage struct {
	Count   uint64
	Results []*dcimdomain.DeviceType
}

type DeviceTypeDependent struct {
	ID      shared.ID
	Display string
}

// PositionedDevice is the typed projection needed to validate a DeviceType
// height transition against every mounted instance.
type PositionedDevice struct {
	ID                    shared.ID
	DeviceTypeID          shared.ID
	RackID                shared.ID
	PositionHalfUnits     uint32
	Face                  string
	StoredHeightHalfUnits uint32
	StoredFullDepth       bool
	RackStartingUnit      uint32
	RackUHeight           uint32
}

type InterfaceTemplateDeletion struct {
	ID             shared.ID
	Representation string
	Snapshot       dcimdomain.InterfaceTemplateSnapshot
}

type DeviceTypeRepository interface {
	List(context.Context, DeviceTypeListCriteria) (DeviceTypePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.DeviceType, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.DeviceType, error)
	Create(context.Context, *dcimdomain.DeviceType) error
	Update(context.Context, *dcimdomain.DeviceType) error
	ListPositionedDevicesForUpdate(context.Context) ([]PositionedDevice, error)
	FindDeviceUsingDeviceType(context.Context, shared.ID) (*DeviceTypeDependent, error)
	ListInterfaceTemplatesForUpdate(
		context.Context,
		shared.ID,
	) ([]InterfaceTemplateDeletion, error)
	DeleteInterfaceTemplate(context.Context, shared.ID) error
	Delete(context.Context, *dcimdomain.DeviceType) error
}

type DeviceTypeManufacturerReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.Manufacturer, error)
}
