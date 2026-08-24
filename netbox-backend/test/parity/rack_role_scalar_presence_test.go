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

func TestRackRoleScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})

	var rowsBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.RackRoleRow{}).Count(&rowsBefore).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesBefore).Error)
	restCreateRejections := []struct {
		name string
		body map[string]any
		want map[string]any
	}{
		{
			name: "required fields omitted",
			body: map[string]any{},
			want: map[string]any{
				"name": []any{"This field is required."},
				"slug": []any{"This field is required."},
			},
		},
		{
			name: "name null",
			body: map[string]any{"name": nil, "slug": "rejected-name-null"},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "slug null",
			body: map[string]any{"name": "Rejected Slug Null", "slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "color null",
			body: map[string]any{
				"name": "Rejected Color Null", "slug": "rejected-color-null", "color": nil,
			},
			want: map[string]any{"color": []any{"This field may not be null."}},
		},
		{
			name: "description null",
			body: map[string]any{
				"name": "Rejected Description Null", "slug": "rejected-description-null",
				"description": nil,
			},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "name blank",
			body: map[string]any{"name": "", "slug": "rejected-name-blank"},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "slug blank",
			body: map[string]any{"name": "Rejected Slug Blank", "slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "color blank",
			body: map[string]any{
				"name": "Rejected Color Blank", "slug": "rejected-color-blank", "color": "",
			},
			want: map[string]any{"color": []any{"This field may not be blank."}},
		},
		{
			name: "color uppercase",
			body: map[string]any{
				"name": "Rejected Color Uppercase", "slug": "rejected-color-uppercase",
				"color": "ABCDEF",
			},
			want: map[string]any{"color": []any{"Enter a valid hexadecimal RGB color code."}},
		},
	}
	for _, test := range restCreateRejections {
		test := test
		t.Run("REST create/"+test.name, func(t *testing.T) {
			got := requestJSON(
				t, environment.router, http.MethodPost, "/api/dcim/rack-roles",
				test.body, http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
		})
	}

	blank := ""
	uppercase := "ABCDEF"
	grpcCreateRejections := []struct {
		name  string
		input *dcimv1.RackRoleInput
	}{
		{name: "required fields omitted", input: &dcimv1.RackRoleInput{}},
		{
			name:  "name blank",
			input: &dcimv1.RackRoleInput{Name: &blank, Slug: pointer("rejected-grpc-name-blank")},
		},
		{
			name:  "slug blank",
			input: &dcimv1.RackRoleInput{Name: pointer("Rejected gRPC Slug Blank"), Slug: &blank},
		},
		{
			name: "color blank",
			input: &dcimv1.RackRoleInput{
				Name: pointer("Rejected gRPC Color Blank"),
				Slug: pointer("rejected-grpc-color-blank"), Color: &blank,
			},
		},
		{
			name: "color uppercase",
			input: &dcimv1.RackRoleInput{
				Name: pointer("Rejected gRPC Color Uppercase"),
				Slug: pointer("rejected-grpc-color-uppercase"), Color: &uppercase,
			},
		},
	}
	for _, test := range grpcCreateRejections {
		test := test
		t.Run("gRPC create/"+test.name, func(t *testing.T) {
			_, err := environment.dcim.CreateRackRole(
				environment.ctx, &dcimv1.CreateRackRoleRequest{RackRole: test.input},
			)
			requireRackRoleGRPCInvalid(t, err)
		})
	}

	var rowsAfter, changesAfter int64
	require.NoError(t, environment.db.Model(&dcimrow.RackRoleRow{}).Count(&rowsAfter).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	require.Equal(t, rowsBefore, rowsAfter)
	require.Equal(t, changesBefore, changesAfter)

	createdWithOmission := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-roles",
		map[string]any{"name": "  REST Default Rack Role  ", "slug": "  rest-default-rack-role  "},
		http.StatusCreated,
	)
	defaultRoleID := jsonID(t, createdWithOmission["id"])
	requireRackRoleRESTScalars(
		t, createdWithOmission, "REST Default Rack Role", "rest-default-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)
	defaultState := loadParityRackRolePresenceState(t, environment, defaultRoleID)
	require.Equal(t, rowsAfter+1, defaultState.roleCount)
	require.Equal(t, int64(1), defaultState.changeCount)
	require.Equal(t, changesAfter+1, defaultState.totalChangeCount)
	defaultGRPCRead, err := environment.dcim.GetRackRole(
		environment.ctx, &dcimv1.GetRackRoleRequest{Id: defaultRoleID},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, defaultGRPCRead.RackRole, "REST Default Rack Role", "rest-default-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)

	createdByGRPC, err := environment.dcim.CreateRackRole(
		environment.ctx,
		&dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{
			Name: pointer("  gRPC Default Rack Role  "), Slug: pointer("  grpc-default-rack-role  "),
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, createdByGRPC.RackRole, "gRPC Default Rack Role", "grpc-default-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)
	grpcDefaultState := loadParityRackRolePresenceState(t, environment, createdByGRPC.RackRole.Id)
	require.Equal(t, defaultState.roleCount+1, grpcDefaultState.roleCount)
	require.Equal(t, int64(1), grpcDefaultState.changeCount)
	require.Equal(t, defaultState.totalChangeCount+1, grpcDefaultState.totalChangeCount)
	restRead := requestJSON(
		t, environment.router, http.MethodGet,
		"/api/dcim/rack-roles/"+strconv.FormatInt(createdByGRPC.RackRole.Id, 10), nil, http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restRead, "gRPC Default Rack Role", "grpc-default-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)

	createdBlankByREST := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-roles",
		map[string]any{
			"name": "REST Blank Rack Role", "slug": "rest-blank-rack-role",
			"description": "",
		},
		http.StatusCreated,
	)
	requireRackRoleRESTScalars(
		t, createdBlankByREST, "REST Blank Rack Role", "rest-blank-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)
	restBlankState := loadParityRackRolePresenceState(
		t, environment, jsonID(t, createdBlankByREST["id"]),
	)
	require.Equal(t, grpcDefaultState.roleCount+1, restBlankState.roleCount)
	require.Equal(t, int64(1), restBlankState.changeCount)
	require.Equal(t, grpcDefaultState.totalChangeCount+1, restBlankState.totalChangeCount)

	createdBlankByGRPC, err := environment.dcim.CreateRackRole(
		environment.ctx,
		&dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{
			Name: pointer("gRPC Blank Rack Role"), Slug: pointer("grpc-blank-rack-role"),
			Description: &blank,
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, createdBlankByGRPC.RackRole, "gRPC Blank Rack Role", "grpc-blank-rack-role",
		domaindcim.RackRoleDefaultColor, "",
	)
	grpcBlankState := loadParityRackRolePresenceState(t, environment, createdBlankByGRPC.RackRole.Id)
	require.Equal(t, restBlankState.roleCount+1, grpcBlankState.roleCount)
	require.Equal(t, int64(1), grpcBlankState.changeCount)
	require.Equal(t, restBlankState.totalChangeCount+1, grpcBlankState.totalChangeCount)

	createdConcreteByGRPC, err := environment.dcim.CreateRackRole(
		environment.ctx,
		&dcimv1.CreateRackRoleRequest{RackRole: &dcimv1.RackRoleInput{
			Name: pointer("  gRPC Concrete Rack Role  "),
			Slug: pointer("  grpc-concrete-rack-role  "), Color: pointer("  4455dd  "),
			Description: pointer("  gRPC concrete description  "),
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, createdConcreteByGRPC.RackRole, "gRPC Concrete Rack Role", "grpc-concrete-rack-role",
		"4455dd", "gRPC concrete description",
	)
	grpcConcreteState := loadParityRackRolePresenceState(t, environment, createdConcreteByGRPC.RackRole.Id)
	require.Equal(t, grpcBlankState.roleCount+1, grpcConcreteState.roleCount)
	require.Equal(t, int64(1), grpcConcreteState.changeCount)
	require.Equal(t, grpcBlankState.totalChangeCount+1, grpcConcreteState.totalChangeCount)

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/rack-roles",
		map[string]any{
			"name": "  Presence Rack Role  ", "slug": "  presence-rack-role  ",
			"color": "  123abc  ", "description": "  created description  ",
		},
		http.StatusCreated,
	)
	roleID := jsonID(t, created["id"])
	requireRackRoleRESTScalars(
		t, created, "Presence Rack Role", "presence-rack-role", "123abc", "created description",
	)
	createdState := loadParityRackRolePresenceState(t, environment, roleID)
	require.Equal(t, grpcConcreteState.roleCount+1, createdState.roleCount)
	require.Equal(t, int64(1), createdState.changeCount)
	require.Equal(t, grpcConcreteState.totalChangeCount+1, createdState.totalChangeCount)

	itemPath := "/api/dcim/rack-roles/" + strconv.FormatInt(roleID, 10)
	replacedByREST := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"name": "  REST Replaced Rack Role  ", "slug": "  rest-replaced-rack-role  ",
		},
		http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, replacedByREST, "REST Replaced Rack Role", "rest-replaced-rack-role",
		"123abc", "created description",
	)
	afterRESTPut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, createdState, afterRESTPut)

	replacedByGRPC, err := environment.dcim.ReplaceRackRole(
		environment.ctx,
		&dcimv1.ReplaceRackRoleRequest{Id: roleID, RackRole: &dcimv1.RackRoleInput{
			Name: pointer("  gRPC Replaced Rack Role  "), Slug: pointer("  grpc-replaced-rack-role  "),
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, replacedByGRPC.RackRole, "gRPC Replaced Rack Role", "grpc-replaced-rack-role",
		"123abc", "created description",
	)
	afterGRPCPut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTPut, afterGRPCPut)

	restMutationRejections := []struct {
		name   string
		method string
		body   map[string]any
		want   map[string]any
	}{
		{
			name: "PUT identity omitted", method: http.MethodPut, body: map[string]any{},
			want: map[string]any{
				"name": []any{"This field is required."},
				"slug": []any{"This field is required."},
			},
		},
		{
			name: "PUT name null", method: http.MethodPut,
			body: map[string]any{"name": nil, "slug": "grpc-replaced-rack-role"},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "PUT slug null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Rack Role", "slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "PUT color null", method: http.MethodPut,
			body: map[string]any{
				"name": "gRPC Replaced Rack Role", "slug": "grpc-replaced-rack-role", "color": nil,
			},
			want: map[string]any{"color": []any{"This field may not be null."}},
		},
		{
			name: "PUT description null", method: http.MethodPut,
			body: map[string]any{
				"name": "gRPC Replaced Rack Role", "slug": "grpc-replaced-rack-role",
				"description": nil,
			},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PUT name blank", method: http.MethodPut,
			body: map[string]any{"name": "", "slug": "grpc-replaced-rack-role"},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "PUT slug blank", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Rack Role", "slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "PUT color blank", method: http.MethodPut,
			body: map[string]any{
				"name": "gRPC Replaced Rack Role", "slug": "grpc-replaced-rack-role", "color": "",
			},
			want: map[string]any{"color": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH name null", method: http.MethodPatch, body: map[string]any{"name": nil},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "PATCH slug null", method: http.MethodPatch, body: map[string]any{"slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "PATCH color null", method: http.MethodPatch, body: map[string]any{"color": nil},
			want: map[string]any{"color": []any{"This field may not be null."}},
		},
		{
			name: "PATCH description null", method: http.MethodPatch,
			body: map[string]any{"description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PATCH name blank", method: http.MethodPatch, body: map[string]any{"name": ""},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH slug blank", method: http.MethodPatch, body: map[string]any{"slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH color blank", method: http.MethodPatch, body: map[string]any{"color": ""},
			want: map[string]any{"color": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH invalid color", method: http.MethodPatch,
			body: map[string]any{"color": "#123abc"},
			want: map[string]any{"color": []any{"Enter a valid hexadecimal RGB color code."}},
		},
	}
	for _, test := range restMutationRejections {
		test := test
		t.Run("REST mutation/"+test.name, func(t *testing.T) {
			before := loadParityRackRolePresenceState(t, environment, roleID)
			got := requestJSON(
				t, environment.router, test.method, itemPath, test.body, http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
			require.Equal(t, before, loadParityRackRolePresenceState(t, environment, roleID))
		})
	}

	grpcReplaceRejections := []struct {
		name  string
		input *dcimv1.RackRoleInput
	}{
		{name: "required identity omitted", input: &dcimv1.RackRoleInput{}},
		{
			name:  "name omitted",
			input: &dcimv1.RackRoleInput{Slug: pointer("grpc-replaced-rack-role")},
		},
		{
			name:  "slug omitted",
			input: &dcimv1.RackRoleInput{Name: pointer("gRPC Replaced Rack Role")},
		},
		{
			name: "name blank",
			input: &dcimv1.RackRoleInput{
				Name: &blank, Slug: pointer("grpc-replaced-rack-role"),
			},
		},
		{
			name: "slug blank",
			input: &dcimv1.RackRoleInput{
				Name: pointer("gRPC Replaced Rack Role"), Slug: &blank,
			},
		},
		{
			name: "color blank",
			input: &dcimv1.RackRoleInput{
				Name: pointer("gRPC Replaced Rack Role"),
				Slug: pointer("grpc-replaced-rack-role"), Color: &blank,
			},
		},
	}
	// Proto3 optional strings represent omission with nil pointers and cannot
	// express PUT null. REST owns the explicit-null PUT cells above; a masked
	// absent PATCH scalar owns gRPC explicit-null intent below.
	for _, test := range grpcReplaceRejections {
		test := test
		t.Run("gRPC replace/"+test.name, func(t *testing.T) {
			before := loadParityRackRolePresenceState(t, environment, roleID)
			_, err := environment.dcim.ReplaceRackRole(
				environment.ctx,
				&dcimv1.ReplaceRackRoleRequest{Id: roleID, RackRole: test.input},
			)
			requireRackRoleGRPCInvalid(t, err)
			require.Equal(t, before, loadParityRackRolePresenceState(t, environment, roleID))
		})
	}

	for _, field := range []string{"name", "slug", "color", "description"} {
		field := field
		t.Run("gRPC FieldMask null/"+field, func(t *testing.T) {
			before := loadParityRackRolePresenceState(t, environment, roleID)
			_, err := environment.dcim.UpdateRackRole(
				environment.ctx,
				&dcimv1.UpdateRackRoleRequest{
					Id: roleID, RackRole: &dcimv1.RackRoleInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireRackRoleGRPCInvalid(t, err)
			require.Equal(t, before, loadParityRackRolePresenceState(t, environment, roleID))
		})
	}

	for _, field := range []string{"name", "slug", "color"} {
		field := field
		t.Run("gRPC PATCH blank/"+field, func(t *testing.T) {
			before := loadParityRackRolePresenceState(t, environment, roleID)
			input := &dcimv1.RackRoleInput{}
			setRackRoleProtoScalar(input, field, &blank)
			_, err := environment.dcim.UpdateRackRole(
				environment.ctx,
				&dcimv1.UpdateRackRoleRequest{
					Id: roleID, RackRole: input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireRackRoleGRPCInvalid(t, err)
			require.Equal(t, before, loadParityRackRolePresenceState(t, environment, roleID))
		})
	}

	beforeRESTConcretePut := loadParityRackRolePresenceState(t, environment, roleID)
	restConcretePut := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"name": "  REST Concrete PUT Rack Role  ",
			"slug": "  rest-concrete-put-rack-role  ", "color": "  00aa11  ",
			"description": "  REST concrete PUT description  ",
		},
		http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restConcretePut, "REST Concrete PUT Rack Role", "rest-concrete-put-rack-role",
		"00aa11", "REST concrete PUT description",
	)
	afterRESTConcretePut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, beforeRESTConcretePut, afterRESTConcretePut)

	restBlankPut := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"name": "REST Blank PUT Rack Role", "slug": "rest-blank-put-rack-role",
			"description": "",
		},
		http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restBlankPut, "REST Blank PUT Rack Role", "rest-blank-put-rack-role", "00aa11", "",
	)
	afterRESTBlankPut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTConcretePut, afterRESTBlankPut)

	grpcBlankPut, err := environment.dcim.ReplaceRackRole(
		environment.ctx,
		&dcimv1.ReplaceRackRoleRequest{Id: roleID, RackRole: &dcimv1.RackRoleInput{
			Name: pointer("gRPC Blank PUT Rack Role"), Slug: pointer("grpc-blank-put-rack-role"),
			Description: &blank,
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, grpcBlankPut.RackRole, "gRPC Blank PUT Rack Role", "grpc-blank-put-rack-role",
		"00aa11", "",
	)
	afterGRPCBlankPut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTBlankPut, afterGRPCBlankPut)

	grpcConcretePut, err := environment.dcim.ReplaceRackRole(
		environment.ctx,
		&dcimv1.ReplaceRackRoleRequest{Id: roleID, RackRole: &dcimv1.RackRoleInput{
			Name: pointer("  gRPC Concrete PUT Rack Role  "),
			Slug: pointer("  grpc-concrete-put-rack-role  "), Color: pointer("  1122aa  "),
			Description: pointer("  gRPC concrete PUT description  "),
		}},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, grpcConcretePut.RackRole, "gRPC Concrete PUT Rack Role", "grpc-concrete-put-rack-role",
		"1122aa", "gRPC concrete PUT description",
	)
	afterGRPCConcretePut := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterGRPCBlankPut, afterGRPCConcretePut)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireRackRoleRESTScalars(
		t, restRead, "gRPC Concrete PUT Rack Role", "grpc-concrete-put-rack-role",
		"1122aa", "gRPC concrete PUT description",
	)

	beforeRESTOmissionPatch := loadParityRackRolePresenceState(t, environment, roleID)
	restOmissionPatch := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"description": "  REST sibling patch  "}, http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restOmissionPatch, "gRPC Concrete PUT Rack Role", "grpc-concrete-put-rack-role",
		"1122aa", "REST sibling patch",
	)
	afterRESTOmissionPatch := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, beforeRESTOmissionPatch, afterRESTOmissionPatch)

	restConcretePatch := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{
			"name": "  REST Patched Rack Role  ", "slug": "  rest-patched-rack-role  ",
			"color": "  4455dd  ",
		},
		http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restConcretePatch, "REST Patched Rack Role", "rest-patched-rack-role",
		"4455dd", "REST sibling patch",
	)
	afterRESTConcretePatch := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTOmissionPatch, afterRESTConcretePatch)

	restClearedPatch := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"description": ""}, http.StatusOK,
	)
	requireRackRoleRESTScalars(
		t, restClearedPatch, "REST Patched Rack Role", "rest-patched-rack-role", "4455dd", "",
	)
	afterRESTClearedPatch := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTConcretePatch, afterRESTClearedPatch)

	grpcColorPatch, err := environment.dcim.UpdateRackRole(
		environment.ctx,
		&dcimv1.UpdateRackRoleRequest{
			Id: roleID, RackRole: &dcimv1.RackRoleInput{Color: pointer("  2233bb  ")},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"color"}},
		},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, grpcColorPatch.RackRole, "REST Patched Rack Role", "rest-patched-rack-role",
		"2233bb", "",
	)
	afterGRPCColorPatch := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterRESTClearedPatch, afterGRPCColorPatch)

	grpcConcretePatch, err := environment.dcim.UpdateRackRole(
		environment.ctx,
		&dcimv1.UpdateRackRoleRequest{
			Id: roleID,
			RackRole: &dcimv1.RackRoleInput{
				Name: pointer("  gRPC Patched Rack Role  "),
				Slug: pointer("  grpc-patched-rack-role  "), Color: pointer("  3344cc  "),
				Description: pointer("  gRPC patched description  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "slug", "color", "description"}},
		},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, grpcConcretePatch.RackRole, "gRPC Patched Rack Role", "grpc-patched-rack-role",
		"3344cc", "gRPC patched description",
	)
	afterGRPCConcretePatch := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterGRPCColorPatch, afterGRPCConcretePatch)

	clearedByGRPC, err := environment.dcim.UpdateRackRole(
		environment.ctx,
		&dcimv1.UpdateRackRoleRequest{
			Id: roleID, RackRole: &dcimv1.RackRoleInput{Description: &blank},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}},
		},
	)
	require.NoError(t, err)
	requireRackRoleProtoScalars(
		t, clearedByGRPC.RackRole, "gRPC Patched Rack Role", "grpc-patched-rack-role", "3344cc", "",
	)
	afterGRPCClear := loadParityRackRolePresenceState(t, environment, roleID)
	requireRackRoleParityUpdateRecorded(t, afterGRPCConcretePatch, afterGRPCClear)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireRackRoleRESTScalars(
		t, restRead, "gRPC Patched Rack Role", "grpc-patched-rack-role", "3344cc", "",
	)
}

