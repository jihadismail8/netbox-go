package dcim

import (
	"encoding/json"
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
	applicationdcim "netbox-go/internal/application/dcim"
	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceTypedWorkflowInstantiatesTemplatesAndCascadesChildren(t *testing.T) {
	db, _ := newSiteTestDatabase(t, nil)
	seedChangeActor(t, db)
	graph := seedDeviceWorkflowGraph(t, db)
	service := newPostgresDeviceService(t, db, graph)

	device, err := service.CreateDevice(
		t.Context(),
		deviceTypePrincipal(),
		applicationdcim.CreateDeviceCommand{
			DeviceType: applicationdcim.FieldValue(graph.deviceType.ID()),
			Role:       applicationdcim.FieldValue(graph.deviceRole.ID()),
			Name:       applicationdcim.FieldValue(" edge-01 "),
			Site:       applicationdcim.FieldValue(graph.site.ID()),
			Status:     applicationdcim.FieldValue("active"),
			AssetTag:   applicationdcim.NullField[string](),
		},
	)
	require.NoError(t, err)
	assert.True(t, device.ID().IsValid())
	assert.Equal(t, "edge-01", device.Display())
	assert.True(t, device.AssetTag().IsNull())

	loaded, err := graph.devices.Get(t.Context(), device.ID())
	require.NoError(t, err)
	assert.Equal(t, device.ID(), loaded.ID())
	assert.Equal(t, graph.deviceType.ID(), loaded.DeviceType().ID())
	assert.Equal(t, graph.deviceRole.ID(), loaded.Role().ID)
	assert.Equal(t, graph.site.ID(), loaded.Site().ID())
	assert.Equal(t, uint64(2), loaded.InterfaceCount())

	page, err := graph.devices.List(t.Context(), applicationdcim.DeviceListCriteria{
		Limit: 50, SiteIDs: []int64{graph.site.ID().Int64()},
		DeviceTypeSlugs: []string{graph.deviceType.Slug().String()},
		RoleSlugs:       []string{graph.deviceRole.Slug().String()},
		Names:           []string{"edge-01"},
		Statuses:        []domaindcim.DeviceStatus{domaindcim.DeviceStatusActive},
		Ordering: []applicationdcim.DeviceSort{
			{Field: applicationdcim.DeviceSortName},
			{Field: applicationdcim.DeviceSortID},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(1), page.Count)
	require.Len(t, page.Results, 1)
	assert.Equal(t, device.ID(), page.Results[0].ID())

	interfaces, err := graph.interfaces.List(
		t.Context(),
		applicationdcim.InterfaceListCriteria{
			Limit:     50,
			DeviceIDs: []int64{device.ID().Int64()},
			Ordering: []applicationdcim.InterfaceSort{
				{Field: applicationdcim.InterfaceSortName},
				{Field: applicationdcim.InterfaceSortID},
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, interfaces.Results, 2)
	assert.Equal(t, []string{"eth0", "eth1"}, []string{
		interfaces.Results[0].Name(), interfaces.Results[1].Name(),
	})
	assignedType := domaindcim.InterfaceObjectType
	assignedID := interfaces.Results[0].ID().Int64()
	address := ipamrow.IPAddressRow{
		RowMetadata: dcimrow.RowMetadata{
			Created: repositoryCreatedAt.Time, LastUpdated: repositoryCreatedAt.Time,
		},
		Address: "192.0.2.1/24", Status: "active",
		AssignedObjectType: &assignedType, AssignedObjectID: &assignedID,
	}
	require.NoError(t, db.Create(&address).Error)

	var createChanges []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&createChanges).Error)
	require.Len(t, createChanges, 3)
	assert.Equal(t, []string{
		domaindcim.DeviceObjectType,
		domaindcim.InterfaceObjectType,
		domaindcim.InterfaceObjectType,
	}, []string{createChanges[0].Kind, createChanges[1].Kind, createChanges[2].Kind})

	require.NoError(t, db.Where("1 = 1").Delete(&postgreschangelog.ChangeRow{}).Error)
	require.NoError(t, service.DeleteDevice(
		t.Context(),
		deviceTypePrincipal(),
		applicationdcim.DeleteDeviceCommand{ID: device.ID()},
	))

	_, err = graph.devices.Get(t.Context(), device.ID())
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonNotFound))
	var interfaceCount int64
	require.NoError(t, db.Model(&dcimrow.InterfaceRow{}).
		Where("device_id = ?", device.ID().Int64()).Count(&interfaceCount).Error)
	assert.Zero(t, interfaceCount)
	var addressCount int64
	require.NoError(t, db.Model(&ipamrow.IPAddressRow{}).
		Where("id = ?", address.ID).Count(&addressCount).Error)
	assert.Zero(t, addressCount)

	var deleteChanges []postgreschangelog.ChangeRow
	require.NoError(t, db.Order("id").Find(&deleteChanges).Error)
	require.Len(t, deleteChanges, 4)
	assert.Equal(t, []string{
		"ipam.ipaddress",
		domaindcim.InterfaceObjectType,
		domaindcim.InterfaceObjectType,
		domaindcim.DeviceObjectType,
	}, []string{
		deleteChanges[0].Kind,
		deleteChanges[1].Kind,
		deleteChanges[2].Kind,
		deleteChanges[3].Kind,
	})
	var before map[string]any
	require.NoError(t, json.Unmarshal(deleteChanges[3].BeforeData, &before))
	assert.Equal(t, "edge-01", before["name"])
}

type deviceWorkflowGraph struct {
	site        *domaindcim.Site
	deviceType  *domaindcim.DeviceType
	deviceRole  *domaindcim.DeviceRole
	devices     *DeviceRepository
	interfaces  *InterfaceRepository
	templates   *InterfaceTemplateRepository
	deviceTypes *DeviceTypeRepository
	roles       *DeviceRoleRepository
	sites       *SiteRepository
	racks       *RackRepository
}

func seedDeviceWorkflowGraph(t *testing.T, db *gorm.DB) deviceWorkflowGraph {
	t.Helper()
	sites := NewSiteRepository(db)
	manufacturers := NewManufacturerRepository(db)
	deviceTypes := NewDeviceTypeRepository(db)
	roles := NewDeviceRoleRepository(db)
	templates := NewInterfaceTemplateRepository(db)

	site := newSiteFixture(t, "Primary", "primary", "")
	require.NoError(t, sites.Create(t.Context(), site))
	manufacturer := newManufacturerFixture(t, "Acme", "acme")
	require.NoError(t, manufacturers.Create(t.Context(), manufacturer))
	deviceType := newDeviceTypeFixture(
		t,
		manufacturer,
		"Router 9000",
		"router-9000",
		"1.5",
		domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
	)
	require.NoError(t, deviceTypes.Create(t.Context(), deviceType))
	deviceRole := newDeviceRoleFixture(
		t,
		domaindcim.RootDeviceRoleParent(),
		"Core Router",
		"core-router",
	)
	require.NoError(t, roles.Create(t.Context(), deviceRole))
	for index, name := range []string{"eth0", "eth1"} {
		template := newInterfaceTemplateRepositoryFixture(
			t, deviceType, name, "1000base-t", true, index == 0,
		)
		require.NoError(t, templates.Create(t.Context(), template))
	}
	return deviceWorkflowGraph{
		site: site, deviceType: deviceType, deviceRole: deviceRole,
		devices: NewDeviceRepository(db), interfaces: NewInterfaceRepository(db),
		templates: templates, deviceTypes: deviceTypes, roles: roles,
		sites: sites, racks: NewRackRepository(db),
	}
}

func newPostgresDeviceService(
	t *testing.T,
	db *gorm.DB,
	graph deviceWorkflowGraph,
) *applicationdcim.DeviceService {
	t.Helper()
	recorder := postgreschangelog.NewRecorder(db)
	unitOfWork := postgresTransaction.NewUnitOfWork(db)
	ipAddresses, err := applicationipam.NewIPAddressService(
		ipampostgres.NewIPAddressRepository(db),
		ipampostgres.NewVRFRepository(db),
		graph.interfaces,
		unitOfWork,
		recorder,
		authz.AllowAll{},
		deviceTypeRepositoryClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	interfaces, err := applicationdcim.NewInterfaceService(
		graph.interfaces,
		graph.devices,
		ipAddresses,
		unitOfWork,
		recorder,
		authz.AllowAll{},
		deviceTypeRepositoryClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	service, err := applicationdcim.NewDeviceService(
		graph.devices,
		graph.deviceTypes,
		graph.roles,
		graph.sites,
		graph.racks,
		graph.templates,
		graph.interfaces,
		interfaces,
		unitOfWork,
		recorder,
		authz.AllowAll{},
		deviceTypeRepositoryClock{now: repositoryUpdatedAt},
	)
	require.NoError(t, err)
	return service
}
