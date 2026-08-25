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

func TestDeviceTypeScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	manufacturerA := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{"name": "DeviceType Presence Manufacturer A", "slug": "device-type-presence-a"},
		http.StatusCreated,
	)
	manufacturerB := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{"name": "DeviceType Presence Manufacturer B", "slug": "device-type-presence-b"},
		http.StatusCreated,
	)
	manufacturerAID := jsonID(t, manufacturerA["id"])
	manufacturerBID := jsonID(t, manufacturerB["id"])

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/device-types",
		map[string]any{
			"manufacturer": manufacturerAID,
			"model":        "  REST Default Device Type  ",
			"slug":         "  rest-default-device-type  ",
		},
		http.StatusCreated,
	)
	deviceTypeID := jsonID(t, created["id"])
	itemPath := "/api/dcim/device-types/" + strconv.FormatInt(deviceTypeID, 10)
	requireDeviceTypeRESTScalars(
		t, created, manufacturerAID, "REST Default Device Type", "rest-default-device-type",
		"", 1, false, true, nil, "", "",
	)
	createdState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	require.Equal(t, int64(1), createdState.deviceTypeCount)
	require.Equal(t, int64(1), createdState.changeCount)
	require.Zero(t, createdState.deviceCount)
	require.Zero(t, createdState.templateCount)

	grpcRead, err := environment.dcim.GetDeviceType(
		environment.ctx, &dcimv1.GetDeviceTypeRequest{Id: deviceTypeID},
	)
	require.NoError(t, err)
	requireDeviceTypeProtoScalars(
		t, grpcRead.DeviceType, manufacturerAID, "REST Default Device Type",
		"rest-default-device-type", "", "1", false, true, "", "", "",
	)

	concrete, err := environment.dcim.UpdateDeviceType(
		environment.ctx,
		&dcimv1.UpdateDeviceTypeRequest{
			Id: deviceTypeID,
			DeviceType: &dcimv1.DeviceTypeInput{
				Manufacturer:           &manufacturerBID,
				Model:                  pointer("  gRPC Concrete Device Type  "),
				Slug:                   pointer("  grpc-concrete-device-type  "),
				PartNumber:             pointer("  PN-9000  "),
				UHeight:                pointer("2.5"),
				ExcludeFromUtilization: pointer(true),
				IsFullDepth:            pointer(false),
				Airflow:                pointer("front-to-rear"),
				Description:            pointer("  concrete description  "),
				Comments:               pointer("  concrete comments  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: deviceTypeScalarFields()},
		},
	)
	require.NoError(t, err)
	requireDeviceTypeProtoScalars(
		t, concrete.DeviceType, manufacturerBID, "gRPC Concrete Device Type",
		"grpc-concrete-device-type", "PN-9000", "2.5", true, false,
		"front-to-rear", "concrete description", "concrete comments",
	)
	concreteState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, createdState, concreteState)

	restRead := requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireDeviceTypeRESTScalars(
		t, restRead, manufacturerBID, "gRPC Concrete Device Type",
		"grpc-concrete-device-type", "PN-9000", 2.5, true, false,
		pointer("front-to-rear"), "concrete description", "concrete comments",
	)

	replaced := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"manufacturer": manufacturerAID,
			"model":        "  REST Replaced Device Type  ",
			"slug":         "  rest-replaced-device-type  ",
		},
		http.StatusOK,
	)
	requireDeviceTypeRESTScalars(
		t, replaced, manufacturerAID, "REST Replaced Device Type",
		"rest-replaced-device-type", "PN-9000", 1, true, false,
		pointer("front-to-rear"), "concrete description", "concrete comments",
	)
	replacedState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, concreteState, replacedState)
	require.Equal(t, float64(1), replacedState.row.UHeight, "PUT omission must reset only height")

	grpcReplaced, err := environment.dcim.ReplaceDeviceType(
		environment.ctx,
		&dcimv1.ReplaceDeviceTypeRequest{
			Id: deviceTypeID,
			DeviceType: &dcimv1.DeviceTypeInput{
				Manufacturer: &manufacturerBID,
				Model:        pointer("  gRPC Replaced Device Type  "),
				Slug:         pointer("  grpc-replaced-device-type  "),
			},
		},
	)
	require.NoError(t, err)
	requireDeviceTypeProtoScalars(
		t, grpcReplaced.DeviceType, manufacturerBID, "gRPC Replaced Device Type",
		"grpc-replaced-device-type", "PN-9000", "1", true, false,
		"front-to-rear", "concrete description", "concrete comments",
	)
	grpcReplacedState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, replacedState, grpcReplacedState)

	cleared := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"part_number": "", "u_height": 0,
			"exclude_from_utilization": false, "is_full_depth": false,
			"airflow": "", "description": "", "comments": "",
		},
		http.StatusOK,
	)
	requireDeviceTypeRESTScalars(
		t, cleared, manufacturerBID, "gRPC Replaced Device Type",
		"grpc-replaced-device-type", "", 0, false, false, nil, "", "",
	)
	blankState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, grpcReplacedState, blankState)
	require.NotNil(t, blankState.row.Airflow)
	require.Empty(t, *blankState.row.Airflow, "blank airflow must remain distinct in storage")

	nullAirflow := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"airflow": nil}, http.StatusOK,
	)
	require.Nil(t, nullAirflow["airflow"])
	nullState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, blankState, nullState)
	require.Nil(t, nullState.row.Airflow)

	setAirflow, err := environment.dcim.UpdateDeviceType(
		environment.ctx,
		&dcimv1.UpdateDeviceTypeRequest{
			Id:         deviceTypeID,
			DeviceType: &dcimv1.DeviceTypeInput{Airflow: pointer("rear-to-front")},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"airflow"}},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "rear-to-front", setAirflow.DeviceType.Airflow)
	setAirflowState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, nullState, setAirflowState)

	clearedByGRPC, err := environment.dcim.UpdateDeviceType(
		environment.ctx,
		&dcimv1.UpdateDeviceTypeRequest{
			Id: deviceTypeID, DeviceType: &dcimv1.DeviceTypeInput{},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"airflow"}},
		},
	)
	require.NoError(t, err)
	require.Empty(t, clearedByGRPC.DeviceType.Airflow)
	grpcNullState := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
	requireDeviceTypeParityUpdateRecorded(t, setAirflowState, grpcNullState)
	require.Nil(t, grpcNullState.row.Airflow)

	for _, field := range []string{
		"manufacturer", "model", "slug", "part_number", "u_height",
		"exclude_from_utilization", "is_full_depth", "description", "comments",
	} {
		field := field
		t.Run("REST PATCH null rejection/"+field, func(t *testing.T) {
			before := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
			requestJSON(
				t, environment.router, http.MethodPatch, itemPath,
				map[string]any{field: nil}, http.StatusBadRequest,
			)
			require.Equal(t, before, loadParityDeviceTypePresenceState(t, environment, deviceTypeID))
		})
		t.Run("gRPC masked absent rejection/"+field, func(t *testing.T) {
			before := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
			_, updateErr := environment.dcim.UpdateDeviceType(
				environment.ctx,
				&dcimv1.UpdateDeviceTypeRequest{
					Id: deviceTypeID, DeviceType: &dcimv1.DeviceTypeInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireDeviceTypeGRPCInvalid(t, updateErr)
			require.Equal(t, before, loadParityDeviceTypePresenceState(t, environment, deviceTypeID))
		})
	}

	t.Run("required PUT fields omitted", func(t *testing.T) {
		before := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
		requestJSON(t, environment.router, http.MethodPut, itemPath, map[string]any{}, http.StatusBadRequest)
		_, replaceErr := environment.dcim.ReplaceDeviceType(
			environment.ctx,
			&dcimv1.ReplaceDeviceTypeRequest{Id: deviceTypeID, DeviceType: &dcimv1.DeviceTypeInput{}},
		)
		requireDeviceTypeGRPCInvalid(t, replaceErr)
		require.Equal(t, before, loadParityDeviceTypePresenceState(t, environment, deviceTypeID))
	})

	t.Run("unknown Manufacturer and mask fail closed", func(t *testing.T) {
		before := loadParityDeviceTypePresenceState(t, environment, deviceTypeID)
		requestJSON(
			t, environment.router, http.MethodPatch, itemPath,
			map[string]any{"manufacturer": int64(999999)}, http.StatusBadRequest,
		)
		_, updateErr := environment.dcim.UpdateDeviceType(
			environment.ctx,
			&dcimv1.UpdateDeviceTypeRequest{
				Id:         deviceTypeID,
				DeviceType: &dcimv1.DeviceTypeInput{Model: pointer("must not persist")},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
			},
		)
		requireDeviceTypeGRPCInvalid(t, updateErr)
		require.Equal(t, before, loadParityDeviceTypePresenceState(t, environment, deviceTypeID))
	})
}

func deviceTypeScalarFields() []string {
	return []string{
		"manufacturer", "model", "slug", "part_number", "u_height",
		"exclude_from_utilization", "is_full_depth", "airflow", "description", "comments",
	}
}

type parityDeviceTypePresenceState struct {
	row             dcimrow.DeviceTypeRow
	deviceTypeCount int64
	deviceCount     int64
	templateCount   int64
	changeCount     int64
	totalChanges    int64
}

func loadParityDeviceTypePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityDeviceTypePresenceState {
	t.Helper()
	var state parityDeviceTypePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceTypeRow{}).Count(&state.deviceTypeCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.DeviceRow{}).Count(&state.deviceCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.InterfaceTemplateRow{}).Count(&state.templateCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.DeviceTypeObjectType, id,
	).Count(&state.changeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChanges).Error)
	return state
}

