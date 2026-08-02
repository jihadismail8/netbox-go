package dcim

import (
	"context"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type InterfaceTemplateListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []InterfaceTemplateSort
	DeviceTypeIDs         []int64
	Names                 []string
	Types                 []dcimdomain.InterfaceType
	Enabled               *bool
	MgmtOnly              *bool
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type InterfaceTemplatePage struct {
	Count   uint64
	Results []*dcimdomain.InterfaceTemplate
}

type InterfaceTemplateRepository interface {
	List(context.Context, InterfaceTemplateListCriteria) (InterfaceTemplatePage, error)
	Get(context.Context, shared.ID) (*dcimdomain.InterfaceTemplate, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.InterfaceTemplate, error)
	Create(context.Context, *dcimdomain.InterfaceTemplate) error
	Update(context.Context, *dcimdomain.InterfaceTemplate) error
	Delete(context.Context, *dcimdomain.InterfaceTemplate) error
}

type InterfaceTemplateDeviceTypeReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.DeviceType, error)
}
