package ipam_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	postgresipam "netbox-go/internal/adapters/postgres/ipam"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgrestransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/application/changelog"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestPrefixServiceEnforcesGlobalAndPerVRFUniqueness(t *testing.T) {
	db, service, _, vrfs := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	principal := testPrincipal()

	global, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("192.0.2.0/24", nil))
	require.NoError(t, err)
	assert.Equal(t, domainipam.PrefixStatusActive, global.Status())
	assert.False(t, global.IsPool())
	assert.False(t, global.MarkUtilized())
	assert.True(t, global.VRF().IsNull())

	_, err = service.CreatePrefix(t.Context(), principal, createPrefixCommand("192.0.2.0/24", nil))
	require.Error(t, err)
	assert.Equal(t, "prefix", shared.ViolationsOf(err)[0].Field)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "global table")

	nonUnique := createApplicationVRF(t, vrfs, "Non-unique", "65000:1", false)
	for range 2 {
		_, err = service.CreatePrefix(
			t.Context(), principal, createPrefixCommand("198.51.100.0/24", idPointer(nonUnique.ID())),
		)
		require.NoError(t, err)
	}

	unique := createApplicationVRF(t, vrfs, "Unique", "65000:2", true)
	_, err = service.CreatePrefix(
		t.Context(), principal, createPrefixCommand("203.0.113.0/24", idPointer(unique.ID())),
	)
	require.NoError(t, err)
	_, err = service.CreatePrefix(
		t.Context(), principal, createPrefixCommand("203.0.113.0/24", idPointer(unique.ID())),
	)
	require.Error(t, err)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "VRF Unique (65000:2)")

	var prefixCount, changeCount int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&prefixCount).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changeCount).Error)
	assert.Equal(t, int64(4), prefixCount)
	assert.Equal(t, int64(4), changeCount)
}

func TestPrefixServiceProjectsHierarchyAndExactListSemantics(t *testing.T) {
	_, service, _, _ := newPrefixApplicationService(t, nil, &authz.AllowAll{})
	principal := testPrincipal()
	outer, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("10.0.0.0/8", nil))
	require.NoError(t, err)
	inner, err := service.CreatePrefix(t.Context(), principal, createPrefixCommand("10.1.0.0/16", nil))
	require.NoError(t, err)
	_, err = service.CreatePrefix(t.Context(), principal, createPrefixCommand("2001:db8::/32", nil))
	require.NoError(t, err)

	outer, err = service.GetPrefix(t.Context(), principal, applicationipam.GetPrefixQuery{ID: outer.ID()})
	require.NoError(t, err)
	inner, err = service.GetPrefix(t.Context(), principal, applicationipam.GetPrefixQuery{ID: inner.ID()})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), outer.Children())
	assert.Equal(t, uint32(0), outer.Depth())
	assert.Equal(t, uint64(0), inner.Children())
	assert.Equal(t, uint32(1), inner.Depth())

	zero := uint32(0)
	within := "10.0.0.0/8"
	page, err := service.ListPrefixes(t.Context(), principal, applicationipam.ListPrefixesQuery{
		Limit: zero, LimitPresent: true, IDs: []int64{-1, outer.ID().Int64(), inner.ID().Int64()},
		Prefixes: []string{"invalid", "10.0.0.0/8", "10.1.2.3/16"},
		Family:   int64PointerForPrefix(4), WithinInclude: &within,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, applicationipam.MaximumPrefixPageLimit, applicationipam.ListPrefixesQuery{
		LimitPresent: true,
	}.EffectiveLimit())
	assert.Equal(t, "10.0.0.0/8", page.Results[0].Display())
	assert.Equal(t, "10.1.0.0/16", page.Results[1].Display())
}

func TestPrefixServiceRollsBackMutationAndChangeTogether(t *testing.T) {
	db, service, _, _ := newPrefixApplicationService(
		t, failingPrefixRecorder{}, &authz.AllowAll{},
	)
	_, err := service.CreatePrefix(
		t.Context(), testPrincipal(), createPrefixCommand("192.0.2.0/24", nil),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))

	var count int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&count).Error)
	assert.Zero(t, count)
}

func TestPrefixServiceDeniedCreateStopsBeforeWrite(t *testing.T) {
	db, service, _, _ := newPrefixApplicationService(
		t, nil, &trackingAuthorizer{denyAction: authz.Add},
	)
	_, err := service.CreatePrefix(
		t.Context(), testPrincipal(), createPrefixCommand("192.0.2.0/24", nil),
	)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonForbidden))
	var count int64
	require.NoError(t, db.Model(&ipamrow.PrefixRow{}).Count(&count).Error)
	assert.Zero(t, count)
}

func newPrefixApplicationService(
	t *testing.T,
	recorder changelog.Recorder,
	authorizer authz.ResourceAuthorizer,
) (*gorm.DB, *applicationipam.PrefixService, *postgresipam.PrefixRepository, *postgresipam.VRFRepository) {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", name)),
		&gorm.Config{TranslateError: true, Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	models := append(identityrows.Models(), dcimrow.Models()...)
	models = append(models, ipamrow.Models()...)
	models = append(models, &postgreschangelog.ChangeRow{})
	require.NoError(t, db.AutoMigrate(models...))
	require.NoError(t, db.Create(&identityrows.UserRow{
		ID: 17, Username: "operator", PasswordHash: "test", IsActive: true,
		Permissions: datatypes.JSON([]byte(`[]`)),
		Created:     applicationCreatedAt.Time, Updated: applicationUpdatedAt.Time,
	}).Error)
	if recorder == nil {
		recorder = postgreschangelog.NewRecorder(db)
	}
	prefixes := postgresipam.NewPrefixRepository(db)
	vrfs := postgresipam.NewVRFRepository(db)
	service, err := applicationipam.NewPrefixService(
		prefixes, vrfs, postgrestransaction.NewUnitOfWork(db), recorder,
		authorizer, fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)
	return db, service, prefixes, vrfs
}

func createApplicationVRF(
	t *testing.T,
	repository *postgresipam.VRFRepository,
	name, rdValue string,
	enforceUnique bool,
) *domainipam.VRF {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	vrf, err := domainipam.NewVRF(domainipam.VRFValues{
		Name: name, RD: domainipam.NonNullRouteDistinguisher(rd), EnforceUnique: enforceUnique,
	}, applicationCreatedAt)
	require.NoError(t, err)
	require.NoError(t, repository.Create(t.Context(), vrf))
	return vrf
}

func createPrefixCommand(value string, vrfID *shared.ID) applicationipam.CreatePrefixCommand {
	command := applicationipam.CreatePrefixCommand{Prefix: applicationipam.FieldValue(value)}
	if vrfID != nil {
		command.VRF = applicationipam.FieldValue(vrfID.Int64())
	}
	return command
}

func idPointer(value shared.ID) *shared.ID { return &value }

func int64PointerForPrefix(value int64) *int64 { return &value }

type failingPrefixRecorder struct{}

func (failingPrefixRecorder) Record(context.Context, changelog.Change) error {
	return errors.New("forced Prefix change failure")
}
