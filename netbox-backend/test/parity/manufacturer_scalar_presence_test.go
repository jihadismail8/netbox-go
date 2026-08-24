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

func TestManufacturerScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})

	var rowsBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.ManufacturerRow{}).Count(&rowsBefore).Error)
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
	}
	for _, test := range restCreateRejections {
		test := test
		t.Run("REST create/"+test.name, func(t *testing.T) {
			got := requestJSON(
				t,
				environment.router,
				http.MethodPost,
				"/api/dcim/manufacturers",
				test.body,
				http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
		})
	}

	blank := ""
	grpcCreateRejections := []struct {
		name  string
		input *dcimv1.ManufacturerInput
	}{
		{name: "required fields omitted", input: &dcimv1.ManufacturerInput{}},
		{
			name: "name blank",
			input: &dcimv1.ManufacturerInput{
				Name: &blank, Slug: pointer("rejected-grpc-name-blank"),
			},
		},
		{
			name: "slug blank",
			input: &dcimv1.ManufacturerInput{
				Name: pointer("Rejected gRPC Slug Blank"), Slug: &blank,
			},
		},
	}
	for _, test := range grpcCreateRejections {
		test := test
		t.Run("gRPC create/"+test.name, func(t *testing.T) {
			_, err := environment.dcim.CreateManufacturer(
				environment.ctx,
				&dcimv1.CreateManufacturerRequest{Manufacturer: test.input},
			)
			requireManufacturerGRPCInvalid(t, err)
		})
	}

	var rowsAfter, changesAfter int64
	require.NoError(t, environment.db.Model(&dcimrow.ManufacturerRow{}).Count(&rowsAfter).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	require.Equal(t, rowsBefore, rowsAfter)
	require.Equal(t, changesBefore, changesAfter)

	createdWithOmission := requestJSON(
		t,
		environment.router,
		http.MethodPost,
		"/api/dcim/manufacturers",
		map[string]any{
			"name": "  REST Default Manufacturer  ",
			"slug": "  rest-default-manufacturer  ",
		},
		http.StatusCreated,
	)
	defaultManufacturerID := jsonID(t, createdWithOmission["id"])
	requireManufacturerRESTScalars(
		t,
		createdWithOmission,
		"REST Default Manufacturer",
		"rest-default-manufacturer",
		"",
	)
	defaultState := loadParityManufacturerPresenceState(t, environment, defaultManufacturerID)
	require.Equal(t, rowsAfter+1, defaultState.manufacturerCount)
	require.Equal(t, int64(1), defaultState.changeCount)
	require.Equal(t, changesAfter+1, defaultState.totalChangeCount)
	defaultGRPCRead, err := environment.dcim.GetManufacturer(
		environment.ctx,
		&dcimv1.GetManufacturerRequest{Id: defaultManufacturerID},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		defaultGRPCRead.Manufacturer,
		"REST Default Manufacturer",
		"rest-default-manufacturer",
		"",
	)

	createdWithBlank := requestJSON(
		t,
		environment.router,
		http.MethodPost,
		"/api/dcim/manufacturers",
		map[string]any{
			"name": "REST Blank Manufacturer", "slug": "rest-blank-manufacturer",
			"description": "",
		},
		http.StatusCreated,
	)
	blankManufacturerID := jsonID(t, createdWithBlank["id"])
	requireManufacturerRESTScalars(
		t,
		createdWithBlank,
		"REST Blank Manufacturer",
		"rest-blank-manufacturer",
		"",
	)
	blankState := loadParityManufacturerPresenceState(t, environment, blankManufacturerID)
	require.Equal(t, defaultState.manufacturerCount+1, blankState.manufacturerCount)
	require.Equal(t, int64(1), blankState.changeCount)
	require.Equal(t, defaultState.totalChangeCount+1, blankState.totalChangeCount)

	created := requestJSON(
		t,
		environment.router,
		http.MethodPost,
		"/api/dcim/manufacturers",
		map[string]any{
			"name": "  Presence Manufacturer  ", "slug": "  presence-manufacturer  ",
			"description": "  created description  ",
		},
		http.StatusCreated,
	)
	manufacturerID := jsonID(t, created["id"])
	requireManufacturerRESTScalars(
		t,
		created,
		"Presence Manufacturer",
		"presence-manufacturer",
		"created description",
	)
	createdState := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	require.Equal(t, blankState.manufacturerCount+1, createdState.manufacturerCount)
	require.Equal(t, int64(1), createdState.changeCount)
	require.Equal(t, blankState.totalChangeCount+1, createdState.totalChangeCount)
	grpcRead, err := environment.dcim.GetManufacturer(
		environment.ctx,
		&dcimv1.GetManufacturerRequest{Id: manufacturerID},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		grpcRead.Manufacturer,
		"Presence Manufacturer",
		"presence-manufacturer",
		"created description",
	)

	itemPath := "/api/dcim/manufacturers/" + strconv.FormatInt(manufacturerID, 10)
	replacedByREST := requestJSON(
		t,
		environment.router,
		http.MethodPut,
		itemPath,
		map[string]any{
			"name": "  REST Replaced Manufacturer  ",
			"slug": "  rest-replaced-manufacturer  ",
		},
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		replacedByREST,
		"REST Replaced Manufacturer",
		"rest-replaced-manufacturer",
		"created description",
	)
	afterRESTPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, createdState, afterRESTPut)

	replacedByGRPC, err := environment.dcim.ReplaceManufacturer(
		environment.ctx,
		&dcimv1.ReplaceManufacturerRequest{
			Id: manufacturerID,
			Manufacturer: &dcimv1.ManufacturerInput{
				Name: pointer("  gRPC Replaced Manufacturer  "),
				Slug: pointer("  grpc-replaced-manufacturer  "),
			},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		replacedByGRPC.Manufacturer,
		"gRPC Replaced Manufacturer",
		"grpc-replaced-manufacturer",
		"created description",
	)
	afterGRPCPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, afterRESTPut, afterGRPCPut)
	restRead := requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Replaced Manufacturer",
		"grpc-replaced-manufacturer",
		"created description",
	)

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
			body: map[string]any{"name": nil, "slug": "grpc-replaced-manufacturer"},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "PUT slug null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Manufacturer", "slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "PUT description null", method: http.MethodPut,
			body: map[string]any{
				"name": "gRPC Replaced Manufacturer", "slug": "grpc-replaced-manufacturer",
				"description": nil,
			},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PUT name blank", method: http.MethodPut,
			body: map[string]any{"name": "", "slug": "grpc-replaced-manufacturer"},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "PUT slug blank", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Manufacturer", "slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH name null", method: http.MethodPatch,
			body: map[string]any{"name": nil},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "PATCH slug null", method: http.MethodPatch,
			body: map[string]any{"slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "PATCH description null", method: http.MethodPatch,
			body: map[string]any{"description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PATCH name blank", method: http.MethodPatch,
			body: map[string]any{"name": ""},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "PATCH slug blank", method: http.MethodPatch,
			body: map[string]any{"slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
	}
	for _, test := range restMutationRejections {
		test := test
		t.Run("REST mutation/"+test.name, func(t *testing.T) {
			before := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			got := requestJSON(
				t,
				environment.router,
				test.method,
				itemPath,
				test.body,
				http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
			after := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			require.Equal(t, before, after)
		})
	}

	grpcReplaceRejections := []struct {
		name  string
		input *dcimv1.ManufacturerInput
	}{
		{name: "required identity omitted", input: &dcimv1.ManufacturerInput{}},
		{
			name:  "name omitted",
			input: &dcimv1.ManufacturerInput{Slug: pointer("grpc-replaced-manufacturer")},
		},
		{
			name:  "slug omitted",
			input: &dcimv1.ManufacturerInput{Name: pointer("gRPC Replaced Manufacturer")},
		},
		{
			name: "name blank",
			input: &dcimv1.ManufacturerInput{
				Name: &blank, Slug: pointer("grpc-replaced-manufacturer"),
			},
		},
		{
			name: "slug blank",
			input: &dcimv1.ManufacturerInput{
				Name: pointer("gRPC Replaced Manufacturer"), Slug: &blank,
			},
		},
	}
	// Proto3 optional strings represent omission with nil pointers. They cannot
	// express PUT null; REST owns the explicit-null PUT cases above, while gRPC
	// FieldMask plus an absent scalar owns explicit null for PATCH below.
	for _, test := range grpcReplaceRejections {
		test := test
		t.Run("gRPC replace/"+test.name, func(t *testing.T) {
			before := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			_, err := environment.dcim.ReplaceManufacturer(
				environment.ctx,
				&dcimv1.ReplaceManufacturerRequest{
					Id: manufacturerID, Manufacturer: test.input,
				},
			)
			requireManufacturerGRPCInvalid(t, err)
			after := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			require.Equal(t, before, after)
		})
	}

	for _, field := range []string{"name", "slug", "description"} {
		field := field
		t.Run("gRPC FieldMask null/"+field, func(t *testing.T) {
			before := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			_, err := environment.dcim.UpdateManufacturer(
				environment.ctx,
				&dcimv1.UpdateManufacturerRequest{
					Id:           manufacturerID,
					Manufacturer: &dcimv1.ManufacturerInput{},
					UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireManufacturerGRPCInvalid(t, err)
			after := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			require.Equal(t, before, after)
		})
	}

	for _, field := range []string{"name", "slug"} {
		field := field
		t.Run("gRPC PATCH blank/"+field, func(t *testing.T) {
			before := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			input := &dcimv1.ManufacturerInput{}
			setManufacturerProtoScalar(input, field, &blank)
			_, err := environment.dcim.UpdateManufacturer(
				environment.ctx,
				&dcimv1.UpdateManufacturerRequest{
					Id: manufacturerID, Manufacturer: input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireManufacturerGRPCInvalid(t, err)
			after := loadParityManufacturerPresenceState(t, environment, manufacturerID)
			require.Equal(t, before, after)
		})
	}

	beforeRESTBlankPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	restBlankPut := requestJSON(
		t,
		environment.router,
		http.MethodPut,
		itemPath,
		map[string]any{
			"name": "REST Blank PUT Manufacturer", "slug": "rest-blank-put-manufacturer",
			"description": "",
		},
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		restBlankPut,
		"REST Blank PUT Manufacturer",
		"rest-blank-put-manufacturer",
		"",
	)
	afterRESTBlankPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeRESTBlankPut, afterRESTBlankPut)
	grpcRead, err = environment.dcim.GetManufacturer(
		environment.ctx,
		&dcimv1.GetManufacturerRequest{Id: manufacturerID},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		grpcRead.Manufacturer,
		"REST Blank PUT Manufacturer",
		"rest-blank-put-manufacturer",
		"",
	)

	beforeRESTConcretePut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	restConcretePut := requestJSON(
		t,
		environment.router,
		http.MethodPut,
		itemPath,
		map[string]any{
			"name":        "  REST Concrete PUT Manufacturer  ",
			"slug":        "  rest-concrete-put-manufacturer  ",
			"description": "  REST concrete description  ",
		},
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		restConcretePut,
		"REST Concrete PUT Manufacturer",
		"rest-concrete-put-manufacturer",
		"REST concrete description",
	)
	afterRESTConcretePut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeRESTConcretePut, afterRESTConcretePut)

	beforeGRPCBlankPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	grpcBlankPut, err := environment.dcim.ReplaceManufacturer(
		environment.ctx,
		&dcimv1.ReplaceManufacturerRequest{
			Id: manufacturerID,
			Manufacturer: &dcimv1.ManufacturerInput{
				Name: pointer("gRPC Blank PUT Manufacturer"),
				Slug: pointer("grpc-blank-put-manufacturer"), Description: &blank,
			},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		grpcBlankPut.Manufacturer,
		"gRPC Blank PUT Manufacturer",
		"grpc-blank-put-manufacturer",
		"",
	)
	afterGRPCBlankPut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeGRPCBlankPut, afterGRPCBlankPut)

	beforeGRPCConcretePut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	grpcConcretePut, err := environment.dcim.ReplaceManufacturer(
		environment.ctx,
		&dcimv1.ReplaceManufacturerRequest{
			Id: manufacturerID,
			Manufacturer: &dcimv1.ManufacturerInput{
				Name:        pointer("  gRPC Concrete PUT Manufacturer  "),
				Slug:        pointer("  grpc-concrete-put-manufacturer  "),
				Description: pointer("  gRPC concrete description  "),
			},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		grpcConcretePut.Manufacturer,
		"gRPC Concrete PUT Manufacturer",
		"grpc-concrete-put-manufacturer",
		"gRPC concrete description",
	)
	afterGRPCConcretePut := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeGRPCConcretePut, afterGRPCConcretePut)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Concrete PUT Manufacturer",
		"grpc-concrete-put-manufacturer",
		"gRPC concrete description",
	)

	beforeRESTConcretePatch := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	restConcretePatch := requestJSON(
		t,
		environment.router,
		http.MethodPatch,
		itemPath,
		map[string]any{
			"name": "  REST Patched Manufacturer  ", "slug": "  rest-patched-manufacturer  ",
			"description": "  REST patched description  ",
		},
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		restConcretePatch,
		"REST Patched Manufacturer",
		"rest-patched-manufacturer",
		"REST patched description",
	)
	afterRESTConcretePatch := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeRESTConcretePatch, afterRESTConcretePatch)
	grpcRead, err = environment.dcim.GetManufacturer(
		environment.ctx,
		&dcimv1.GetManufacturerRequest{Id: manufacturerID},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		grpcRead.Manufacturer,
		"REST Patched Manufacturer",
		"rest-patched-manufacturer",
		"REST patched description",
	)

	beforeRESTClear := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	clearedByREST := requestJSON(
		t,
		environment.router,
		http.MethodPatch,
		itemPath,
		map[string]any{"description": ""},
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		clearedByREST,
		"REST Patched Manufacturer",
		"rest-patched-manufacturer",
		"",
	)
	afterRESTClear := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, beforeRESTClear, afterRESTClear)

	setByGRPC, err := environment.dcim.UpdateManufacturer(
		environment.ctx,
		&dcimv1.UpdateManufacturerRequest{
			Id: manufacturerID,
			Manufacturer: &dcimv1.ManufacturerInput{
				Name:        pointer("  gRPC Patched Manufacturer  "),
				Slug:        pointer("  grpc-patched-manufacturer  "),
				Description: pointer("  gRPC patched description  "),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name", "slug", "description"}},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		setByGRPC.Manufacturer,
		"gRPC Patched Manufacturer",
		"grpc-patched-manufacturer",
		"gRPC patched description",
	)
	afterGRPCSet := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, afterRESTClear, afterGRPCSet)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Patched Manufacturer",
		"grpc-patched-manufacturer",
		"gRPC patched description",
	)

	clearedByGRPC, err := environment.dcim.UpdateManufacturer(
		environment.ctx,
		&dcimv1.UpdateManufacturerRequest{
			Id:           manufacturerID,
			Manufacturer: &dcimv1.ManufacturerInput{Description: &blank},
			UpdateMask:   &fieldmaskpb.FieldMask{Paths: []string{"description"}},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		clearedByGRPC.Manufacturer,
		"gRPC Patched Manufacturer",
		"grpc-patched-manufacturer",
		"",
	)
	afterGRPCClear := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	requireManufacturerUpdateRecorded(t, afterGRPCSet, afterGRPCClear)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Patched Manufacturer",
		"grpc-patched-manufacturer",
		"",
	)

	beforeGRPCCreate := loadParityManufacturerPresenceState(t, environment, manufacturerID)
	createdByGRPC, err := environment.dcim.CreateManufacturer(
		environment.ctx,
		&dcimv1.CreateManufacturerRequest{
			Manufacturer: &dcimv1.ManufacturerInput{
				Name: pointer("  gRPC Default Manufacturer  "),
				Slug: pointer("  grpc-default-manufacturer  "),
			},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		createdByGRPC.Manufacturer,
		"gRPC Default Manufacturer",
		"grpc-default-manufacturer",
		"",
	)
	createdByGRPCState := loadParityManufacturerPresenceState(
		t,
		environment,
		createdByGRPC.Manufacturer.Id,
	)
	require.Equal(t, beforeGRPCCreate.manufacturerCount+1, createdByGRPCState.manufacturerCount)
	require.Equal(t, int64(1), createdByGRPCState.changeCount)
	require.Equal(t, beforeGRPCCreate.totalChangeCount+1, createdByGRPCState.totalChangeCount)
	restRead = requestJSON(
		t,
		environment.router,
		http.MethodGet,
		"/api/dcim/manufacturers/"+strconv.FormatInt(createdByGRPC.Manufacturer.Id, 10),
		nil,
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Default Manufacturer",
		"grpc-default-manufacturer",
		"",
	)

	createdBlankByGRPC, err := environment.dcim.CreateManufacturer(
		environment.ctx,
		&dcimv1.CreateManufacturerRequest{
			Manufacturer: &dcimv1.ManufacturerInput{
				Name: pointer("gRPC Blank Manufacturer"),
				Slug: pointer("grpc-blank-manufacturer"), Description: &blank,
			},
		},
	)
	require.NoError(t, err)
	requireManufacturerProtoScalars(
		t,
		createdBlankByGRPC.Manufacturer,
		"gRPC Blank Manufacturer",
		"grpc-blank-manufacturer",
		"",
	)
	createdBlankByGRPCState := loadParityManufacturerPresenceState(
		t,
		environment,
		createdBlankByGRPC.Manufacturer.Id,
	)
	require.Equal(t, createdByGRPCState.manufacturerCount+1, createdBlankByGRPCState.manufacturerCount)
	require.Equal(t, int64(1), createdBlankByGRPCState.changeCount)
	require.Equal(t, createdByGRPCState.totalChangeCount+1, createdBlankByGRPCState.totalChangeCount)
	restRead = requestJSON(
		t,
		environment.router,
		http.MethodGet,
		"/api/dcim/manufacturers/"+strconv.FormatInt(createdBlankByGRPC.Manufacturer.Id, 10),
		nil,
		http.StatusOK,
	)
	requireManufacturerRESTScalars(
		t,
		restRead,
		"gRPC Blank Manufacturer",
		"grpc-blank-manufacturer",
		"",
	)
}

type parityManufacturerPresenceState struct {
	row               dcimrow.ManufacturerRow
	manufacturerCount int64
	changeCount       int64
	totalChangeCount  int64
}

func loadParityManufacturerPresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) parityManufacturerPresenceState {
	t.Helper()
	var state parityManufacturerPresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(
		t,
		environment.db.Model(&dcimrow.ManufacturerRow{}).Count(&state.manufacturerCount).Error,
	)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.ManufacturerObjectType, id,
	).Count(&state.changeCount).Error)
	require.NoError(
		t,
		environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error,
	)
	return state
}

