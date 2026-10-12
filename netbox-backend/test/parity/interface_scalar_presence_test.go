package parity

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	grpcadapter "netbox-go/internal/adapters/grpc/workflow"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	restadapter "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
	"netbox-go/internal/platform/composition"
)

func TestInterfaceScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	deviceID := seedParityInterfacePresenceDevice(t, environment)
	core := composition.NewCoreWithAuthorizer(environment.db, authz.AllowAll{})
	observer := &interfacePresenceErrorObserver{InterfaceService: core.Interfaces}
	router := gin.New()
	restadapter.NewInterfaceRESTHandler(observer).Register(router, func(c *gin.Context) {
		restadapter.SetPrincipal(c, environment.principal)
		c.Next()
	})
	rpc := grpcadapter.NewInterfaceRPCHandler(observer)
	before := loadParityInterfacePresenceState(t, environment)
	require.Empty(t, before.interfaces)
	require.Empty(t, before.changes)

	created := requestJSON(t, router, http.MethodPost, "/api/dcim/interfaces", map[string]any{
		"device": deviceID, "name": "  REST default  ", "type": "1000base-t",
	}, http.StatusCreated)
	id := jsonID(t, created["id"])
	itemPath := "/api/dcim/interfaces/" + strconv.FormatInt(id, 10)
	state := requireInterfaceParityMutation(t, environment, before, id, true)
	row := parityInterfaceRow(t, state, id)
	require.Equal(t, "REST default", row.Name)
	require.Empty(t, row.Label)
	require.True(t, row.Enabled)
	require.False(t, row.MgmtOnly)
	require.Nil(t, row.MTU)
	require.Nil(t, row.Speed)
	require.Nil(t, row.Duplex)
	require.Empty(t, row.Description)
	read, err := rpc.GetInterface(environment.ctx, &dcimv1.GetInterfaceRequest{Id: id})
	require.NoError(t, err)
	requireInterfaceParityProjection(t, created, read.Interface, row)

	grpcCreated, err := rpc.CreateInterface(environment.ctx, &dcimv1.CreateInterfaceRequest{Interface: &dcimv1.InterfaceInput{
		Device: &deviceID, Name: pointer("  gRPC concrete  "), Label: pointer("  LAN  "), Type: pointer("10gbase-sr"),
		Enabled: pointer(false), MgmtOnly: pointer(true), Mtu: wrapperspb.Int32(1), Speed: wrapperspb.Int64(2147483647),
		Duplex: wrapperspb.String("full"), Description: pointer("  gRPC description  "),
	}})
	require.NoError(t, err)
	state = requireInterfaceParityMutation(t, environment, state, grpcCreated.Interface.Id, true)
	grpcRow := parityInterfaceRow(t, state, grpcCreated.Interface.Id)
	require.Equal(t, "gRPC concrete", grpcRow.Name)
	require.Equal(t, "LAN", grpcRow.Label)
	require.False(t, grpcRow.Enabled)
	require.True(t, grpcRow.MgmtOnly)
	require.Equal(t, int64(1), *grpcRow.MTU)
	require.Equal(t, int64(2147483647), *grpcRow.Speed)
	require.Equal(t, "full", *grpcRow.Duplex)
	require.Equal(t, "gRPC description", grpcRow.Description)
	restRead := requestJSON(t, router, http.MethodGet, "/api/dcim/interfaces/"+strconv.FormatInt(grpcRow.ID, 10), nil, http.StatusOK)
	requireInterfaceParityProjection(t, restRead, grpcCreated.Interface, grpcRow)

	patched, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{
		Device: &deviceID, Name: pointer("  updated  "), Label: pointer("  uplink  "), Type: pointer("10gbase-sr"),
		Enabled: pointer(false), MgmtOnly: pointer(true), Mtu: wrapperspb.Int32(65536), Speed: wrapperspb.Int64(0),
		Duplex: wrapperspb.String(""), Description: pointer("  updated description  "),
	}, UpdateMask: &fieldmaskpb.FieldMask{Paths: interfacePresenceParityFields()}})
	require.NoError(t, err)
	state = requireInterfaceParityMutation(t, environment, state, id, false)
	row = parityInterfaceRow(t, state, id)
	require.Equal(t, "updated", row.Name)
	require.Equal(t, "uplink", row.Label)
	require.False(t, row.Enabled)
	require.True(t, row.MgmtOnly)
	require.Equal(t, int64(65536), *row.MTU)
	require.Equal(t, int64(0), *row.Speed)
	require.Equal(t, "", *row.Duplex, "stored blank is distinct from null")
	require.Equal(t, "updated description", row.Description)
	restRead = requestJSON(t, router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireInterfaceParityProjection(t, restRead, patched.Interface, row)
	require.Nil(t, restRead["duplex"], "REST cannot distinguish stored blank from null")
	require.NotNil(t, patched.Interface.Duplex)

	for _, transport := range []string{"REST", "gRPC"} {
		t.Run("PUT preserves every optional omission/"+transport, func(t *testing.T) {
			before := loadParityInterfacePresenceState(t, environment)
			prior := parityInterfaceRow(t, before, id)
			if transport == "REST" {
				requestJSON(t, router, http.MethodPut, itemPath, map[string]any{"device": deviceID, "name": "REST PUT", "type": "25gbase-sr"}, http.StatusOK)
			} else {
				_, err := rpc.ReplaceInterface(environment.ctx, &dcimv1.ReplaceInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{
					Device: &deviceID, Name: pointer("gRPC PUT"), Type: pointer("25gbase-sr"),
				}})
				require.NoError(t, err)
			}
			after := requireInterfaceParityMutation(t, environment, before, id, false)
			current := parityInterfaceRow(t, after, id)
			require.Equal(t, transport+" PUT", current.Name)
			require.Equal(t, "25gbase-sr", current.Type)
			prior.Name, prior.Type, prior.LastUpdated = current.Name, current.Type, current.LastUpdated
			require.Equal(t, prior, current)
			grpcRead, err := rpc.GetInterface(environment.ctx, &dcimv1.GetInterfaceRequest{Id: id})
			require.NoError(t, err)
			requireInterfaceParityProjection(t, requestJSON(t, router, http.MethodGet, itemPath, nil, http.StatusOK), grpcRead.Interface, current)
		})
	}

	for _, transport := range []string{"REST", "gRPC"} {
		t.Run("nullable PATCH clearing and false zero blank/"+transport, func(t *testing.T) {
			before := loadParityInterfacePresenceState(t, environment)
			if transport == "REST" {
				requestJSON(t, router, http.MethodPatch, itemPath, map[string]any{"mtu": nil, "speed": nil, "duplex": nil}, http.StatusOK)
			} else {
				_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"mtu", "speed", "duplex"}}})
				require.NoError(t, err)
			}
			after := requireInterfaceParityMutation(t, environment, before, id, false)
			current := parityInterfaceRow(t, after, id)
			require.Nil(t, current.MTU)
			require.Nil(t, current.Speed)
			require.Nil(t, current.Duplex)
			grpcRead, err := rpc.GetInterface(environment.ctx, &dcimv1.GetInterfaceRequest{Id: id})
			require.NoError(t, err)
			requireInterfaceParityProjection(t, requestJSON(t, router, http.MethodGet, itemPath, nil, http.StatusOK), grpcRead.Interface, current)
			before = after
			if transport == "REST" {
				requestJSON(t, router, http.MethodPatch, itemPath, map[string]any{"label": "", "enabled": false, "mgmt_only": false, "mtu": 1, "speed": 0, "duplex": "", "description": ""}, http.StatusOK)
			} else {
				_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{
					Label: pointer(""), Enabled: pointer(false), MgmtOnly: pointer(false), Mtu: wrapperspb.Int32(1), Speed: wrapperspb.Int64(0), Duplex: wrapperspb.String(""), Description: pointer(""),
				}})
				require.NoError(t, err)
			}
			after = requireInterfaceParityMutation(t, environment, before, id, false)
			current = parityInterfaceRow(t, after, id)
			require.False(t, current.Enabled)
			require.False(t, current.MgmtOnly)
			require.Equal(t, int64(1), *current.MTU)
			require.Equal(t, int64(0), *current.Speed)
			require.Equal(t, "", *current.Duplex)
			require.Empty(t, current.Label)
			require.Empty(t, current.Description)
		})
	}

	for _, field := range []string{"device", "name", "label", "type", "enabled", "mgmt_only", "description"} {
		t.Run("nonnullable null shared rejection/"+field, func(t *testing.T) {
			before := loadParityInterfacePresenceState(t, environment)
			response := requestJSON(t, router, http.MethodPatch, itemPath, map[string]any{field: nil}, http.StatusBadRequest)
			restViolations := observer.violations()
			require.Len(t, restViolations, 1)
			require.Equal(t, field, restViolations[0].Field)
			if field == "type" {
				require.Equal(t, "blank", restViolations[0].Reason)
			}
			require.Equal(t, []any{restViolations[0].Description}, response[field])
			_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}}})
			requireInterfacePresenceRPCInvalid(t, err)
			require.Equal(t, restViolations, observer.violations())
			require.Equal(t, before, loadParityInterfacePresenceState(t, environment))
			for _, method := range []string{http.MethodPost, http.MethodPut} {
				body := map[string]any{"device": deviceID, "name": "rejected", "type": "1000base-t"}
				body[field] = nil
				path := itemPath
				if method == http.MethodPost {
					path = "/api/dcim/interfaces"
				}
				requestJSON(t, router, method, path, body, http.StatusBadRequest)
				require.Equal(t, before, loadParityInterfacePresenceState(t, environment))
			}
		})
	}

	for _, test := range []struct {
		name  string
		field string
		value any
		input *dcimv1.InterfaceInput
		text  string
	}{
		{"negative MTU", "mtu", -1, &dcimv1.InterfaceInput{Mtu: wrapperspb.Int32(-1)}, "Ensure this value is greater than or equal to 1."},
		{"zero MTU", "mtu", 0, &dcimv1.InterfaceInput{Mtu: wrapperspb.Int32(0)}, "Ensure this value is greater than or equal to 1."},
		{"high MTU", "mtu", 65537, &dcimv1.InterfaceInput{Mtu: wrapperspb.Int32(65537)}, "Ensure this value is less than or equal to 65536."},
		{"negative speed", "speed", -1, &dcimv1.InterfaceInput{Speed: wrapperspb.Int64(-1)}, "Ensure this value is greater than or equal to 0."},
		{"high speed", "speed", int64(2147483648), &dcimv1.InterfaceInput{Speed: wrapperspb.Int64(2147483648)}, "Ensure this value is less than or equal to 2147483647."},
		{"type whitespace", "type", " 1000base-t ", &dcimv1.InterfaceInput{Type: pointer(" 1000base-t ")}, ""},
		{"duplex whitespace", "duplex", " full ", &dcimv1.InterfaceInput{Duplex: wrapperspb.String(" full ")}, ""},
		{"unknown owner", "device", int64(999999), &dcimv1.InterfaceInput{Device: pointer(int64(999999))}, "Select a valid choice."},
	} {
		t.Run("equivalent shared field error/"+test.name, func(t *testing.T) {
			before := loadParityInterfacePresenceState(t, environment)
			response := requestJSON(t, router, http.MethodPatch, itemPath, map[string]any{test.field: test.value}, http.StatusBadRequest)
			restViolations := observer.violations()
			require.Len(t, restViolations, 1)
			require.Equal(t, test.field, restViolations[0].Field)
			if test.text != "" {
				require.Equal(t, test.text, restViolations[0].Description)
			}
			require.Equal(t, []any{restViolations[0].Description}, response[test.field])
			_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: test.input, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{test.field}}})
			requireInterfacePresenceRPCInvalid(t, err)
			require.Equal(t, restViolations, observer.violations())
			require.Equal(t, before, loadParityInterfacePresenceState(t, environment))
		})
	}
	t.Run("required omissions and deterministic aggregate errors", func(t *testing.T) {
		before := loadParityInterfacePresenceState(t, environment)
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			path := itemPath
			if method == http.MethodPost {
				path = "/api/dcim/interfaces"
			}
			requestJSON(t, router, method, path, map[string]any{}, http.StatusBadRequest)
			restViolations := observer.violations()
			var err error
			if method == http.MethodPost {
				_, err = rpc.CreateInterface(environment.ctx, &dcimv1.CreateInterfaceRequest{})
			} else {
				_, err = rpc.ReplaceInterface(environment.ctx, &dcimv1.ReplaceInterfaceRequest{Id: id})
			}
			requireInterfacePresenceRPCInvalid(t, err)
			require.Equal(t, restViolations, observer.violations())
		}
		body := map[string]any{"device": 999999, "name": " ", "enabled": nil, "mtu": -1, "speed": 2147483648, "duplex": " full ", "description": strings.Repeat("x", 201)}
		response := requestJSON(t, router, http.MethodPatch, itemPath, body, http.StatusBadRequest)
		restViolations := observer.violations()
		var fields []string
		for _, violation := range restViolations {
			fields = append(fields, violation.Field)
			require.Equal(t, []any{violation.Description}, response[violation.Field])
		}
		require.Equal(t, []string{"device", "name", "enabled", "mtu", "speed", "duplex", "description"}, fields)
		_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{
			Device: pointer(int64(999999)), Name: pointer(" "), Mtu: wrapperspb.Int32(-1), Speed: wrapperspb.Int64(2147483648), Duplex: wrapperspb.String(" full "), Description: pointer(strings.Repeat("x", 201)),
		}, UpdateMask: &fieldmaskpb.FieldMask{Paths: fields}})
		requireInterfacePresenceRPCInvalid(t, err)
		require.Equal(t, restViolations, observer.violations())
		require.Equal(t, before, loadParityInterfacePresenceState(t, environment))
	})
	t.Run("undeclared fields and masks fail before mutation", func(t *testing.T) {
		before := loadParityInterfacePresenceState(t, environment)
		requestJSON(t, router, http.MethodPatch, itemPath, map[string]any{"bridge": 7}, http.StatusBadRequest)
		_, err := rpc.UpdateInterface(environment.ctx, &dcimv1.UpdateInterfaceRequest{Id: id, Interface: &dcimv1.InterfaceInput{Description: pointer("must not persist")}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"bridge"}}})
		requireInterfacePresenceRPCInvalid(t, err)
		require.Equal(t, before, loadParityInterfacePresenceState(t, environment))
	})
}

