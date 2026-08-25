package parity

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	"netbox-go/internal/application/authz"
	domaindcim "netbox-go/internal/domain/dcim"
)

func TestInterfaceTemplateScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	manufacturer := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{
			"name": "InterfaceTemplate Presence Manufacturer",
			"slug": "interface-template-presence-manufacturer",
		},
		http.StatusCreated,
	)
	manufacturerID := jsonID(t, manufacturer["id"])
	deviceTypeA := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/device-types",
		map[string]any{
			"manufacturer": manufacturerID,
			"model":        "InterfaceTemplate Presence Device Type A",
			"slug":         "interface-template-presence-device-type-a",
		},
		http.StatusCreated,
	)
	deviceTypeB := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/device-types",
		map[string]any{
			"manufacturer": manufacturerID,
			"model":        "InterfaceTemplate Presence Device Type B",
			"slug":         "interface-template-presence-device-type-b",
		},
		http.StatusCreated,
	)
	deviceTypeAID := jsonID(t, deviceTypeA["id"])
	deviceTypeBID := jsonID(t, deviceTypeB["id"])

	var templatesBefore, devicesBefore, interfacesBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&templatesBefore).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&devicesBefore).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&interfacesBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	require.Zero(t, templatesBefore)
	require.Zero(t, devicesBefore, "scalar presence requires zero Devices")
	require.Zero(t, interfacesBefore, "scalar presence requires zero Interfaces")
	require.Equal(t, int64(3), changesBefore, "the three public fixture creates each record once")

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/interface-templates",
		map[string]any{
			"device_type": deviceTypeAID,
			"name":        "  REST Default Interface  ",
			"type":        "1000base-t",
		},
		http.StatusCreated,
	)
	templateID := jsonID(t, created["id"])
	itemPath := "/api/dcim/interface-templates/" + strconv.FormatInt(templateID, 10)
	requireInterfaceTemplateRESTScalars(
		t, created, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"REST Default Interface", "", "1000base-t", true, false, "",
	)
	createdState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	require.Equal(t, templatesBefore+1, createdState.interfaceTemplateCount)
	require.Equal(t, int64(1), createdState.targetChangeCount)
	require.Equal(t, int64(1), createdState.allInterfaceTemplateChangeCount)
	require.Equal(t, changesBefore+1, createdState.totalChangeCount)
	require.Zero(t, createdState.deviceCount)
	require.Zero(t, createdState.interfaceCount)

	grpcRead, err := environment.dcim.GetInterfaceTemplate(
		environment.ctx, &dcimv1.GetInterfaceTemplateRequest{Id: templateID},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, grpcRead.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "REST Default Interface", "",
		"1000base-t", true, false, "",
	)

	grpcCreated, err := environment.dcim.CreateInterfaceTemplate(
		environment.ctx,
		&dcimv1.CreateInterfaceTemplateRequest{
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				DeviceType: &deviceTypeAID,
				Name:       pointer("  gRPC Concrete Interface  "),
				Label:      pointer("  secondary  "),
				Type:       pointer("10gbase-sr"),
				Enabled:    pointer(false),
				MgmtOnly:   pointer(true),
				Description: pointer(
					"  gRPC concrete description  ",
				),
			},
		},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, grpcCreated.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Concrete Interface",
		"secondary", "10gbase-sr", false, true, "gRPC concrete description",
	)
	grpcCreatedPath := "/api/dcim/interface-templates/" +
		strconv.FormatInt(grpcCreated.InterfaceTemplate.Id, 10)
	grpcCreatedRESTRead := requestJSON(
		t, environment.router, http.MethodGet, grpcCreatedPath, nil, http.StatusOK,
	)
	requireInterfaceTemplateRESTScalars(
		t, grpcCreatedRESTRead, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Concrete Interface",
		"secondary", "10gbase-sr", false, true, "gRPC concrete description",
	)
	grpcCreatedState := loadParityInterfaceTemplatePresenceState(
		t, environment, grpcCreated.InterfaceTemplate.Id,
	)
	require.Equal(t, createdState.interfaceTemplateCount+1, grpcCreatedState.interfaceTemplateCount)
	require.Equal(t, int64(1), grpcCreatedState.targetChangeCount)
	require.Equal(
		t, createdState.allInterfaceTemplateChangeCount+1,
		grpcCreatedState.allInterfaceTemplateChangeCount,
	)
	require.Equal(t, createdState.totalChangeCount+1, grpcCreatedState.totalChangeCount)
	require.Zero(t, grpcCreatedState.deviceCount)
	require.Zero(t, grpcCreatedState.interfaceCount)

	beforeConcrete := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	concrete, err := environment.dcim.UpdateInterfaceTemplate(
		environment.ctx,
		&dcimv1.UpdateInterfaceTemplateRequest{
			Id: templateID,
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				DeviceType:  &deviceTypeAID,
				Name:        pointer("  gRPC Patched Interface  "),
				Label:       pointer("  WAN  "),
				Type:        pointer("25gbase-sr"),
				Enabled:     pointer(false),
				MgmtOnly:    pointer(true),
				Description: pointer("  gRPC patched description  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: interfaceTemplateScalarFields()},
		},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, concrete.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Patched Interface", "WAN",
		"25gbase-sr", false, true, "gRPC patched description",
	)
	concreteState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, beforeConcrete, concreteState)
	restRead := requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireInterfaceTemplateRESTScalars(
		t, restRead, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"gRPC Patched Interface", "WAN", "25gbase-sr", false, true,
		"gRPC patched description",
	)

	replacedByREST := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"device_type": deviceTypeAID,
			"name":        "  REST Replaced Interface  ",
			"type":        "40gbase-sr4",
		},
		http.StatusOK,
	)
	requireInterfaceTemplateRESTScalars(
		t, replacedByREST, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"REST Replaced Interface", "WAN", "40gbase-sr4", false, true,
		"gRPC patched description",
	)
	replacedByRESTState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, concreteState, replacedByRESTState)
	grpcRead, err = environment.dcim.GetInterfaceTemplate(
		environment.ctx, &dcimv1.GetInterfaceTemplateRequest{Id: templateID},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, grpcRead.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "REST Replaced Interface", "WAN",
		"40gbase-sr4", false, true, "gRPC patched description",
	)

	replacedByGRPC, err := environment.dcim.ReplaceInterfaceTemplate(
		environment.ctx,
		&dcimv1.ReplaceInterfaceTemplateRequest{
			Id: templateID,
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				DeviceType: &deviceTypeAID,
				Name:       pointer("  gRPC Replaced Interface  "),
				Type:       pointer("50gbase-sr"),
			},
		},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, replacedByGRPC.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Replaced Interface", "WAN",
		"50gbase-sr", false, true, "gRPC patched description",
	)
	replacedByGRPCState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, replacedByRESTState, replacedByGRPCState)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireInterfaceTemplateRESTScalars(
		t, restRead, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"gRPC Replaced Interface", "WAN", "50gbase-sr", false, true,
		"gRPC patched description",
	)

	clearedByGRPC, err := environment.dcim.UpdateInterfaceTemplate(
		environment.ctx,
		&dcimv1.UpdateInterfaceTemplateRequest{
			Id: templateID,
			InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
				Label: pointer(""), Enabled: pointer(false), MgmtOnly: pointer(false),
				Description: pointer(""),
			},
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"label", "enabled", "mgmt_only", "description"},
			},
		},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, clearedByGRPC.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Replaced Interface", "",
		"50gbase-sr", false, false, "",
	)
	clearedByGRPCState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, replacedByGRPCState, clearedByGRPCState)

	restoredByREST := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"label": "  REST WAN  ", "enabled": true, "mgmt_only": true,
			"description": "  REST restored description  ",
		},
		http.StatusOK,
	)
	requireInterfaceTemplateRESTScalars(
		t, restoredByREST, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"gRPC Replaced Interface", "REST WAN", "50gbase-sr", true, true,
		"REST restored description",
	)
	restoredByRESTState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, clearedByGRPCState, restoredByRESTState)

	clearedByREST := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"label": "", "enabled": false, "mgmt_only": false, "description": "",
		},
		http.StatusOK,
	)
	requireInterfaceTemplateRESTScalars(
		t, clearedByREST, deviceTypeAID, "InterfaceTemplate Presence Device Type A",
		"gRPC Replaced Interface", "", "50gbase-sr", false, false, "",
	)
	clearedByRESTState := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
	requireInterfaceTemplateParityUpdateRecorded(t, restoredByRESTState, clearedByRESTState)
	grpcRead, err = environment.dcim.GetInterfaceTemplate(
		environment.ctx, &dcimv1.GetInterfaceTemplateRequest{Id: templateID},
	)
	require.NoError(t, err)
	requireInterfaceTemplateProtoScalars(
		t, grpcRead.InterfaceTemplate, deviceTypeAID,
		"InterfaceTemplate Presence Device Type A", "gRPC Replaced Interface", "",
		"50gbase-sr", false, false, "",
	)

	for _, field := range interfaceTemplateScalarFields() {
		field := field
		t.Run("REST POST null/"+field, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			body := map[string]any{
				"device_type": deviceTypeAID,
				"name":        "Rejected POST " + field,
				"type":        "1000base-t",
			}
			body[field] = nil
			requestJSON(
				t, environment.router, http.MethodPost, "/api/dcim/interface-templates",
				body, http.StatusBadRequest,
			)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
		t.Run("REST PUT null/"+field, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			body := map[string]any{
				"device_type": deviceTypeAID,
				"name":        "gRPC Replaced Interface",
				"type":        "50gbase-sr",
			}
			body[field] = nil
			requestJSON(t, environment.router, http.MethodPut, itemPath, body, http.StatusBadRequest)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
		t.Run("REST PATCH null/"+field, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				map[string]any{field: nil}, http.StatusBadRequest,
			)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
		t.Run("gRPC masked absent/"+field, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			_, updateErr := environment.dcim.UpdateInterfaceTemplate(
				environment.ctx,
				&dcimv1.UpdateInterfaceTemplateRequest{
					Id: templateID, InterfaceTemplate: &dcimv1.InterfaceTemplateInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireInterfaceTemplateGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
	}

	t.Run("required fields omitted", func(t *testing.T) {
		before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
		requestJSON(
			t, environment.router, http.MethodPost, "/api/dcim/interface-templates",
			map[string]any{}, http.StatusBadRequest,
		)
		requestJSON(t, environment.router, http.MethodPut, itemPath, map[string]any{}, http.StatusBadRequest)
		_, createErr := environment.dcim.CreateInterfaceTemplate(
			environment.ctx,
			&dcimv1.CreateInterfaceTemplateRequest{
				InterfaceTemplate: &dcimv1.InterfaceTemplateInput{},
			},
		)
		requireInterfaceTemplateGRPCInvalid(t, createErr)
		_, replaceErr := environment.dcim.ReplaceInterfaceTemplate(
			environment.ctx,
			&dcimv1.ReplaceInterfaceTemplateRequest{
				Id: templateID, InterfaceTemplate: &dcimv1.InterfaceTemplateInput{},
			},
		)
		requireInterfaceTemplateGRPCInvalid(t, replaceErr)
		require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
	})

	for _, rejected := range []struct {
		name string
		body map[string]any
	}{
		{name: "zero DeviceType", body: map[string]any{"device_type": int64(0)}},
		{name: "blank name", body: map[string]any{"name": "   "}},
		{name: "untrimmed type", body: map[string]any{"type": " 50gbase-sr "}},
		{name: "unknown DeviceType", body: map[string]any{"device_type": int64(999999)}},
		{name: "different known DeviceType", body: map[string]any{"device_type": deviceTypeBID}},
	} {
		rejected := rejected
		t.Run("REST rejection/"+rejected.name, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				rejected.body, http.StatusBadRequest,
			)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
	}

	for _, rejected := range []struct {
		name  string
		value int64
	}{
		{name: "zero DeviceType", value: 0},
		{name: "unknown DeviceType", value: 999999},
		{name: "different known DeviceType", value: deviceTypeBID},
	} {
		rejected := rejected
		t.Run("gRPC relationship rejection/"+rejected.name, func(t *testing.T) {
			before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
			_, updateErr := environment.dcim.UpdateInterfaceTemplate(
				environment.ctx,
				&dcimv1.UpdateInterfaceTemplateRequest{
					Id: templateID,
					InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
						DeviceType: &rejected.value,
					},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"device_type"}},
				},
			)
			requireInterfaceTemplateGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
		})
	}

	t.Run("gRPC invalid scalar and unknown mask fail closed", func(t *testing.T) {
		before := loadParityInterfaceTemplatePresenceState(t, environment, templateID)
		for _, input := range []*dcimv1.InterfaceTemplateInput{
			{Name: pointer("   ")},
			{Type: pointer(" 50gbase-sr ")},
		} {
			field := "name"
			if input.Type != nil {
				field = "type"
			}
			_, updateErr := environment.dcim.UpdateInterfaceTemplate(
				environment.ctx,
				&dcimv1.UpdateInterfaceTemplateRequest{
					Id: templateID, InterfaceTemplate: input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireInterfaceTemplateGRPCInvalid(t, updateErr)
		}
		_, maskErr := environment.dcim.UpdateInterfaceTemplate(
			environment.ctx,
			&dcimv1.UpdateInterfaceTemplateRequest{
				Id: templateID,
				InterfaceTemplate: &dcimv1.InterfaceTemplateInput{
					Description: pointer("must not persist"),
				},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
			},
		)
		requireInterfaceTemplateGRPCInvalid(t, maskErr)
		require.Equal(t, before, loadParityInterfaceTemplatePresenceState(t, environment, templateID))
	})
}

func interfaceTemplateScalarFields() []string {
	return []string{
		"device_type", "name", "label", "type", "enabled", "mgmt_only", "description",
	}
}

type parityInterfaceTemplatePresenceState struct {
	row                             dcimrow.InterfaceTemplateRow
	deviceTypeCount                 int64
	interfaceTemplateCount          int64
	deviceCount                     int64
	interfaceCount                  int64
	targetChangeCount               int64
	allInterfaceTemplateChangeCount int64
	deviceTypeChangeCount           int64
	totalChangeCount                int64
}

func loadParityInterfaceTemplatePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityInterfaceTemplatePresenceState {
	t.Helper()
	var state parityInterfaceTemplatePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceTypeRow{}).Count(&state.deviceTypeCount).Error)
	require.NoError(
		t,
		environment.db.Model(&dcimrow.InterfaceTemplateRow{}).
			Count(&state.interfaceTemplateCount).Error,
	)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&state.interfaceCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.InterfaceTemplateObjectType, id,
	).Count(&state.targetChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.InterfaceTemplateObjectType,
	).Count(&state.allInterfaceTemplateChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceTypeObjectType,
	).Count(&state.deviceTypeChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireInterfaceTemplateParityUpdateRecorded(
	t *testing.T,
	before parityInterfaceTemplatePresenceState,
	after parityInterfaceTemplatePresenceState,
) {
	t.Helper()
	require.Equal(t, before.deviceTypeCount, after.deviceTypeCount)
	require.Equal(t, before.interfaceTemplateCount, after.interfaceTemplateCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, before.interfaceCount, after.interfaceCount)
	require.Zero(t, after.deviceCount, "scalar presence requires zero Devices")
	require.Zero(t, after.interfaceCount, "scalar presence requires zero Interfaces")
	require.Equal(t, before.targetChangeCount+1, after.targetChangeCount)
	require.Equal(
		t, before.allInterfaceTemplateChangeCount+1,
		after.allInterfaceTemplateChangeCount,
	)
	require.Equal(t, before.deviceTypeChangeCount, after.deviceTypeChangeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireInterfaceTemplateGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireInterfaceTemplateProtoScalars(
	t *testing.T,
	template *dcimv1.InterfaceTemplate,
	deviceTypeID int64,
	deviceTypeDisplay string,
	name string,
	label string,
	interfaceType string,
	enabled bool,
	mgmtOnly bool,
	description string,
) {
	t.Helper()
	require.NotNil(t, template)
	require.NotNil(t, template.DeviceType)
	require.Equal(t, deviceTypeID, template.DeviceType.Id)
	require.Equal(t, deviceTypeDisplay, template.DeviceType.Display)
	require.Equal(t, name, template.Name)
	require.Equal(t, label, template.Label)
	require.Equal(t, interfaceType, template.Type)
	require.Equal(t, enabled, template.Enabled)
	require.Equal(t, mgmtOnly, template.MgmtOnly)
	require.Equal(t, description, template.Description)
}

func requireInterfaceTemplateRESTScalars(
	t *testing.T,
	template map[string]any,
	deviceTypeID int64,
	deviceTypeDisplay string,
	name string,
	label string,
	interfaceType string,
	enabled bool,
	mgmtOnly bool,
	description string,
) {
	t.Helper()
	deviceType, ok := template["device_type"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(deviceTypeID), deviceType["id"])
	require.Equal(t, deviceTypeDisplay, deviceType["display"])
	typeChoice, ok := template["type"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, interfaceType, typeChoice["value"])
	require.NotEmpty(t, typeChoice["label"])
	require.Equal(t, name, template["name"])
	require.Equal(t, label, template["label"])
	require.Equal(t, enabled, template["enabled"])
	require.Equal(t, mgmtOnly, template["mgmt_only"])
	require.Equal(t, description, template["description"])
}
