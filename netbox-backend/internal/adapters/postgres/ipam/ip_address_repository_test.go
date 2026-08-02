package ipam

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestIPAddressRepositoryPreservesHostsAndAppliesNetworkFilters(t *testing.T) {
	db := newVRFTestDatabase(t)
	rows := []ipamrow.IPAddressRow{
		{RowMetadata: prefixTestMetadata(), Address: "192.0.2.17/24", Status: "active", DNSName: "edge.example"},
		{RowMetadata: prefixTestMetadata(), Address: "192.0.2.17/32", Status: "reserved"},
		{RowMetadata: prefixTestMetadata(), Address: "192.0.3.1/24", Status: "active"},
		{RowMetadata: prefixTestMetadata(), Address: "2001:db8::1/64", Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)
	repository := NewIPAddressRepository(db)

	host := requiredIPAddressRepositoryFilter(t, "192.0.2.17")
	page, err := repository.List(t.Context(), applicationipam.IPAddressListCriteria{
		Limit: 50, AddressesPresent: true,
		Addresses: []applicationipam.IPAddressFilter{
			{Address: host, Valid: true},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	assert.Equal(t, "192.0.2.17/24", page.Results[0].Display())
	assert.Equal(t, "192.0.2.17/32", page.Results[1].Display())

	exact := requiredIPAddressRepositoryFilter(t, "192.0.2.17/32")
	page, err = repository.List(t.Context(), applicationipam.IPAddressListCriteria{
		Limit: 50, AddressesPresent: true,
		Addresses: []applicationipam.IPAddressFilter{
			{Address: exact, ExplicitMask: true, Valid: true},
		},
	})
	require.NoError(t, err)
	require.Len(t, page.Results, 1)
	assert.Equal(t, "192.0.2.17/32", page.Results[0].Display())

	parent := requiredIPAddressParentFilter(t, "192.0.2.99/24")
	page, err = repository.List(t.Context(), applicationipam.IPAddressListCriteria{
		Limit: 50, Parent: parent,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)

	family := int64(6)
	page, err = repository.List(t.Context(), applicationipam.IPAddressListCriteria{
		Limit: 50, Family: &family, Query: "2001:DB8",
	})
	require.NoError(t, err)
	require.Len(t, page.Results, 1)
	assert.Equal(t, uint32(6), page.Results[0].Family())
}

func TestIPAddressRepositoryRestoresTypedAssignmentAndCascadeOrder(t *testing.T) {
	db := newVRFTestDatabase(t)
	networkInterface, device := seedIPAddressInterfaceGraph(t, db)
	objectType := domainipam.IPAddressAssignmentType
	interfaceID := networkInterface.ID
	rows := []ipamrow.IPAddressRow{
		{
			RowMetadata: prefixTestMetadata(), Address: "192.0.2.2/24",
			Status: "active", AssignedObjectType: &objectType,
			AssignedObjectID: &interfaceID,
		},
		{
			RowMetadata: prefixTestMetadata(), Address: "192.0.2.1/24",
			Status: "active", AssignedObjectType: &objectType,
			AssignedObjectID: &interfaceID,
		},
		{RowMetadata: prefixTestMetadata(), Address: "192.0.2.3/24", Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)
	repository := NewIPAddressRepository(db)

	loaded, err := repository.Get(t.Context(), shared.ID(rows[0].ID))
	require.NoError(t, err)
	assignment, present := loaded.Assignment().Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(networkInterface.ID), assignment.ID())
	assert.Equal(t, "Ethernet1 (uplink)", assignment.Display())
	assert.Equal(t, "edge-01 (EDGE-01)", assignment.Device().Display())

	assigned := true
	page, err := repository.List(t.Context(), applicationipam.IPAddressListCriteria{
		Limit: 50, Assigned: &assigned,
		InterfaceIDs: []int64{-1, networkInterface.ID},
		DeviceIDs:    []int64{-1, device.ID},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)

	cascade, err := repository.ListAssignedToInterfaceForUpdate(
		t.Context(), shared.ID(networkInterface.ID),
	)
	require.NoError(t, err)
	require.Len(t, cascade, 2)
	assert.Less(t, cascade[0].ID(), cascade[1].ID())
}

func TestIPAddressRepositoryFindDuplicatesIgnoresPrefixLength(t *testing.T) {
	db := newVRFTestDatabase(t)
	repository := NewIPAddressRepository(db)
	first := newRepositoryIPAddressFixture(t, "198.51.100.9/24", "")
	require.NoError(t, repository.Create(t.Context(), first))

	otherMask := requiredIPAddressRepositoryFilter(t, "198.51.100.9/32")
	duplicates, err := repository.FindDuplicates(
		t.Context(), domainipam.NullVRFReference(), otherMask, 0,
	)
	require.NoError(t, err)
	require.Len(t, duplicates, 1)
	assert.Equal(t, first.ID(), duplicates[0].ID())

	duplicates, err = repository.FindDuplicates(
		t.Context(), domainipam.NullVRFReference(), otherMask, first.ID(),
	)
	require.NoError(t, err)
	assert.Empty(t, duplicates)
}

func requiredIPAddressRepositoryFilter(
	t *testing.T,
	value string,
) domainipam.HostAddress {
	t.Helper()
	if address, err := domainipam.ParseHostAddress(value); err == nil {
		return address
	}
	host, err := netip.ParseAddr(value)
	require.NoError(t, err)
	bits := 128
	if host.Is4() {
		bits = 32
	}
	address, err := domainipam.ParseHostAddress(
		netip.PrefixFrom(host, bits).String(),
	)
	require.NoError(t, err)
	return address
}

func requiredIPAddressParentFilter(
	t *testing.T,
	value string,
) *applicationipam.IPAddressParentFilter {
	t.Helper()
	network, err := netip.ParsePrefix(value)
	require.NoError(t, err)
	return &applicationipam.IPAddressParentFilter{
		Network: network.Masked(), Valid: true,
	}
}

func newRepositoryIPAddressFixture(
	t *testing.T,
	value string,
	role domainipam.IPAddressRole,
) *domainipam.IPAddress {
	t.Helper()
	nullableRole := domainipam.NullIPAddressRole()
	if role != "" {
		nullableRole = domainipam.NonNullIPAddressRole(role)
	}
	address, err := domainipam.NewIPAddress(domainipam.IPAddressValues{
		Address: value, VRF: domainipam.NullVRFReference(),
		Status: domainipam.IPAddressStatusActive.String(), Role: nullableRole,
	}, vrfRepositoryTime)
	require.NoError(t, err)
	return address
}

func seedIPAddressInterfaceGraph(
	t *testing.T,
	db *gorm.DB,
) (dcimrow.InterfaceRow, dcimrow.DeviceRow) {
	t.Helper()
	metadata := prefixTestMetadata()
	manufacturer := dcimrow.ManufacturerRow{
		RowMetadata: metadata, Name: "Vendor", Slug: "vendor",
	}
	require.NoError(t, db.Create(&manufacturer).Error)
	deviceType := dcimrow.DeviceTypeRow{
		RowMetadata: metadata, ManufacturerID: manufacturer.ID,
		Model: "Router", Slug: "router", UHeight: 1, IsFullDepth: true,
	}
	require.NoError(t, db.Create(&deviceType).Error)
	site := dcimrow.SiteRow{
		RowMetadata: metadata, Name: "Site", Slug: "site", Status: "active",
	}
	require.NoError(t, db.Create(&site).Error)
	role := dcimrow.DeviceRoleRow{
		RowMetadata: metadata, Name: "Router", Slug: "router",
		Color: "112233",
	}
	require.NoError(t, db.Create(&role).Error)
	name, assetTag := "edge-01", "EDGE-01"
	device := dcimrow.DeviceRow{
		RowMetadata: metadata, DeviceTypeID: deviceType.ID, RoleID: role.ID,
		Name: &name, SiteID: site.ID, Status: "active", AssetTag: &assetTag,
	}
	require.NoError(t, db.Create(&device).Error)
	networkInterface := dcimrow.InterfaceRow{
		RowMetadata: metadata, DeviceID: device.ID, Name: "Ethernet1",
		Label: "uplink", Type: "1000base-t", Enabled: true,
	}
	require.NoError(t, db.Create(&networkInterface).Error)
	return networkInterface, device
}