func seedParityInterfacePresenceDevice(t *testing.T, environment *profileParityEnvironment) int64 {
	t.Helper()
	now := time.Date(2026, time.October, 12, 20, 0, 0, 0, time.UTC)
	metadata := dcimrow.RowMetadata{Created: now, LastUpdated: now}
	manufacturer := dcimrow.ManufacturerRow{RowMetadata: metadata, Name: "Interface Presence Vendor", Slug: "interface-presence-vendor"}
	require.NoError(t, environment.db.Create(&manufacturer).Error)
	deviceType := dcimrow.DeviceTypeRow{RowMetadata: metadata, ManufacturerID: manufacturer.ID, Model: "Interface Presence Type", Slug: "interface-presence-type", UHeight: 1, IsFullDepth: true}
	require.NoError(t, environment.db.Create(&deviceType).Error)
	site := dcimrow.SiteRow{RowMetadata: metadata, Name: "Interface Presence Site", Slug: "interface-presence-site", Status: "active"}
	require.NoError(t, environment.db.Create(&site).Error)
	role := dcimrow.DeviceRoleRow{RowMetadata: metadata, Name: "Interface Presence Role", Slug: "interface-presence-role", Color: "9e9e9e"}
	require.NoError(t, environment.db.Create(&role).Error)
	name := "interface-presence-device"
	device := dcimrow.DeviceRow{RowMetadata: metadata, DeviceTypeID: deviceType.ID, RoleID: role.ID, SiteID: site.ID, Name: &name, Status: "active"}
	require.NoError(t, environment.db.Create(&device).Error)
	return device.ID
}

