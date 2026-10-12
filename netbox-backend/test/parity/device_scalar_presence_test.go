package parity

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	"netbox-go/internal/application/authz"
	domaindcim "netbox-go/internal/domain/dcim"
)

func TestDeviceScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	site := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/sites",
		map[string]any{"name": "Device Presence Site", "slug": "device-presence-site"},
		http.StatusCreated,
	)
	manufacturer := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{
			"name": "Device Presence Manufacturer", "slug": "device-presence-manufacturer",
		},
		http.StatusCreated,
	)
	role := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/device-roles",
		map[string]any{
			"name": "Device Presence Role", "slug": "device-presence-role", "color": "00ff00",
		},
		http.StatusCreated,
	)
	siteID := jsonID(t, site["id"])
	manufacturerID := jsonID(t, manufacturer["id"])
	roleID := jsonID(t, role["id"])
	deviceType := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/device-types",
		map[string]any{
			"manufacturer": manufacturerID,
			"model":        "Device Presence Type",
			"slug":         "device-presence-type",
			"airflow":      "front-to-rear",
		},
		http.StatusCreated,
	)
	rack := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/racks",
		map[string]any{"site": siteID, "name": "Device Presence Rack"},
		http.StatusCreated,
	)
	deviceTypeID := jsonID(t, deviceType["id"])
	rackID := jsonID(t, rack["id"])
	requireDevicePresenceComponentsZero(t, environment)

	defaultExpected := parityDeviceExpected{
		deviceTypeID: deviceTypeID, deviceTypeDisplay: "Device Presence Type",
		roleID: roleID, roleDisplay: "Device Presence Role",
		siteID: siteID, siteDisplay: "Device Presence Site",
		status: "active", airflow: devicePresenceString("front-to-rear"),
	}
	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/devices",
		map[string]any{"device_type": deviceTypeID, "role": roleID, "site": siteID},
		http.StatusCreated,
	)
	deviceID := jsonID(t, created["id"])
	itemPath := "/api/dcim/devices/" + strconv.FormatInt(deviceID, 10)
	requireDeviceRESTPresence(t, created, defaultExpected)
	createdState := loadParityDevicePresenceState(t, environment, deviceID)
	require.Equal(t, int64(1), createdState.deviceCount)
	require.Equal(t, int64(1), createdState.targetDeviceChangeCount)
	require.Equal(t, int64(1), createdState.allDeviceChangeCount)
	requireDevicePresenceNoComponentSideEffects(t, createdState)
	requireDevicePresenceRow(t, createdState.row, defaultExpected)

	grpcRead, err := environment.dcim.GetDevice(
		environment.ctx, &dcimv1.GetDeviceRequest{Id: deviceID},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcRead.Device, defaultExpected)

	grpcCreatedExpected := parityDeviceExpected{
		deviceTypeID: deviceTypeID, deviceTypeDisplay: "Device Presence Type",
		roleID: roleID, roleDisplay: "Device Presence Role",
		name:   devicePresenceString("gRPC Device"),
		siteID: siteID, siteDisplay: "Device Presence Site",
		status: "planned", serial: "GRPC-SERIAL",
		assetTag:    devicePresenceString("GRPC-ASSET"),
		airflow:     devicePresenceString("top-to-bottom"),
		description: "gRPC description", comments: "gRPC comments",
	}
	grpcCreated, err := environment.dcim.CreateDevice(
		environment.ctx,
		&dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{
			DeviceType:  &deviceTypeID,
			Role:        &roleID,
			Name:        wrapperspb.String("  gRPC Device  "),
			Site:        &siteID,
			Status:      pointer("planned"),
			Serial:      pointer("  GRPC-SERIAL  "),
			AssetTag:    wrapperspb.String("  GRPC-ASSET  "),
			Airflow:     pointer("top-to-bottom"),
			Description: pointer("  gRPC description  "),
			Comments:    pointer("  gRPC comments  "),
		}},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcCreated.Device, grpcCreatedExpected)
	grpcCreatedREST := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/dcim/devices/"+strconv.FormatInt(grpcCreated.Device.Id, 10), nil, http.StatusOK,
	)
	requireDeviceRESTPresence(t, grpcCreatedREST, grpcCreatedExpected)
	grpcCreatedState := loadParityDevicePresenceState(t, environment, grpcCreated.Device.Id)
	require.Equal(t, createdState.deviceCount+1, grpcCreatedState.deviceCount)
	require.Equal(t, int64(1), grpcCreatedState.targetDeviceChangeCount)
	require.Equal(t, createdState.allDeviceChangeCount+1, grpcCreatedState.allDeviceChangeCount)
	require.Equal(t, createdState.totalChangeCount+1, grpcCreatedState.totalChangeCount)
	requireDevicePresenceNoComponentSideEffects(t, grpcCreatedState)
	requireDevicePresenceRow(t, grpcCreatedState.row, grpcCreatedExpected)

	concreteExpected := defaultExpected
	concreteExpected.name = devicePresenceString("gRPC Concrete Device")
	concreteExpected.rackID = &rackID
	concreteExpected.position = devicePresenceFloat(10.5)
	concreteExpected.face = "front"
	concreteExpected.status = "planned"
	concreteExpected.serial = "SERIAL-I10"
	concreteExpected.assetTag = devicePresenceString("ASSET-I10")
	concreteExpected.airflow = devicePresenceString("rear-to-front")
	concreteExpected.description = "concrete description"
	concreteExpected.comments = "concrete comments"
	beforeConcrete := loadParityDevicePresenceState(t, environment, deviceID)
	concrete, err := environment.dcim.UpdateDevice(
		environment.ctx,
		&dcimv1.UpdateDeviceRequest{
			Id: deviceID,
			Device: &dcimv1.DeviceInput{
				Name: wrapperspb.String("  gRPC Concrete Device  "),
				Rack: wrapperspb.Int64(rackID), Position: wrapperspb.String("10.5"),
				Face: wrapperspb.String("front"), Status: pointer("planned"),
				Serial: pointer("  SERIAL-I10  "), AssetTag: wrapperspb.String("  ASSET-I10  "),
				Airflow: pointer("rear-to-front"), Description: pointer("  concrete description  "),
				Comments: pointer("  concrete comments  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
				"name", "rack", "position", "face", "status", "serial",
				"asset_tag", "airflow", "description", "comments",
			}},
		},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, concrete.Device, concreteExpected)
	concreteState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, beforeConcrete, concreteState)
	requireDevicePresenceRow(t, concreteState.row, concreteExpected)
	requireDeviceRESTPresence(
		t,
		requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK),
		concreteExpected,
	)

	replacedExpected := concreteExpected
	replacedExpected.rackID = nil
	replacedExpected.position = nil
	replacedExpected.face = ""
	replaced := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{"device_type": deviceTypeID, "role": roleID, "site": siteID},
		http.StatusOK,
	)
	requireDeviceRESTPresence(t, replaced, replacedExpected)
	replacedState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, concreteState, replacedState)
	requireDevicePresenceRow(t, replacedState.row, replacedExpected)
	grpcRead, err = environment.dcim.GetDevice(
		environment.ctx, &dcimv1.GetDeviceRequest{Id: deviceID},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcRead.Device, replacedExpected)

	placedAgainExpected := replacedExpected
	placedAgainExpected.rackID = &rackID
	placedAgainExpected.position = devicePresenceFloat(20.5)
	placedAgainExpected.face = "rear"
	placedAgain := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"rack": rackID, "position": 20.5, "face": "rear"},
		http.StatusOK,
	)
	requireDeviceRESTPresence(t, placedAgain, placedAgainExpected)
	placedAgainState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, replacedState, placedAgainState)
	requireDevicePresenceRow(t, placedAgainState.row, placedAgainExpected)
	grpcRead, err = environment.dcim.GetDevice(
		environment.ctx, &dcimv1.GetDeviceRequest{Id: deviceID},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcRead.Device, placedAgainExpected)

	grpcReplacedExpected := placedAgainExpected
	grpcReplacedExpected.rackID = nil
	grpcReplacedExpected.position = nil
	grpcReplacedExpected.face = ""
	grpcReplaced, err := environment.dcim.ReplaceDevice(
		environment.ctx,
		&dcimv1.ReplaceDeviceRequest{Id: deviceID, Device: &dcimv1.DeviceInput{
			DeviceType: &deviceTypeID, Role: &roleID, Site: &siteID,
		}},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcReplaced.Device, grpcReplacedExpected)
	grpcReplacedState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, placedAgainState, grpcReplacedState)
	requireDevicePresenceRow(t, grpcReplacedState.row, grpcReplacedExpected)
	requireDeviceRESTPresence(
		t,
		requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK),
		grpcReplacedExpected,
	)

	clearedExpected := grpcReplacedExpected
	clearedExpected.name = nil
	clearedExpected.status = "inventory"
	clearedExpected.serial = ""
	clearedExpected.assetTag = nil
	clearedExpected.airflow = devicePresenceString("")
	clearedExpected.description = ""
	clearedExpected.comments = ""
	cleared := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"name": nil, "rack": nil, "position": nil, "face": nil,
			"status": "inventory", "serial": "", "asset_tag": nil,
			"airflow": nil, "description": "", "comments": "",
		},
		http.StatusOK,
	)
	requireDeviceRESTPresence(t, cleared, clearedExpected)
	clearedState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, grpcReplacedState, clearedState)
	requireDevicePresenceRow(t, clearedState.row, clearedExpected)
	require.NotNil(t, clearedState.row.Airflow)
	require.Empty(t, *clearedState.row.Airflow, "PATCH airflow:null stores blank and does not re-inherit")
	grpcRead, err = environment.dcim.GetDevice(
		environment.ctx, &dcimv1.GetDeviceRequest{Id: deviceID},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcRead.Device, clearedExpected)

	blankNullableExpected := clearedExpected
	blankNullableExpected.name = devicePresenceString("")
	blankNullableExpected.assetTag = devicePresenceString("")
	blankNullable := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"name": "", "asset_tag": ""}, http.StatusOK,
	)
	requireDeviceRESTPresence(t, blankNullable, blankNullableExpected)
	blankNullableState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, clearedState, blankNullableState)
	requireDevicePresenceRow(t, blankNullableState.row, blankNullableExpected)
	require.NotNil(t, blankNullableState.row.Name)
	require.NotNil(t, blankNullableState.row.AssetTag)

	grpcPatchedExpected := blankNullableExpected
	grpcPatchedExpected.name = devicePresenceString("gRPC PATCH Device")
	grpcPatchedExpected.rackID = &rackID
	grpcPatchedExpected.position = devicePresenceFloat(30)
	grpcPatchedExpected.face = "front"
	grpcPatchedExpected.status = "staged"
	grpcPatchedExpected.serial = "PATCH-SERIAL"
	grpcPatchedExpected.assetTag = devicePresenceString("PATCH-ASSET")
	grpcPatchedExpected.airflow = devicePresenceString("left-to-right")
	grpcPatchedExpected.description = "patch description"
	grpcPatchedExpected.comments = "patch comments"
	grpcPatched, err := environment.dcim.UpdateDevice(
		environment.ctx,
		&dcimv1.UpdateDeviceRequest{
			Id: deviceID,
			Device: &dcimv1.DeviceInput{
				Name: wrapperspb.String("  gRPC PATCH Device  "),
				Rack: wrapperspb.Int64(rackID), Position: wrapperspb.String("30"),
				Face: wrapperspb.String("front"), Status: pointer("staged"),
				Serial: pointer("  PATCH-SERIAL  "), AssetTag: wrapperspb.String("  PATCH-ASSET  "),
				Airflow: pointer("left-to-right"), Description: pointer("  patch description  "),
				Comments: pointer("  patch comments  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
				"name", "rack", "position", "face", "status", "serial",
				"asset_tag", "airflow", "description", "comments",
			}},
		},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcPatched.Device, grpcPatchedExpected)
	grpcPatchedState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, blankNullableState, grpcPatchedState)
	requireDevicePresenceRow(t, grpcPatchedState.row, grpcPatchedExpected)
	requireDeviceRESTPresence(
		t,
		requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK),
		grpcPatchedExpected,
	)

	grpcClearedExpected := grpcPatchedExpected
	grpcClearedExpected.name = nil
	grpcClearedExpected.rackID = nil
	grpcClearedExpected.position = nil
	grpcClearedExpected.face = ""
	grpcClearedExpected.assetTag = nil
	grpcClearedExpected.airflow = devicePresenceString("")
	grpcCleared, err := environment.dcim.UpdateDevice(
		environment.ctx,
		&dcimv1.UpdateDeviceRequest{
			Id: deviceID, Device: &dcimv1.DeviceInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
				"name", "rack", "position", "face", "asset_tag", "airflow",
			}},
		},
	)
	require.NoError(t, err)
	requireDeviceProtoPresence(t, grpcCleared.Device, grpcClearedExpected)
	grpcClearedState := loadParityDevicePresenceState(t, environment, deviceID)
	requireDeviceParityMutationRecorded(t, grpcPatchedState, grpcClearedState)
	requireDevicePresenceRow(t, grpcClearedState.row, grpcClearedExpected)
	require.NotNil(t, grpcClearedState.row.Airflow)
	require.Empty(t, *grpcClearedState.row.Airflow)
	requireDeviceRESTPresence(
		t,
		requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK),
		grpcClearedExpected,
	)

	t.Run("required create and replace fields", func(t *testing.T) {
		before := loadParityDevicePresenceState(t, environment, deviceID)
		requestJSON(
			t, environment.router, http.MethodPost, "/api/dcim/devices",
			map[string]any{}, http.StatusBadRequest,
		)
		requestJSON(t, environment.router, http.MethodPut, itemPath, map[string]any{}, http.StatusBadRequest)
		_, createErr := environment.dcim.CreateDevice(
			environment.ctx, &dcimv1.CreateDeviceRequest{Device: &dcimv1.DeviceInput{}},
		)
		requireDevicePresenceGRPCInvalid(t, createErr)
		_, replaceErr := environment.dcim.ReplaceDevice(
			environment.ctx,
			&dcimv1.ReplaceDeviceRequest{Id: deviceID, Device: &dcimv1.DeviceInput{}},
		)
		requireDevicePresenceGRPCInvalid(t, replaceErr)
		require.Equal(t, before, loadParityDevicePresenceState(t, environment, deviceID))
	})

	for _, field := range []string{"device_type", "role", "site"} {
		field := field
		t.Run("required PATCH value/"+field, func(t *testing.T) {
			before := loadParityDevicePresenceState(t, environment, deviceID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				map[string]any{field: nil}, http.StatusBadRequest,
			)
			_, updateErr := environment.dcim.UpdateDevice(
				environment.ctx,
				&dcimv1.UpdateDeviceRequest{
					Id: deviceID, Device: &dcimv1.DeviceInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireDevicePresenceGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityDevicePresenceState(t, environment, deviceID))
		})
	}

	for _, rejected := range []struct {
		name  string
		field string
		value string
		input *dcimv1.DeviceInput
	}{
		{
			name: "status whitespace", field: "status", value: " active ",
			input: &dcimv1.DeviceInput{Status: pointer(" active ")},
		},
		{
			name: "face whitespace", field: "face", value: " front ",
			input: &dcimv1.DeviceInput{Face: wrapperspb.String(" front ")},
		},
		{
			name: "airflow whitespace", field: "airflow", value: " front-to-rear ",
			input: &dcimv1.DeviceInput{Airflow: pointer(" front-to-rear ")},
		},
	} {
		rejected := rejected
		t.Run("exact choice/"+rejected.name, func(t *testing.T) {
			before := loadParityDevicePresenceState(t, environment, deviceID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				map[string]any{rejected.field: rejected.value}, http.StatusBadRequest,
			)
			_, updateErr := environment.dcim.UpdateDevice(
				environment.ctx,
				&dcimv1.UpdateDeviceRequest{
					Id: deviceID, Device: rejected.input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{rejected.field}},
				},
			)
			requireDevicePresenceGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityDevicePresenceState(t, environment, deviceID))
		})
	}
}

