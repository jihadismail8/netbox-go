package ipam_test

import (
	"testing"

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
		fixedClock{now: applicationUpdatedAt},
	)
	require.NoError(t, err)
	return db, service, repository, vrfs, first, second
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
