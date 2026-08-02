package ipam

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var vrfRepositoryTime = shared.NewTimestamp(time.Date(2026, time.July, 22, 10, 0, 0, 0, time.UTC))

func TestVRFRepositoryMapsExistingRowsAndTypedCounts(t *testing.T) {
	db := newVRFTestDatabase(t)
	repository := NewVRFRepository(db)
	vrf := newVRFFixture(t, "Tenant Blue", "65000:100")
	require.NoError(t, repository.Create(t.Context(), vrf))

	require.NoError(t, db.Create(&ipamrow.PrefixRow{
		RowMetadata: ipamrow.RowMetadata{Created: vrfRepositoryTime.Time, LastUpdated: vrfRepositoryTime.Time},
		Prefix:      "10.0.0.0/24",
		VRFID:       int64Pointer(vrf.ID().Int64()),
		Status:      "active",
		Description: "",
		Comments:    "",
	}).Error)
	require.NoError(t, db.Create(&ipamrow.IPAddressRow{
		RowMetadata: ipamrow.RowMetadata{Created: vrfRepositoryTime.Time, LastUpdated: vrfRepositoryTime.Time},
		Address:     "10.0.0.1/24",
		VRFID:       int64Pointer(vrf.ID().Int64()),
		Status:      "active",
		Description: "",
		Comments:    "",
	}).Error)

	loaded, err := repository.Get(t.Context(), vrf.ID())
	require.NoError(t, err, "cause: %v", errors.Unwrap(err))
	assert.Equal(t, "Tenant Blue", loaded.Name())
	assert.Equal(t, "65000:100", requiredRepositoryRD(t, loaded.RD()))
	assert.Equal(t, "Tenant Blue (65000:100)", loaded.Display())
	assert.Equal(t, uint64(1), loaded.PrefixCount())
	assert.Equal(t, uint64(1), loaded.IPAddressCount())

	_, err = repository.Get(t.Context(), shared.ID(9999))
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonNotFound))
}

func TestVRFRepositoryAppliesRepeatedFiltersSignedIDsAndDefaultOrdering(t *testing.T) {
	db := newVRFTestDatabase(t)
	repository := NewVRFRepository(db)
	rows := []ipamrow.VRFRow{
		{RowMetadata: metadata(), Name: "Zulu", RD: stringPointer("9:9"), EnforceUnique: true},
		{RowMetadata: metadata(), Name: "Alpha", RD: stringPointer("2:2"), EnforceUnique: true},
		{RowMetadata: metadata(), Name: "Alpha", RD: stringPointer("1:1"), EnforceUnique: true},
		{RowMetadata: metadata(), Name: "Beta", RD: stringPointer("3:3"), EnforceUnique: false},
	}
	require.NoError(t, db.Create(&rows).Error)

	rdOne, err := domainipam.ParseRouteDistinguisher("1:1")
	require.NoError(t, err)
	rdTwo, err := domainipam.ParseRouteDistinguisher("2:2")
	require.NoError(t, err)
	page, err := repository.List(t.Context(), applicationipam.VRFListCriteria{
		Limit: 50,
		IDs:   []int64{-1, rows[1].ID, rows[2].ID},
		Names: []string{"Alpha", "Does not exist"},
		RDs:   []domainipam.RouteDistinguisher{rdOne, rdTwo},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, "1:1", requiredRepositoryRD(t, page.Results[0].RD()))
	assert.Equal(t, "2:2", requiredRepositoryRD(t, page.Results[1].RD()))

	empty, err := repository.List(t.Context(), applicationipam.VRFListCriteria{
		Limit: 50,
		IDs:   []int64{-7, 0},
	})
	require.NoError(t, err)
	assert.Zero(t, empty.Count)
	assert.Empty(t, empty.Results)
}

func TestVRFRepositoryCountsBeforePaginationAndEscapesSearchWildcards(t *testing.T) {
	db := newVRFTestDatabase(t)
	rows := []ipamrow.VRFRow{
		{RowMetadata: metadata(), Name: "Percent % One", RD: stringPointer("10:1"), EnforceUnique: true},
		{RowMetadata: metadata(), Name: "Percent % Two", RD: stringPointer("10:2"), EnforceUnique: true},
		{RowMetadata: metadata(), Name: "Ordinary", RD: stringPointer("10:3"), EnforceUnique: true},
	}
	require.NoError(t, db.Create(&rows).Error)

	page, err := NewVRFRepository(db).List(t.Context(), applicationipam.VRFListCriteria{
		Limit:  1,
		Offset: 1,
		Query:  "%",
		Ordering: []applicationipam.VRFSort{
			{Field: applicationipam.VRFSortName},
			{Field: applicationipam.VRFSortRD},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, "Percent % Two", page.Results[0].Name())
}

func TestVRFRepositoryTranslatesUniqueAndProtectedConstraints(t *testing.T) {
	db := newVRFTestDatabase(t)
	repository := NewVRFRepository(db)
	vrf := newVRFFixture(t, "Original", "65000:99")
	require.NoError(t, repository.Create(t.Context(), vrf))

	duplicate := newVRFFixture(t, "Duplicate", "65000:99")
	err := repository.Create(t.Context(), duplicate)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Empty(t, shared.ViolationsOf(err), "translated errors must not guess a lost constraint")
	assert.False(t, duplicate.ID().IsValid())

	require.NoError(t, db.Create(&ipamrow.PrefixRow{
		RowMetadata: metadata(), Prefix: "192.0.2.0/24", VRFID: int64Pointer(vrf.ID().Int64()),
		Status: "active", Description: "", Comments: "",
	}).Error)
	err = repository.Delete(t.Context(), vrf)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonProtected))
}

func TestTranslateVRFMutationErrorPreservesRouteDistinguisherViolation(t *testing.T) {
	err := translateVRFMutationError(
		"create VRF",
		errors.New(`duplicate key value violates unique constraint "uq_go_vrf_rd"`),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, []shared.FieldViolation{{
		Field: "rd", Reason: "unique",
		Description: "VRF with this route distinguisher already exists.",
	}}, shared.ViolationsOf(err))
}

func newVRFFixture(t *testing.T, name, rdValue string) *domainipam.VRF {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	vrf, err := domainipam.NewVRF(domainipam.VRFValues{
		Name: name, RD: domainipam.NonNullRouteDistinguisher(rd), EnforceUnique: true,
		Description: "VRF description", Comments: "VRF comments",
	}, vrfRepositoryTime)
	require.NoError(t, err)
	return vrf
}

func requiredRepositoryRD(t *testing.T, nullable domainipam.NullableRouteDistinguisher) string {
	t.Helper()
	rd, present := nullable.Get()
	require.True(t, present)
	return rd.String()
}

func metadata() ipamrow.RowMetadata {
	return ipamrow.RowMetadata{Created: vrfRepositoryTime.Time, LastUpdated: vrfRepositoryTime.Time}
}

func stringPointer(value string) *string { return &value }

func int64Pointer(value int64) *int64 { return &value }

func newVRFTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", name)),
		&gorm.Config{TranslateError: true},
	)
	require.NoError(t, err)
	models := append(identityrows.Models(), dcimrow.Models()...)
	models = append(models, ipamrow.Models()...)
	models = append(models, &postgreschangelog.ChangeRow{})
	require.NoError(t, db.AutoMigrate(models...))
	return db
}