type parityDeviceExpected struct {
	deviceTypeID      int64
	deviceTypeDisplay string
	roleID            int64
	roleDisplay       string
	name              *string
	siteID            int64
	siteDisplay       string
	rackID            *int64
	position          *float64
	face              string
	status            string
	serial            string
	assetTag          *string
	airflow           *string
	description       string
	comments          string
}

type parityDevicePresenceState struct {
	row                          dcimrow.DeviceRow
	deviceCount                  int64
	templateCount                int64
	interfaceCount               int64
	targetDeviceChangeCount      int64
	allDeviceChangeCount         int64
	interfaceTemplateChangeCount int64
	interfaceChangeCount         int64
	totalChangeCount             int64
}

func loadParityDevicePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityDevicePresenceState {
	t.Helper()
	var state parityDevicePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&state.templateCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&state.interfaceCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceObjectType, id,
	).Count(&state.targetDeviceChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceObjectType,
	).Count(&state.allDeviceChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.InterfaceTemplateObjectType,
	).Count(&state.interfaceTemplateChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.InterfaceObjectType,
	).Count(&state.interfaceChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireDeviceParityMutationRecorded(
	t *testing.T,
	before parityDevicePresenceState,
	after parityDevicePresenceState,
) {
	t.Helper()
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, before.templateCount, after.templateCount)
	require.Equal(t, before.interfaceCount, after.interfaceCount)
	require.Equal(t, before.targetDeviceChangeCount+1, after.targetDeviceChangeCount)
	require.Equal(t, before.allDeviceChangeCount+1, after.allDeviceChangeCount)
	require.Equal(t, before.interfaceTemplateChangeCount, after.interfaceTemplateChangeCount)
	require.Equal(t, before.interfaceChangeCount, after.interfaceChangeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
	requireDevicePresenceNoComponentSideEffects(t, after)
}

func requireDevicePresenceComponentsZero(
	t *testing.T,
	environment *profileParityEnvironment,
) {
	t.Helper()
	var templates, interfaces, templateChanges, interfaceChanges int64
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&templates).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceRow{}).Count(&interfaces).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.InterfaceTemplateObjectType,
	).Count(&templateChanges).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.InterfaceObjectType,
	).Count(&interfaceChanges).Error)
	require.Zero(t, templates, "I10 fixtures contain zero InterfaceTemplates")
	require.Zero(t, interfaces, "I10 creates no Interfaces")
	require.Zero(t, templateChanges, "I10 records no InterfaceTemplate changes")
	require.Zero(t, interfaceChanges, "I10 records no Interface changes")
}

