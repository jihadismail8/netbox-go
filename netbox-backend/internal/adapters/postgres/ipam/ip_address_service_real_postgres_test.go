package ipam

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	postgresdcim "netbox-go/internal/adapters/postgres/dcim"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrow "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestIPAddressServiceConcurrentConditionalUniquenessPostgres(t *testing.T) {
	db, service, principal, _ := newIPAddressConcurrencyPostgres(t)
	now := time.Now().UTC()
	vrf := ipamrow.VRFRow{
		RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Unique VRF", EnforceUnique: true,
	}
	require.NoError(t, db.Create(&vrf).Error)

	type result struct {
		address *domainipam.IPAddress
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for index := range 2 {
		index := index
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			address, createErr := service.CreateIPAddress(
				t.Context(), principal, applicationipam.CreateIPAddressCommand{
					Address: applicationipam.FieldValue("192.0.2.10/24"),
					VRF:     applicationipam.FieldValue(vrf.ID),
					Description: applicationipam.FieldValue(
						fmt.Sprintf("writer-%d", index),
					),
				},
			)
			results <- result{address: address, err: createErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	successes, uniquenessRejections := 0, 0
	for createResult := range results {
		if createResult.err == nil {
			require.NotNil(t, createResult.address)
			successes++
			continue
		}
		require.True(t, shared.HasReason(
			createResult.err, shared.ErrorReasonValidation,
		), createResult.err)
		violations := shared.ViolationsOf(createResult.err)
		require.Len(t, violations, 1)
		require.Equal(t, "address", violations[0].Field)
		require.Equal(t, "unique", violations[0].Reason)
		uniquenessRejections++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, uniquenessRejections)

	var addresses int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).Count(&addresses).Error)
	require.Equal(t, int64(1), addresses)
	var changes int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).
		Where("kind = ? AND action = ?", domainipam.IPAddressObjectType, "create").
		Count(&changes).Error)
	require.Equal(t, int64(1), changes, "the rejected duplicate must not leave a change row")
}

func TestIPAddressServiceConcurrentAssignmentAuditChainPostgres(t *testing.T) {
	db, service, principal, interfaceIDs := newIPAddressConcurrencyPostgres(t)
	address, err := service.CreateIPAddress(
		t.Context(), principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("198.51.100.20/24"),
		},
	)
	require.NoError(t, err)

	type result struct {
		address *domainipam.IPAddress
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, len(interfaceIDs))
	var workers sync.WaitGroup
	for _, interfaceID := range interfaceIDs {
		interfaceID := interfaceID
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			assigned, assignErr := service.AssignIPAddress(
				t.Context(), principal, applicationipam.AssignIPAddressCommand{
					ID: address.ID(), InterfaceID: interfaceID,
				},
			)
			results <- result{address: assigned, err: assignErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	for assignResult := range results {
		require.NoError(t, assignResult.err)
		require.NotNil(t, assignResult.address)
		assignment, present := assignResult.address.Assignment().Get()
		require.True(t, present)
		require.Contains(t, interfaceIDs, assignment.ID())
	}

	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Where(
		"kind = ? AND object_id = ? AND action = ?",
		domainipam.IPAddressObjectType, address.ID().Int64(), "update",
	).Order("id").Find(&changes).Error)
	require.Len(t, changes, 2)
	firstBefore := decodeIPAddressChangeData(t, changes[0].BeforeData)
	firstAfter := decodeIPAddressChangeData(t, changes[0].AfterData)
	secondBefore := decodeIPAddressChangeData(t, changes[1].BeforeData)
	secondAfter := decodeIPAddressChangeData(t, changes[1].AfterData)
	require.Nil(t, firstBefore["assigned_object_id"])
	require.Equal(t, firstAfter["assigned_object_id"], secondBefore["assigned_object_id"],
		"the second transaction must observe and audit the first committed assignment")

	auditedTargets := []int64{
		changeJSONInt64(t, firstAfter["assigned_object_id"]),
		changeJSONInt64(t, secondAfter["assigned_object_id"]),
	}
	sort.Slice(auditedTargets, func(i, j int) bool { return auditedTargets[i] < auditedTargets[j] })
	expectedTargets := []int64{interfaceIDs[0].Int64(), interfaceIDs[1].Int64()}
	sort.Slice(expectedTargets, func(i, j int) bool { return expectedTargets[i] < expectedTargets[j] })
	require.Equal(t, expectedTargets, auditedTargets)

	loaded, err := service.GetIPAddress(
		t.Context(), principal, applicationipam.GetIPAddressQuery{ID: address.ID()},
	)
	require.NoError(t, err)
	assignment, present := loaded.Assignment().Get()
	require.True(t, present)
	require.Equal(t, changeJSONInt64(t, secondAfter["assigned_object_id"]), assignment.ID().Int64())
	var persisted ipamrow.IPAddressRow
	require.NoError(t, db.First(&persisted, address.ID().Int64()).Error)
	require.NotNil(t, persisted.AssignedObjectID)
	require.Equal(t, assignment.ID().Int64(), *persisted.AssignedObjectID)
}

