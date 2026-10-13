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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	grpcadapter "netbox-go/internal/adapters/grpc/workflow"
	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrow "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	restadapter "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/application/authz"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
	"netbox-go/internal/platform/composition"
)

func TestVRFScalarPresenceRESTGRPCParity(t *testing.T) {
	environment, observer, rpc := newVRFPresenceParityEnvironment(t)
	const collection = "/api/ipam/vrfs/"
	state := loadVRFPresenceParityState(t, environment)
	require.Empty(t, state.vrfs)
	require.Empty(t, state.changes)

	defaults := requestJSON(t, environment.router, http.MethodPost, collection, map[string]any{"name": "  REST default  "}, http.StatusCreated)
	defaultID := jsonID(t, defaults["id"])
	state = requireVRFPresenceParityMutation(t, environment, state, defaultID, true)
	defaultRow := vrfPresenceParityRow(t, state, defaultID)
	require.Nil(t, defaultRow.RD)
	require.True(t, defaultRow.EnforceUnique)
	require.Empty(t, defaultRow.Description)
	require.Empty(t, defaultRow.Comments)
	read, err := rpc.GetVRF(environment.ctx, &ipamv1.GetVRFRequest{Id: defaultID})
	require.NoError(t, err)
	requireVRFPresenceParityProjection(t, defaults, read.Vrf, defaultRow, false)

	grpcDefault, err := rpc.CreateVRF(environment.ctx, &ipamv1.CreateVRFRequest{Vrf: &ipamv1.VRFInput{Name: pointer("  gRPC default  ")}})
	require.NoError(t, err)
	state = requireVRFPresenceParityMutation(t, environment, state, grpcDefault.Vrf.Id, true)
	grpcDefaultRow := vrfPresenceParityRow(t, state, grpcDefault.Vrf.Id)
	require.Nil(t, grpcDefaultRow.RD)
	require.True(t, grpcDefaultRow.EnforceUnique)
	require.Empty(t, grpcDefaultRow.Description)
	require.Empty(t, grpcDefaultRow.Comments)
	requireVRFPresenceParityProjection(t, requestJSON(t, environment.router, http.MethodGet, vrfPresenceParityPath(grpcDefault.Vrf.Id), nil, http.StatusOK), grpcDefault.Vrf, grpcDefaultRow, true)

	// POST/PUT protobuf absence represents omission, not an independently representable null.
	nullCreated := requestJSON(t, environment.router, http.MethodPost, collection, map[string]any{"name": "REST explicit null RD", "rd": nil}, http.StatusCreated)
	state = requireVRFPresenceParityMutation(t, environment, state, jsonID(t, nullCreated["id"]), true)
	require.Nil(t, nullCreated["rd"])
	require.Nil(t, vrfPresenceParityRow(t, state, jsonID(t, nullCreated["id"])).RD)

	for _, transport := range []string{"REST", "gRPC"} {
		t.Run("accepted presence/"+transport, func(t *testing.T) {
			before := loadVRFPresenceParityState(t, environment)
			input := &ipamv1.VRFInput{
				Name: pointer("  " + transport + " concrete  "), Rd: wrapperspb.String("  " + transport + " arbitrary RD  "),
				EnforceUnique: pointer(false), Description: pointer("  concrete description  "), Comments: pointer("  " + strings.Repeat("c", 240) + "  "),
			}
			var id int64
			if transport == "REST" {
				response := requestJSON(t, environment.router, http.MethodPost, collection, vrfPresenceParityBody(input), http.StatusCreated)
				id = jsonID(t, response["id"])
				read, err := rpc.GetVRF(environment.ctx, &ipamv1.GetVRFRequest{Id: id})
				require.NoError(t, err)
				after := requireVRFPresenceParityMutation(t, environment, before, id, true)
				requireVRFPresenceParityProjection(t, response, read.Vrf, vrfPresenceParityRow(t, after, id), false)
			} else {
				response, err := rpc.CreateVRF(environment.ctx, &ipamv1.CreateVRFRequest{Vrf: input})
				require.NoError(t, err)
				id = response.Vrf.Id
				after := requireVRFPresenceParityMutation(t, environment, before, id, true)
				requireVRFPresenceParityProjection(t, requestJSON(t, environment.router, http.MethodGet, vrfPresenceParityPath(id), nil, http.StatusOK), response.Vrf, vrfPresenceParityRow(t, after, id), true)
			}
			current := loadVRFPresenceParityState(t, environment)
			row := vrfPresenceParityRow(t, current, id)
			require.Equal(t, transport+" concrete", row.Name)
			require.Equal(t, transport+" arbitrary RD", *row.RD)
			require.False(t, row.EnforceUnique)
			require.Equal(t, "concrete description", row.Description)
			require.Equal(t, strings.Repeat("c", 240), row.Comments)

			for _, method := range []string{http.MethodPut, http.MethodPatch} {
				prior := vrfPresenceParityRow(t, current, id)
				name := transport + " " + method
				input := &ipamv1.VRFInput{Name: pointer("  " + name + "  ")}
				mutateVRFPresenceParity(t, environment, rpc, transport, method, id, input)
				current = requireVRFPresenceParityMutation(t, environment, current, id, false)
				row = vrfPresenceParityRow(t, current, id)
				require.Equal(t, name, row.Name)
				prior.Name, prior.LastUpdated = row.Name, row.LastUpdated
				require.Equal(t, prior, row, "all four omitted optional fields must retain concrete RD, false and text")
				requireVRFPresenceParityRead(t, environment, rpc, row)
			}

			// A supplied blank is not null. Only one blank RD is stored at a time.
			mutateVRFPresenceParity(t, environment, rpc, transport, http.MethodPut, id, &ipamv1.VRFInput{
				Name: pointer("Blank RD " + transport), Rd: wrapperspb.String("  "), EnforceUnique: pointer(false), Description: pointer("  "), Comments: pointer("  "),
			})
			current = requireVRFPresenceParityMutation(t, environment, current, id, false)
			row = vrfPresenceParityRow(t, current, id)
			require.NotNil(t, row.RD)
			require.Empty(t, *row.RD)
			require.False(t, row.EnforceUnique)
			require.Empty(t, row.Description)
			require.Empty(t, row.Comments)
			requireVRFPresenceParityRead(t, environment, rpc, row)
			if transport == "REST" {
				response := requestJSON(t, environment.router, http.MethodPatch, vrfPresenceParityPath(id), map[string]any{"rd": nil, "enforce_unique": false}, http.StatusOK)
				require.Nil(t, response["rd"])
				require.Equal(t, float64(0), response["ipaddress_count"])
				require.Equal(t, float64(0), response["prefix_count"])
			} else {
				response, err := rpc.UpdateVRF(environment.ctx, &ipamv1.UpdateVRFRequest{Id: id, Vrf: &ipamv1.VRFInput{EnforceUnique: pointer(false)}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"rd", "enforce_unique"}}})
				require.NoError(t, err)
				require.Nil(t, response.Vrf.Rd)
			}
			current = requireVRFPresenceParityMutation(t, environment, current, id, false)
			row = vrfPresenceParityRow(t, current, id)
			require.Nil(t, row.RD)
			require.False(t, row.EnforceUnique)
			requireVRFPresenceParityRead(t, environment, rpc, row)
			for _, value := range []bool{true, false} {
				prior := row
				mutateVRFPresenceParity(t, environment, rpc, transport, http.MethodPatch, id, &ipamv1.VRFInput{EnforceUnique: pointer(value)})
				current = requireVRFPresenceParityMutation(t, environment, current, id, false)
				row = vrfPresenceParityRow(t, current, id)
				require.Equal(t, value, row.EnforceUnique)
				prior.EnforceUnique, prior.LastUpdated = value, row.LastUpdated
				require.Equal(t, prior, row)
				requireVRFPresenceParityRead(t, environment, rpc, row)
			}
		})
	}

	for _, field := range []string{"name", "enforce_unique", "description", "comments"} {
		t.Run("shared explicit-null rejection/"+field, func(t *testing.T) {
			before := loadVRFPresenceParityState(t, environment)
			observer.remember(nil)
			response := requestJSON(t, environment.router, http.MethodPatch, vrfPresenceParityPath(defaultID), map[string]any{field: nil}, http.StatusBadRequest)
			violations := observer.violations()
			require.Equal(t, []shared.FieldViolation{{Field: field, Reason: "null", Description: "This field may not be null."}}, violations)
			requireVRFPresenceRESTViolations(t, response, violations)
			observer.remember(nil)
			_, err := rpc.UpdateVRF(environment.ctx, &ipamv1.UpdateVRFRequest{Id: defaultID, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}}})
			requireVRFPresenceRPCInvalid(t, err)
			require.Equal(t, violations, observer.violations(), "known mask with absent wrapper must reach shared explicit-null validation")
			require.Equal(t, before, loadVRFPresenceParityState(t, environment))
			for _, method := range []string{http.MethodPost, http.MethodPut} {
				body := map[string]any{"name": "Rejected VRF", field: nil}
				path := vrfPresenceParityPath(defaultID)
				if method == http.MethodPost {
					path = collection
				}
				observer.remember(nil)
				response := requestJSON(t, environment.router, method, path, body, http.StatusBadRequest)
				require.Equal(t, violations, observer.violations())
				requireVRFPresenceRESTViolations(t, response, violations)
				require.Equal(t, before, loadVRFPresenceParityState(t, environment))
			}
		})
	}

	for _, test := range []struct {
		name     string
		input    *ipamv1.VRFInput
		expected []string
	}{
		{"blank name", &ipamv1.VRFInput{Name: pointer("  ")}, []string{"name/blank"}},
		{"name limit", &ipamv1.VRFInput{Name: pointer(strings.Repeat("名", 101))}, []string{"name/max_length"}},
		{"RD limit", &ipamv1.VRFInput{Rd: wrapperspb.String(strings.Repeat("界", 22))}, []string{"rd/max_length"}},
		{"description limit", &ipamv1.VRFInput{Description: pointer(strings.Repeat("界", 201))}, []string{"description/max_length"}},
		{"name NUL", &ipamv1.VRFInput{Name: pointer("bad\x00name")}, []string{"name/null_characters_not_allowed"}},
		{"RD NUL", &ipamv1.VRFInput{Rd: wrapperspb.String("bad\x00RD")}, []string{"rd/null_characters_not_allowed"}},
		{"description NUL", &ipamv1.VRFInput{Description: pointer("bad\x00description")}, []string{"description/null_characters_not_allowed"}},
		{"comments NUL", &ipamv1.VRFInput{Comments: pointer("bad\x00comments")}, []string{"comments/null_characters_not_allowed"}},
		{"independent RD length and NUL", &ipamv1.VRFInput{Rd: wrapperspb.String(strings.Repeat("r", 21) + "\x00")}, []string{"rd/max_length", "rd/null_characters_not_allowed"}},
		{"aggregate errors", &ipamv1.VRFInput{Name: pointer(" "), Rd: wrapperspb.String(strings.Repeat("r", 21) + "\x00"), Description: pointer(strings.Repeat("d", 200) + "\x00"), Comments: pointer("bad\x00comments")}, []string{"name/blank", "rd/max_length", "rd/null_characters_not_allowed", "description/max_length", "description/null_characters_not_allowed", "comments/null_characters_not_allowed"}},
	} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			t.Run("shared validation/"+method+"/"+test.name, func(t *testing.T) {
				before := loadVRFPresenceParityState(t, environment)
				input := proto.Clone(test.input).(*ipamv1.VRFInput)
				if method != http.MethodPatch && input.Name == nil {
					input.Name = pointer("Rejected VRF")
				}
				path := vrfPresenceParityPath(defaultID)
				if method == http.MethodPost {
					path = collection
				}
				observer.remember(nil)
				response := requestJSON(t, environment.router, method, path, vrfPresenceParityBody(input), http.StatusBadRequest)
				violations := observer.violations()
				var keys []string
				for _, violation := range violations {
					keys = append(keys, violation.Field+"/"+violation.Reason)
				}
				require.Equal(t, test.expected, keys, "exact field/reason order, no duplicates or fabricated update_mask error")
				requireVRFPresenceRESTViolations(t, response, violations)
				observer.remember(nil)
				var err error
				switch method {
				case http.MethodPost:
					_, err = rpc.CreateVRF(environment.ctx, &ipamv1.CreateVRFRequest{Vrf: input})
				case http.MethodPut:
					_, err = rpc.ReplaceVRF(environment.ctx, &ipamv1.ReplaceVRFRequest{Id: defaultID, Vrf: input})
				default:
					_, err = rpc.UpdateVRF(environment.ctx, &ipamv1.UpdateVRFRequest{Id: defaultID, Vrf: input})
				}
				requireVRFPresenceRPCInvalid(t, err)
				require.Equal(t, violations, observer.violations())
				require.Equal(t, before, loadVRFPresenceParityState(t, environment))
			})
		}
	}

	t.Run("required omission parity and undeclared mask containment", func(t *testing.T) {
		before := loadVRFPresenceParityState(t, environment)
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			path := vrfPresenceParityPath(defaultID)
			if method == http.MethodPost {
				path = collection
			}
			observer.remember(nil)
			requestJSON(t, environment.router, method, path, map[string]any{}, http.StatusBadRequest)
			violations := observer.violations()
			require.Equal(t, []shared.FieldViolation{{Field: "name", Reason: "required", Description: "This field is required."}}, violations)
			observer.remember(nil)
			var err error
			if method == http.MethodPost {
				_, err = rpc.CreateVRF(environment.ctx, &ipamv1.CreateVRFRequest{})
			} else {
				_, err = rpc.ReplaceVRF(environment.ctx, &ipamv1.ReplaceVRFRequest{Id: defaultID})
			}
			requireVRFPresenceRPCInvalid(t, err)
			require.Equal(t, violations, observer.violations())
		}
		observer.remember(nil)
		requestJSON(t, environment.router, http.MethodPatch, vrfPresenceParityPath(defaultID), map[string]any{"tenant": 7}, http.StatusBadRequest)
		require.Empty(t, observer.violations(), "unknown REST fields must not reach the shared service")
		observer.remember(nil)
		_, err := rpc.UpdateVRF(environment.ctx, &ipamv1.UpdateVRFRequest{Id: defaultID, Vrf: &ipamv1.VRFInput{Description: pointer("must not persist")}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"tenant"}}})
		requireVRFPresenceRPCInvalid(t, err)
		require.Empty(t, observer.violations(), "unknown masks must not reach the shared service")
		require.Equal(t, before, loadVRFPresenceParityState(t, environment))
	})
}