func requireDevicePresenceNoComponentSideEffects(
	t *testing.T,
	state parityDevicePresenceState,
) {
	t.Helper()
	require.Zero(t, state.templateCount, "I10 fixtures contain zero InterfaceTemplates")
	require.Zero(t, state.interfaceCount, "I10 creates no Interfaces")
	require.Zero(t, state.interfaceTemplateChangeCount, "I10 records no InterfaceTemplate changes")
	require.Zero(t, state.interfaceChangeCount, "I10 records no Interface changes")
}

func requireDevicePresenceGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireDevicePresenceRow(
	t *testing.T,
	row dcimrow.DeviceRow,
	expected parityDeviceExpected,
) {
	t.Helper()
	require.Equal(t, expected.deviceTypeID, row.DeviceTypeID)
	require.Equal(t, expected.roleID, row.RoleID)
	require.Equal(t, expected.name, row.Name)
	require.Equal(t, expected.siteID, row.SiteID)
	require.Equal(t, expected.rackID, row.RackID)
	require.Equal(t, expected.position, row.Position)
	require.Equal(t, expected.face, row.Face)
	require.Equal(t, expected.status, row.Status)
	require.Equal(t, expected.serial, row.Serial)
	require.Equal(t, expected.assetTag, row.AssetTag)
	require.Equal(t, expected.airflow, row.Airflow)
	require.Equal(t, expected.description, row.Description)
	require.Equal(t, expected.comments, row.Comments)
}