func TestPostgresIPAddressScalarPresenceDurability(t *testing.T) {
	db, service, principal, interfaceIDs := newIPAddressScalarPostgres(t)
	ctx := t.Context()
	now := time.Now().UTC()
	vrf := ipamrow.VRFRow{
		RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now},
		Name:        "Scalar presence VRF", RD: stringPointer("64512:44"),
	}
	require.NoError(t, db.Create(&vrf).Error)

	omittedRole, err := service.CreateIPAddress(
		ctx, principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.51"),
		},
	)
	require.NoError(t, err)
	requirePostgresIPAddressRole(t, db, omittedRole.ID(), nil)

	nullRole, err := service.CreateIPAddress(
		ctx, principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.52"),
			Role:    applicationipam.NullField[string](),
		},
	)
	require.NoError(t, err)
	requirePostgresIPAddressRole(t, db, nullRole.ID(), stringPointer(""))

	blankRole, err := service.CreateIPAddress(
		ctx, principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.53"),
			Role:    applicationipam.FieldValue(""),
		},
	)
	require.NoError(t, err)
	requirePostgresIPAddressRole(t, db, blankRole.ID(), stringPointer(""))

	netmaskAddress, err := service.CreateIPAddress(
		ctx, principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.55/255.255.255.0"),
		},
	)
	require.NoError(t, err)
	require.Equal(t, "192.0.2.55/24", netmaskAddress.Address().String())
	require.Equal(
		t, "192.0.2.55/24",
		requirePostgresIPAddressRow(t, db, netmaskAddress.ID()).Address,
	)

	preserved, err := service.CreateIPAddress(
		ctx, principal, applicationipam.CreateIPAddressCommand{
			Address:     applicationipam.FieldValue("192.0.2.54/24"),
			VRF:         applicationipam.FieldValue(vrf.ID),
			Status:      applicationipam.FieldValue("reserved"),
			Role:        applicationipam.FieldValue("loopback"),
			DNSName:     applicationipam.FieldValue("Edge.EXAMPLE.Test"),
			Description: applicationipam.FieldValue(" durable description "),
			Comments:    applicationipam.FieldValue(" durable comments "),
			AssignedObjectType: applicationipam.FieldValue(
				domainipam.IPAddressAssignmentType,
			),
			AssignedObjectID: applicationipam.FieldValue(interfaceIDs[0].Int64()),
		},
	)
	require.NoError(t, err)

	replaced, err := service.ReplaceIPAddress(
		ctx, principal, applicationipam.ReplaceIPAddressCommand{
			ID: preserved.ID(),
			CreateIPAddressCommand: applicationipam.CreateIPAddressCommand{
				Address: applicationipam.FieldValue("2001:db8::54"),
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "2001:db8::54/128", replaced.Address().String())
	require.Equal(t, domainipam.IPAddressStatusReserved, replaced.Status())
	role, rolePresent := replaced.Role().Get()
	require.True(t, rolePresent)
	require.Equal(t, domainipam.IPAddressRoleLoopback, role)
	require.Equal(t, "edge.example.test", replaced.DNSName())
	require.Equal(t, "durable description", replaced.Description())
	require.Equal(t, "durable comments", replaced.Comments())
	loadedVRF, vrfPresent := replaced.VRF().Get()
	require.True(t, vrfPresent)
	require.Equal(t, shared.ID(vrf.ID), loadedVRF.ID())
	assignment, assigned := replaced.Assignment().Get()
	require.True(t, assigned)
	require.Equal(t, interfaceIDs[0], assignment.ID())
	persisted := requirePostgresIPAddressRow(t, db, preserved.ID())
	require.Equal(t, "2001:db8::54/128", persisted.Address)
	require.NotNil(t, persisted.VRFID)
	require.Equal(t, vrf.ID, *persisted.VRFID)
	require.Equal(t, "reserved", persisted.Status)
	require.NotNil(t, persisted.Role)
	require.Equal(t, "loopback", *persisted.Role)
	require.Equal(t, "edge.example.test", persisted.DNSName)
	require.Equal(t, "durable description", persisted.Description)
	require.Equal(t, "durable comments", persisted.Comments)
	require.NotNil(t, persisted.AssignedObjectType)
	require.Equal(t, domainipam.IPAddressAssignmentType, *persisted.AssignedObjectType)
	require.NotNil(t, persisted.AssignedObjectID)
	require.Equal(t, interfaceIDs[0].Int64(), *persisted.AssignedObjectID)

	cleared, err := service.UpdateIPAddress(
		ctx, principal, applicationipam.UpdateIPAddressCommand{
			ID: preserved.ID(), Role: applicationipam.NullField[string](),
			DNSName:     applicationipam.FieldValue(""),
			Description: applicationipam.FieldValue(""),
			Comments:    applicationipam.FieldValue(""),
		},
	)
	require.NoError(t, err)
	clearedRole, rolePresent := cleared.Role().Get()
	require.True(t, rolePresent)
	require.Empty(t, clearedRole.String())
	require.Empty(t, cleared.DNSName())
	require.Empty(t, cleared.Description())
	require.Empty(t, cleared.Comments())
	assignment, assigned = cleared.Assignment().Get()
	require.True(t, assigned)
	require.Equal(t, interfaceIDs[0], assignment.ID())
	persisted = requirePostgresIPAddressRow(t, db, preserved.ID())
	require.NotNil(t, persisted.Role)
	require.Empty(t, *persisted.Role)
	require.Empty(t, persisted.DNSName)
	require.Empty(t, persisted.Description)
	require.Empty(t, persisted.Comments)
	require.NotNil(t, persisted.VRFID)
	require.Equal(t, vrf.ID, *persisted.VRFID)
	require.NotNil(t, persisted.AssignedObjectID)
	require.Equal(t, interfaceIDs[0].Int64(), *persisted.AssignedObjectID)

	beforeInvalid := requirePostgresIPAddressRow(t, db, preserved.ID())
	beforeChanges := countIPAddressChanges(t, db, preserved.ID())
	_, err = service.UpdateIPAddress(
		ctx, principal, applicationipam.UpdateIPAddressCommand{
			ID: preserved.ID(), Status: applicationipam.NullField[string](),
		},
	)
	requireIPAddressScalarViolation(t, err, "status", "blank")
	afterInvalid := requirePostgresIPAddressRow(t, db, preserved.ID())
	require.Equal(t, beforeInvalid, afterInvalid)
	require.Equal(t, beforeChanges, countIPAddressChanges(t, db, preserved.ID()))

	for _, test := range []struct {
		name    string
		command applicationipam.UpdateIPAddressCommand
		field   string
		reason  string
	}{
		{
			name: "address whitespace",
			command: applicationipam.UpdateIPAddressCommand{
				ID: preserved.ID(), Address: applicationipam.FieldValue(" 2001:db8::54"),
			},
			field: "address", reason: "invalid",
		},
		{
			name: "status boolean-like choice",
			command: applicationipam.UpdateIPAddressCommand{
				ID: preserved.ID(), Status: applicationipam.FieldValue("true"),
			},
			field: "status", reason: "invalid_choice",
		},
		{
			name: "role integer-like choice",
			command: applicationipam.UpdateIPAddressCommand{
				ID: preserved.ID(), Role: applicationipam.FieldValue("001"),
			},
			field: "role", reason: "invalid_choice",
		},
		{
			name: "description null character",
			command: applicationipam.UpdateIPAddressCommand{
				ID: preserved.ID(), Description: applicationipam.FieldValue("contains\x00null"),
			},
			field: "description", reason: "invalid",
		},
	} {
		t.Run("rejection preserves durable state/"+test.name, func(t *testing.T) {
			_, err := service.UpdateIPAddress(ctx, principal, test.command)
			requireIPAddressScalarViolation(t, err, test.field, test.reason)
			require.Equal(t, beforeInvalid, requirePostgresIPAddressRow(t, db, preserved.ID()))
			require.Equal(t, beforeChanges, countIPAddressChanges(t, db, preserved.ID()))
		})
	}

	_, err = service.ReplaceIPAddress(
		ctx, principal, applicationipam.ReplaceIPAddressCommand{
			ID: preserved.ID(),
			CreateIPAddressCommand: applicationipam.CreateIPAddressCommand{
				Address: applicationipam.NullField[string](),
			},
		},
	)
	requireIPAddressScalarViolation(t, err, "address", "null")
	require.Equal(t, beforeInvalid, requirePostgresIPAddressRow(t, db, preserved.ID()))
	require.Equal(t, beforeChanges, countIPAddressChanges(t, db, preserved.ID()))
}

func requirePostgresIPAddressRole(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
	expected *string,
) {
	t.Helper()
	row := requirePostgresIPAddressRow(t, db, id)
	if expected == nil {
		require.Nil(t, row.Role)
		return
	}
	require.NotNil(t, row.Role)
	require.Equal(t, *expected, *row.Role)
}

func requirePostgresIPAddressRow(
	t *testing.T,
	db *gorm.DB,
	id shared.ID,
) ipamrow.IPAddressRow {
	t.Helper()
	var row ipamrow.IPAddressRow
	require.NoError(t, db.First(&row, id.Int64()).Error)
	row.VRF = nil
	row.AssignedInterface = nil
	return row
}

func countIPAddressChanges(t *testing.T, db *gorm.DB, id shared.ID) int64 {
	t.Helper()
	var count int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domainipam.IPAddressObjectType, id.Int64(),
	).Count(&count).Error)
	return count
}

