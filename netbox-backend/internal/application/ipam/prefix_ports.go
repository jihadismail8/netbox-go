package ipam

import (
	"context"

	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type PrefixListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []PrefixSort
	VRFIDs                []int64
	VRFRDs                []string
	Prefixes              []domainipam.PrefixNetwork
	PrefixesPresent       bool
	Family                *int64
	Statuses              []domainipam.PrefixStatus
	Within                *PrefixNetworkFilter
	WithinInclude         *PrefixNetworkFilter
	Contains              *PrefixNetworkFilter
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type PrefixPage struct {
	Count   uint64
	Results []*domainipam.Prefix
}

type PrefixRepository interface {
	List(context.Context, PrefixListCriteria) (PrefixPage, error)
	Get(context.Context, shared.ID) (*domainipam.Prefix, error)
	GetForUpdate(context.Context, shared.ID) (*domainipam.Prefix, error)
	Create(context.Context, *domainipam.Prefix) error
	Update(context.Context, *domainipam.Prefix) error
	Delete(context.Context, *domainipam.Prefix) error
	LockUniqueness(context.Context, domainipam.NullableVRFReference, domainipam.PrefixNetwork) error
	FindDuplicate(context.Context, domainipam.NullableVRFReference, domainipam.PrefixNetwork, shared.ID) (*domainipam.Prefix, error)
}

type PrefixVRFReader interface {
	Get(context.Context, shared.ID) (*domainipam.VRF, error)
}
