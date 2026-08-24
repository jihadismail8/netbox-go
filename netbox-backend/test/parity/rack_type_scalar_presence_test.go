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

func TestRackTypeScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})
	manufacturer := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/manufacturers",
		map[string]any{"name": "RackType Presence Manufacturer", "slug": "rack-type-presence-manufacturer"},
		http.StatusCreated,
	)
	manufacturerID := jsonID(t, manufacturer["id"])

	var rowsBefore, racksBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.RackTypeRow{}).Count(&rowsBefore).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackRow{}).Count(&racksBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	require.Zero(t, racksBefore, "the scalar-presence parity proof requires zero referencing Racks")

	for _, test := range []struct {
		name string
		body map[string]any
	}{
		{name: "required fields omitted", body: map[string]any{}},
		{name: "manufacturer null", body: rackTypeRESTCreateBody(nil, "Rejected Manufacturer", "rejected-manufacturer", "4-post-cabinet")},
		{name: "model null", body: rackTypeRESTCreateBody(manufacturerID, nil, "rejected-model", "4-post-cabinet")},
		{name: "slug null", body: rackTypeRESTCreateBody(manufacturerID, "Rejected Slug", nil, "4-post-cabinet")},
		{name: "form factor null", body: rackTypeRESTCreateBody(manufacturerID, "Rejected Form Factor", "rejected-form-factor", nil)},
		{name: "width null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Width", "rejected-width", "4-post-cabinet", "width", nil)},
		{name: "u height null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected U Height", "rejected-u-height", "4-post-cabinet", "u_height", nil)},
		{name: "starting unit null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Starting Unit", "rejected-starting-unit", "4-post-cabinet", "starting_unit", nil)},
		{name: "desc units null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Desc Units", "rejected-desc-units", "4-post-cabinet", "desc_units", nil)},
		{name: "description null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Description", "rejected-description", "4-post-cabinet", "description", nil)},
		{name: "comments null", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Comments", "rejected-comments", "4-post-cabinet", "comments", nil)},
		{name: "blank model", body: rackTypeRESTCreateBody(manufacturerID, "", "rejected-blank-model", "4-post-cabinet")},
		{name: "blank slug", body: rackTypeRESTCreateBody(manufacturerID, "Rejected Blank Slug", "", "4-post-cabinet")},
		{name: "blank form factor", body: rackTypeRESTCreateBody(manufacturerID, "Rejected Blank Form Factor", "rejected-blank-form-factor", "")},
		{name: "zero width", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Zero Width", "rejected-zero-width", "4-post-cabinet", "width", 0)},
		{name: "zero u height", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Zero U Height", "rejected-zero-u-height", "4-post-cabinet", "u_height", 0)},
		{name: "zero starting unit", body: rackTypeRESTCreateBodyWith(manufacturerID, "Rejected Zero Starting Unit", "rejected-zero-starting-unit", "4-post-cabinet", "starting_unit", 0)},
	} {
		test := test
		t.Run("REST create/"+test.name, func(t *testing.T) {
			requestJSON(t, environment.router, http.MethodPost, "/api/dcim/rack-types", test.body, http.StatusBadRequest)
		})
	}

	blank := ""
	zero := uint32(0)
	for _, test := range []struct {
		name  string
		input *dcimv1.RackTypeInput
	}{
		{name: "required fields omitted", input: &dcimv1.RackTypeInput{}},
		{name: "blank model", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: &blank, Slug: pointer("grpc-blank-model"), FormFactor: pointer("4-post-cabinet")}},
		{name: "blank slug", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: pointer("gRPC Blank Slug"), Slug: &blank, FormFactor: pointer("4-post-cabinet")}},
		{name: "blank form factor", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: pointer("gRPC Blank Form Factor"), Slug: pointer("grpc-blank-form-factor"), FormFactor: &blank}},
		{name: "zero width", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: pointer("gRPC Zero Width"), Slug: pointer("grpc-zero-width"), FormFactor: pointer("4-post-cabinet"), Width: &zero}},
		{name: "zero u height", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: pointer("gRPC Zero U Height"), Slug: pointer("grpc-zero-u-height"), FormFactor: pointer("4-post-cabinet"), UHeight: &zero}},
		{name: "zero starting unit", input: &dcimv1.RackTypeInput{Manufacturer: &manufacturerID, Model: pointer("gRPC Zero Starting Unit"), Slug: pointer("grpc-zero-starting-unit"), FormFactor: pointer("4-post-cabinet"), StartingUnit: &zero}},
	} {
		test := test
		t.Run("gRPC create/"+test.name, func(t *testing.T) {
			_, err := environment.dcim.CreateRackType(
				environment.ctx, &dcimv1.CreateRackTypeRequest{RackType: test.input},
			)
			requireRackTypeGRPCInvalid(t, err)
		})
	}

	var rowsAfterRejections, racksAfterRejections, changesAfterRejections int64
	require.NoError(t, environment.db.Model(&dcimrow.RackTypeRow{}).Count(&rowsAfterRejections).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackRow{}).Count(&racksAfterRejections).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfterRejections).Error)
	require.Equal(t, rowsBefore, rowsAfterRejections)
	require.Equal(t, racksBefore, racksAfterRejections)
	require.Equal(t, changesBefore, changesAfterRejections)

	createdByREST := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-types",
		map[string]any{
			"manufacturer": manufacturerID, "model": "  REST Default Rack Type  ",
			"slug": "  rest-default-rack-type  ", "form_factor": "4-post-cabinet",
		},
		http.StatusCreated,
	)
	rackTypeID := jsonID(t, createdByREST["id"])
	requireRackTypeRESTScalars(t, createdByREST, manufacturerID, "REST Default Rack Type", "rest-default-rack-type", "4-post-cabinet", 19, 42, 1, false, "", "")
	createdState := loadParityRackTypePresenceState(t, environment, rackTypeID)
	require.Equal(t, rowsAfterRejections+1, createdState.rackTypeCount)
	require.Equal(t, int64(1), createdState.rackTypeChangeCount)
	require.Equal(t, changesAfterRejections+1, createdState.totalChangeCount)
	require.Zero(t, createdState.rackCount)
	require.Zero(t, createdState.rackChangeCount)

	grpcRead, err := environment.dcim.GetRackType(
		environment.ctx, &dcimv1.GetRackTypeRequest{Id: rackTypeID},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, grpcRead.RackType, manufacturerID, "REST Default Rack Type", "rest-default-rack-type", "4-post-cabinet", 19, 42, 1, false, "", "")

	createdByGRPC, err := environment.dcim.CreateRackType(
		environment.ctx,
		&dcimv1.CreateRackTypeRequest{RackType: &dcimv1.RackTypeInput{
			Manufacturer: &manufacturerID, Model: pointer("  gRPC Concrete Rack Type  "),
			Slug: pointer("  grpc-concrete-rack-type  "), FormFactor: pointer("wall-cabinet"),
			Width: pointer(uint32(23)), UHeight: pointer(uint32(48)), StartingUnit: pointer(uint32(2)),
			DescUnits: pointer(true), Description: pointer("  gRPC description  "),
			Comments: pointer("  gRPC comments  "),
		}},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, createdByGRPC.RackType, manufacturerID, "gRPC Concrete Rack Type", "grpc-concrete-rack-type", "wall-cabinet", 23, 48, 2, true, "gRPC description", "gRPC comments")
	grpcCreatedRESTRead := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/dcim/rack-types/"+strconv.FormatInt(createdByGRPC.RackType.Id, 10), nil, http.StatusOK,
	)
	requireRackTypeRESTScalars(t, grpcCreatedRESTRead, manufacturerID, "gRPC Concrete Rack Type", "grpc-concrete-rack-type", "wall-cabinet", 23, 48, 2, true, "gRPC description", "gRPC comments")
	beforeRESTPut := loadParityRackTypePresenceState(t, environment, rackTypeID)

	itemPath := "/api/dcim/rack-types/" + strconv.FormatInt(rackTypeID, 10)
	replacedByREST := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"manufacturer": manufacturerID, "model": "  REST Replaced Rack Type  ",
			"slug": "  rest-replaced-rack-type  ",
		},
		http.StatusOK,
	)
	requireRackTypeRESTScalars(t, replacedByREST, manufacturerID, "REST Replaced Rack Type", "rest-replaced-rack-type", "4-post-cabinet", 19, 42, 1, false, "", "")
	afterRESTPut := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, beforeRESTPut, afterRESTPut)

	replacedByGRPC, err := environment.dcim.ReplaceRackType(
		environment.ctx,
		&dcimv1.ReplaceRackTypeRequest{Id: rackTypeID, RackType: &dcimv1.RackTypeInput{
			Manufacturer: &manufacturerID, Model: pointer("  gRPC Replaced Rack Type  "),
			Slug: pointer("  grpc-replaced-rack-type  "),
		}},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, replacedByGRPC.RackType, manufacturerID, "gRPC Replaced Rack Type", "grpc-replaced-rack-type", "4-post-cabinet", 19, 42, 1, false, "", "")
	afterGRPCPut := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterRESTPut, afterGRPCPut)

	t.Run("REST PUT required fields omitted", func(t *testing.T) {
		before := loadParityRackTypePresenceState(t, environment, rackTypeID)
		response := requestJSON(
			t, environment.router, http.MethodPut, itemPath, map[string]any{}, http.StatusBadRequest,
		)
		require.Equal(t, map[string]any{
			"manufacturer": []any{"This field is required."},
			"model":        []any{"This field is required."},
			"slug":         []any{"This field is required."},
		}, response)
		require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
	})

	restOmission := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"description": "  REST omission proof  "}, http.StatusOK,
	)
	requireRackTypeRESTScalars(
		t, restOmission, manufacturerID, "gRPC Replaced Rack Type", "grpc-replaced-rack-type",
		"4-post-cabinet", 19, 42, 1, false, "REST omission proof", "",
	)
	afterRESTOmission := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterGRPCPut, afterRESTOmission)
	grpcRead, err = environment.dcim.GetRackType(
		environment.ctx, &dcimv1.GetRackTypeRequest{Id: rackTypeID},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(
		t, grpcRead.RackType, manufacturerID, "gRPC Replaced Rack Type", "grpc-replaced-rack-type",
		"4-post-cabinet", 19, 42, 1, false, "REST omission proof", "",
	)

	grpcOmission, err := environment.dcim.UpdateRackType(
		environment.ctx,
		&dcimv1.UpdateRackTypeRequest{
			Id: rackTypeID,
			RackType: &dcimv1.RackTypeInput{
				Comments: pointer("  gRPC omission proof  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"comments"}},
		},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(
		t, grpcOmission.RackType, manufacturerID, "gRPC Replaced Rack Type", "grpc-replaced-rack-type",
		"4-post-cabinet", 19, 42, 1, false, "REST omission proof", "gRPC omission proof",
	)
	afterGRPCOmission := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterRESTOmission, afterGRPCOmission)
	restRead := requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireRackTypeRESTScalars(
		t, restRead, manufacturerID, "gRPC Replaced Rack Type", "grpc-replaced-rack-type",
		"4-post-cabinet", 19, 42, 1, false, "REST omission proof", "gRPC omission proof",
	)

	restNullFields := []string{
		"manufacturer", "model", "slug", "form_factor", "width", "u_height",
		"starting_unit", "desc_units", "description", "comments",
	}
	for _, field := range restNullFields {
		field := field
		t.Run("REST PATCH null/"+field, func(t *testing.T) {
			before := loadParityRackTypePresenceState(t, environment, rackTypeID)
			requestJSON(t, environment.router, http.MethodPatch, itemPath, map[string]any{field: nil}, http.StatusBadRequest)
			require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
		})
		t.Run("REST PUT null/"+field, func(t *testing.T) {
			before := loadParityRackTypePresenceState(t, environment, rackTypeID)
			body := map[string]any{
				"manufacturer": manufacturerID, "model": "gRPC Replaced Rack Type",
				"slug": "grpc-replaced-rack-type",
			}
			body[field] = nil
			requestJSON(t, environment.router, http.MethodPut, itemPath, body, http.StatusBadRequest)
			require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
		})
	}

	for _, test := range []struct {
		name string
		body map[string]any
	}{
		{name: "blank model", body: map[string]any{"model": ""}},
		{name: "blank slug", body: map[string]any{"slug": ""}},
		{name: "blank form factor", body: map[string]any{"form_factor": ""}},
		{name: "untrimmed form factor", body: map[string]any{"form_factor": " 4-post-cabinet "}},
		{name: "zero width", body: map[string]any{"width": 0}},
		{name: "zero u height", body: map[string]any{"u_height": 0}},
		{name: "zero starting unit", body: map[string]any{"starting_unit": 0}},
	} {
		test := test
		t.Run("REST PATCH invalid/"+test.name, func(t *testing.T) {
			before := loadParityRackTypePresenceState(t, environment, rackTypeID)
			requestJSON(t, environment.router, http.MethodPatch, itemPath, test.body, http.StatusBadRequest)
			require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
		})
	}

	for _, input := range []*dcimv1.RackTypeInput{
		{},
		{Manufacturer: &manufacturerID, Slug: pointer("grpc-replaced-rack-type")},
		{Manufacturer: &manufacturerID, Model: pointer("gRPC Replaced Rack Type")},
		{Model: pointer("gRPC Replaced Rack Type"), Slug: pointer("grpc-replaced-rack-type")},
	} {
		before := loadParityRackTypePresenceState(t, environment, rackTypeID)
		_, err := environment.dcim.ReplaceRackType(
			environment.ctx, &dcimv1.ReplaceRackTypeRequest{Id: rackTypeID, RackType: input},
		)
		requireRackTypeGRPCInvalid(t, err)
		require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
	}

	for _, field := range restNullFields {
		field := field
		t.Run("gRPC FieldMask null/"+field, func(t *testing.T) {
			before := loadParityRackTypePresenceState(t, environment, rackTypeID)
			_, err := environment.dcim.UpdateRackType(
				environment.ctx,
				&dcimv1.UpdateRackTypeRequest{
					Id: rackTypeID, RackType: &dcimv1.RackTypeInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireRackTypeGRPCInvalid(t, err)
			require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
		})
	}
	t.Run("gRPC unknown mask path fails closed", func(t *testing.T) {
		before := loadParityRackTypePresenceState(t, environment, rackTypeID)
		_, err := environment.dcim.UpdateRackType(
			environment.ctx,
			&dcimv1.UpdateRackTypeRequest{
				Id: rackTypeID, RackType: &dcimv1.RackTypeInput{Model: pointer("must not persist")},
				UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unknown"}},
			},
		)
		requireRackTypeGRPCInvalid(t, err)
		require.Equal(t, before, loadParityRackTypePresenceState(t, environment, rackTypeID))
	})

	restConcrete := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"manufacturer": manufacturerID, "model": "  REST Patched Rack Type  ",
			"slug": "  rest-patched-rack-type  ", "form_factor": "wall-frame",
			"width": 21, "u_height": 50, "starting_unit": 3, "desc_units": true,
			"description": "  REST patched description  ", "comments": "  REST patched comments  ",
		},
		http.StatusOK,
	)
	requireRackTypeRESTScalars(t, restConcrete, manufacturerID, "REST Patched Rack Type", "rest-patched-rack-type", "wall-frame", 21, 50, 3, true, "REST patched description", "REST patched comments")
	afterRESTConcrete := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterGRPCOmission, afterRESTConcrete)
	grpcRead, err = environment.dcim.GetRackType(environment.ctx, &dcimv1.GetRackTypeRequest{Id: rackTypeID})
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, grpcRead.RackType, manufacturerID, "REST Patched Rack Type", "rest-patched-rack-type", "wall-frame", 21, 50, 3, true, "REST patched description", "REST patched comments")

	grpcConcrete, err := environment.dcim.UpdateRackType(
		environment.ctx,
		&dcimv1.UpdateRackTypeRequest{
			Id: rackTypeID,
			RackType: &dcimv1.RackTypeInput{
				Manufacturer: &manufacturerID, Model: pointer("  gRPC Patched Rack Type  "),
				Slug: pointer("  grpc-patched-rack-type  "), FormFactor: pointer("wall-cabinet"),
				Width: pointer(uint32(23)), UHeight: pointer(uint32(52)), StartingUnit: pointer(uint32(4)),
				DescUnits: pointer(true), Description: pointer("  gRPC patched description  "),
				Comments: pointer("  gRPC patched comments  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: restNullFields},
		},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, grpcConcrete.RackType, manufacturerID, "gRPC Patched Rack Type", "grpc-patched-rack-type", "wall-cabinet", 23, 52, 4, true, "gRPC patched description", "gRPC patched comments")
	afterGRPCConcrete := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterRESTConcrete, afterGRPCConcrete)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireRackTypeRESTScalars(t, restRead, manufacturerID, "gRPC Patched Rack Type", "grpc-patched-rack-type", "wall-cabinet", 23, 52, 4, true, "gRPC patched description", "gRPC patched comments")

	restCleared := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"desc_units": false, "description": "", "comments": ""}, http.StatusOK,
	)
	requireRackTypeRESTScalars(t, restCleared, manufacturerID, "gRPC Patched Rack Type", "grpc-patched-rack-type", "wall-cabinet", 23, 52, 4, false, "", "")
	afterRESTClear := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterGRPCConcrete, afterRESTClear)

	grpcCleared, err := environment.dcim.UpdateRackType(
		environment.ctx,
		&dcimv1.UpdateRackTypeRequest{
			Id: rackTypeID,
			RackType: &dcimv1.RackTypeInput{
				DescUnits: pointer(false), Description: &blank, Comments: &blank,
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"desc_units", "description", "comments"}},
		},
	)
	require.NoError(t, err)
	requireRackTypeProtoScalars(t, grpcCleared.RackType, manufacturerID, "gRPC Patched Rack Type", "grpc-patched-rack-type", "wall-cabinet", 23, 52, 4, false, "", "")
	afterGRPCClear := loadParityRackTypePresenceState(t, environment, rackTypeID)
	requireRackTypeParityUpdateRecorded(t, afterRESTClear, afterGRPCClear)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireRackTypeRESTScalars(t, restRead, manufacturerID, "gRPC Patched Rack Type", "grpc-patched-rack-type", "wall-cabinet", 23, 52, 4, false, "", "")
}

