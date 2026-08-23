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

func TestSiteScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newProfileParityEnvironment(t, authz.AllowAll{})

	var rowsBefore, changesBefore int64
	require.NoError(t, environment.db.Model(&dcimrow.SiteRow{}).Count(&rowsBefore).Error)
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
			name: "status null",
			body: map[string]any{"name": "Rejected Status Null", "slug": "rejected-status-null", "status": nil},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "facility null",
			body: map[string]any{"name": "Rejected Facility Null", "slug": "rejected-facility-null", "facility": nil},
			want: map[string]any{"facility": []any{"This field may not be null."}},
		},
		{
			name: "description null",
			body: map[string]any{"name": "Rejected Description Null", "slug": "rejected-description-null", "description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "comments null",
			body: map[string]any{"name": "Rejected Comments Null", "slug": "rejected-comments-null", "comments": nil},
			want: map[string]any{"comments": []any{"This field may not be null."}},
		},
		{
			name: "status is not transport-trimmed",
			body: map[string]any{"name": "Rejected Status Spaces", "slug": "rejected-status-spaces", "status": " active "},
			want: map[string]any{"status": []any{" active  is not a valid choice."}},
		},
	}
	for _, test := range restCreateRejections {
		test := test
		t.Run("REST create/"+test.name, func(t *testing.T) {
			got := requestJSON(
				t, environment.router, http.MethodPost, "/api/dcim/sites",
				test.body, http.StatusBadRequest,
			)
			require.Equal(t, test.want, got)
		})
	}

	_, grpcErr := environment.dcim.CreateSite(
		environment.ctx,
		&dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{}},
	)
	requireSiteGRPCInvalid(t, grpcErr)
	invalidStatus := " active "
	_, grpcErr = environment.dcim.CreateSite(
		environment.ctx,
		&dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{
			Name: pointer("Rejected gRPC Status"), Slug: pointer("rejected-grpc-status"),
			Status: &invalidStatus,
		}},
	)
	requireSiteGRPCInvalid(t, grpcErr)

	var rowsAfter, changesAfter int64
	require.NoError(t, environment.db.Model(&dcimrow.SiteRow{}).Count(&rowsAfter).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&changesAfter).Error)
	require.Equal(t, rowsBefore, rowsAfter)
	require.Equal(t, changesBefore, changesAfter)

	createdWithOmissions := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/sites",
		map[string]any{"name": "  REST Default Site  ", "slug": "  rest-default-site  "},
		http.StatusCreated,
	)
	defaultSiteID := jsonID(t, createdWithOmissions["id"])
	require.Equal(t, "REST Default Site", createdWithOmissions["name"])
	require.Equal(t, "rest-default-site", createdWithOmissions["slug"])
	requireChoiceValue(t, createdWithOmissions["status"], "active")
	require.Empty(t, createdWithOmissions["facility"])
	require.Empty(t, createdWithOmissions["description"])
	require.Empty(t, createdWithOmissions["comments"])
	defaultState := loadParitySitePresenceState(t, environment, defaultSiteID)
	require.Equal(t, rowsAfter+1, defaultState.siteCount)
	require.Equal(t, int64(1), defaultState.changeCount)
	require.Equal(t, changesAfter+1, defaultState.totalChangeCount)
	defaultGRPCRead, err := environment.dcim.GetSite(
		environment.ctx,
		&dcimv1.GetSiteRequest{Id: defaultSiteID},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, defaultGRPCRead.Site,
		"REST Default Site", "rest-default-site", "active", "", "", "",
	)

	createdWithBlankOptionals := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/sites",
		map[string]any{
			"name": "REST Blank Optional Site", "slug": "rest-blank-optional-site",
			"facility": "", "description": "", "comments": "",
		},
		http.StatusCreated,
	)
	blankOptionalSiteID := jsonID(t, createdWithBlankOptionals["id"])
	requireChoiceValue(t, createdWithBlankOptionals["status"], "active")
	require.Empty(t, createdWithBlankOptionals["facility"])
	require.Empty(t, createdWithBlankOptionals["description"])
	require.Empty(t, createdWithBlankOptionals["comments"])
	blankOptionalState := loadParitySitePresenceState(t, environment, blankOptionalSiteID)
	require.Equal(t, defaultState.siteCount+1, blankOptionalState.siteCount)
	require.Equal(t, int64(1), blankOptionalState.changeCount)
	require.Equal(t, defaultState.totalChangeCount+1, blankOptionalState.totalChangeCount)
	blankOptionalGRPCRead, err := environment.dcim.GetSite(
		environment.ctx,
		&dcimv1.GetSiteRequest{Id: blankOptionalSiteID},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, blankOptionalGRPCRead.Site,
		"REST Blank Optional Site", "rest-blank-optional-site", "active", "", "", "",
	)

	created := requestJSON(
		t, environment.router, http.MethodPost, "/api/dcim/sites",
		map[string]any{
			"name": "  Presence Site  ", "slug": "  presence-site  ",
			"status": "staging", "facility": "  P1  ",
			"description": "  created description  ", "comments": "  created comments  ",
		},
		http.StatusCreated,
	)
	siteID := jsonID(t, created["id"])
	require.Equal(t, "Presence Site", created["name"])
	require.Equal(t, "presence-site", created["slug"])
	requireChoiceValue(t, created["status"], "staging")
	require.Equal(t, "P1", created["facility"])
	require.Equal(t, "created description", created["description"])
	require.Equal(t, "created comments", created["comments"])
	createdState := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, int64(1), createdState.changeCount)
	require.Equal(t, blankOptionalState.siteCount+1, createdState.siteCount)
	require.Equal(t, blankOptionalState.totalChangeCount+1, createdState.totalChangeCount)

	grpcRead, err := environment.dcim.GetSite(environment.ctx, &dcimv1.GetSiteRequest{Id: siteID})
	require.NoError(t, err)
	requireSiteProtoScalars(t, grpcRead.Site, "Presence Site", "presence-site", "staging", "P1", "created description", "created comments")

	itemPath := "/api/dcim/sites/" + strconv.FormatInt(siteID, 10)
	replaced := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{"name": "  REST Replaced Site  ", "slug": "  rest-replaced-site  "},
		http.StatusOK,
	)
	require.Equal(t, "REST Replaced Site", replaced["name"])
	require.Equal(t, "rest-replaced-site", replaced["slug"])
	requireChoiceValue(t, replaced["status"], "staging")
	require.Equal(t, "P1", replaced["facility"])
	require.Equal(t, "created description", replaced["description"])
	require.Equal(t, "created comments", replaced["comments"])
	afterRESTPut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, createdState.siteCount, afterRESTPut.siteCount)
	require.Equal(t, createdState.changeCount+1, afterRESTPut.changeCount)
	require.Equal(t, createdState.totalChangeCount+1, afterRESTPut.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterRESTPut.row.Created)

	replacedByGRPC, err := environment.dcim.ReplaceSite(
		environment.ctx,
		&dcimv1.ReplaceSiteRequest{Id: siteID, Site: &dcimv1.SiteInput{
			Name: pointer("  gRPC Replaced Site  "), Slug: pointer("  grpc-replaced-site  "),
		}},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(t, replacedByGRPC.Site, "gRPC Replaced Site", "grpc-replaced-site", "staging", "P1", "created description", "created comments")
	afterGRPCPut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, afterRESTPut.siteCount, afterGRPCPut.siteCount)
	require.Equal(t, afterRESTPut.changeCount+1, afterGRPCPut.changeCount)
	require.Equal(t, afterRESTPut.totalChangeCount+1, afterGRPCPut.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterGRPCPut.row.Created)
	restRead := requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	require.Equal(t, "gRPC Replaced Site", restRead["name"])
	require.Equal(t, "P1", restRead["facility"])

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
			body: map[string]any{"name": nil, "slug": "grpc-replaced-site"},
			want: map[string]any{"name": []any{"This field may not be null."}},
		},
		{
			name: "PUT slug null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": nil},
			want: map[string]any{"slug": []any{"This field may not be null."}},
		},
		{
			name: "PUT status null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": "grpc-replaced-site", "status": nil},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{
			name: "PUT facility null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": "grpc-replaced-site", "facility": nil},
			want: map[string]any{"facility": []any{"This field may not be null."}},
		},
		{
			name: "PUT description null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": "grpc-replaced-site", "description": nil},
			want: map[string]any{"description": []any{"This field may not be null."}},
		},
		{
			name: "PUT comments null", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": "grpc-replaced-site", "comments": nil},
			want: map[string]any{"comments": []any{"This field may not be null."}},
		},
		{
			name: "PUT name blank", method: http.MethodPut,
			body: map[string]any{"name": "", "slug": "grpc-replaced-site"},
			want: map[string]any{"name": []any{"This field may not be blank."}},
		},
		{
			name: "PUT slug blank", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": ""},
			want: map[string]any{"slug": []any{"This field may not be blank."}},
		},
		{
			name: "PUT status blank", method: http.MethodPut,
			body: map[string]any{"name": "gRPC Replaced Site", "slug": "grpc-replaced-site", "status": ""},
			want: map[string]any{"status": []any{"This field may not be blank."}},
		},
		{name: "PATCH name null", method: http.MethodPatch, body: map[string]any{"name": nil}, want: map[string]any{"name": []any{"This field may not be null."}}},
		{name: "PATCH slug null", method: http.MethodPatch, body: map[string]any{"slug": nil}, want: map[string]any{"slug": []any{"This field may not be null."}}},
		{name: "PATCH status null", method: http.MethodPatch, body: map[string]any{"status": nil}, want: map[string]any{"status": []any{"This field may not be blank."}}},
		{name: "PATCH facility null", method: http.MethodPatch, body: map[string]any{"facility": nil}, want: map[string]any{"facility": []any{"This field may not be null."}}},
		{name: "PATCH description null", method: http.MethodPatch, body: map[string]any{"description": nil}, want: map[string]any{"description": []any{"This field may not be null."}}},
		{name: "PATCH comments null", method: http.MethodPatch, body: map[string]any{"comments": nil}, want: map[string]any{"comments": []any{"This field may not be null."}}},
		{name: "PATCH name blank", method: http.MethodPatch, body: map[string]any{"name": ""}, want: map[string]any{"name": []any{"This field may not be blank."}}},
		{name: "PATCH slug blank", method: http.MethodPatch, body: map[string]any{"slug": ""}, want: map[string]any{"slug": []any{"This field may not be blank."}}},
		{name: "PATCH status blank", method: http.MethodPatch, body: map[string]any{"status": ""}, want: map[string]any{"status": []any{"This field may not be blank."}}},
		{name: "PATCH status spaced invalid choice", method: http.MethodPatch, body: map[string]any{"status": " active "}, want: map[string]any{"status": []any{" active  is not a valid choice."}}},
	}
	for _, test := range restMutationRejections {
		test := test
		t.Run("REST mutation/"+test.name, func(t *testing.T) {
			before := loadParitySitePresenceState(t, environment, siteID)
			got := requestJSON(t, environment.router, test.method, itemPath, test.body, http.StatusBadRequest)
			require.Equal(t, test.want, got)
			after := loadParitySitePresenceState(t, environment, siteID)
			require.Equal(t, before, after)
		})
	}

	blankScalar := ""
	grpcReplaceRejections := []struct {
		name  string
		input *dcimv1.SiteInput
	}{
		{name: "required identity omitted", input: &dcimv1.SiteInput{}},
		{name: "name omitted", input: &dcimv1.SiteInput{Slug: pointer("grpc-replaced-site")}},
		{name: "slug omitted", input: &dcimv1.SiteInput{Name: pointer("gRPC Replaced Site")}},
		{
			name: "name blank",
			input: &dcimv1.SiteInput{
				Name: &blankScalar, Slug: pointer("grpc-replaced-site"),
			},
		},
		{
			name: "slug blank",
			input: &dcimv1.SiteInput{
				Name: pointer("gRPC Replaced Site"), Slug: &blankScalar,
			},
		},
		{
			name: "status blank",
			input: &dcimv1.SiteInput{
				Name: pointer("gRPC Replaced Site"), Slug: pointer("grpc-replaced-site"),
				Status: &blankScalar,
			},
		},
	}
	// Proto3 optional strings represent omission with nil pointers. They cannot
	// express PUT null; REST owns the six explicit-null cases above, while gRPC
	// FieldMask plus an absent scalar owns explicit null for PATCH below.
	for _, test := range grpcReplaceRejections {
		test := test
		t.Run("gRPC replace/"+test.name, func(t *testing.T) {
			before := loadParitySitePresenceState(t, environment, siteID)
			_, err := environment.dcim.ReplaceSite(
				environment.ctx,
				&dcimv1.ReplaceSiteRequest{Id: siteID, Site: test.input},
			)
			requireSiteGRPCInvalid(t, err)
			after := loadParitySitePresenceState(t, environment, siteID)
			require.Equal(t, before, after)
		})
	}

	for _, field := range []string{"name", "slug", "status", "facility", "description", "comments"} {
		field := field
		t.Run("gRPC FieldMask null/"+field, func(t *testing.T) {
			before := loadParitySitePresenceState(t, environment, siteID)
			_, err := environment.dcim.UpdateSite(
				environment.ctx,
				&dcimv1.UpdateSiteRequest{
					Id: siteID, Site: &dcimv1.SiteInput{},
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireSiteGRPCInvalid(t, err)
			after := loadParitySitePresenceState(t, environment, siteID)
			require.Equal(t, before, after)
		})
	}

	for _, field := range []string{"name", "slug", "status"} {
		field := field
		t.Run("gRPC blank/"+field, func(t *testing.T) {
			before := loadParitySitePresenceState(t, environment, siteID)
			blank := ""
			input := &dcimv1.SiteInput{}
			setSiteProtoScalar(input, field, &blank)
			_, err := environment.dcim.UpdateSite(
				environment.ctx,
				&dcimv1.UpdateSiteRequest{
					Id: siteID, Site: input,
					UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}},
				},
			)
			requireSiteGRPCInvalid(t, err)
			after := loadParitySitePresenceState(t, environment, siteID)
			require.Equal(t, before, after)
		})
	}

	beforeUnsupportedMask := loadParitySitePresenceState(t, environment, siteID)
	_, grpcErr = environment.dcim.UpdateSite(
		environment.ctx,
		&dcimv1.UpdateSiteRequest{
			Id: siteID, Site: &dcimv1.SiteInput{Description: pointer("ignored")},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
		},
	)
	requireSiteGRPCInvalid(t, grpcErr)
	require.Equal(t, beforeUnsupportedMask, loadParitySitePresenceState(t, environment, siteID))

	beforeRESTBlankPut := loadParitySitePresenceState(t, environment, siteID)
	restBlankPut := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"name": "REST Blank PUT Site", "slug": "rest-blank-put-site",
			"facility": "", "description": "", "comments": "",
		},
		http.StatusOK,
	)
	require.Equal(t, "REST Blank PUT Site", restBlankPut["name"])
	requireChoiceValue(t, restBlankPut["status"], "staging")
	require.Empty(t, restBlankPut["facility"])
	require.Empty(t, restBlankPut["description"])
	require.Empty(t, restBlankPut["comments"])
	afterRESTBlankPut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, beforeRESTBlankPut.siteCount, afterRESTBlankPut.siteCount)
	require.Equal(t, beforeRESTBlankPut.changeCount+1, afterRESTBlankPut.changeCount)
	require.Equal(t, beforeRESTBlankPut.totalChangeCount+1, afterRESTBlankPut.totalChangeCount)
	require.Equal(t, beforeRESTBlankPut.row.Created, afterRESTBlankPut.row.Created)
	grpcRead, err = environment.dcim.GetSite(environment.ctx, &dcimv1.GetSiteRequest{Id: siteID})
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, grpcRead.Site,
		"REST Blank PUT Site", "rest-blank-put-site", "staging", "", "", "",
	)

	beforeRESTConcretePut := loadParitySitePresenceState(t, environment, siteID)
	restConcretePut := requestJSON(
		t, environment.router, http.MethodPut, itemPath,
		map[string]any{
			"name": "  REST Concrete PUT Site  ", "slug": "  rest-concrete-put-site  ",
			"status": "planned", "facility": "  R2  ",
			"description": "  REST concrete description  ",
			"comments":    "  REST concrete comments  ",
		},
		http.StatusOK,
	)
	require.Equal(t, "REST Concrete PUT Site", restConcretePut["name"])
	require.Equal(t, "rest-concrete-put-site", restConcretePut["slug"])
	requireChoiceValue(t, restConcretePut["status"], "planned")
	require.Equal(t, "R2", restConcretePut["facility"])
	require.Equal(t, "REST concrete description", restConcretePut["description"])
	require.Equal(t, "REST concrete comments", restConcretePut["comments"])
	afterRESTConcretePut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, beforeRESTConcretePut.siteCount, afterRESTConcretePut.siteCount)
	require.Equal(t, beforeRESTConcretePut.changeCount+1, afterRESTConcretePut.changeCount)
	require.Equal(t, beforeRESTConcretePut.totalChangeCount+1, afterRESTConcretePut.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterRESTConcretePut.row.Created)
	grpcRead, err = environment.dcim.GetSite(environment.ctx, &dcimv1.GetSiteRequest{Id: siteID})
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, grpcRead.Site,
		"REST Concrete PUT Site", "rest-concrete-put-site", "planned", "R2",
		"REST concrete description", "REST concrete comments",
	)

	beforeGRPCBlankPut := loadParitySitePresenceState(t, environment, siteID)
	grpcBlankPut, err := environment.dcim.ReplaceSite(
		environment.ctx,
		&dcimv1.ReplaceSiteRequest{Id: siteID, Site: &dcimv1.SiteInput{
			Name: pointer("gRPC Blank PUT Site"), Slug: pointer("grpc-blank-put-site"),
			Facility: &blankScalar, Description: &blankScalar, Comments: &blankScalar,
		}},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, grpcBlankPut.Site,
		"gRPC Blank PUT Site", "grpc-blank-put-site", "planned", "", "", "",
	)
	afterGRPCBlankPut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, beforeGRPCBlankPut.siteCount, afterGRPCBlankPut.siteCount)
	require.Equal(t, beforeGRPCBlankPut.changeCount+1, afterGRPCBlankPut.changeCount)
	require.Equal(t, beforeGRPCBlankPut.totalChangeCount+1, afterGRPCBlankPut.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterGRPCBlankPut.row.Created)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireChoiceValue(t, restRead["status"], "planned")
	require.Empty(t, restRead["facility"])
	require.Empty(t, restRead["description"])
	require.Empty(t, restRead["comments"])

	beforeGRPCConcretePut := loadParitySitePresenceState(t, environment, siteID)
	grpcConcretePut, err := environment.dcim.ReplaceSite(
		environment.ctx,
		&dcimv1.ReplaceSiteRequest{Id: siteID, Site: &dcimv1.SiteInput{
			Name: pointer("  gRPC Concrete PUT Site  "), Slug: pointer("  grpc-concrete-put-site  "),
			Status: pointer("retired"), Facility: pointer("  G3  "),
			Description: pointer("  gRPC concrete description  "),
			Comments:    pointer("  gRPC concrete comments  "),
		}},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(
		t, grpcConcretePut.Site,
		"gRPC Concrete PUT Site", "grpc-concrete-put-site", "retired", "G3",
		"gRPC concrete description", "gRPC concrete comments",
	)
	afterGRPCConcretePut := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, beforeGRPCConcretePut.siteCount, afterGRPCConcretePut.siteCount)
	require.Equal(t, beforeGRPCConcretePut.changeCount+1, afterGRPCConcretePut.changeCount)
	require.Equal(t, beforeGRPCConcretePut.totalChangeCount+1, afterGRPCConcretePut.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterGRPCConcretePut.row.Created)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	requireChoiceValue(t, restRead["status"], "retired")
	require.Equal(t, "G3", restRead["facility"])
	require.Equal(t, "gRPC concrete description", restRead["description"])
	require.Equal(t, "gRPC concrete comments", restRead["comments"])

	beforeRESTClear := loadParitySitePresenceState(t, environment, siteID)
	clearedByREST := requestJSON(
		t, environment.router, http.MethodPatch, itemPath,
		map[string]any{"facility": "", "description": "", "comments": ""},
		http.StatusOK,
	)
	require.Empty(t, clearedByREST["facility"])
	require.Empty(t, clearedByREST["description"])
	require.Empty(t, clearedByREST["comments"])
	afterRESTClear := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, beforeRESTClear.siteCount, afterRESTClear.siteCount)
	require.Equal(t, beforeRESTClear.changeCount+1, afterRESTClear.changeCount)
	require.Equal(t, beforeRESTClear.totalChangeCount+1, afterRESTClear.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterRESTClear.row.Created)
	grpcRead, err = environment.dcim.GetSite(environment.ctx, &dcimv1.GetSiteRequest{Id: siteID})
	require.NoError(t, err)
	require.Empty(t, grpcRead.Site.Facility)
	require.Empty(t, grpcRead.Site.Description)
	require.Empty(t, grpcRead.Site.Comments)

	setByGRPC, err := environment.dcim.UpdateSite(
		environment.ctx,
		&dcimv1.UpdateSiteRequest{
			Id: siteID,
			Site: &dcimv1.SiteInput{
				Facility: pointer("  G2  "), Description: pointer("  gRPC description  "),
				Comments: pointer("  gRPC comments  "), Status: pointer("decommissioning"),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"facility", "description", "comments", "status"}},
		},
	)
	require.NoError(t, err)
	require.Equal(t, "G2", setByGRPC.Site.Facility)
	require.Equal(t, "gRPC description", setByGRPC.Site.Description)
	require.Equal(t, "gRPC comments", setByGRPC.Site.Comments)
	require.Equal(t, "decommissioning", setByGRPC.Site.Status)
	afterGRPCSet := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, afterRESTClear.siteCount, afterGRPCSet.siteCount)
	require.Equal(t, afterRESTClear.changeCount+1, afterGRPCSet.changeCount)
	require.Equal(t, afterRESTClear.totalChangeCount+1, afterGRPCSet.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterGRPCSet.row.Created)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	require.Equal(t, "G2", restRead["facility"])
	require.Equal(t, "gRPC description", restRead["description"])
	require.Equal(t, "gRPC comments", restRead["comments"])
	requireChoiceValue(t, restRead["status"], "decommissioning")

	blank := ""
	clearedByGRPC, err := environment.dcim.UpdateSite(
		environment.ctx,
		&dcimv1.UpdateSiteRequest{
			Id:         siteID,
			Site:       &dcimv1.SiteInput{Facility: &blank, Description: &blank, Comments: &blank},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"facility", "description", "comments"}},
		},
	)
	require.NoError(t, err)
	require.Empty(t, clearedByGRPC.Site.Facility)
	require.Empty(t, clearedByGRPC.Site.Description)
	require.Empty(t, clearedByGRPC.Site.Comments)
	afterGRPCClear := loadParitySitePresenceState(t, environment, siteID)
	require.Equal(t, afterGRPCSet.siteCount, afterGRPCClear.siteCount)
	require.Equal(t, afterGRPCSet.changeCount+1, afterGRPCClear.changeCount)
	require.Equal(t, afterGRPCSet.totalChangeCount+1, afterGRPCClear.totalChangeCount)
	require.Equal(t, createdState.row.Created, afterGRPCClear.row.Created)
	restRead = requestJSON(t, environment.router, http.MethodGet, itemPath, nil, http.StatusOK)
	require.Empty(t, restRead["facility"])
	require.Empty(t, restRead["description"])
	require.Empty(t, restRead["comments"])

	beforeGRPCCreate := loadParitySitePresenceState(t, environment, siteID)
	createdByGRPC, err := environment.dcim.CreateSite(
		environment.ctx,
		&dcimv1.CreateSiteRequest{Site: &dcimv1.SiteInput{
			Name: pointer("  gRPC Default Site  "), Slug: pointer("  grpc-default-site  "),
		}},
	)
	require.NoError(t, err)
	requireSiteProtoScalars(t, createdByGRPC.Site, "gRPC Default Site", "grpc-default-site", "active", "", "", "")
	createdByGRPCState := loadParitySitePresenceState(t, environment, createdByGRPC.Site.Id)
	require.Equal(t, beforeGRPCCreate.siteCount+1, createdByGRPCState.siteCount)
	require.Equal(t, int64(1), createdByGRPCState.changeCount)
	require.Equal(t, beforeGRPCCreate.totalChangeCount+1, createdByGRPCState.totalChangeCount)
	restRead = requestJSON(
		t, environment.router, http.MethodGet,
		"/api/dcim/sites/"+strconv.FormatInt(createdByGRPC.Site.Id, 10), nil, http.StatusOK,
	)
	requireChoiceValue(t, restRead["status"], "active")
	require.Empty(t, restRead["facility"])
	require.Empty(t, restRead["description"])
	require.Empty(t, restRead["comments"])
}

