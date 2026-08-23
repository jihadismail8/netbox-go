package ipam_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	postgresdcim "netbox-go/internal/adapters/postgres/dcim"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	postgresipam "netbox-go/internal/adapters/postgres/ipam"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgrestransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationipam "netbox-go/internal/application/ipam"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestIPAddressServiceEnforcesConditionalHostUniqueness(t *testing.T) {
	db, service, _, vrfs, _, _ := newIPAddressApplicationService(t)
	principal := testPrincipal()

	first, err := service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand("192.0.2.17/24", "", nil),
	)
	require.NoError(t, err)
	assert.Equal(t, "192.0.2.17/24", first.Display())
	assert.True(t, first.Role().IsNull())

	_, err = service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand("192.0.2.17/32", "", nil),
	)
	require.Error(t, err)
	assert.Equal(t, "address", shared.ViolationsOf(err)[0].Field)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "global table")

	for _, role := range []domainipam.IPAddressRole{
		domainipam.IPAddressRoleAnycast,
		domainipam.IPAddressRoleVIP,
		domainipam.IPAddressRoleVRRP,
	} {
		_, err = service.CreateIPAddress(
			t.Context(), principal,
			createIPAddressCommand("198.51.100.9/24", role, nil),
		)
		require.NoError(t, err)
	}

	nonUnique := createApplicationVRF(t, vrfs, "Shared", "65000:90", false)
	for range 2 {
		_, err = service.CreateIPAddress(
			t.Context(), principal,
			createIPAddressCommand(
				"203.0.113.8/24", "", ipAddressIDPointer(nonUnique.ID()),
			),
		)
		require.NoError(t, err)
	}
	unique := createApplicationVRF(t, vrfs, "Unique IP", "65000:91", true)
	_, err = service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand(
			"203.0.113.9/24", "", ipAddressIDPointer(unique.ID()),
		),
	)
	require.NoError(t, err)
	_, err = service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand(
			"203.0.113.9/32", "", ipAddressIDPointer(unique.ID()),
		),
	)
	require.Error(t, err)
	assert.Contains(t, shared.ViolationsOf(err)[0].Description, "Unique IP")

	var persisted, changes int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).Count(&persisted).Error)
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changes).Error)
	assert.Equal(t, persisted, changes)
}

func TestIPAddressServiceAssignmentPartialPatchRollbackAndCascade(t *testing.T) {
	db, service, repository, _, firstInterface, secondInterface :=
		newIPAddressApplicationService(t)
	principal := testPrincipal()

	network, err := service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand("192.0.2.0/24", "", nil),
	)
	require.NoError(t, err)
	var changesBefore int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	_, err = service.AssignIPAddress(
		t.Context(), principal, applicationipam.AssignIPAddressCommand{
			ID: network.ID(), InterfaceID: shared.ID(firstInterface.ID),
		},
	)
	require.Error(t, err)
	assert.Equal(t, "non_field_errors", shared.ViolationsOf(err)[0].Field)
	reloaded, err := repository.Get(t.Context(), network.ID())
	require.NoError(t, err)
	assert.True(t, reloaded.Assignment().IsNull())
	var changesAfter int64
	require.NoError(t, db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	assert.Equal(t, changesBefore, changesAfter)

	address, err := service.CreateIPAddress(
		t.Context(), principal,
		createIPAddressCommand("192.0.2.1/24", "", nil),
	)
	require.NoError(t, err)
	address, err = service.AssignIPAddress(
		t.Context(), principal, applicationipam.AssignIPAddressCommand{
			ID: address.ID(), InterfaceID: shared.ID(firstInterface.ID),
		},
	)
	require.NoError(t, err)
	assignment, present := address.Assignment().Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(firstInterface.ID), assignment.ID())

	address, err = service.UpdateIPAddress(
		t.Context(), principal, applicationipam.UpdateIPAddressCommand{
			ID:               address.ID(),
			AssignedObjectID: applicationipam.FieldValue(secondInterface.ID),
		},
	)
	require.NoError(t, err, "ID-only patch must retain the assignment type")
	assignment, present = address.Assignment().Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(secondInterface.ID), assignment.ID())

	address, err = service.UpdateIPAddress(
		t.Context(), principal, applicationipam.UpdateIPAddressCommand{
			ID: address.ID(),
			AssignedObjectType: applicationipam.FieldValue(
				domainipam.IPAddressAssignmentType,
			),
		},
	)
	require.NoError(t, err, "type-only patch must retain the Interface ID")
	assignment, present = address.Assignment().Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(secondInterface.ID), assignment.ID())

	changes, err := service.DeleteAssignedToInterface(
		t.Context(), shared.ID(secondInterface.ID), applicationUpdatedAt,
	)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, address.ID(), changes[0].ID)
	assert.Equal(t, domainipam.IPAddressObjectType, changes[0].ObjectType)
	_, err = repository.Get(t.Context(), address.ID())
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonNotFound))
}

