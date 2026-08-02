package ipam

import (
	"context"

	ipamdomain "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type VRFListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []VRFSort
	Names                 []string
	RDs                   []ipamdomain.RouteDistinguisher
	EnforceUnique         *bool
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	// DeferPagination requests all SQL-filtered rows so a custom authorizer
	// without a complete object-ID scope can be applied before count/page.
	DeferPagination bool
}

type VRFPage struct {
	Count   uint64
	Results []*ipamdomain.VRF
}

type VRFRepository interface {
	List(context.Context, VRFListCriteria) (VRFPage, error)
	Get(context.Context, shared.ID) (*ipamdomain.VRF, error)
	GetForUpdate(context.Context, shared.ID) (*ipamdomain.VRF, error)
	Create(context.Context, *ipamdomain.VRF) error
	Update(context.Context, *ipamdomain.VRF) error
	Delete(context.Context, *ipamdomain.VRF) error
}
