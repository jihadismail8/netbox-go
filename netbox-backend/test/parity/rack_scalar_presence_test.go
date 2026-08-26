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

func TestRackScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	site := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/sites",
		map[string]any{
			"name": "Rack Presence Site", "slug": "rack-presence-site",
		},
		http.StatusCreated,
	)
	manufacturer := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{
			"name": "Rack Presence Manufacturer", "slug": "rack-presence-manufacturer",
		},
		http.StatusCreated,
	)
	role := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-roles",
		map[string]any{
			"name": "Rack Presence Role", "slug": "rack-presence-role", "color": "00ff00",
		},
		http.StatusCreated,
	)
	siteID := jsonID(t, site["id"])
	manufacturerID := jsonID(t, manufacturer["id"])
	roleID := jsonID(t, role["id"])
	rackType := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-types",
		map[string]any{
			"manufacturer":  manufacturerID,
			"model":         "Rack Presence Type",
			"slug":          "rack-presence-type",
			"form_factor":   "wall-frame",
			"width":         23,
			"u_height":      24,
			"starting_unit": 3,
			"desc_units":    true,
		},
		http.StatusCreated,
	)
	rackTypeID := jsonID(t, rackType["id"])

	var racksBefore, devicesBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.RackRow{}).Count(&racksBefore).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&devicesBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	require.Zero(t, racksBefore)
	require.Zero(t, devicesBefore, "scalar presence requires zero Devices")
	require.Equal(t, int64(4), changesBefore, "the four public fixture creates each record once")

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/racks",
		map[string]any{"site": siteID, "name": "  REST Default Rack  "},
		http.StatusCreated,
	)
	rackID := jsonID(t, created["id"])
	itemPath := "/api/dcim/racks/" + strconv.FormatInt(rackID, 10)
	requireRackRESTScalars(
		t, created, siteID, "Rack Presence Site", "REST Default Rack",
		nil, nil, "active", nil, "", nil, nil, 19, 42, 1, false, nil, "", "",
	)
	createdState := loadParityRackPresenceState(t, environment, rackID)
	require.Equal(t, racksBefore+1, createdState.rackCount)
	require.Zero(t, createdState.deviceCount)
	require.Equal(t, int64(1), createdState.targetRackChangeCount)
	require.Equal(t, int64(1), createdState.allRackChangeCount)
	require.Zero(t, createdState.deviceChangeCount)
	require.Equal(t, changesBefore+1, createdState.totalChangeCount)

	grpcRead, err := environment.dcim.GetRack(
		environment.ctx, &dcimv1.GetRackRequest{Id: rackID},
	)
	require.NoError(t, err)
	requireRackProtoScalars(
		t, grpcRead.Rack, siteID, "Rack Presence Site", "REST Default Rack",
		"", nil, "active", nil, "", "", "", 19, 42, 1, false, "", "", "",
	)

	grpcCreated, err := environment.dcim.CreateRack(
		environment.ctx,
		&dcimv1.CreateRackRequest{Rack: &dcimv1.RackInput{
			Site: pointer(siteID), Name: pointer("  gRPC Concrete Rack  "),
			FacilityId: pointer("  FAC-GRPC  "), Status: pointer("planned"),
			Role: wrapperspb.Int64(roleID), Serial: pointer("  SERIAL-GRPC  "),
			AssetTag: pointer("  ASSET-GRPC  "), FormFactor: pointer("wall-cabinet"),
			Width: pointer(uint32(23)), UHeight: pointer(uint32(80)),
			StartingUnit: pointer(uint32(32767)), DescUnits: pointer(true),
			Airflow: pointer("rear-to-front"), Description: pointer("  gRPC description  "),
			Comments: pointer("  gRPC comments  "),
		}},
	)
	require.NoError(t, err)
	requireRackProtoScalars(
		t, grpcCreated.Rack, siteID, "Rack Presence Site", "gRPC Concrete Rack",
		"FAC-GRPC", nil, "planned", &roleID, "SERIAL-GRPC", "ASSET-GRPC",
		"wall-cabinet", 23, 80, 32767, true, "rear-to-front", "gRPC description",
		"gRPC comments",
	)
	grpcCreatedREST := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/dcim/racks/"+strconv.FormatInt(grpcCreated.Rack.Id, 10), nil, http.StatusOK,
	)
	requireRackRESTScalars(
		t, grpcCreatedREST, siteID, "Rack Presence Site", "gRPC Concrete Rack",
		rackPresenceString("FAC-GRPC"), nil, "planned", &roleID, "SERIAL-GRPC",
		rackPresenceString("ASSET-GRPC"), rackPresenceString("wall-cabinet"),
		23, 80, 32767, true, rackPresenceString("rear-to-front"),
		"gRPC description", "gRPC comments",
	)
	grpcCreatedState := loadParityRackPresenceState(t, environment, grpcCreated.Rack.Id)
	require.Equal(t, createdState.rackCount+1, grpcCreatedState.rackCount)
	require.Equal(t, int64(1), grpcCreatedState.targetRackChangeCount)
	require.Equal(t, createdState.allRackChangeCount+1, grpcCreatedState.allRackChangeCount)
	require.Equal(t, createdState.totalChangeCount+1, grpcCreatedState.totalChangeCount)
	require.Zero(t, grpcCreatedState.deviceCount)
	require.Zero(t, grpcCreatedState.deviceChangeCount)

	beforeConcrete := loadParityRackPresenceState(t, environment, rackID)
	concrete, err := environment.dcim.UpdateRack(
		environment.ctx,
		&dcimv1.UpdateRackRequest{
			Id: rackID,
			Rack: &dcimv1.RackInput{
				FacilityId: pointer("  FAC-REST-GRPC  "), Status: pointer("planned"),
				Role: wrapperspb.Int64(roleID), Serial: pointer("  SERIAL-I9  "),
				AssetTag: pointer("  ASSET-I9  "), FormFactor: pointer("wall-cabinet"),
				Width: pointer(uint32(23)), UHeight: pointer(uint32(100)),
				StartingUnit: pointer(uint32(32767)), DescUnits: pointer(true),
				Airflow: pointer("rear-to-front"), Description: pointer("  concrete description  "),
				Comments: pointer("  concrete comments  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{
				"facility_id", "status", "role", "serial", "asset_tag", "form_factor",
				"width", "u_height", "starting_unit", "desc_units", "airflow",
				"description", "comments",
			}},
		},
	)
	require.NoError(t, err)
	requireRackProtoScalars(
		t, concrete.Rack, siteID, "Rack Presence Site", "REST Default Rack",
		"FAC-REST-GRPC", nil, "planned", &roleID, "SERIAL-I9", "ASSET-I9",
		"wall-cabinet", 23, 100, 32767, true, "rear-to-front",
		"concrete description", "concrete comments",
	)
	concreteState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, beforeConcrete, concreteState)

	owned := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"site": siteID, "name": "REST Default Rack", "facility_id": "FAC-REST-GRPC",
			"rack_type": rackTypeID, "form_factor": "4-post-cabinet", "width": 19,
			"u_height": 80, "starting_unit": 7, "desc_units": false,
		},
		http.StatusOK,
	)
	requireRackRESTScalars(
		t, owned, siteID, "Rack Presence Site", "REST Default Rack",
		rackPresenceString("FAC-REST-GRPC"), &rackTypeID, "planned", &roleID,
		"SERIAL-I9", rackPresenceString("ASSET-I9"), rackPresenceString("wall-frame"),
		23, 24, 3, true, rackPresenceString("rear-to-front"),
		"concrete description", "concrete comments",
	)
	ownedState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, concreteState, ownedState)

	replaced := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{"site": siteID, "name": "  REST PUT Rack  "},
		http.StatusOK,
	)
	requireRackRESTScalars(
		t, replaced, siteID, "Rack Presence Site", "REST PUT Rack",
		nil, nil, "planned", &roleID, "SERIAL-I9", rackPresenceString("ASSET-I9"),
		rackPresenceString("wall-frame"), 23, 24, 3, true,
		rackPresenceString("rear-to-front"), "concrete description", "concrete comments",
	)
	replacedState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, ownedState, replacedState)

	replacedByGRPC, err := environment.dcim.ReplaceRack(
		environment.ctx,
		&dcimv1.ReplaceRackRequest{
			Id:   rackID,
			Rack: &dcimv1.RackInput{Site: pointer(siteID), Name: pointer("  gRPC PUT Rack  ")},
		},
	)
	require.NoError(t, err)
	requireRackProtoScalars(
		t, replacedByGRPC.Rack, siteID, "Rack Presence Site", "gRPC PUT Rack",
		"", nil, "planned", &roleID, "SERIAL-I9", "ASSET-I9", "wall-frame",
		23, 24, 3, true, "rear-to-front", "concrete description", "concrete comments",
	)
	replacedByGRPCState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, replacedState, replacedByGRPCState)

	cleared := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"facility_id": nil, "role": nil, "serial": "", "asset_tag": nil,
			"form_factor": "", "desc_units": false, "airflow": nil,
			"description": "", "comments": "",
		},
		http.StatusOK,
	)
	requireRackRESTScalars(
		t, cleared, siteID, "Rack Presence Site", "gRPC PUT Rack",
		nil, nil, "planned", nil, "", nil, rackPresenceString(""),
		23, 24, 3, false, rackPresenceString(""), "", "",
	)
	clearedState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, replacedByGRPCState, clearedState)
	require.NotNil(t, clearedState.row.Airflow, "REST airflow:null stores explicit blank, not SQL NULL")
	require.Equal(t, "", *clearedState.row.Airflow)
	grpcRead, err = environment.dcim.GetRack(environment.ctx, &dcimv1.GetRackRequest{Id: rackID})
	require.NoError(t, err)
	requireRackProtoScalars(
		t, grpcRead.Rack, siteID, "Rack Presence Site", "gRPC PUT Rack",
		"", nil, "planned", nil, "", "", "", 23, 24, 3, false, "", "", "",
	)

	blank := ""
	blankAirflow, err := environment.dcim.UpdateRack(
		environment.ctx,
		&dcimv1.UpdateRackRequest{
			Id: rackID, Rack: &dcimv1.RackInput{Airflow: &blank},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"airflow"}},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "", blankAirflow.Rack.Airflow)
	blankAirflowState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, clearedState, blankAirflowState)
	require.NotNil(t, blankAirflowState.row.Airflow)
	require.Equal(t, "", *blankAirflowState.row.Airflow)

	reassigned, err := environment.dcim.UpdateRack(
		environment.ctx,
		&dcimv1.UpdateRackRequest{
			Id: rackID, Rack: &dcimv1.RackInput{RackType: wrapperspb.Int64(rackTypeID)},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"rack_type"}},
		},
	)
	require.NoError(t, err)
	require.NotNil(t, reassigned.Rack.RackTypeId)
	require.Equal(t, rackTypeID, reassigned.Rack.RackTypeId.Value)
	reassignedState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, blankAirflowState, reassignedState)

	clearedType, err := environment.dcim.UpdateRack(
		environment.ctx,
		&dcimv1.UpdateRackRequest{
			Id: rackID, Rack: &dcimv1.RackInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"rack_type"}},
		},
	)
	require.NoError(t, err)
	requireRackProtoScalars(
		t, clearedType.Rack, siteID, "Rack Presence Site", "gRPC PUT Rack",
		"", nil, "planned", nil, "", "", "wall-frame", 23, 24, 3, true,
		"", "", "",
	)
	clearedTypeState := loadParityRackPresenceState(t, environment, rackID)
	requireRackParityMutationRecorded(t, reassignedState, clearedTypeState)

	t.Run("required fields omitted", func(t *testing.T) {
		before := loadParityRackPresenceState(t, environment, rackID)
		requestJSON(t, environment.router, http.MethodPost, "/api/dcim/racks", map[string]any{}, http.StatusBadRequest)
		requestJSON(t, environment.router, http.MethodPut, itemPath, map[string]any{}, http.StatusBadRequest)
		_, createErr := environment.dcim.CreateRack(
			environment.ctx, &dcimv1.CreateRackRequest{Rack: &dcimv1.RackInput{}},
		)
		requireRackGRPCInvalid(t, createErr)
		_, replaceErr := environment.dcim.ReplaceRack(
			environment.ctx,
			&dcimv1.ReplaceRackRequest{Id: rackID, Rack: &dcimv1.RackInput{}},
		)
		requireRackGRPCInvalid(t, replaceErr)
		require.Equal(t, before, loadParityRackPresenceState(t, environment, rackID))
	})

	for _, rejected := range []struct {
		name string
		body map[string]any
	}{
		{name: "null Site", body: map[string]any{"site": nil}},
		{name: "null name", body: map[string]any{"name": nil}},
		{name: "null status", body: map[string]any{"status": nil}},
		{name: "null serial", body: map[string]any{"serial": nil}},
		{name: "null width", body: map[string]any{"width": nil}},
		{name: "null U height", body: map[string]any{"u_height": nil}},
		{name: "null starting unit", body: map[string]any{"starting_unit": nil}},
		{name: "null descending units", body: map[string]any{"desc_units": nil}},
		{name: "null description", body: map[string]any{"description": nil}},
		{name: "null comments", body: map[string]any{"comments": nil}},
		{name: "whitespace status", body: map[string]any{"status": " active "}},
		{name: "whitespace form factor", body: map[string]any{"form_factor": " wall-frame "}},
		{name: "whitespace airflow", body: map[string]any{"airflow": " front-to-rear "}},
		{name: "starting unit overflow", body: map[string]any{"starting_unit": 32768}},
		{name: "unknown Site", body: map[string]any{"site": 999997}},
		{name: "unknown RackType", body: map[string]any{"rack_type": 999998}},
		{name: "unknown RackRole", body: map[string]any{"role": 999999}},
	} {
		rejected := rejected
		t.Run("REST rejection/"+rejected.name, func(t *testing.T) {
			before := loadParityRackPresenceState(t, environment, rackID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				rejected.body, http.StatusBadRequest,
			)
			require.Equal(t, before, loadParityRackPresenceState(t, environment, rackID))
		})
	}

	for _, path := range []string{
		"site", "name", "status", "serial", "width", "u_height", "starting_unit",
		"desc_units", "airflow", "description", "comments",
	} {
		path := path
		t.Run("gRPC masked absent/"+path, func(t *testing.T) {
			before := loadParityRackPresenceState(t, environment, rackID)
			_, updateErr := environment.dcim.UpdateRack(
				environment.ctx,
				&dcimv1.UpdateRackRequest{
					Id: rackID, Rack: &dcimv1.RackInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{path}},
				},
			)
			requireRackGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityRackPresenceState(t, environment, rackID))
		})
	}

	t.Run("gRPC validation and unknown mask fail closed", func(t *testing.T) {
		before := loadParityRackPresenceState(t, environment, rackID)
		for _, test := range []struct {
			path  string
			input *dcimv1.RackInput
		}{
			{path: "status", input: &dcimv1.RackInput{Status: pointer(" active ")}},
			{path: "form_factor", input: &dcimv1.RackInput{FormFactor: pointer(" wall-frame ")}},
			{path: "airflow", input: &dcimv1.RackInput{Airflow: pointer(" front-to-rear ")}},
			{path: "starting_unit", input: &dcimv1.RackInput{StartingUnit: pointer(uint32(32768))}},
			{path: "site", input: &dcimv1.RackInput{Site: pointer(int64(999997))}},
			{path: "rack_type", input: &dcimv1.RackInput{RackType: wrapperspb.Int64(999998)}},
			{path: "role", input: &dcimv1.RackInput{Role: wrapperspb.Int64(999999)}},
		} {
			_, updateErr := environment.dcim.UpdateRack(
				environment.ctx,
				&dcimv1.UpdateRackRequest{
					Id: rackID, Rack: test.input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{test.path}},
				},
			)
			requireRackGRPCInvalid(t, updateErr)
		}
		_, maskErr := environment.dcim.UpdateRack(
			environment.ctx,
			&dcimv1.UpdateRackRequest{
				Id: rackID, Rack: &dcimv1.RackInput{Description: pointer("must not persist")},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
			},
		)
		requireRackGRPCInvalid(t, maskErr)
		require.Equal(t, before, loadParityRackPresenceState(t, environment, rackID))
	})
}

