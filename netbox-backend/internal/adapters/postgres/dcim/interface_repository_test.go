package dcim

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipampostgres "netbox-go/internal/adapters/postgres/ipam"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgresTransaction "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	applicationchangelog "netbox-go/internal/application/changelog"
	applicationdcim "netbox-go/internal/application/dcim"
	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceRepositoryMapsFiltersNullabilityCounterAndUniqueness(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	manufacturerRepository := NewManufacturerRepository(db)
	deviceTypeRepository := NewDeviceTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Router", "router")
	require.NoError(t, deviceTypeRepository.Create(t.Context(), deviceType))
	siteID, roleID := seedDeviceTypeDeviceDependencies(t, db)
	deviceName, assetTag := "edge-01", "EDGE-01"
	device := dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: deviceType.ID().Int64(), RoleID: roleID,
		Name: &deviceName, SiteID: siteID, Status: "active", AssetTag: &assetTag,
	}
	require.NoError(t, db.Create(&device).Error)

	repository := NewInterfaceRepository(db)
	fixtures := []*domaindcim.Interface{
		newInterfaceRepositoryFixture(
			t, shared.ID(device.ID), deviceName+" ("+assetTag+")", "Ethernet2",
			"10gbase-sr", false, true,
			domaindcim.NonNullDeviceValue(uint32(9000)),
			domaindcim.NonNullDeviceValue(uint64(10_000_000)),
			domaindcim.NonNullDeviceValue("full"),
		),
		newInterfaceRepositoryFixture(
			t, shared.ID(device.ID), deviceName+" ("+assetTag+")", "Ethernet1",
			"10gbase-sr", false, true,
			domaindcim.NullDeviceValue[uint32](),
			domaindcim.NullDeviceValue[uint64](),
			domaindcim.NonNullDeviceValue(""),
		),
	}
	for _, networkInterface := range fixtures {
		require.NoError(t, repository.Create(t.Context(), networkInterface))
	}
	assignedType := domaindcim.InterfaceObjectType
	assignedID := fixtures[1].ID().Int64()
	require.NoError(t, db.Create(&ipamrow.IPAddressRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Address: "192.0.2.1/24", Status: "active",
		AssignedObjectType: &assignedType, AssignedObjectID: &assignedID,
	}).Error)

	loaded, err := repository.Get(t.Context(), fixtures[1].ID())
	require.NoError(t, err)
	assert.Equal(t, shared.ID(device.ID), loaded.Device().ID())
	assert.Equal(t, "edge-01 (EDGE-01)", loaded.Device().Display())
	assert.True(t, loaded.MTU().IsNull())
	assert.True(t, loaded.Speed().IsNull())
	duplex, present := loaded.Duplex().Get()
	require.True(t, present)
	assert.Empty(t, duplex)
	assert.Equal(t, uint64(1), loaded.IPAddressCount())

	enabled := false
	mgmtOnly := true
	page, err := repository.List(t.Context(), applicationdcim.InterfaceListCriteria{
		Limit: 50, DeviceIDs: []int64{-1, device.ID},
		DeviceNames: []string{"missing", deviceName}, Names: []string{"Ethernet1", "Ethernet2"},
		Types:   []domaindcim.InterfaceType{"10gbase-sr"},
		Enabled: &enabled, MgmtOnly: &mgmtOnly, Query: "description",
		Ordering: []applicationdcim.InterfaceSort{
			{Field: applicationdcim.InterfaceSortDevice},
			{Field: applicationdcim.InterfaceSortName},
			{Field: applicationdcim.InterfaceSortID},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(2), page.Count)
	require.Len(t, page.Results, 2)
	assert.Equal(t, []string{"Ethernet1", "Ethernet2"}, []string{
		page.Results[0].Name(), page.Results[1].Name(),
	})

	duplicate := newInterfaceRepositoryFixture(
		t, shared.ID(device.ID), deviceName, "Ethernet1", "other", true, false,
		domaindcim.NullDeviceValue[uint32](),
		domaindcim.NullDeviceValue[uint64](),
		domaindcim.NullDeviceValue[string](),
	)
	err = repository.Create(t.Context(), duplicate)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonConflict))
	assert.Equal(t, []shared.FieldViolation{{
		Field: "non_field_errors", Reason: "unique_together",
		Description: "The fields device, name must make a unique set.",
	}}, shared.ViolationsOf(err))
}

func TestInterfaceServiceDeletesAssignedIPBeforeParentWithTypedAudit(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository, networkInterface := seedInterfaceRepositoryGraph(t, db, "Ethernet1")
	assignedType := domaindcim.InterfaceObjectType
	assignedID := networkInterface.ID().Int64()
	ipAddress := ipamrow.IPAddressRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Address: "192.0.2.1/24", Status: "active",
		AssignedObjectType: &assignedType, AssignedObjectID: &assignedID,
	}
	require.NoError(t, db.Create(&ipAddress).Error)
	service := newPostgresInterfaceService(
		t, db, repository, postgreschangelog.NewRecorder(db),
	)

	require.NoError(t, service.DeleteInterface(
		t.Context(), rackTypePrincipal(),
		applicationdcim.DeleteInterfaceCommand{ID: networkInterface.ID()},
	))
	_, err := repository.Get(t.Context(), networkInterface.ID())
	require.Error(t, err)
	var addressCount int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).
		Where("id = ?", ipAddress.ID).Count(&addressCount).Error)
	assert.Zero(t, addressCount)
	var changes []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&changes).Error)
	require.Len(t, changes, 2)
	assert.Equal(t, "ipam.ipaddress", changes[0].Kind)
	assert.Equal(t, ipAddress.ID, changes[0].ObjectID)
	assert.Equal(t, domaindcim.InterfaceObjectType, changes[1].Kind)
	assert.Equal(t, networkInterface.ID().Int64(), changes[1].ObjectID)
}

