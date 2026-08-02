package dcim

import (
	"context"

	applicationipam "netbox-go/internal/application/ipam"
	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

type InterfaceListCriteria struct {
	Limit                 uint32
	Offset                uint32
	Query                 string
	IDs                   []int64
	Ordering              []InterfaceSort
	DeviceIDs             []int64
	DeviceNames           []string
	Names                 []string
	Types                 []dcimdomain.InterfaceType
	Enabled               *bool
	MgmtOnly              *bool
	VisibleObjectIDs      []shared.ID
	VisibilityConstrained bool
	DeferPagination       bool
}

type InterfacePage struct {
	Count   uint64
	Results []*dcimdomain.Interface
}

// InterfaceReader is the narrow typed lookup used by IPAddress assignment.
type InterfaceReader interface {
	Get(context.Context, shared.ID) (*dcimdomain.Interface, error)
}

type InterfaceRepository interface {
	InterfaceReader
	List(context.Context, InterfaceListCriteria) (InterfacePage, error)
	GetForUpdate(context.Context, shared.ID) (*dcimdomain.Interface, error)
	ListForDeviceForUpdate(context.Context, shared.ID) ([]*dcimdomain.Interface, error)
	Create(context.Context, *dcimdomain.Interface) error
	Update(context.Context, *dcimdomain.Interface) error
	Delete(context.Context, *dcimdomain.Interface) error
}

type InterfaceDeviceReader interface {
	GetDeviceReference(context.Context, shared.ID) (dcimdomain.DeviceReference, error)
}

type InterfaceIPAddressCascade interface {
	DeleteAssignedToInterface(
		context.Context,
		shared.ID,
		shared.Timestamp,
	) ([]applicationipam.IPAddressCascadeChange, error)
}

// InterfaceCascadeChange is returned to Device deletion in exact audit order:
// all assigned IPAddresses precede their Interface parent.
type InterfaceCascadeChange struct {
	ObjectType     string
	ID             shared.ID
	Representation string
	Before         shared.ObjectSnapshot
}

// DeviceInterfaceCascade is the narrow typed port consumed by Device deletion.
// The caller owns the transaction and records the returned ordered changes.
type DeviceInterfaceCascade interface {
	DeleteForDevice(
		context.Context,
		shared.ID,
		shared.Timestamp,
	) ([]InterfaceCascadeChange, error)
}