func mutateVRFPresenceParity(t *testing.T, environment *profileParityEnvironment, rpc *grpcadapter.VRFRPCHandler, transport, method string, id int64, input *ipamv1.VRFInput) {
	t.Helper()
	if transport == "REST" {
		response := requestJSON(t, environment.router, method, vrfPresenceParityPath(id), vrfPresenceParityBody(input), http.StatusOK)
		require.Equal(t, float64(0), response["ipaddress_count"], "only POST omits annotated counts")
		require.Equal(t, float64(0), response["prefix_count"])
		return
	}
	if method == http.MethodPut {
		_, err := rpc.ReplaceVRF(environment.ctx, &ipamv1.ReplaceVRFRequest{Id: id, Vrf: input})
		require.NoError(t, err)
	} else {
		_, err := rpc.UpdateVRF(environment.ctx, &ipamv1.UpdateVRFRequest{Id: id, Vrf: input})
		require.NoError(t, err)
	}
}

func vrfPresenceParityBody(input *ipamv1.VRFInput) map[string]any {
	body := map[string]any{}
	if input.Name != nil {
		body["name"] = *input.Name
	}
	if input.Rd != nil {
		body["rd"] = input.Rd.Value
	}
	if input.EnforceUnique != nil {
		body["enforce_unique"] = *input.EnforceUnique
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if input.Comments != nil {
		body["comments"] = *input.Comments
	}
	return body
}

func vrfPresenceParityPath(id int64) string {
	return "/api/ipam/vrfs/" + strconv.FormatInt(id, 10) + "/"
}

type vrfPresenceParityState struct {
	vrfs    []ipamrow.VRFRow
	changes []postgreschangelog.ChangeRow
}

func loadVRFPresenceParityState(t *testing.T, environment *profileParityEnvironment) vrfPresenceParityState {
	t.Helper()
	var state vrfPresenceParityState
	require.NoError(t, environment.db.Order("id").Find(&state.vrfs).Error)
	require.NoError(t, environment.db.Order("id").Find(&state.changes).Error)
	for _, model := range []any{&ipamrow.PrefixRow{}, &ipamrow.IPAddressRow{}} {
		var count int64
		require.NoError(t, environment.db.Model(model).Count(&count).Error)
		require.Zero(t, count)
	}
	return state
}

func vrfPresenceParityRow(t *testing.T, state vrfPresenceParityState, id int64) ipamrow.VRFRow {
	t.Helper()
	for _, row := range state.vrfs {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("VRF %d is absent", id)
	return ipamrow.VRFRow{}
}

func requireVRFPresenceParityMutation(t *testing.T, environment *profileParityEnvironment, before vrfPresenceParityState, id int64, create bool) vrfPresenceParityState {
	t.Helper()
	after := loadVRFPresenceParityState(t, environment)
	count := len(before.vrfs)
	if create {
		count++
	}
	require.Len(t, after.vrfs, count)
	for _, prior := range before.vrfs {
		if prior.ID != id {
			require.Equal(t, prior, vrfPresenceParityRow(t, after, prior.ID))
		}
	}
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	current := vrfPresenceParityRow(t, after, id)
	change := after.changes[len(after.changes)-1]
	require.Positive(t, change.ID)
	require.Equal(t, environment.principal.ID, change.ActorID)
	require.Equal(t, domainipam.VRFObjectType, change.Kind)
	require.Equal(t, id, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(current.LastUpdated))
	if create {
		require.Equal(t, "create", change.Action)
		require.Empty(t, change.BeforeData)
	} else {
		require.Equal(t, "update", change.Action)
		prior := vrfPresenceParityRow(t, before, id)
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

func requireVRFPresenceParityRead(t *testing.T, environment *profileParityEnvironment, rpc *grpcadapter.VRFRPCHandler, row ipamrow.VRFRow) {
	t.Helper()
	read, err := rpc.GetVRF(environment.ctx, &ipamv1.GetVRFRequest{Id: row.ID})
	require.NoError(t, err)
	requireVRFPresenceParityProjection(t, requestJSON(t, environment.router, http.MethodGet, vrfPresenceParityPath(row.ID), nil, http.StatusOK), read.Vrf, row, true)
}

func requireVRFPresenceParityProjection(t *testing.T, rest map[string]any, rpc *ipamv1.VRF, row ipamrow.VRFRow, counts bool) {
	t.Helper()
	require.NotNil(t, rpc)
	require.Equal(t, row.ID, rpc.Id)
	require.Equal(t, float64(row.ID), rest["id"])
	require.Equal(t, vrfPresenceParityPath(row.ID), rpc.Url)
	require.Equal(t, rpc.Url, rest["url"])
	require.Equal(t, rpc.Display, rest["display"])
	require.Equal(t, row.Name, rpc.Name)
	require.Equal(t, row.Name, rest["name"])
	require.Equal(t, row.EnforceUnique, rpc.EnforceUnique)
	require.Equal(t, row.EnforceUnique, rest["enforce_unique"])
	require.Equal(t, row.Description, rpc.Description)
	require.Equal(t, row.Description, rest["description"])
	require.Equal(t, row.Comments, rpc.Comments)
	require.Equal(t, row.Comments, rest["comments"])
	if row.RD == nil {
		require.Nil(t, rpc.Rd)
		require.Nil(t, rest["rd"])
	} else {
		require.NotNil(t, rpc.Rd)
		require.Equal(t, *row.RD, rpc.Rd.Value)
		require.Equal(t, *row.RD, rest["rd"])
	}
	require.NotNil(t, rpc.Created)
	require.NotNil(t, rpc.LastUpdated)
	require.NoError(t, rpc.Created.CheckValid())
	require.NoError(t, rpc.LastUpdated.CheckValid())
	require.True(t, row.Created.Equal(rpc.Created.AsTime()))
	require.True(t, row.LastUpdated.Equal(rpc.LastUpdated.AsTime()))
	for field, expected := range map[string]time.Time{"created": row.Created, "last_updated": row.LastUpdated} {
		text, ok := rest[field].(string)
		require.True(t, ok, "newly persisted timestamp must be an RFC3339 string")
		actual, err := time.Parse(time.RFC3339Nano, text)
		require.NoError(t, err)
		require.True(t, expected.Equal(actual))
	}
	require.Zero(t, rpc.IpaddressCount)
	require.Zero(t, rpc.PrefixCount)
	if counts {
		require.Equal(t, float64(0), rest["ipaddress_count"])
		require.Equal(t, float64(0), rest["prefix_count"])
	} else {
		require.NotContains(t, rest, "ipaddress_count")
		require.NotContains(t, rest, "prefix_count")
	}
}

func requireVRFPresenceRESTViolations(t *testing.T, response map[string]any, violations []shared.FieldViolation) {
	t.Helper()
	expected := map[string]any{}
	for _, violation := range violations {
		if expected[violation.Field] == nil {
			expected[violation.Field] = []any{}
		}
		expected[violation.Field] = append(expected[violation.Field].([]any), violation.Description)
	}
	require.Equal(t, expected, response)
}

func requireVRFPresenceRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

// Diagnostics are observed at the actual shared service boundary, not invented on the gRPC wire.
type vrfPresenceParityObserver struct {
	restadapter.VRFService
	mutex sync.Mutex
	err   error
}

func (observer *vrfPresenceParityObserver) remember(err error) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.err = err
}

func (observer *vrfPresenceParityObserver) violations() []shared.FieldViolation {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	return append([]shared.FieldViolation(nil), shared.ViolationsOf(observer.err)...)
}

func (observer *vrfPresenceParityObserver) CreateVRF(ctx context.Context, principal identity.Principal, command applicationipam.CreateVRFCommand) (*domainipam.VRF, error) {
	result, err := observer.VRFService.CreateVRF(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *vrfPresenceParityObserver) ReplaceVRF(ctx context.Context, principal identity.Principal, command applicationipam.ReplaceVRFCommand) (*domainipam.VRF, error) {
	result, err := observer.VRFService.ReplaceVRF(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *vrfPresenceParityObserver) UpdateVRF(ctx context.Context, principal identity.Principal, command applicationipam.UpdateVRFCommand) (*domainipam.VRF, error) {
	result, err := observer.VRFService.UpdateVRF(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func newVRFPresenceParityEnvironment(t *testing.T) (*profileParityEnvironment, *vrfPresenceParityObserver, *grpcadapter.VRFRPCHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared&_foreign_keys=on"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	entries := []bootstrap.Entry{{Name: "go_identity_users", Model: &identityrow.UserRow{}}}
	for _, descriptor := range dcimrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	for _, descriptor := range ipamrow.Descriptors() {
		entries = append(entries, bootstrap.Entry{Name: descriptor.Name, Model: descriptor.Model, Dependencies: descriptor.Dependencies})
	}
	entries = append(entries, bootstrap.Entry{Name: "go_object_changes", Model: &postgreschangelog.ChangeRow{}, Dependencies: []string{"go_identity_users"}})
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(entries))
	require.Empty(t, result.Existing)
	now := time.Date(2026, time.October, 13, 0, 40, 0, 0, time.UTC)
	require.NoError(t, db.Create(&identityrow.UserRow{ID: 1, Username: "vrf-parity-test", PasswordHash: "not-used-by-tests", IsActive: true, IsSuperuser: true, Permissions: []byte(`[]`), Created: now, Updated: now}).Error)
	principal := identity.Principal{ID: 1, Username: "vrf-parity-test", IsSuperuser: true}
	core := composition.NewCoreWithAuthorizer(db, authz.AllowAll{})
	observer := &vrfPresenceParityObserver{VRFService: core.VRFs}
	router := gin.New()
	restadapter.NewVRFRESTHandler(observer).Register(router, func(c *gin.Context) {
		restadapter.SetPrincipal(c, principal)
		c.Next()
	})
	environment := &profileParityEnvironment{db: db, router: router, principal: principal, ctx: identity.WithPrincipal(t.Context(), principal)}
	return environment, observer, grpcadapter.NewVRFRPCHandler(observer)
}
