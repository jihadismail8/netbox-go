package parity

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

type parityScenario struct {
	name       string
	path       string
	fields     []string
	references map[string]string // semantic key -> REST DTO field
	create     func(map[string]int64) map[string]any
	grpc       func(context.Context, int64, string) (map[string]any, map[string]any, error)
}

// TestCanonicalResourcesHaveRESTGRPCSemanticParity proves that the two public
// transports are adapters over one application state. Each of the 13 resources
// is created through REST, read and patched through gRPC, and read back through
// REST. The comparison includes canonical identity, URL/display, writable
// state, and relationship identity (not transport-specific JSON/protobuf shape).
func TestCanonicalResourcesHaveRESTGRPCSemanticParity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:all_resource_parity?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(dcimrow.Models()...))
	require.NoError(t, db.AutoMigrate(ipamrow.Models()...))
	require.NoError(t, db.AutoMigrate(&postgreschangelog.ChangeRow{}))

	core := composition.NewCoreWithAuthorizer(db, authz.AllowAll{})
	principal := identity.Principal{ID: 1, Username: "admin", IsSuperuser: true}
	ctx := identity.WithPrincipal(t.Context(), principal)
	dcim := newParityDCIMServer(core)
	ipam := newParityIPAMServer(core)
	router := newParityRESTRouter(core, principal)

	mask := func() *fieldmaskpb.FieldMask {
		return &fieldmaskpb.FieldMask{Paths: []string{"description"}}
	}
	base := func(id int64, url, display, description string) map[string]any {
		return map[string]any{"id": id, "url": url, "display": display, "description": description}
	}
	refID := func(reference interface{ GetId() int64 }) int64 {
		if reference == nil {
			return 0
		}
		return reference.GetId()
	}
	valueID := func(value *wrapperspb.Int64Value) any {
		if value == nil {
			return nil
		}
		return value.Value
	}
	valueString := func(value *wrapperspb.StringValue) any {
		if value == nil {
			return nil
		}
		return value.Value
	}

	scenarios := []parityScenario{
		{
			name: "site", path: "/api/dcim/sites",
			fields: []string{"name", "slug", "status", "facility", "description"},
			create: func(map[string]int64) map[string]any {
				return map[string]any{"name": "Parity Site", "slug": "parity-site", "status": "active", "facility": "MOW1", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetSite(ctx, &dcimv1.GetSiteRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.Site) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["name"], out["slug"], out["status"], out["facility"] = v.Name, v.Slug, v.Status, v.Facility
					return out
				}
				updated, err := dcim.UpdateSite(ctx, &dcimv1.UpdateSiteRequest{Id: id, Site: &dcimv1.SiteInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Site), semantic(updated.Site), nil
			},
		},
		{
			name: "manufacturer", path: "/api/dcim/manufacturers",
			fields: []string{"name", "slug", "description"},
			create: func(map[string]int64) map[string]any {
				return map[string]any{"name": "Parity Networks", "slug": "parity-networks", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetManufacturer(ctx, &dcimv1.GetManufacturerRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.Manufacturer) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["name"], out["slug"] = v.Name, v.Slug
					return out
				}
				updated, err := dcim.UpdateManufacturer(ctx, &dcimv1.UpdateManufacturerRequest{Id: id, Manufacturer: &dcimv1.ManufacturerInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Manufacturer), semantic(updated.Manufacturer), nil
			},
		},
		{
			name: "rack_role", path: "/api/dcim/rack-roles",
			fields: []string{"name", "slug", "color", "description"},
			create: func(map[string]int64) map[string]any {
				return map[string]any{"name": "Parity Production", "slug": "parity-production", "color": "a1b2c3", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetRackRole(ctx, &dcimv1.GetRackRoleRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.RackRole) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["name"], out["slug"], out["color"] = v.Name, v.Slug, v.Color
					return out
				}
				updated, err := dcim.UpdateRackRole(ctx, &dcimv1.UpdateRackRoleRequest{Id: id, RackRole: &dcimv1.RackRoleInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.RackRole), semantic(updated.RackRole), nil
			},
		},
		{
			name: "rack_type", path: "/api/dcim/rack-types",
			fields:     []string{"model", "slug", "form_factor", "width", "u_height", "starting_unit", "desc_units", "description"},
			references: map[string]string{"manufacturer_id": "manufacturer"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"manufacturer": ids["manufacturer"], "model": "Parity 48U", "slug": "parity-48u", "form_factor": "4-post-cabinet", "width": 19, "u_height": 48, "starting_unit": 1, "desc_units": false, "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetRackType(ctx, &dcimv1.GetRackTypeRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.RackType) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["manufacturer_id"], out["model"], out["slug"], out["form_factor"] = refID(v.Manufacturer), v.Model, v.Slug, v.FormFactor
					out["width"], out["u_height"], out["starting_unit"], out["desc_units"] = v.Width, v.UHeight, v.StartingUnit, v.DescUnits
					return out
				}
				updated, err := dcim.UpdateRackType(ctx, &dcimv1.UpdateRackTypeRequest{Id: id, RackType: &dcimv1.RackTypeInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.RackType), semantic(updated.RackType), nil
			},
		},
		{
			name: "rack", path: "/api/dcim/racks",
			fields:     []string{"name", "facility_id", "status", "form_factor", "width", "u_height", "starting_unit", "desc_units", "description"},
			references: map[string]string{"site_id": "site", "rack_type_id": "rack_type", "role_id": "role"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"site": ids["site"], "name": "R42", "facility_id": "MOW1-R42", "rack_type": ids["rack_type"], "role": ids["rack_role"], "status": "active", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetRack(ctx, &dcimv1.GetRackRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.Rack) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["site_id"], out["rack_type_id"], out["role_id"] = refID(v.Site), valueID(v.RackTypeId), valueID(v.RoleId)
					out["name"], out["facility_id"], out["status"], out["form_factor"] = v.Name, v.FacilityId, v.Status, v.FormFactor
					out["width"], out["u_height"], out["starting_unit"], out["desc_units"] = v.Width, v.UHeight, v.StartingUnit, v.DescUnits
					return out
				}
				updated, err := dcim.UpdateRack(ctx, &dcimv1.UpdateRackRequest{Id: id, Rack: &dcimv1.RackInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Rack), semantic(updated.Rack), nil
			},
		},
		{
			name: "device_role", path: "/api/dcim/device-roles",
			fields:     []string{"name", "slug", "color", "vm_role", "description"},
			references: map[string]string{"parent_id": "parent"},
			create: func(map[string]int64) map[string]any {
				return map[string]any{"parent": nil, "name": "Parity Router", "slug": "parity-router", "color": "112233", "vm_role": false, "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetDeviceRole(ctx, &dcimv1.GetDeviceRoleRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.DeviceRole) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["parent_id"], out["name"], out["slug"], out["color"], out["vm_role"] = valueID(v.ParentId), v.Name, v.Slug, v.Color, v.VmRole
					return out
				}
				updated, err := dcim.UpdateDeviceRole(ctx, &dcimv1.UpdateDeviceRoleRequest{Id: id, DeviceRole: &dcimv1.DeviceRoleInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.DeviceRole), semantic(updated.DeviceRole), nil
			},
		},
		{
			name: "device_type", path: "/api/dcim/device-types",
			fields:     []string{"model", "slug", "part_number", "u_height", "exclude_from_utilization", "is_full_depth", "airflow", "description"},
			references: map[string]string{"manufacturer_id": "manufacturer"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"manufacturer": ids["manufacturer"], "model": "Parity Router 1", "slug": "parity-router-1", "part_number": "PR-1", "u_height": 1.5, "exclude_from_utilization": false, "is_full_depth": true, "airflow": "front-to-rear", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetDeviceType(ctx, &dcimv1.GetDeviceTypeRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.DeviceType) map[string]any {
					height, _ := strconv.ParseFloat(v.UHeight, 64)
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["manufacturer_id"], out["model"], out["slug"], out["part_number"], out["u_height"] = refID(v.Manufacturer), v.Model, v.Slug, v.PartNumber, height
					out["exclude_from_utilization"], out["is_full_depth"], out["airflow"] = v.ExcludeFromUtilization, v.IsFullDepth, v.Airflow
					return out
				}
				updated, err := dcim.UpdateDeviceType(ctx, &dcimv1.UpdateDeviceTypeRequest{Id: id, DeviceType: &dcimv1.DeviceTypeInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.DeviceType), semantic(updated.DeviceType), nil
			},
		},
		{
			name: "interface_template", path: "/api/dcim/interface-templates",
			fields:     []string{"name", "label", "type", "enabled", "mgmt_only", "description"},
			references: map[string]string{"device_type_id": "device_type"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"device_type": ids["device_type"], "name": "xe-0/0/0", "label": "WAN template", "type": "1000base-x-sfp", "enabled": true, "mgmt_only": false, "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetInterfaceTemplate(ctx, &dcimv1.GetInterfaceTemplateRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.InterfaceTemplate) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["device_type_id"], out["name"], out["label"], out["type"], out["enabled"], out["mgmt_only"] = refID(v.DeviceType), v.Name, v.Label, v.Type, v.Enabled, v.MgmtOnly
					return out
				}
				updated, err := dcim.UpdateInterfaceTemplate(ctx, &dcimv1.UpdateInterfaceTemplateRequest{Id: id, InterfaceTemplate: &dcimv1.InterfaceTemplateInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.InterfaceTemplate), semantic(updated.InterfaceTemplate), nil
			},
		},
		{
			name: "device", path: "/api/dcim/devices",
			fields:     []string{"name", "position", "face", "status", "serial", "airflow", "description"},
			references: map[string]string{"device_type_id": "device_type", "role_id": "role", "site_id": "site", "rack_id": "rack"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"device_type": ids["device_type"], "role": ids["device_role"], "name": "parity-rtr-01", "site": ids["site"], "rack": ids["rack"], "position": 10, "face": "front", "status": "active", "serial": "PARITY001", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetDevice(ctx, &dcimv1.GetDeviceRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.Device) map[string]any {
					position, _ := strconv.ParseFloat(v.Position, 64)
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["device_type_id"], out["role_id"], out["site_id"], out["rack_id"] = refID(v.DeviceType), refID(v.Role), refID(v.Site), valueID(v.RackId)
					out["name"], out["position"], out["face"], out["status"], out["serial"], out["airflow"] = v.Name, position, v.Face, v.Status, v.Serial, v.Airflow
					return out
				}
				updated, err := dcim.UpdateDevice(ctx, &dcimv1.UpdateDeviceRequest{Id: id, Device: &dcimv1.DeviceInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Device), semantic(updated.Device), nil
			},
		},
		{
			name: "interface", path: "/api/dcim/interfaces",
			fields:     []string{"name", "label", "type", "enabled", "mgmt_only", "mtu", "speed", "duplex", "description"},
			references: map[string]string{"device_id": "device"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"device": ids["device"], "name": "mgmt0", "label": "Management", "type": "1000base-t", "enabled": true, "mgmt_only": true, "mtu": 1500, "speed": 1000000, "duplex": "full", "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := dcim.GetInterface(ctx, &dcimv1.GetInterfaceRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *dcimv1.Interface) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["device_id"], out["name"], out["label"], out["type"], out["enabled"], out["mgmt_only"] = refID(v.Device), v.Name, v.Label, v.Type, v.Enabled, v.MgmtOnly
					out["mtu"], out["speed"], out["duplex"] = valueID32(v.Mtu), valueID(v.Speed), valueString(v.Duplex)
					return out
				}
				updated, err := dcim.UpdateInterface(ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Interface), semantic(updated.Interface), nil
			},
		},
		{
			name: "vrf", path: "/api/ipam/vrfs",
			fields: []string{"name", "rd", "enforce_unique", "description"},
			create: func(map[string]int64) map[string]any {
				return map[string]any{"name": "Parity VRF", "rd": "65000:42", "enforce_unique": true, "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := ipam.GetVRF(ctx, &ipamv1.GetVRFRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *ipamv1.VRF) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["name"], out["rd"], out["enforce_unique"] = v.Name, valueString(v.Rd), v.EnforceUnique
					return out
				}
				updated, err := ipam.UpdateVRF(ctx, &ipamv1.UpdateVRFRequest{Id: id, Vrf: &ipamv1.VRFInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Vrf), semantic(updated.Vrf), nil
			},
		},
		{
			name: "prefix", path: "/api/ipam/prefixes",
			fields:     []string{"prefix", "status", "is_pool", "mark_utilized", "description", "family"},
			references: map[string]string{"vrf_id": "vrf"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"prefix": "10.42.0.0/24", "vrf": ids["vrf"], "status": "active", "is_pool": false, "mark_utilized": false, "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := ipam.GetPrefix(ctx, &ipamv1.GetPrefixRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *ipamv1.Prefix) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["vrf_id"], out["prefix"], out["status"], out["is_pool"], out["mark_utilized"], out["family"] = valueID(v.VrfId), v.Prefix, v.Status, v.IsPool, v.MarkUtilized, v.Family
					return out
				}
				updated, err := ipam.UpdatePrefix(ctx, &ipamv1.UpdatePrefixRequest{Id: id, Prefix: &ipamv1.PrefixInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.Prefix), semantic(updated.Prefix), nil
			},
		},
		{
			name: "ip_address", path: "/api/ipam/ip-addresses",
			fields:     []string{"address", "status", "dns_name", "assigned_object_type", "assigned_object_id", "description", "family"},
			references: map[string]string{"vrf_id": "vrf"},
			create: func(ids map[string]int64) map[string]any {
				return map[string]any{"address": "10.42.0.10/24", "vrf": ids["vrf"], "status": "active", "dns_name": "RTR01.Example.Test.", "assigned_object_type": "dcim.interface", "assigned_object_id": ids["interface"], "description": "created through REST"}
			},
			grpc: func(ctx context.Context, id int64, description string) (map[string]any, map[string]any, error) {
				got, err := ipam.GetIPAddress(ctx, &ipamv1.GetIPAddressRequest{Id: id})
				if err != nil {
					return nil, nil, err
				}
				semantic := func(v *ipamv1.IPAddress) map[string]any {
					out := base(v.Id, v.Url, v.Display, v.Description)
					out["vrf_id"], out["address"], out["status"], out["dns_name"], out["assigned_object_type"], out["assigned_object_id"], out["family"] = valueID(v.VrfId), v.Address, v.Status, v.DnsName, valueString(v.AssignedObjectType), valueID(v.AssignedObjectId), v.Family
					return out
				}
				updated, err := ipam.UpdateIPAddress(ctx, &ipamv1.UpdateIPAddressRequest{Id: id, IpAddress: &ipamv1.IPAddressInput{Description: pointer(description)}, UpdateMask: mask()})
				if err != nil {
					return nil, nil, err
				}
				return semantic(got.IpAddress), semantic(updated.IpAddress), nil
			},
		},
	}

	ids := map[string]int64{}
	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			created := requestJSON(t, router, http.MethodPost, scenario.path, scenario.create(ids), http.StatusCreated)
			id := jsonID(t, created["id"])
			ids[scenario.name] = id
			createdSemantic := restSemantics(t, created, scenario)

			description := "updated through gRPC: " + scenario.name
			before, after, err := scenario.grpc(ctx, id, description)
			require.NoError(t, err)
			require.Equal(t, canonicalMap(createdSemantic), canonicalMap(before), "REST create and gRPC get diverged")
			require.Equal(t, description, after["description"])

			readBack := requestJSON(t, router, http.MethodGet, scenario.path+"/"+strconv.FormatInt(id, 10), nil, http.StatusOK)
			require.Equal(t, canonicalMap(after), canonicalMap(restSemantics(t, readBack, scenario)), "gRPC patch and REST get diverged")
		})
	}
}

func requestJSON(t *testing.T, router http.Handler, method, path string, body map[string]any, wantStatus int) map[string]any {
	t.Helper()
	path = strings.TrimRight(path, "/") + "/"
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		require.NoError(t, err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(encoded))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, wantStatus, response.Code, response.Body.String())
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &decoded), response.Body.String())
	return decoded
}