type paritySitePresenceState struct {
	row              dcimrow.SiteRow
	siteCount        int64
	changeCount      int64
	totalChangeCount int64
}

func loadParitySitePresenceState(
	t *testing.T,
	environment *profileParityEnvironment,
	id int64,
) paritySitePresenceState {
	t.Helper()
	var state paritySitePresenceState
	require.NoError(t, environment.db.First(&state.row, id).Error)
	require.NoError(t, environment.db.Model(&dcimrow.SiteRow{}).Count(&state.siteCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Where(
		"kind = ? AND object_id = ?", domaindcim.SiteObjectType, id,
	).Count(&state.changeCount).Error)
	require.NoError(t, environment.db.Model(&postgreschangelog.ChangeRow{}).Count(&state.totalChangeCount).Error)
	return state
}

func requireSiteGRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(
		t,
		"Invalid input.",
		status.Convert(err).Message(),
		"public gRPC deliberately exposes the generic validation envelope; mapper and application matrices pin the field detail",
	)
}

func requireSiteProtoScalars(
	t *testing.T,
	site *dcimv1.Site,
	name, slug, siteStatus, facility, description, comments string,
) {
	t.Helper()
	require.NotNil(t, site)
	require.Equal(t, name, site.Name)
	require.Equal(t, slug, site.Slug)
	require.Equal(t, siteStatus, site.Status)
	require.Equal(t, facility, site.Facility)
	require.Equal(t, description, site.Description)
	require.Equal(t, comments, site.Comments)
}

func setSiteProtoScalar(input *dcimv1.SiteInput, field string, value *string) {
	switch field {
	case "name":
		input.Name = value
	case "slug":
		input.Slug = value
	case "status":
		input.Status = value
	case "facility":
		input.Facility = value
	case "description":
		input.Description = value
	case "comments":
		input.Comments = value
	}
}
