package parity

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

type grpcErrorScenario struct {
	name        string
	invalid     func() error
	notFound    func() error
	deniedView  func() error
	deniedWrite func() error
}

// TestCanonicalGRPCErrorsAllProfileResources proves every canonical resource
// reaches the shared validation, lookup, and authorization taxonomy. This is
// deliberately adapter-level: assertions are on public gRPC status codes.
func TestCanonicalGRPCErrorsAllProfileResources(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	permissionCore := composition.NewCoreWithAuthorizer(environment.db, authz.PermissionAuthorizer{})
	deniedDCIM := newParityDCIMServer(permissionCore)
	deniedIPAM := newParityIPAMServer(permissionCore)
	deniedContext := identity.WithPrincipal(t.Context(), identity.Principal{ID: 2, Username: "denied"})
	missingID := int64(987654321)

	scenarios := []grpcErrorScenario{
		{
			name: "site",
			invalid: func() error {
				_, err := environment.dcim.CreateSite(environment.ctx, &dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetSite(environment.ctx, &dcimv1.GetSiteRequest{Id: missingID})
				return err
			},
			deniedView: func() error { _, err := deniedDCIM.ListSites(deniedContext, &dcimv1.ListSitesRequest{}); return err },
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateSite(deniedContext, &dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{}})
				return err
			},
		},
		{
			name: "manufacturer",
			invalid: func() error {
				_, err := environment.dcim.CreateManufacturer(environment.ctx, &dcimv1.CreateManufacturerRequest{Manufacturer: &dcimv1.ManufacturerInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetManufacturer(environment.ctx, &dcimv1.GetManufacturerRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListManufacturers(deniedContext, &dcimv1.ListManufacturersRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateManufacturer(deniedContext, &dcimv1.CreateManufacturerRequest{Manufacturer: &dcimv1.ManufacturerInput{}})
				return err
			},
		},
		{
			name: "rack_role",
			invalid: func() error {
				_, err := environment.dcim.CreateRackRole(environment.ctx, &dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetRackRole(environment.ctx, &dcimv1.GetRackRoleRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListRackRoles(deniedContext, &dcimv1.ListRackRolesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateRackRole(deniedContext, &dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{}})
				return err
			},
		},
		{
			name: "rack_type",
			invalid: func() error {
				_, err := environment.dcim.CreateRackType(environment.ctx, &dcimv1.CreateRackTypeRequest{RackType: &dcimv1.RackTypeInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetRackType(environment.ctx, &dcimv1.GetRackTypeRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListRackTypes(deniedContext, &dcimv1.ListRackTypesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateRackType(deniedContext, &dcimv1.CreateRackTypeRequest{RackType: &dcimv1.RackTypeInput{}})
				return err
			},
		},
		{
			name: "rack",
			invalid: func() error {
				_, err := environment.dcim.CreateRack(environment.ctx, &dcimv1.CreateRackRequest{Rack: &dcimv1.RackInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetRack(environment.ctx, &dcimv1.GetRackRequest{Id: missingID})
				return err
			},
			deniedView: func() error { _, err := deniedDCIM.ListRacks(deniedContext, &dcimv1.ListRacksRequest{}); return err },
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateRack(deniedContext, &dcimv1.CreateRackRequest{Rack: &dcimv1.RackInput{}})
				return err
			},
		},
		{
			name: "device_role",
			invalid: func() error {
				_, err := environment.dcim.CreateDeviceRole(environment.ctx, &dcimv1.CreateDeviceRoleRequest{DeviceRole: &dcimv1.DeviceRoleInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetDeviceRole(environment.ctx, &dcimv1.GetDeviceRoleRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListDeviceRoles(deniedContext, &dcimv1.ListDeviceRolesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateDeviceRole(deniedContext, &dcimv1.CreateDeviceRoleRequest{DeviceRole: &dcimv1.DeviceRoleInput{}})
				return err
			},
		},
		{
			name: "device_type",
			invalid: func() error {
				_, err := environment.dcim.CreateDeviceType(environment.ctx, &dcimv1.CreateDeviceTypeRequest{DeviceType: &dcimv1.DeviceTypeInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetDeviceType(environment.ctx, &dcimv1.GetDeviceTypeRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListDeviceTypes(deniedContext, &dcimv1.ListDeviceTypesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateDeviceType(deniedContext, &dcimv1.CreateDeviceTypeRequest{DeviceType: &dcimv1.DeviceTypeInput{}})
				return err
			},
		},
		{
			name: "interface_template",
			invalid: func() error {
				_, err := environment.dcim.CreateInterfaceTemplate(environment.ctx, &dcimv1.CreateInterfaceTemplateRequest{InterfaceTemplate: &dcimv1.InterfaceTemplateInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetInterfaceTemplate(environment.ctx, &dcimv1.GetInterfaceTemplateRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListInterfaceTemplates(deniedContext, &dcimv1.ListInterfaceTemplatesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateInterfaceTemplate(deniedContext, &dcimv1.CreateInterfaceTemplateRequest{InterfaceTemplate: &dcimv1.InterfaceTemplateInput{}})
				return err
			},
		},
		{
			name: "device",
			invalid: func() error {
				_, err := environment.dcim.CreateDevice(environment.ctx, &dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetDevice(environment.ctx, &dcimv1.GetDeviceRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListDevices(deniedContext, &dcimv1.ListDevicesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateDevice(deniedContext, &dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{}})
				return err
			},
		},
		{
			name: "interface",
			invalid: func() error {
				_, err := environment.dcim.CreateInterface(environment.ctx, &dcimv1.CreateInterfaceRequest{Interface: &dcimv1.InterfaceInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.dcim.GetInterface(environment.ctx, &dcimv1.GetInterfaceRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedDCIM.ListInterfaces(deniedContext, &dcimv1.ListInterfacesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedDCIM.CreateInterface(deniedContext, &dcimv1.CreateInterfaceRequest{Interface: &dcimv1.InterfaceInput{}})
				return err
			},
		},
		{
			name: "vrf",
			invalid: func() error {
				_, err := environment.ipam.CreateVRF(environment.ctx, &ipamv1.CreateVRFRequest{Vrf: &ipamv1.VRFInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.ipam.GetVRF(environment.ctx, &ipamv1.GetVRFRequest{Id: missingID})
				return err
			},
			deniedView: func() error { _, err := deniedIPAM.ListVRFs(deniedContext, &ipamv1.ListVRFsRequest{}); return err },
			deniedWrite: func() error {
				_, err := deniedIPAM.CreateVRF(deniedContext, &ipamv1.CreateVRFRequest{Vrf: &ipamv1.VRFInput{}})
				return err
			},
		},
		{
			name: "prefix",
			invalid: func() error {
				_, err := environment.ipam.CreatePrefix(environment.ctx, &ipamv1.CreatePrefixRequest{Prefix: &ipamv1.PrefixInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.ipam.GetPrefix(environment.ctx, &ipamv1.GetPrefixRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedIPAM.ListPrefixes(deniedContext, &ipamv1.ListPrefixesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedIPAM.CreatePrefix(deniedContext, &ipamv1.CreatePrefixRequest{Prefix: &ipamv1.PrefixInput{}})
				return err
			},
		},
		{
			name: "ip_address",
			invalid: func() error {
				_, err := environment.ipam.CreateIPAddress(environment.ctx, &ipamv1.CreateIPAddressRequest{IpAddress: &ipamv1.IPAddressInput{}})
				return err
			},
			notFound: func() error {
				_, err := environment.ipam.GetIPAddress(environment.ctx, &ipamv1.GetIPAddressRequest{Id: missingID})
				return err
			},
			deniedView: func() error {
				_, err := deniedIPAM.ListIPAddresses(deniedContext, &ipamv1.ListIPAddressesRequest{})
				return err
			},
			deniedWrite: func() error {
				_, err := deniedIPAM.CreateIPAddress(deniedContext, &ipamv1.CreateIPAddressRequest{IpAddress: &ipamv1.IPAddressInput{}})
				return err
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			require.Equal(t, codes.InvalidArgument, status.Code(scenario.invalid()))
			require.Equal(t, codes.NotFound, status.Code(scenario.notFound()))
			require.Equal(t, codes.PermissionDenied, status.Code(scenario.deniedView()))
			require.Equal(t, codes.PermissionDenied, status.Code(scenario.deniedWrite()))
		})
	}

	_, err := environment.dcim.ListSites(t.Context(), &dcimv1.ListSitesRequest{})
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestGRPCDeviceCreationRollbackIsVisibleToREST(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	fixtures := environment.seedProfileFixtures(t)
	first := requestJSON(t, environment.router, http.MethodPost, "/api/dcim/interface-templates", map[string]any{
		"device_type": fixtures.deviceType,
		"name":        "duplicate",
		"type":        "other",
	}, http.StatusCreated)
	second := requestJSON(t, environment.router, http.MethodPost, "/api/dcim/interface-templates", map[string]any{
		"device_type": fixtures.deviceType,
		"name":        "unique",
		"type":        "other",
	}, http.StatusCreated)
	require.NoError(t, environment.db.Migrator().DropIndex(&dcimrow.InterfaceTemplateRow{}, "uq_go_interface_template_name"))
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).
		Where("id = ?", jsonID(t, second["id"])).
		Update("name", first["name"]).Error)

	var devicesBefore, interfacesBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&devicesBefore).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&interfacesBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)

	_, err := environment.dcim.CreateDevice(environment.ctx, &dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{
		DeviceType: &fixtures.deviceType,
		Role:       &fixtures.deviceRole,
		Name:       wrapperspb.String("rollback-device"),
		Site:       &fixtures.site,
	}})
	require.Equal(t, codes.AlreadyExists, status.Code(err))

	var devicesAfter, interfacesAfter, changesAfter int64
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&devicesAfter).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&interfacesAfter).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	require.Equal(t, devicesBefore, devicesAfter)
	require.Equal(t, interfacesBefore, interfacesAfter)
	require.Equal(t, changesBefore, changesAfter)

	response := requestJSON(t, environment.router, http.MethodGet, "/api/dcim/devices", nil, http.StatusOK)
	require.Equal(t, float64(devicesBefore), response["count"])
}

func TestIPAddressAssignUnassignSharedStateAcrossRESTAndGRPC(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	fixtures := environment.seedProfileFixtures(t)
	created := requestJSON(t, environment.router, http.MethodPost, "/api/ipam/ip-addresses", map[string]any{
		"address": "192.0.2.25/24",
		"vrf":     fixtures.vrf,
	}, http.StatusCreated)
	addressID := jsonID(t, created["id"])
	require.Nil(t, created["assigned_object_id"])

	assigned, err := environment.ipam.AssignIPAddress(environment.ctx, &ipamv1.AssignIPAddressRequest{Id: addressID, InterfaceId: fixtures.iface})
	require.NoError(t, err)
	require.Equal(t, fixtures.iface, assigned.IpAddress.AssignedObjectId.Value)
	require.Equal(t, "dcim.interface", assigned.IpAddress.AssignedObjectType.Value)

	restAssigned := requestJSON(t, environment.router, http.MethodGet, "/api/ipam/ip-addresses/"+strconv.FormatInt(addressID, 10), nil, http.StatusOK)
	require.Equal(t, float64(fixtures.iface), restAssigned["assigned_object_id"])
	reference, ok := restAssigned["assigned_object"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(fixtures.iface), reference["id"])
	interfaceState := requestJSON(t, environment.router, http.MethodGet, "/api/dcim/interfaces/"+strconv.FormatInt(fixtures.iface, 10), nil, http.StatusOK)
	require.Equal(t, float64(1), interfaceState["count_ipaddresses"])

	requestJSON(t, environment.router, http.MethodPatch, "/api/ipam/ip-addresses/"+strconv.FormatInt(addressID, 10), map[string]any{
		"assigned_object_type": nil,
		"assigned_object_id":   nil,
	}, http.StatusOK)
	grpcUnassigned, err := environment.ipam.GetIPAddress(environment.ctx, &ipamv1.GetIPAddressRequest{Id: addressID})
	require.NoError(t, err)
	require.Nil(t, grpcUnassigned.IpAddress.AssignedObjectType)
	require.Nil(t, grpcUnassigned.IpAddress.AssignedObjectId)

	_, err = environment.ipam.AssignIPAddress(environment.ctx, &ipamv1.AssignIPAddressRequest{Id: addressID, InterfaceId: fixtures.iface})
	require.NoError(t, err)
	unassigned, err := environment.ipam.UnassignIPAddress(environment.ctx, &ipamv1.UnassignIPAddressRequest{Id: addressID})
	require.NoError(t, err)
	require.Nil(t, unassigned.IpAddress.AssignedObjectType)
	require.Nil(t, unassigned.IpAddress.AssignedObjectId)
	restUnassigned := requestJSON(t, environment.router, http.MethodGet, "/api/ipam/ip-addresses/"+strconv.FormatInt(addressID, 10), nil, http.StatusOK)
	require.Nil(t, restUnassigned["assigned_object_type"])
	require.Nil(t, restUnassigned["assigned_object_id"])
	interfaceState = requestJSON(t, environment.router, http.MethodGet, "/api/dcim/interfaces/"+strconv.FormatInt(fixtures.iface, 10), nil, http.StatusOK)
	require.Equal(t, float64(0), interfaceState["count_ipaddresses"])

	_, err = environment.ipam.AssignIPAddress(environment.ctx, &ipamv1.AssignIPAddressRequest{Id: addressID, InterfaceId: 999999})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	stillUnassigned, err := environment.ipam.GetIPAddress(environment.ctx, &ipamv1.GetIPAddressRequest{Id: addressID})
	require.NoError(t, err)
	require.Nil(t, stillUnassigned.IpAddress.AssignedObjectId, "failed assignment must roll back without partial state")
}