func restSemantics(t *testing.T, body map[string]any, scenario parityScenario) map[string]any {
	t.Helper()
	out := map[string]any{}
	for _, field := range append([]string{"id", "url", "display"}, scenario.fields...) {
		value, ok := body[field]
		require.True(t, ok, "REST response omitted %q: %#v", field, body)
		if envelope, choice := value.(map[string]any); choice && len(envelope) == 2 {
			if choiceValue, hasValue := envelope["value"]; hasValue {
				if _, hasLabel := envelope["label"]; hasLabel {
					value = choiceValue
				}
			}
		}
		out[field] = value
	}
	for semanticKey, field := range scenario.references {
		value, ok := body[field]
		require.True(t, ok, "REST response omitted relationship %q: %#v", field, body)
		if value == nil {
			out[semanticKey] = nil
			continue
		}
		reference, ok := value.(map[string]any)
		require.True(t, ok, "REST relationship %q was not an object: %#v", field, value)
		out[semanticKey] = reference["id"]
	}
	return out
}

func canonicalMap(input map[string]any) map[string]any {
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = canonicalScalar(value)
	}
	return out
}

func canonicalScalar(value any) any {
	switch number := value.(type) {
	case float64:
		if math.Trunc(number) == number {
			return int64(number)
		}
		return number
	case float32:
		return float64(number)
	case int:
		return int64(number)
	case int32:
		return int64(number)
	case uint32:
		return int64(number)
	case uint64:
		if number <= math.MaxInt64 {
			return int64(number)
		}
		return strconv.FormatUint(number, 10)
	case string:
		return strings.TrimSpace(number)
	default:
		return value
	}
}

func jsonID(t *testing.T, value any) int64 {
	t.Helper()
	number, ok := value.(float64)
	require.True(t, ok, "JSON ID is not numeric: %T", value)
	require.Greater(t, number, float64(0))
	return int64(number)
}

func pointer[T any](value T) *T { return &value }

func valueID32(value *wrapperspb.Int32Value) any {
	if value == nil {
		return nil
	}
	return int64(value.Value)
}