func requireDeviceProtoPresence(
	t *testing.T,
	device *dcimv1.Device,
	expected parityDeviceExpected,
) {
	t.Helper()
	require.NotNil(t, device)
	require.NotNil(t, device.DeviceType)
	require.Equal(t, expected.deviceTypeID, device.DeviceType.Id)
	require.Equal(t, expected.deviceTypeDisplay, device.DeviceType.Display)
	require.NotNil(t, device.Role)
	require.Equal(t, expected.roleID, device.Role.Id)
	require.Equal(t, expected.roleDisplay, device.Role.Display)
	require.Equal(t, devicePresenceValue(expected.name), device.Name)
	require.NotNil(t, device.Site)
	require.Equal(t, expected.siteID, device.Site.Id)
	require.Equal(t, expected.siteDisplay, device.Site.Display)
	if expected.rackID == nil {
		require.Nil(t, device.RackId)
	} else {
		require.NotNil(t, device.RackId)
		require.Equal(t, *expected.rackID, device.RackId.Value)
	}
	if expected.position == nil {
		require.Empty(t, device.Position)
	} else {
		require.Equal(t, strconv.FormatFloat(*expected.position, 'f', -1, 64), device.Position)
	}
	require.Equal(t, expected.face, device.Face)
	require.Equal(t, expected.status, device.Status)
	require.Equal(t, expected.serial, device.Serial)
	require.Equal(t, devicePresenceValue(expected.assetTag), device.AssetTag)
	require.Equal(t, devicePresenceValue(expected.airflow), device.Airflow)
	require.Equal(t, expected.description, device.Description)
	require.Equal(t, expected.comments, device.Comments)
	require.NotNil(t, device.Created)
	require.NotNil(t, device.LastUpdated)
	require.Zero(t, device.InterfaceCount)
}