func requireIPAddressScalarViolation(
	t *testing.T,
	err error,
	field string,
	reason string,
) {
	t.Helper()
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 1)
	require.Equal(t, field, violations[0].Field)
	require.Equal(t, reason, violations[0].Reason)
}

func newIPAddressConcurrencyPostgres(
	t *testing.T,
) (*gorm.DB, *applicationipam.IPAddressService, identity.Principal, []shared.ID) {
	return newIPAddressPostgresService(t, false)
}

func newIPAddressScalarPostgres(
	t *testing.T,
) (*gorm.DB, *applicationipam.IPAddressService, identity.Principal, []shared.ID) {
	return newIPAddressPostgresService(t, true)
}

func newIPAddressPostgresService(
	t *testing.T,
	advanceClock bool,
) (*gorm.DB, *applicationipam.IPAddressService, identity.Principal, []shared.ID) {
	t.Helper()
	dsn := os.Getenv("NETBOX_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}
	base, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	baseSQL, err := base.DB()
	require.NoError(t, err)
	schema := fmt.Sprintf("ipam_address_concurrency_%d", time.Now().UnixNano())
	require.NoError(t, base.Exec(`CREATE SCHEMA "`+schema+`"`).Error)

	db, err := gorm.Open(
		postgres.Open(ipamPostgresDSNWithSearchPath(t, dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Error("could not close the scoped PostgreSQL test connection")
		}
		if err := base.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Error("could not drop the scoped PostgreSQL test schema")
		}
		var remaining int64
		if err := base.Raw(
			"SELECT count(*) FROM pg_namespace WHERE nspname = ?", schema,
		).Scan(&remaining).Error; err != nil || remaining != 0 {
			t.Error("the scoped PostgreSQL test schema remained after cleanup")
		}
		if err := baseSQL.Close(); err != nil {
			t.Error("could not close the PostgreSQL cleanup connection")
		}
	})

	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	changeDependencies := []string{"go_identity_users"}
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
		changeDependencies = append(changeDependencies, descriptor.Name)
	}
	for _, descriptor := range ipamrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{
			Name: descriptor.Name, Model: descriptor.Model,
			Dependencies: descriptor.Dependencies,
		})
		changeDependencies = append(changeDependencies, descriptor.Name)
	}
	entries = append(entries, bootstrap.Entry{
		Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{},
		Dependencies: changeDependencies,
	})
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	_, err = bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)

	now := time.Now().UTC()
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "postgres-test", Email: "postgres-test@invalid.example",
		PasswordHash: "not-used-by-tests", IsStaff: true, IsSuperuser: true,
		IsActive: true, Permissions: []byte(`[]`), Created: now, Updated: now,
	}).Error)
	interfaceIDs := seedIPAddressConcurrencyInterfaces(t, db, now)
	var clock shared.Clock = ipamPostgresConcurrencyClock{now: shared.NewTimestamp(now)}
	if advanceClock {
		clock = &ipamPostgresAdvancingClock{now: shared.NewTimestamp(now)}
	}
	service, err := applicationipam.NewIPAddressService(
		NewIPAddressRepository(db),
		NewVRFRepository(db),
		postgresdcim.NewInterfaceRepository(db),
		postgresTransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		authz.AllowAll{},
		clock,
	)
	require.NoError(t, err)
	return db, service, identity.Principal{
		ID: 1, Username: "postgres-test", IsSuperuser: true,
	}, interfaceIDs
}