func TestReplaceIPAddressPreservesOmittedState(t *testing.T) {
	_, service, _, vrfs, firstInterface, _ := newIPAddressApplicationService(t)
	principal := testPrincipal()
	vrf := createApplicationVRF(t, vrfs, "Preserved", "65000:101", false)

	created, err := service.CreateIPAddress(
		t.Context(), principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.17/24"),
			VRF:     applicationipam.FieldValue(vrf.ID().Int64()),
			Status: applicationipam.FieldValue(
				domainipam.IPAddressStatusReserved.String(),
			),
			Role: applicationipam.FieldValue(
				domainipam.IPAddressRoleLoopback.String(),
			),
			DNSName:     applicationipam.FieldValue(" Original.Example. "),
			Description: applicationipam.FieldValue(" original description "),
			Comments:    applicationipam.FieldValue(" original comments "),
			AssignedObjectType: applicationipam.FieldValue(
				domainipam.IPAddressAssignmentType,
			),
			AssignedObjectID: applicationipam.FieldValue(firstInterface.ID),
		},
	)
	require.NoError(t, err)

	replaced, err := service.ReplaceIPAddress(
		t.Context(), principal, applicationipam.ReplaceIPAddressCommand{
			ID: created.ID(),
			CreateIPAddressCommand: applicationipam.CreateIPAddressCommand{
				Address: applicationipam.FieldValue("198.51.100.8"),
			},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "198.51.100.8/32", replaced.Display())
	assert.Equal(t, domainipam.IPAddressStatusReserved, replaced.Status())
	role, present := replaced.Role().Get()
	require.True(t, present)
	assert.Equal(t, domainipam.IPAddressRoleLoopback, role)
	assert.Equal(t, "original.example.", replaced.DNSName())
	assert.Equal(t, "original description", replaced.Description())
	assert.Equal(t, "original comments", replaced.Comments())
	resolvedVRF, present := replaced.VRF().Get()
	require.True(t, present)
	assert.Equal(t, vrf.ID(), resolvedVRF.ID())
	assignment, present := replaced.Assignment().Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(firstInterface.ID), assignment.ID())
}

