package ipam

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestPrefixRepositoryRestoresVRFAndHierarchyWithinNamespace(t *testing.T) {
	db := newVRFTestDatabase(t)
	vrf := ipamrow.VRFRow{
		RowMetadata: prefixTestMetadata(), Name: "Tenant", RD: stringPointer("65000:10"), EnforceUnique: true,
	}
	require.NoError(t, db.Create(&vrf).Error)
	rows := []ipamrow.PrefixRow{
		{RowMetadata: prefixTestMetadata(), Prefix: "10.0.0.0/8", Status: "container"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.1.0.0/16", Status: "active"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.1.2.0/24", Status: "active"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.0.0.0/8", VRFID: int64Pointer(vrf.ID), Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)
	repository := NewPrefixRepository(db)

	outer, err := repository.Get(t.Context(), shared.ID(rows[0].ID))
	require.NoError(t, err)
	assert.Equal(t, uint64(2), outer.Children())
	assert.Zero(t, outer.Depth())
	inner, err := repository.Get(t.Context(), shared.ID(rows[2].ID))
	require.NoError(t, err)
	assert.Zero(t, inner.Children())
	assert.Equal(t, uint32(2), inner.Depth())
	tenant, err := repository.Get(t.Context(), shared.ID(rows[3].ID))
	require.NoError(t, err)
	assert.Zero(t, tenant.Children(), "global Prefixes must not enter a VRF hierarchy")
	reference, present := tenant.VRF().Get()
	require.True(t, present)
	assert.Equal(t, "Tenant (65000:10)", reference.Display())
	assert.True(t, reference.EnforceUnique())
}

func TestPrefixRepositoryAppliesContainmentSearchRepeatedAndSignedFilters(t *testing.T) {
	db := newVRFTestDatabase(t)
	rows := []ipamrow.PrefixRow{
		{RowMetadata: prefixTestMetadata(), Prefix: "10.0.0.0/8", Status: "container", Description: "parent network"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.1.0.0/16", Status: "active"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.1.2.3/32", Status: "reserved"},
		{RowMetadata: prefixTestMetadata(), Prefix: "2001:db8::/32", Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)
	repository := NewPrefixRepository(db)

	withinNetwork := requiredRepositoryPrefixFilter(t, "10.0.0.0/8")
	within, err := repository.List(t.Context(), applicationipam.PrefixListCriteria{
		Limit: 50,
		IDs:   []int64{-1, rows[0].ID, rows[1].ID, rows[2].ID},
		Statuses: []domainipam.PrefixStatus{
			domainipam.PrefixStatusActive, domainipam.PrefixStatusReserved,
		},
		Within: &applicationipam.PrefixNetworkFilter{Network: withinNetwork, ExplicitMask: true, Valid: true},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), within.Count)
	assert.Equal(t, "10.1.0.0/16", within.Results[0].Display())
	assert.Equal(t, "10.1.2.3/32", within.Results[1].Display())

	host := requiredRepositoryPrefixFilter(t, "10.1.2.3")
	contains, err := repository.List(t.Context(), applicationipam.PrefixListCriteria{
		Limit: 50,
		Contains: &applicationipam.PrefixNetworkFilter{
			Network: host, ExplicitMask: false, Valid: true,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), contains.Count, "bare hosts use strict containment")

	exactHost, err := repository.List(t.Context(), applicationipam.PrefixListCriteria{
		Limit: 50,
		Contains: &applicationipam.PrefixNetworkFilter{
			Network: host, ExplicitMask: true, Valid: true,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(3), exactHost.Count, "explicit host prefixes include equality")

	search, err := repository.List(t.Context(), applicationipam.PrefixListCriteria{
		Limit: 50, Query: "10.1.2.3",
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(3), search.Count, "network search returns containing Prefixes")
}

func TestPrefixRepositoryDefaultOrderIsGlobalThenVRFPrefixAndID(t *testing.T) {
	db := newVRFTestDatabase(t)
	vrf := ipamrow.VRFRow{
		RowMetadata: prefixTestMetadata(), Name: "Tenant", RD: stringPointer("65000:20"), EnforceUnique: true,
	}
	require.NoError(t, db.Create(&vrf).Error)
	rows := []ipamrow.PrefixRow{
		{RowMetadata: prefixTestMetadata(), Prefix: "10.2.0.0/16", VRFID: int64Pointer(vrf.ID), Status: "active"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.1.0.0/16", Status: "active"},
		{RowMetadata: prefixTestMetadata(), Prefix: "10.0.0.0/8", Status: "active"},
	}
	require.NoError(t, db.Create(&rows).Error)
	page, err := NewPrefixRepository(db).List(t.Context(), applicationipam.PrefixListCriteria{Limit: 50})
	require.NoError(t, err)
	require.Len(t, page.Results, 3)
	assert.Equal(t, "10.0.0.0/8", page.Results[0].Display())
	assert.Equal(t, "10.1.0.0/16", page.Results[1].Display())
	assert.Equal(t, "10.2.0.0/16", page.Results[2].Display())
}

func TestPrefixRepositoryFindDuplicateHonorsNullableVRFScope(t *testing.T) {
	db := newVRFTestDatabase(t)
	repository := NewPrefixRepository(db)
	prefix := newRepositoryPrefixFixture(t, "192.0.2.0/24", domainipam.NullVRFReference())
	require.NoError(t, repository.Create(t.Context(), prefix))

	duplicate, err := repository.FindDuplicate(
		t.Context(), domainipam.NullVRFReference(), prefix.Network(), 0,
	)
	require.NoError(t, err)
	require.NotNil(t, duplicate)
	assert.Equal(t, prefix.ID(), duplicate.ID())

	none, err := repository.FindDuplicate(
		t.Context(), domainipam.NullVRFReference(), prefix.Network(), prefix.ID(),
	)
	require.NoError(t, err)
	assert.Nil(t, none)
}

func newRepositoryPrefixFixture(
	t *testing.T,
	value string,
	vrf domainipam.NullableVRFReference,
) *domainipam.Prefix {
	t.Helper()
	prefix, err := domainipam.NewPrefix(domainipam.PrefixValues{
		Prefix: value, VRF: vrf, Status: domainipam.PrefixStatusActive.String(),
	}, vrfRepositoryTime)
	require.NoError(t, err)
	return prefix
}

func requiredRepositoryPrefixFilter(t *testing.T, value string) domainipam.PrefixNetwork {
	t.Helper()
	network, _, err := domainipam.ParsePrefixFilter(value)
	require.NoError(t, err)
	return network
}

func prefixTestMetadata() ipamrow.RowMetadata {
	now := time.Date(2026, time.July, 22, 15, 0, 0, 0, time.UTC)
	return ipamrow.RowMetadata{Created: now, LastUpdated: now}
}