func requireDeviceTypeParityUpdateRecorded(
	t *testing.T,
	before parityDeviceTypePresenceState,
	after parityDeviceTypePresenceState,
) {
	t.Helper()
	require.Equal(t, before.deviceTypeCount, after.deviceTypeCount)
	require.Equal(t, before.deviceCount, after.deviceCount)
	require.Equal(t, before.templateCount, after.templateCount)
	require.Zero(t, after.deviceCount, "scalar presence requires zero referencing Devices")
	require.Zero(t, after.templateCount, "scalar presence claims no template behavior")
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChanges+1, after.totalChanges)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireDeviceTypeGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireDeviceTypeProtoScalars(
	t *testing.T,
	deviceType *dcimv1.DeviceType,
	manufacturer int64,
	model string,
	slug string,
	partNumber string,
	uHeight string,
	excluded bool,
	fullDepth bool,
	airflow string,
	description string,
	comments string,
) {
	t.Helper()
	require.NotNil(t, deviceType)
	require.NotNil(t, deviceType.Manufacturer)
	require.Equal(t, manufacturer, deviceType.Manufacturer.Id)
	require.Equal(t, model, deviceType.Model)
	require.Equal(t, slug, deviceType.Slug)
	require.Equal(t, partNumber, deviceType.PartNumber)
	require.Equal(t, uHeight, deviceType.UHeight)
	require.Equal(t, excluded, deviceType.ExcludeFromUtilization)
	require.Equal(t, fullDepth, deviceType.IsFullDepth)
	require.Equal(t, airflow, deviceType.Airflow)
	require.Equal(t, description, deviceType.Description)
	require.Equal(t, comments, deviceType.Comments)
}

func requireDeviceTypeRESTScalars(
	t *testing.T,
	deviceType map[string]any,
	manufacturer int64,
	model string,
	slug string,
	partNumber string,
	uHeight float64,
	excluded bool,
	fullDepth bool,
	airflow *string,
	description string,
	comments string,
) {
	t.Helper()
	manufacturerReference, ok := deviceType["manufacturer"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(manufacturer), manufacturerReference["id"])
	require.Equal(t, model, deviceType["model"])
	require.Equal(t, slug, deviceType["slug"])
	require.Equal(t, partNumber, deviceType["part_number"])
	require.Equal(t, uHeight, deviceType["u_height"])
	require.Equal(t, excluded, deviceType["exclude_from_utilization"])
	require.Equal(t, fullDepth, deviceType["is_full_depth"])
	if airflow == nil {
		require.Nil(t, deviceType["airflow"])
	} else {
		choice, choiceOK := deviceType["airflow"].(map[string]any)
		require.True(t, choiceOK)
		require.Equal(t, *airflow, choice["value"])
	}
	require.Equal(t, description, deviceType["description"])
	require.Equal(t, comments, deviceType["comments"])
}