func seedIPAddressConcurrencyInterfaces(t *testing.T, db *gorm.DB, now time.Time) []shared.ID {
	t.Helper()
	metadata := dcimrow.RowMetadata{Created: now, LastUpdated: now}
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
		RowMetadata: metadata, Name: "Router", Slug: "router", Color: "112233",
	}
	require.NoError(t, db.Create(&role).Error)
	name := "edge-01"
	device := dcimrow.DeviceRow{
		RowMetadata: metadata, DeviceTypeID: deviceType.ID, RoleID: role.ID,
		Name: &name, SiteID: site.ID, Status: "active",
	}
	require.NoError(t, db.Create(&device).Error)
	interfaces := []dcimrow.InterfaceRow{
		{RowMetadata: metadata, DeviceID: device.ID, Name: "Ethernet1", Type: "1000base-t", Enabled: true},
		{RowMetadata: metadata, DeviceID: device.ID, Name: "Ethernet2", Type: "1000base-t", Enabled: true},
	}
	require.NoError(t, db.Create(&interfaces).Error)
	return []shared.ID{shared.ID(interfaces[0].ID), shared.ID(interfaces[1].ID)}
}

func ipamPostgresDSNWithSearchPath(t *testing.T, dsn, schema string) string {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err == nil && (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		query := parsed.Query()
		query.Set("search_path", schema)
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}
	if strings.Contains(dsn, "=") {
		return strings.TrimSpace(dsn) + " search_path=" + schema
	}
	t.Fatalf("NETBOX_TEST_POSTGRES_DSN must be a postgres URL or keyword DSN")
	return ""
}

func decodeIPAddressChangeData(t *testing.T, encoded []byte) map[string]any {
	t.Helper()
	var data map[string]any
	require.NoError(t, json.Unmarshal(encoded, &data))
	return data
}

func changeJSONInt64(t *testing.T, value any) int64 {
	t.Helper()
	number, ok := value.(float64)
	require.True(t, ok, "expected JSON number, got %T (%v)", value, value)
	return int64(number)
}

type ipamPostgresConcurrencyClock struct{ now shared.Timestamp }

func (clock ipamPostgresConcurrencyClock) Now() shared.Timestamp { return clock.now }

type ipamPostgresAdvancingClock struct {
	mu  sync.Mutex
	now shared.Timestamp
}

func (clock *ipamPostgresAdvancingClock) Now() shared.Timestamp {
	clock.mu.Lock()
	defer clock.mu.Unlock()
	current := clock.now
	clock.now = shared.NewTimestamp(clock.now.Add(time.Second))
	return current
}