type parityRackRolePresenceState struct {
	row              dcimrow.RackRoleRow
	roleCount        int64
	changeCount      int64
	totalChangeCount int64
}

func loadParityRackRolePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityRackRolePresenceState {
	t.Helper()
	var state parityRackRolePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.RackRoleRow{}).Count(&state.roleCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.RackRoleObjectType, id,
	).Count(&state.changeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireRackRoleParityUpdateRecorded(
	t *testing.T,
	before parityRackRolePresenceState,
	after parityRackRolePresenceState,
) {
	t.Helper()
	require.Equal(t, before.roleCount, after.roleCount)
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireRackRoleGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(
		t,
		"Invalid input.",
		status.Convert(err).Message(),
		"public gRPC exposes the generic validation envelope; mapper and application tests pin field detail",
	)
}

func requireRackRoleProtoScalars(
	t *testing.T,
	role *dcimv1.RackRole,
	name string,
	slug string,
	color string,
	description string,
) {
	t.Helper()
	require.NotNil(t, role)
	require.Equal(t, name, role.Name)
	require.Equal(t, slug, role.Slug)
	require.Equal(t, color, role.Color)
	require.Equal(t, description, role.Description)
}

func requireRackRoleRESTScalars(
	t *testing.T,
	role map[string]any,
	name string,
	slug string,
	color string,
	description string,
) {
	t.Helper()
	require.Equal(t, name, role["name"])
	require.Equal(t, slug, role["slug"])
	require.Equal(t, color, role["color"])
	require.Equal(t, description, role["description"])
}

func setRackRoleProtoScalar(input *dcimv1.RackRoleInput, field string, value *string) {
	switch field {
	case "name":
		input.Name = value
	case "slug":
		input.Slug = value
	case "color":
		input.Color = value
	case "description":
		input.Description = value
	}
}