type parityInterfacePresenceState struct {
	interfaces []dcimrow.InterfaceRow
	devices    []dcimrow.DeviceRow
	changes    []postgreschangelog.ChangeRow
}

func loadParityInterfacePresenceState(t *testing.T, environment *profileParityEnvironment) parityInterfacePresenceState {
	t.Helper()
	var state parityInterfacePresenceState
	require.NoError(t, environment.db.Order("id").Find(&state.interfaces).Error)
	require.NoError(t, environment.db.Order("id").Find(&state.devices).Error)
	require.NoError(t, environment.db.Order("id").Find(&state.changes).Error)
	var templates, addresses int64
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&templates).Error)
	require.NoError(t, environment.db.Model(&ipamrow.IPAddressRow{}).Count(&addresses).Error)
	require.Zero(t, templates)
	require.Zero(t, addresses)
	return state
}

func parityInterfaceRow(t *testing.T, state parityInterfacePresenceState, id int64) dcimrow.InterfaceRow {
	t.Helper()
	for _, row := range state.interfaces {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("Interface %d is absent", id)
	return dcimrow.InterfaceRow{}
}

func requireInterfaceParityMutation(t *testing.T, environment *profileParityEnvironment, before parityInterfacePresenceState, id int64, create bool) parityInterfacePresenceState {
	t.Helper()
	after := loadParityInterfacePresenceState(t, environment)
	require.Equal(t, before.devices, after.devices)
	count := len(before.interfaces)
	if create {
		count++
	}
	require.Len(t, after.interfaces, count)
	for _, prior := range before.interfaces {
		if prior.ID != id {
			require.Equal(t, prior, parityInterfaceRow(t, after, prior.ID))
		}
	}
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	current := parityInterfaceRow(t, after, id)
	change := after.changes[len(after.changes)-1]
	require.Equal(t, environment.principal.ID, change.ActorID)
	require.Equal(t, domaindcim.InterfaceObjectType, change.Kind)
	require.Equal(t, id, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(current.LastUpdated))
	if create {
		require.Equal(t, "create", change.Action)
		require.Empty(t, change.BeforeData)
	} else {
		require.Equal(t, "update", change.Action)
		prior := parityInterfaceRow(t, before, id)
		require.Equal(t, prior.Created, current.Created)
		encoded, err := json.Marshal(prior)
		require.NoError(t, err)
		require.JSONEq(t, string(encoded), string(change.BeforeData))
	}
	encoded, err := json.Marshal(current)
	require.NoError(t, err)
	require.JSONEq(t, string(encoded), string(change.AfterData))
	return after
}

func requireInterfaceParityProjection(t *testing.T, rest map[string]any, rpc *dcimv1.Interface, row dcimrow.InterfaceRow) {
	t.Helper()
	require.NotNil(t, rpc)
	require.NotNil(t, rpc.Device)
	require.Equal(t, row.ID, rpc.Id)
	require.Equal(t, float64(row.ID), rest["id"])
	require.Equal(t, row.DeviceID, rpc.Device.Id)
	require.Equal(t, float64(row.DeviceID), rest["device"].(map[string]any)["id"])
	require.Equal(t, "interface-presence-device", rpc.Device.Display)
	require.Equal(t, row.Name, rpc.Name)
	require.Equal(t, row.Name, rest["name"])
	require.Equal(t, row.Label, rpc.Label)
	require.Equal(t, row.Label, rest["label"])
	require.Equal(t, row.Type, rpc.Type)
	require.Equal(t, row.Type, rest["type"].(map[string]any)["value"])
	require.NotEmpty(t, rest["type"].(map[string]any)["label"])
	require.Equal(t, row.Enabled, rpc.Enabled)
	require.Equal(t, row.Enabled, rest["enabled"])
	require.Equal(t, row.MgmtOnly, rpc.MgmtOnly)
	require.Equal(t, row.MgmtOnly, rest["mgmt_only"])
	require.Equal(t, row.Description, rpc.Description)
	require.Equal(t, row.Description, rest["description"])
	if row.MTU == nil {
		require.Nil(t, rest["mtu"])
		require.Nil(t, rpc.Mtu)
	} else {
		require.NotNil(t, rpc.Mtu)
		require.Equal(t, float64(*row.MTU), rest["mtu"])
		require.Equal(t, int32(*row.MTU), rpc.Mtu.Value)
	}
	if row.Speed == nil {
		require.Nil(t, rest["speed"])
		require.Nil(t, rpc.Speed)
	} else {
		require.NotNil(t, rpc.Speed)
		require.Equal(t, float64(*row.Speed), rest["speed"])
		require.Equal(t, *row.Speed, rpc.Speed.Value)
	}
	if row.Duplex == nil {
		require.Nil(t, rest["duplex"])
		require.Nil(t, rpc.Duplex)
	} else {
		require.NotNil(t, rpc.Duplex)
		require.Equal(t, *row.Duplex, rpc.Duplex.Value)
		if *row.Duplex == "" {
			require.Nil(t, rest["duplex"])
		} else {
			require.Equal(t, *row.Duplex, rest["duplex"].(map[string]any)["value"])
			require.NotEmpty(t, rest["duplex"].(map[string]any)["label"])
		}
	}
	require.Equal(t, float64(0), rest["count_ipaddresses"])
	require.Zero(t, rpc.CountIpaddresses)
}

func interfacePresenceParityFields() []string {
	return []string{"device", "name", "label", "type", "enabled", "mgmt_only", "mtu", "speed", "duplex", "description"}
}

func requireInterfacePresenceRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

// Field diagnostics are observed at the shared application boundary, not added to the gRPC wire contract.
type interfacePresenceErrorObserver struct {
	restadapter.InterfaceService
	mutex sync.Mutex
	err   error
}

func (observer *interfacePresenceErrorObserver) remember(err error) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.err = err
}

func (observer *interfacePresenceErrorObserver) violations() []shared.FieldViolation {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	return append([]shared.FieldViolation(nil), shared.ViolationsOf(observer.err)...)
}

func (observer *interfacePresenceErrorObserver) CreateInterface(ctx context.Context, principal identity.Principal, command applicationdcim.CreateInterfaceCommand) (*domaindcim.Interface, error) {
	result, err := observer.InterfaceService.CreateInterface(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *interfacePresenceErrorObserver) ReplaceInterface(ctx context.Context, principal identity.Principal, command applicationdcim.ReplaceInterfaceCommand) (*domaindcim.Interface, error) {
	result, err := observer.InterfaceService.ReplaceInterface(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *interfacePresenceErrorObserver) UpdateInterface(ctx context.Context, principal identity.Principal, command applicationdcim.UpdateInterfaceCommand) (*domaindcim.Interface, error) {
	result, err := observer.InterfaceService.UpdateInterface(ctx, principal, command)
	observer.remember(err)
	return result, err
}
