package ipam

import (
	"context"

	domaindcim "netbox-go/internal/domain/dcim"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type IPAddressListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []IPAddressSort
	VRFIDs                []int64
	VRFRDs                []string
	Addresses             []IPAddressFilter
	AddressesPresent      bool
	Family                *int64
	Parent                *IPAddressParentFilter
	Statuses              []domainipam.IPAddressStatus
	Assigned              *bool
	InterfaceIDs          []int64
	DeviceIDs             []int64
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type IPAddressPage struct {
	Count   uint64
	Results []*domainipam.IPAddress
}

type IPAddressRepository interface {
	List(context.Context, IPAddressListCriteria) (IPAddressPage, error)
	Get(context.Context, shared.ID) (*domainipam.IPAddress, error)
	GetForUpdate(context.Context, shared.ID) (*domainipam.IPAddress, error)
	Create(context.Context, *domainipam.IPAddress) error
	Update(context.Context, *domainipam.IPAddress) error
	Delete(context.Context, *domainipam.IPAddress) error
	LockUniqueness(
		context.Context,
		domainipam.NullableVRFReference,
		domainipam.HostAddress,
	) error
	FindDuplicates(
		context.Context,
		domainipam.NullableVRFReference,
		domainipam.HostAddress,
		shared.ID,
	) ([]*domainipam.IPAddress, error)
	ListAssignedToInterfaceForUpdate(
		context.Context,
		shared.ID,
	) ([]*domainipam.IPAddress, error)
}

type IPAddressVRFReader interface {
	Get(context.Context, shared.ID) (*domainipam.VRF, error)
}

// IPAddressInterfaceReader intentionally returns the typed DCIM aggregate.
// The concrete DCIM repository satisfies this port without an application
// package dependency in either direction.
type IPAddressInterfaceReader interface {
	Get(context.Context, shared.ID) (*domaindcim.Interface, error)
}