func TestInterfaceDeleteRollsBackCascadeWhenAuditFails(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	repository, networkInterface := seedInterfaceRepositoryGraph(t, db, "Ethernet1")
	assignedType := domaindcim.InterfaceObjectType
	assignedID := networkInterface.ID().Int64()
	ipAddress := ipamrow.IPAddressRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Address: "192.0.2.1/24", Status: "active",
		AssignedObjectType: &assignedType, AssignedObjectID: &assignedID,
	}
	require.NoError(t, db.Create(&ipAddress).Error)
	sentinel := errors.New("forced Interface audit failure")
	service := newPostgresInterfaceService(t, db, repository, &failOnRecorder{
		delegate: postgreschangelog.NewRecorder(db), failAt: 1, err: sentinel,
	})

	err := service.DeleteInterface(
		t.Context(), rackTypePrincipal(),
		applicationdcim.DeleteInterfaceCommand{ID: networkInterface.ID()},
	)
	require.ErrorIs(t, err, sentinel)
	_, getErr := repository.Get(t.Context(), networkInterface.ID())
	require.NoError(t, getErr)
	var addressCount int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).
		Where("id = ?", ipAddress.ID).Count(&addressCount).Error)
	assert.Equal(t, int64(1), addressCount)
	assertTableCount(t, db, &postgreschangelog.ChangeRow{}, 0)
}

func seedInterfaceRepositoryGraph(
	t *testing.T,
	db *gorm.DB,
	name string,
) (*InterfaceRepository, *domaindcim.Interface) {
	t.Helper()
	manufacturerRepository := NewManufacturerRepository(db)
	deviceTypeRepository := NewDeviceTypeRepository(db)
	manufacturer := newManufacturerFixture(t, "Vendor", "vendor")
	require.NoError(t, manufacturerRepository.Create(t.Context(), manufacturer))
	deviceType := newInterfaceTemplateDeviceTypeFixture(t, manufacturer, "Router", "router")
	require.NoError(t, deviceTypeRepository.Create(t.Context(), deviceType))
	siteID, roleID := seedDeviceTypeDeviceDependencies(t, db)
	deviceName := "edge-01"
	device := dcimrow.DeviceRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		DeviceTypeID: deviceType.ID().Int64(), RoleID: roleID,
		Name: &deviceName, SiteID: siteID, Status: "active",
	}
	require.NoError(t, db.Create(&device).Error)
	repository := NewInterfaceRepository(db)
	networkInterface := newInterfaceRepositoryFixture(
		t, shared.ID(device.ID), deviceName, name, "1000base-t", true, false,
		domaindcim.NullDeviceValue[uint32](),
		domaindcim.NullDeviceValue[uint64](),
		domaindcim.NullDeviceValue[string](),
	)
	require.NoError(t, repository.Create(t.Context(), networkInterface))
	return repository, networkInterface
}

func newPostgresInterfaceService(
	t *testing.T,
	db *gorm.DB,
	repository *InterfaceRepository,
	recorder applicationchangelog.Recorder,
) *applicationdcim.InterfaceService {
	t.Helper()
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	ipAddressService, err := applicationipam.NewIPAddressService(
		ipampostgres.NewIPAddressRepository(db), ipampostgres.NewVRFRepository(db),
		repository, unitOfWork, recorder, authz.AllowAll{},
		interfaceTemplateRepositoryClock{},
	)
	require.NoError(t, err)
	service, err := applicationdcim.NewInterfaceService(
		repository, repository, ipAddressService, unitOfWork, recorder,
		authz.AllowAll{}, interfaceTemplateRepositoryClock{},
	)
	require.NoError(t, err)
	return service
}

func newInterfaceRepositoryFixture(
	t *testing.T,
	deviceID shared.ID,
	deviceDisplay string,
	name string,
	interfaceType string,
	enabled bool,
	mgmtOnly bool,
	mtu domaindcim.DeviceNullable[uint32],
	speed domaindcim.DeviceNullable[uint64],
	duplex domaindcim.DeviceNullable[string],
) *domaindcim.Interface {
	t.Helper()
	deviceName := "edge-01"
	reference, err := domaindcim.NewDeviceReference(
		deviceID, domaindcim.NonNullDeviceValue(deviceName), deviceDisplay,
	)
	require.NoError(t, err)
	networkInterface, err := domaindcim.NewInterface(domaindcim.InterfaceValues{
		Device: reference, Name: name, Type: interfaceType,
		Enabled: enabled, MgmtOnly: mgmtOnly, MTU: mtu, Speed: speed, Duplex: duplex,
		Description: "Interface description",
	}, repositoryCreatedAt)
	require.NoError(t, err)
	return networkInterface
}