func TestIPAddressScalarValidationLeavesStateUnchanged(t *testing.T) {
	clock := &advancingIPAddressClock{now: applicationUpdatedAt}
	db, service, repository, _, firstInterface, _ :=
		newIPAddressApplicationServiceWithClock(t, clock)
	principal := testPrincipal()
	created, err := service.CreateIPAddress(
		t.Context(), principal, applicationipam.CreateIPAddressCommand{
			Address: applicationipam.FieldValue("192.0.2.19/32"),
			Status: applicationipam.FieldValue(
				domainipam.IPAddressStatusReserved.String(),
			),
			Role: applicationipam.FieldValue(
				domainipam.IPAddressRoleLoopback.String(),
			),
			DNSName:     applicationipam.FieldValue("unchanged.example"),
			Description: applicationipam.FieldValue("unchanged description"),
			Comments:    applicationipam.FieldValue("unchanged comments"),
			AssignedObjectType: applicationipam.FieldValue(
				domainipam.IPAddressAssignmentType,
			),
			AssignedObjectID: applicationipam.FieldValue(firstInterface.ID),
		},
	)
	require.NoError(t, err)
	wantState := created.State()

	type scalarRejection struct {
		name               string
		field              string
		description        string
		operations         []string
		address            applicationipam.Field[string]
		status             applicationipam.Field[string]
		role               applicationipam.Field[string]
		dnsName            applicationipam.Field[string]
		commandDescription applicationipam.Field[string]
		comments           applicationipam.Field[string]
	}
	allOperations := []string{"create", "replace", "update"}
	createAndReplace := []string{"create", "replace"}
	tests := []scalarRejection{
		{
			name: "address omitted", field: "address",
			description: "This field is required.", operations: createAndReplace,
		},
		{
			name: "address null", field: "address",
			description: "This field may not be null.", operations: allOperations,
			address: applicationipam.NullField[string](),
		},
		{
			name: "address blank", field: "address",
			description: "This field may not be blank.", operations: allOperations,
			address: applicationipam.FieldValue(""),
		},
		{
			name: "address invalid", field: "address",
			description: "Invalid IP address format: invalid",
			operations:  allOperations, address: applicationipam.FieldValue("invalid"),
		},
		{
			name: "address leading whitespace", field: "address",
			description: "Invalid IP address format:  198.51.100.19/32",
			operations:  allOperations,
			address:     applicationipam.FieldValue(" 198.51.100.19/32"),
		},
		{
			name: "status null", field: "status",
			description: "This field may not be blank.", operations: allOperations,
			status: applicationipam.NullField[string](),
		},
		{
			name: "status blank", field: "status",
			description: "This field may not be blank.", operations: allOperations,
			status: applicationipam.FieldValue(""),
		},
		{
			name: "status invalid choice", field: "status",
			description: " invalid  is not a valid choice.", operations: allOperations,
			status: applicationipam.FieldValue(" invalid "),
		},
		{
			name: "status boolean-like choice", field: "status",
			description: "True is not a valid choice.", operations: allOperations,
			status: applicationipam.FieldValue("true"),
		},
		{
			name: "role invalid choice", field: "role",
			description: " invalid  is not a valid choice.", operations: allOperations,
			role: applicationipam.FieldValue(" invalid "),
		},
		{
			name: "role integer-like choice", field: "role",
			description: "1 is not a valid choice.", operations: allOperations,
			role: applicationipam.FieldValue("001"),
		},
		{
			name: "dns name null", field: "dns_name",
			description: "This field may not be null.", operations: allOperations,
			dnsName: applicationipam.NullField[string](),
		},
		{
			name: "dns name invalid", field: "dns_name",
			description: "Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
			operations:  allOperations, dnsName: applicationipam.FieldValue("bad name"),
		},
		{
			name: "description null", field: "description",
			description: "This field may not be null.", operations: allOperations,
			commandDescription: applicationipam.NullField[string](),
		},
		{
			name: "comments null", field: "comments",
			description: "This field may not be null.", operations: allOperations,
			comments: applicationipam.NullField[string](),
		},
		{
			name: "description null character", field: "description",
			description: "Null characters are not allowed.", operations: allOperations,
			commandDescription: applicationipam.FieldValue("contains\x00null"),
		},
		{
			name: "comments null character", field: "comments",
			description: "Null characters are not allowed.", operations: allOperations,
			comments: applicationipam.FieldValue("contains\x00null"),
		},
	}
	mutation := func(operation string, test scalarRejection) error {
		address := test.address
		if test.field != "address" && operation != "update" {
			address = applicationipam.FieldValue("198.51.100.19/32")
		}
		createCommand := applicationipam.CreateIPAddressCommand{
			Address: address, Status: test.status, Role: test.role,
			DNSName: test.dnsName, Description: test.commandDescription,
			Comments: test.comments,
		}
		switch operation {
		case "create":
			_, err := service.CreateIPAddress(t.Context(), principal, createCommand)
			return err
		case "replace":
			_, err := service.ReplaceIPAddress(
				t.Context(), principal, applicationipam.ReplaceIPAddressCommand{
					ID: created.ID(), CreateIPAddressCommand: createCommand,
				},
			)
			return err
		case "update":
			_, err := service.UpdateIPAddress(
				t.Context(), principal, applicationipam.UpdateIPAddressCommand{
					ID: created.ID(), Address: address, Status: test.status,
					Role: test.role, DNSName: test.dnsName,
					Description: test.commandDescription, Comments: test.comments,
				},
			)
			return err
		default:
			t.Fatalf("unsupported scalar rejection operation %q", operation)
			return nil
		}
	}
	for _, test := range tests {
		for _, operation := range test.operations {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				var rowsBefore, changesBefore int64
				require.NoError(
					t,
					db.Model(&ipamrow.IPAddressRow{}).Count(&rowsBefore).Error,
				)
				require.NoError(
					t,
					db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error,
				)
				err := mutation(operation, test)
				require.Error(t, err)
				violations := shared.ViolationsOf(err)
				require.Len(t, violations, 1)
				assert.Equal(t, test.field, violations[0].Field)
				assert.Equal(t, test.description, violations[0].Description)

				reloaded, err := repository.Get(t.Context(), created.ID())
				require.NoError(t, err)
				assert.Equal(t, wantState, reloaded.State())
				var rowsAfter, changesAfter int64
				require.NoError(
					t,
					db.Model(&ipamrow.IPAddressRow{}).Count(&rowsAfter).Error,
				)
				require.NoError(
					t,
					db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error,
				)
				assert.Equal(t, rowsBefore, rowsAfter)
				assert.Equal(t, changesBefore, changesAfter)
			})
		}
	}
	assert.Greater(t, clock.calls, 1, "domain-level rejections must exercise an advancing clock")
}