type parityRackPresenceState struct {
	row                   dcimrow.RackRow
	rackCount             int64
	deviceCount           int64
	targetRackChangeCount int64
	allRackChangeCount    int64
	deviceChangeCount     int64
	totalChangeCount      int64
}

func loadParityRackPresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityRackPresenceState {
	t.Helper()
	var state parityRackPresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackRow{}).Count(&state.rackCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackObjectType, id,
	).Count(&state.targetRackChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.RackObjectType,
	).Count(&state.allRackChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.DeviceObjectType,
	).Count(&state.deviceChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireRackParityMutationRecorded(
	t *testing.T,
	before parityRackPresenceState,
	after parityRackPresenceState,
) {
	t.Helper()
	require.Equal(t, before.rackCount, after.rackCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Zero(t, after.deviceCount, "scalar presence requires zero Devices")
	require.Equal(t, before.targetRackChangeCount+1, after.targetRackChangeCount)
	require.Equal(t, before.allRackChangeCount+1, after.allRackChangeCount)
	require.Equal(t, before.deviceChangeCount, after.deviceChangeCount)
	require.Zero(t, after.deviceChangeCount, "scalar presence records no Device changes")
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireRackGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireRackProtoScalars(
	t *testing.T,
	rack *dcimv1.Rack,
	siteID int64,
	siteDisplay string,
	name string,
	facilityID string,
	rackTypeID *int64,
	statusValue string,
	roleID *int64,
	serial string,
	assetTag string,
	formFactor string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	airflow string,
	description string,
	comments string,
) {
	t.Helper()
	require.NotNil(t, rack)
	require.NotNil(t, rack.Site)
	require.Equal(t, siteID, rack.Site.Id)
	require.Equal(t, siteDisplay, rack.Site.Display)
	require.Equal(t, name, rack.Name)
	require.Equal(t, facilityID, rack.FacilityId)
	if rackTypeID == nil {
		require.Nil(t, rack.RackTypeId)
	} else {
		require.NotNil(t, rack.RackTypeId)
		require.Equal(t, *rackTypeID, rack.RackTypeId.Value)
	}
	require.Equal(t, statusValue, rack.Status)
	if roleID == nil {
		require.Nil(t, rack.RoleId)
	} else {
		require.NotNil(t, rack.RoleId)
		require.Equal(t, *roleID, rack.RoleId.Value)
	}
	require.Equal(t, serial, rack.Serial)
	require.Equal(t, assetTag, rack.AssetTag)
	require.Equal(t, formFactor, rack.FormFactor)
	require.Equal(t, width, rack.Width)
	require.Equal(t, uHeight, rack.UHeight)
	require.Equal(t, startingUnit, rack.StartingUnit)
	require.Equal(t, descUnits, rack.DescUnits)
	require.Equal(t, airflow, rack.Airflow)
	require.Equal(t, description, rack.Description)
	require.Equal(t, comments, rack.Comments)
	require.NotNil(t, rack.Created)
	require.NotNil(t, rack.LastUpdated)
	require.Zero(t, rack.DeviceCount)
}

func requireRackRESTScalars(
	t *testing.T,
	rack map[string]any,
	siteID int64,
	siteDisplay string,
	name string,
	facilityID *string,
	rackTypeID *int64,
	statusValue string,
	roleID *int64,
	serial string,
	assetTag *string,
	formFactor *string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	airflow *string,
	description string,
	comments string,
) {
	t.Helper()
	site, ok := rack["site"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(siteID), site["id"])
	require.Equal(t, siteDisplay, site["display"])
	require.Equal(t, name, rack["name"])
	requireRackRESTNullableString(t, rack["facility_id"], facilityID)
	requireRackRESTNullableReference(t, rack["rack_type"], rackTypeID)
	requireRackRESTStringChoice(t, rack["status"], statusValue)
	requireRackRESTNullableReference(t, rack["role"], roleID)
	require.Equal(t, serial, rack["serial"])
	requireRackRESTNullableString(t, rack["asset_tag"], assetTag)
	requireRackRESTNullableChoice(t, rack["form_factor"], formFactor)
	widthChoice, ok := rack["width"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(width), widthChoice["value"])
	require.NotEmpty(t, widthChoice["label"])
	require.Equal(t, float64(uHeight), rack["u_height"])
	require.Equal(t, float64(startingUnit), rack["starting_unit"])
	require.Equal(t, descUnits, rack["desc_units"])
	requireRackRESTNullableChoice(t, rack["airflow"], airflow)
	require.Equal(t, description, rack["description"])
	require.Equal(t, comments, rack["comments"])
}

func requireRackRESTNullableString(t *testing.T, actual any, expected *string) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual)
		return
	}
	require.Equal(t, *expected, actual)
}

func requireRackRESTNullableReference(t *testing.T, actual any, expected *int64) {
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

func requireRackRESTStringChoice(t *testing.T, actual any, expected string) {
	t.Helper()
	choice, ok := actual.(map[string]any)
	require.True(t, ok)
	require.Equal(t, expected, choice["value"])
	require.NotEmpty(t, choice["label"])
}

func requireRackRESTNullableChoice(t *testing.T, actual any, expected *string) {
	t.Helper()
	if expected == nil || *expected == "" {
		require.Nil(t, actual)
		return
	}
	requireRackRESTStringChoice(t, actual, *expected)
}

func rackPresenceString(value string) *string { return &value }