func requireManufacturerUpdateRecorded(
	t *testing.T,
	before parityManufacturerPresenceState,
	after parityManufacturerPresenceState,
) {
	t.Helper()
	require.Equal(t, before.manufacturerCount, after.manufacturerCount)
	require.Equal(t, before.changeCount+1, after.changeCount)
	require.Equal(t, before.totalChangeCount+1, after.totalChangeCount)
	require.Equal(t, before.row.Created, after.row.Created)
}

func requireManufacturerGRPCInvalid(t *testing.T, err error) {
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

func requireManufacturerProtoScalars(
	t *testing.T,
	manufacturer *dcimv1.Manufacturer,
	name string,
	slug string,
	description string,
) {
	t.Helper()
	require.NotNil(t, manufacturer)
	require.Equal(t, name, manufacturer.Name)
	require.Equal(t, slug, manufacturer.Slug)
	require.Equal(t, description, manufacturer.Description)
}

func requireManufacturerRESTScalars(
	t *testing.T,
	manufacturer map[string]any,
	name string,
	slug string,
	description string,
) {
	t.Helper()
	require.Equal(t, name, manufacturer["name"])
	require.Equal(t, slug, manufacturer["slug"])
	require.Equal(t, description, manufacturer["description"])
}

func setManufacturerProtoScalar(input *dcimv1.ManufacturerInput, field string, value *string) {
	switch field {
	case "name":
		input.Name = value
	case "slug":
		input.Slug = value
	case "description":
		input.Description = value
	}
}