func newIPAddressApplicationService(
	t *testing.T,
) (
	*gorm.DB,
	*applicationipam.IPAddressService,
	*postgresipam.IPAddressRepository,
	*postgresipam.VRFRepository,
	dcimrow.InterfaceRow,
	dcimrow.InterfaceRow,
) {
	return newIPAddressApplicationServiceWithClock(
		t,
		fixedClock{now: applicationUpdatedAt},
	)
}

func newIPAddressApplicationServiceWithClock(
	t *testing.T,
	clock shared.Clock,
) (
	*gorm.DB,
	*applicationipam.IPAddressService,
	*postgresipam.IPAddressRepository,
	*postgresipam.VRFRepository,
	dcimrow.InterfaceRow,
	dcimrow.InterfaceRow,
) {
	t.Helper()
	db, _, _, vrfs := newPrefixApplicationService(
		t, nil, &authz.AllowAll{},
	)
	first, second := seedIPAddressApplicationInterfaceGraph(t, db)
	repository := postgresipam.NewIPAddressRepository(db)
	service, err := applicationipam.NewIPAddressService(
		repository,
		vrfs,
		postgresdcim.NewInterfaceRepository(db),
		postgrestransaction.NewUnitOfWork(db),
		postgreschangelog.NewRecorder(db),
		&authz.AllowAll{},
		clock,
	)
	require.NoError(t, err)
	return db, service, repository, vrfs, first, second
}

type advancingIPAddressClock struct {
	now   shared.Timestamp
	calls int
}

func (clock *advancingIPAddressClock) Now() shared.Timestamp {
	current := clock.now
	clock.calls++
	clock.now = shared.NewTimestamp(clock.now.Add(time.Second))
	return current
}

func createIPAddressCommand(
	address string,
	role domainipam.IPAddressRole,
	vrfID *shared.ID,
) applicationipam.CreateIPAddressCommand {
	command := applicationipam.CreateIPAddressCommand{
		Address: applicationipam.FieldValue(address),
	}
	if role != "" {
		command.Role = applicationipam.FieldValue(role.String())
	}
	if vrfID != nil {
		command.VRF = applicationipam.FieldValue(vrfID.Int64())
	}
	return command
}

func ipAddressIDPointer(value shared.ID) *shared.ID { return &value }

func seedIPAddressApplicationInterfaceGraph(
	t *testing.T,
	db *gorm.DB,
) (dcimrow.InterfaceRow, dcimrow.InterfaceRow) {
	t.Helper()
	metadata := dcimrow.RowMetadata{
		Created: applicationCreatedAt.Time, LastUpdated: applicationUpdatedAt.Time,
	}
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
	name := "edge-01"
	device := dcimrow.DeviceRow{
		RowMetadata: metadata, DeviceTypeID: deviceType.ID, RoleID: role.ID,
		Name: &name, SiteID: site.ID, Status: "active",
	}
	require.NoError(t, db.Create(&device).Error)
	first := dcimrow.InterfaceRow{
		RowMetadata: metadata, DeviceID: device.ID, Name: "Ethernet1",
		Type: "1000base-t", Enabled: true,
	}
	second := dcimrow.InterfaceRow{
		RowMetadata: metadata, DeviceID: device.ID, Name: "Ethernet2",
		Type: "1000base-t", Enabled: true,
	}
	require.NoError(t, db.Create(&first).Error)
	require.NoError(t, db.Create(&second).Error)
	return first, second
}