func requireDeviceRESTPresence(
	t *testing.T,
	device map[string]any,
	expected parityDeviceExpected,
) {
	t.Helper()
	requireRESTDeviceReference(
		t, device["device_type"], expected.deviceTypeID, expected.deviceTypeDisplay,
	)
	requireRESTDeviceReference(t, device["role"], expected.roleID, expected.roleDisplay)
	requireRESTDeviceNullableString(t, device["name"], expected.name)
	requireRESTDeviceReference(t, device["site"], expected.siteID, expected.siteDisplay)
	requireRESTDeviceNullableReference(t, device["rack"], expected.rackID)
	if expected.position == nil {
		require.Nil(t, device["position"])
	} else {
		require.Equal(t, *expected.position, device["position"])
	}
	requireRESTDeviceNullableChoice(t, device["face"], expected.face)
	requireRESTDeviceChoice(t, device["status"], expected.status)
	require.Equal(t, expected.serial, device["serial"])
	requireRESTDeviceNullableString(t, device["asset_tag"], expected.assetTag)
	requireRESTDeviceNullableChoice(t, device["airflow"], devicePresenceValue(expected.airflow))
	require.Equal(t, expected.description, device["description"])
	require.Equal(t, expected.comments, device["comments"])
	require.NotEmpty(t, device["created"])
	require.NotEmpty(t, device["last_updated"])
	require.Equal(t, float64(0), device["interface_count"])
}

func requireRESTDeviceReference(
	t *testing.T,
	actual any,
	id int64,
	display string,
) {
	t.Helper()
	reference, ok := actual.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(id), reference["id"])
	require.Equal(t, display, reference["display"])
}

func requireRESTDeviceNullableReference(t *testing.T, actual any, expected *int64) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	reference, ok := actual.(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(*expected), reference["id"])
	require.NotEmpty(t, reference["display"])
}

func requireRESTDeviceNullableString(t *testing.T, actual any, expected *string) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	require.Equal(t, *expected, actual)
}

func requireRESTDeviceChoice(t *testing.T, actual any, expected string) {
	t.Helper()
	choice, ok := actual.(map[string]any)
	require.True(t, ok)
	require.Equal(t, expected, choice["value"])
	require.NotEmpty(t, choice["label"])
}

func requireRESTDeviceNullableChoice(t *testing.T, actual any, expected string) {
	t.Helper()
	if expected == "" {
		require.Nil(t, actual)
		return
	}
	requireRESTDeviceChoice(t, actual, expected)
}

func devicePresenceString(value string) *string  { return &value }
func devicePresenceFloat(value float64) *float64 { return &value }

func devicePresenceValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
