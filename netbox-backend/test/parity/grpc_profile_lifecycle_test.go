package parity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	grpcadapter "netbox-go/internal/adapters/grpc/workflow"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

type profileParityEnvironment struct {
	db        *gorm.DB
	dcim      *grpcadapter.DCIMServer
	ipam      *grpcadapter.IPAMServer
	router    *gin.Engine
	principal identity.Principal
	ctx       context.Context
}

type profileFixtureIDs struct {
	site, manufacturer, rackRole, rackType, rack int64
	deviceRole, deviceType, device, iface, vrf   int64
}

type lifecycleList struct {
	count uint64
	ids   []int64
}

type lifecycleScenario struct {
	name    string
	path    string
	create  func(context.Context) (int64, error)
	list    func(context.Context) (lifecycleList, error)
	replace func(context.Context, int64) (string, error)
	delete  func(context.Context, int64) error
}

// TestCanonicalGRPCLifecycleParityAllProfileResources closes the positive
// lifecycle gap left by the REST-create/gRPC-patch parity test. Every profile
// resource is created, listed, fully replaced, and deleted through canonical
// gRPC, with the resulting state independently observed through canonical REST.
func TestCanonicalGRPCLifecycleParityAllProfileResources(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	fixtures := environment.seedProfileFixtures(t)
	createdDescription := "created through canonical gRPC"
	replacedDescription := "replaced through canonical gRPC"

	scenarios := []lifecycleScenario{
		{
			name: "site", path: "/api/dcim/sites",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateSite(ctx, &dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{
					Name: pointer("Lifecycle Site"), Slug: pointer("lifecycle-site"), Status: pointer("active"), Description: &createdDescription,
				}})
				if err != nil {
					return 0, err
				}
				return response.Site.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListSites(ctx, &dcimv1.ListSitesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.Site) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceSite(ctx, &dcimv1.ReplaceSiteRequest{Id: id, Site: &dcimv1.SiteInput{
					Name: pointer("Lifecycle Site"), Slug: pointer("lifecycle-site"), Status: pointer("staging"), Facility: pointer("LIFE1"), Description: &replacedDescription,
				}})
				if err != nil {
					return "", err
				}
				return response.Site.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteSite(ctx, &dcimv1.DeleteSiteRequest{Id: id})
				return err
			},
		},
		{
			name: "manufacturer", path: "/api/dcim/manufacturers",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateManufacturer(ctx, &dcimv1.CreateManufacturerRequest{Manufacturer: &dcimv1.ManufacturerInput{Name: pointer("Lifecycle Manufacturer"), Slug: pointer("lifecycle-manufacturer"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Manufacturer.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListManufacturers(ctx, &dcimv1.ListManufacturersRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.Manufacturer) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceManufacturer(ctx, &dcimv1.ReplaceManufacturerRequest{Id: id, Manufacturer: &dcimv1.ManufacturerInput{Name: pointer("Lifecycle Manufacturer"), Slug: pointer("lifecycle-manufacturer"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Manufacturer.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteManufacturer(ctx, &dcimv1.DeleteManufacturerRequest{Id: id})
				return err
			},
		},
		{
			name: "rack_role", path: "/api/dcim/rack-roles",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateRackRole(ctx, &dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{Name: pointer("Lifecycle Rack Role"), Slug: pointer("lifecycle-rack-role"), Color: pointer("aabbcc"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.RackRole.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListRackRoles(ctx, &dcimv1.ListRackRolesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.RackRole) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceRackRole(ctx, &dcimv1.ReplaceRackRoleRequest{Id: id, RackRole: &dcimv1.RackRoleInput{Name: pointer("Lifecycle Rack Role"), Slug: pointer("lifecycle-rack-role"), Color: pointer("ccbbaa"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.RackRole.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteRackRole(ctx, &dcimv1.DeleteRackRoleRequest{Id: id})
				return err
			},
		},
		{
			name: "rack_type", path: "/api/dcim/rack-types",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateRackType(ctx, &dcimv1.CreateRackTypeRequest{RackType: &dcimv1.RackTypeInput{Manufacturer: &fixtures.manufacturer, Model: pointer("Lifecycle Rack Type"), Slug: pointer("lifecycle-rack-type"), FormFactor: pointer("4-post-cabinet"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.RackType.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListRackTypes(ctx, &dcimv1.ListRackTypesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.RackType) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceRackType(ctx, &dcimv1.ReplaceRackTypeRequest{Id: id, RackType: &dcimv1.RackTypeInput{Manufacturer: &fixtures.manufacturer, Model: pointer("Lifecycle Rack Type"), Slug: pointer("lifecycle-rack-type"), FormFactor: pointer("4-post-frame"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.RackType.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteRackType(ctx, &dcimv1.DeleteRackTypeRequest{Id: id})
				return err
			},
		},
		{
			name: "rack", path: "/api/dcim/racks",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateRack(ctx, &dcimv1.CreateRackRequest{Rack: &dcimv1.RackInput{Site: &fixtures.site, Name: pointer("Lifecycle Rack"), RackType: wrapperspb.Int64(fixtures.rackType), Role: wrapperspb.Int64(fixtures.rackRole), Status: pointer("active"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Rack.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListRacks(ctx, &dcimv1.ListRacksRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.Rack) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceRack(ctx, &dcimv1.ReplaceRackRequest{Id: id, Rack: &dcimv1.RackInput{Site: &fixtures.site, Name: pointer("Lifecycle Rack"), RackType: wrapperspb.Int64(fixtures.rackType), Role: wrapperspb.Int64(fixtures.rackRole), Status: pointer("planned"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Rack.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteRack(ctx, &dcimv1.DeleteRackRequest{Id: id})
				return err
			},
		},
		{
			name: "device_role", path: "/api/dcim/device-roles",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateDeviceRole(ctx, &dcimv1.CreateDeviceRoleRequest{DeviceRole: &dcimv1.DeviceRoleInput{Name: pointer("Lifecycle Device Role"), Slug: pointer("lifecycle-device-role"), Color: pointer("112233"), VmRole: pointer(false), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.DeviceRole.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListDeviceRoles(ctx, &dcimv1.ListDeviceRolesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.DeviceRole) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceDeviceRole(ctx, &dcimv1.ReplaceDeviceRoleRequest{Id: id, DeviceRole: &dcimv1.DeviceRoleInput{Name: pointer("Lifecycle Device Role"), Slug: pointer("lifecycle-device-role"), Color: pointer("332211"), VmRole: pointer(true), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.DeviceRole.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteDeviceRole(ctx, &dcimv1.DeleteDeviceRoleRequest{Id: id})
				return err
			},
		},
		{
			name: "device_type", path: "/api/dcim/device-types",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateDeviceType(ctx, &dcimv1.CreateDeviceTypeRequest{DeviceType: &dcimv1.DeviceTypeInput{Manufacturer: &fixtures.manufacturer, Model: pointer("Lifecycle Device Type"), Slug: pointer("lifecycle-device-type"), UHeight: pointer("1.5"), IsFullDepth: pointer(false), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.DeviceType.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListDeviceTypes(ctx, &dcimv1.ListDeviceTypesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.DeviceType) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceDeviceType(ctx, &dcimv1.ReplaceDeviceTypeRequest{Id: id, DeviceType: &dcimv1.DeviceTypeInput{Manufacturer: &fixtures.manufacturer, Model: pointer("Lifecycle Device Type"), Slug: pointer("lifecycle-device-type"), UHeight: pointer("2"), IsFullDepth: pointer(true), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.DeviceType.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteDeviceType(ctx, &dcimv1.DeleteDeviceTypeRequest{Id: id})
				return err
			},
		},
		{
			name: "interface_template", path: "/api/dcim/interface-templates",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateInterfaceTemplate(ctx, &dcimv1.CreateInterfaceTemplateRequest{InterfaceTemplate: &dcimv1.InterfaceTemplateInput{DeviceType: &fixtures.deviceType, Name: pointer("lifecycle-template"), Type: pointer("1000base-t"), Enabled: pointer(true), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.InterfaceTemplate.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListInterfaceTemplates(ctx, &dcimv1.ListInterfaceTemplatesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.InterfaceTemplate) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceInterfaceTemplate(ctx, &dcimv1.ReplaceInterfaceTemplateRequest{Id: id, InterfaceTemplate: &dcimv1.InterfaceTemplateInput{DeviceType: &fixtures.deviceType, Name: pointer("lifecycle-template"), Type: pointer("1000base-t"), Enabled: pointer(false), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.InterfaceTemplate.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteInterfaceTemplate(ctx, &dcimv1.DeleteInterfaceTemplateRequest{Id: id})
				return err
			},
		},
		{
			name: "device", path: "/api/dcim/devices",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateDevice(ctx, &dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{DeviceType: &fixtures.deviceType, Role: &fixtures.deviceRole, Name: wrapperspb.String("lifecycle-device"), Site: &fixtures.site, Status: pointer("active"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Device.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListDevices(ctx, &dcimv1.ListDevicesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.Device) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceDevice(ctx, &dcimv1.ReplaceDeviceRequest{Id: id, Device: &dcimv1.DeviceInput{DeviceType: &fixtures.deviceType, Role: &fixtures.deviceRole, Name: wrapperspb.String("lifecycle-device"), Site: &fixtures.site, Status: pointer("planned"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Device.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteDevice(ctx, &dcimv1.DeleteDeviceRequest{Id: id})
				return err
			},
		},
		{
			name: "interface", path: "/api/dcim/interfaces",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.dcim.CreateInterface(ctx, &dcimv1.CreateInterfaceRequest{Interface: &dcimv1.InterfaceInput{Device: &fixtures.device, Name: pointer("lifecycle-interface"), Type: pointer("1000base-t"), Enabled: pointer(true), Mtu: wrapperspb.Int32(1500), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Interface.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.dcim.ListInterfaces(ctx, &dcimv1.ListInterfacesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *dcimv1.Interface) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.dcim.ReplaceInterface(ctx, &dcimv1.ReplaceInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{Device: &fixtures.device, Name: pointer("lifecycle-interface"), Type: pointer("1000base-t"), Enabled: pointer(false), Mtu: wrapperspb.Int32(9000), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Interface.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.dcim.DeleteInterface(ctx, &dcimv1.DeleteInterfaceRequest{Id: id})
				return err
			},
		},
		{
			name: "vrf", path: "/api/ipam/vrfs",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.ipam.CreateVRF(ctx, &ipamv1.CreateVRFRequest{Vrf: &ipamv1.VRFInput{Name: pointer("Lifecycle VRF"), Rd: wrapperspb.String("65000:999"), EnforceUnique: pointer(true), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Vrf.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.ipam.ListVRFs(ctx, &ipamv1.ListVRFsRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *ipamv1.VRF) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.ipam.ReplaceVRF(ctx, &ipamv1.ReplaceVRFRequest{Id: id, Vrf: &ipamv1.VRFInput{Name: pointer("Lifecycle VRF"), Rd: wrapperspb.String("65000:999"), EnforceUnique: pointer(false), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Vrf.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.ipam.DeleteVRF(ctx, &ipamv1.DeleteVRFRequest{Id: id})
				return err
			},
		},
		{
			name: "prefix", path: "/api/ipam/prefixes",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.ipam.CreatePrefix(ctx, &ipamv1.CreatePrefixRequest{Prefix: &ipamv1.PrefixInput{Prefix: pointer("203.0.113.0/25"), Vrf: wrapperspb.Int64(fixtures.vrf), Status: pointer("active"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.Prefix.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.ipam.ListPrefixes(ctx, &ipamv1.ListPrefixesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *ipamv1.Prefix) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.ipam.ReplacePrefix(ctx, &ipamv1.ReplacePrefixRequest{Id: id, Prefix: &ipamv1.PrefixInput{Prefix: pointer("203.0.113.0/25"), Vrf: wrapperspb.Int64(fixtures.vrf), Status: pointer("reserved"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.Prefix.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.ipam.DeletePrefix(ctx, &ipamv1.DeletePrefixRequest{Id: id})
				return err
			},
		},
		{
			name: "ip_address", path: "/api/ipam/ip-addresses",
			create: func(ctx context.Context) (int64, error) {
				response, err := environment.ipam.CreateIPAddress(ctx, &ipamv1.CreateIPAddressRequest{IpAddress: &ipamv1.IPAddressInput{Address: pointer("203.0.113.10/25"), Vrf: wrapperspb.Int64(fixtures.vrf), Status: pointer("active"), Description: &createdDescription}})
				if err != nil {
					return 0, err
				}
				return response.IpAddress.Id, nil
			},
			list: func(ctx context.Context) (lifecycleList, error) {
				response, err := environment.ipam.ListIPAddresses(ctx, &ipamv1.ListIPAddressesRequest{})
				if err != nil {
					return lifecycleList{}, err
				}
				return lifecycleList{response.Page.Count, collectIDs(response.Results, func(value *ipamv1.IPAddress) int64 { return value.Id })}, nil
			},
			replace: func(ctx context.Context, id int64) (string, error) {
				response, err := environment.ipam.ReplaceIPAddress(ctx, &ipamv1.ReplaceIPAddressRequest{Id: id, IpAddress: &ipamv1.IPAddressInput{Address: pointer("203.0.113.10/25"), Vrf: wrapperspb.Int64(fixtures.vrf), Status: pointer("reserved"), DnsName: pointer("lifecycle.example.test"), Description: &replacedDescription}})
				if err != nil {
					return "", err
				}
				return response.IpAddress.Description, nil
			},
			delete: func(ctx context.Context, id int64) error {
				_, err := environment.ipam.DeleteIPAddress(ctx, &ipamv1.DeleteIPAddressRequest{Id: id})
				return err
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			before, err := scenario.list(environment.ctx)
			require.NoError(t, err)
			id, err := scenario.create(environment.ctx)
			require.NoError(t, err)
			require.Greater(t, id, int64(0))

			afterCreate, err := scenario.list(environment.ctx)
			require.NoError(t, err)
			require.Equal(t, before.count+1, afterCreate.count)
			require.True(t, slices.Contains(afterCreate.ids, id), "gRPC List omitted newly created %s %d", scenario.name, id)
			restCreated := requestJSON(t, environment.router, http.MethodGet, scenario.path+"/"+strconv.FormatInt(id, 10), nil, http.StatusOK)
			require.Equal(t, createdDescription, restCreated["description"])

			description, err := scenario.replace(environment.ctx, id)
			require.NoError(t, err)
			require.Equal(t, replacedDescription, description)
			restReplaced := requestJSON(t, environment.router, http.MethodGet, scenario.path+"/"+strconv.FormatInt(id, 10), nil, http.StatusOK)
			require.Equal(t, replacedDescription, restReplaced["description"])

			require.NoError(t, scenario.delete(environment.ctx, id))
			requireRESTStatus(t, environment.router, http.MethodGet, scenario.path+"/"+strconv.FormatInt(id, 10)+"/", nil, http.StatusNotFound)
			afterDelete, err := scenario.list(environment.ctx)
			require.NoError(t, err)
			require.Equal(t, before.count, afterDelete.count)
			require.False(t, slices.Contains(afterDelete.ids, id))
		})
	}
}

func newProfileParityEnvironment(t *testing.T, authorizer authz.ResourceAuthorizer) *profileParityEnvironment {
	t.Helper()
	gin.SetMode(gin.TestMode)
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(dcimrow.Models()...))
	require.NoError(t, db.AutoMigrate(ipamrow.Models()...))
	require.NoError(t, db.AutoMigrate(&postgreschangelog.ChangeRow{}))
	core := composition.NewCoreWithAuthorizer(db, authorizer)
	principal := identity.Principal{ID: 1, Username: "admin", IsSuperuser: true}
	return &profileParityEnvironment{
		db: db, dcim: newParityDCIMServer(core), ipam: newParityIPAMServer(core),
		router: newParityRESTRouter(core, principal), principal: principal,
		ctx: identity.WithPrincipal(t.Context(), principal),
	}
}

func (environment *profileParityEnvironment) seedProfileFixtures(t *testing.T) profileFixtureIDs {
	t.Helper()
	create := func(path string, data map[string]any) int64 {
		return jsonID(t, requestJSON(t, environment.router, http.MethodPost, path, data, http.StatusCreated)["id"])
	}
	site := create("/api/dcim/sites", map[string]any{"name": "Fixture Site", "slug": "fixture-site"})
	manufacturer := create("/api/dcim/manufacturers", map[string]any{"name": "Fixture Manufacturer", "slug": "fixture-manufacturer"})
	rackRole := create("/api/dcim/rack-roles", map[string]any{"name": "Fixture Rack Role", "slug": "fixture-rack-role"})
	rackType := create("/api/dcim/rack-types", map[string]any{"manufacturer": manufacturer, "model": "Fixture Rack Type", "slug": "fixture-rack-type", "form_factor": "4-post-cabinet"})
	rack := create("/api/dcim/racks", map[string]any{"site": site, "name": "Fixture Rack", "rack_type": rackType, "role": rackRole})
	deviceRole := create("/api/dcim/device-roles", map[string]any{"name": "Fixture Device Role", "slug": "fixture-device-role"})
	deviceType := create("/api/dcim/device-types", map[string]any{"manufacturer": manufacturer, "model": "Fixture Device Type", "slug": "fixture-device-type"})
	device := create("/api/dcim/devices", map[string]any{"device_type": deviceType, "role": deviceRole, "site": site, "name": "fixture-device"})
	iface := create("/api/dcim/interfaces", map[string]any{"device": device, "name": "fixture0", "type": "1000base-t"})
	vrf := create("/api/ipam/vrfs", map[string]any{"name": "Fixture VRF", "rd": "65000:1"})
	return profileFixtureIDs{site, manufacturer, rackRole, rackType, rack, deviceRole, deviceType, device, iface, vrf}
}

func collectIDs[T any](values []T, id func(T) int64) []int64 {
	result := make([]int64, 0, len(values))
	for _, value := range values {
		result = append(result, id(value))
	}
	return result
}

func requireRESTStatus(t *testing.T, handler http.Handler, method, path string, body []byte, expected int) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(string(body)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	require.Equal(t, expected, response.Code, response.Body.String())
}