type parityRackTypePresenceState struct {
	row                 dcimrow.RackTypeRow
	rackTypeCount       int64
	rackCount           int64
	rackTypeChangeCount int64
	rackChangeCount     int64
	totalChangeCount    int64
}

func loadParityRackTypePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityRackTypePresenceState {
	t.Helper()
	var state parityRackTypePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackTypeRow{}).Count(&state.rackTypeCount).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackRow{}).Count(&state.rackCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackTypeObjectType, id,
	).Count(&state.rackTypeChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ?", domaindcim.RackObjectType,
	).Count(&state.rackChangeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireRackTypeParityUpdateRecorded(
	t *testing.T,
	before parityRackTypePresenceState,
	after parityRackTypePresenceState,
) {
	t.Helper()
	require.Equal(t, before.rackTypeCount, after.rackTypeCount)
	require.Equal(t, before.rackCount, after.rackCount)
	require.Zero(t, after.rackCount, "the scalar-presence proof requires zero referencing Racks")
	require.Equal(t, before.rackTypeChangeCount+1, after.rackTypeChangeCount)
	require.Equal(t, before.rackChangeCount, after.rackChangeCount)
	require.Zero(t, after.rackChangeCount, "the scalar-presence proof does not claim propagation")
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireRackTypeGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

func requireRackTypeProtoScalars(
	t *testing.T,
	rackType *dcimv1.RackType,
	manufacturer int64,
	model string,
	slug string,
	formFactor string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	description string,
	comments string,
) {
	t.Helper()
	require.NotNil(t, rackType)
	require.NotNil(t, rackType.Manufacturer)
	require.Equal(t, manufacturer, rackType.Manufacturer.Id)
	require.Equal(t, model, rackType.Model)
	require.Equal(t, slug, rackType.Slug)
	require.Equal(t, formFactor, rackType.FormFactor)
	require.Equal(t, width, rackType.Width)
	require.Equal(t, uHeight, rackType.UHeight)
	require.Equal(t, startingUnit, rackType.StartingUnit)
	require.Equal(t, descUnits, rackType.DescUnits)
	require.Equal(t, description, rackType.Description)
	require.Equal(t, comments, rackType.Comments)
}

func requireRackTypeRESTScalars(
	t *testing.T,
	rackType map[string]any,
	manufacturer int64,
	model string,
	slug string,
	formFactor string,
	width uint32,
	uHeight uint32,
	startingUnit uint32,
	descUnits bool,
	description string,
	comments string,
) {
	t.Helper()
	manufacturerReference, ok := rackType["manufacturer"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(manufacturer), manufacturerReference["id"])
	formFactorChoice, ok := rackType["form_factor"].(map[string]any)
	require.True(t, ok)
	widthChoice, ok := rackType["width"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, model, rackType["model"])
	require.Equal(t, slug, rackType["slug"])
	require.Equal(t, formFactor, formFactorChoice["value"])
	require.Equal(t, float64(width), widthChoice["value"])
	require.Equal(t, float64(uHeight), rackType["u_height"])
	require.Equal(t, float64(startingUnit), rackType["starting_unit"])
	require.Equal(t, descUnits, rackType["desc_units"])
	require.Equal(t, description, rackType["description"])
	require.Equal(t, comments, rackType["comments"])
}

func rackTypeRESTCreateBody(manufacturer any, model any, slug any, formFactor any) map[string]any {
	return map[string]any{
		"manufacturer": manufacturer, "model": model, "slug": slug, "form_factor": formFactor,
	}
}

func rackTypeRESTCreateBodyWith(
	manufacturer any,
	model any,
	slug any,
	formFactor any,
	field string,
	value any,
) map[string]any {
	body := rackTypeRESTCreateBody(manufacturer, model, slug, formFactor)
	body[field] = value
	return body
}
