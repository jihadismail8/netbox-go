package parity

import (
	"context"
	"encoding/json"
	"fmt"
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
	postgresipam "netbox-go/internal/adapters/postgres/ipam"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	postgrestransaction "netbox-go/internal/adapters/postgres/transaction"
	restadapter "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/application/authz"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestPrefixScalarPresenceRESTGRPCParity(t *testing.T) {
	environment := newPrefixPresenceParityEnvironment(t)
	const collection = "/api/ipam/prefixes/"
	state := loadPrefixPresenceParityState(t, environment)
	require.Empty(t, state.prefixes)
	require.Empty(t, state.changes)

	for index, transport := range []string{"REST", "gRPC"} {
		input := &ipamv1.PrefixInput{Prefix: pointer([]string{"192.0.2.0/24", "198.51.100.0/24"}[index])}
		id := createPrefixPresenceParity(t, environment, transport, input)
		state = requirePrefixPresenceParityMutation(t, environment, state, id, true)
		row := prefixPresenceParityRow(t, state, id)
		require.Nil(t, row.VRFID)
		require.Equal(t, "active", row.Status)
		require.False(t, row.IsPool)
		require.False(t, row.MarkUtilized)
		require.Empty(t, row.Description)
		require.Empty(t, row.Comments)
		requirePrefixPresenceParityRead(t, environment, row)
	}
	defaultID := state.prefixes[0].ID

	// POST/PUT protobuf nil represents omission, not independently representable null.
	environment.clock.advance()
	nullCreated := requestJSON(t, environment.router, http.MethodPost, collection, map[string]any{"prefix": "203.0.113.0/24", "vrf": nil}, http.StatusCreated)
	state = requirePrefixPresenceParityMutation(t, environment, state, jsonID(t, nullCreated["id"]), true)
	require.Nil(t, nullCreated["vrf"])
	require.Nil(t, prefixPresenceParityRow(t, state, jsonID(t, nullCreated["id"])).VRFID)
	requirePrefixPresenceParityResponse(t, environment, jsonID(t, nullCreated["id"]), nullCreated, nil)

	for index, transport := range []string{"REST", "gRPC"} {
		t.Run("accepted presence/"+transport, func(t *testing.T) {
			current := loadPrefixPresenceParityState(t, environment)
			input := &ipamv1.PrefixInput{
				Prefix: pointer(fmt.Sprintf("2001:db8:%x::/64", 130+index)), Vrf: wrapperspb.Int64(environment.vrfs[index].ID),
				Status: pointer("reserved"), IsPool: pointer(true), MarkUtilized: pointer(true),
				Description: pointer("  concrete description  "), Comments: pointer("  " + strings.Repeat("c", 240) + "  "),
			}
			id := createPrefixPresenceParity(t, environment, transport, input)
			current = requirePrefixPresenceParityMutation(t, environment, current, id, true)
			row := prefixPresenceParityRow(t, current, id)
			require.Equal(t, *input.Prefix, row.Prefix)
			require.Equal(t, environment.vrfs[index].ID, *row.VRFID)
			require.Equal(t, "reserved", row.Status)
			require.True(t, row.IsPool)
			require.True(t, row.MarkUtilized)
			require.Equal(t, "concrete description", row.Description)
			require.Equal(t, strings.Repeat("c", 240), row.Comments)
			requirePrefixPresenceParityRead(t, environment, row)

			for methodIndex, method := range []string{http.MethodPut, http.MethodPatch} {
				prior := row
				network := fmt.Sprintf("2001:db8:%x::/64", 140+index*10+methodIndex)
				mutatePrefixPresenceParity(t, environment, transport, method, id, &ipamv1.PrefixInput{Prefix: pointer(network)})
				current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
				row = prefixPresenceParityRow(t, current, id)
				require.Equal(t, network, row.Prefix)
				prior.Prefix, prior.LastUpdated = network, row.LastUpdated
				require.Equal(t, prior, row, "all six omitted optional fields preserve concrete VRF, true flags, status and text")
				requirePrefixPresenceParityRead(t, environment, row)
			}
			for _, method := range []string{http.MethodPut, http.MethodPatch} {
				for _, value := range []bool{false, true} {
					prior := row
					input := &ipamv1.PrefixInput{IsPool: pointer(value), MarkUtilized: pointer(value)}
					if method == http.MethodPut {
						input.Prefix = pointer(row.Prefix)
					}
					mutatePrefixPresenceParity(t, environment, transport, method, id, input)
					current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
					row = prefixPresenceParityRow(t, current, id)
					require.Equal(t, value, row.IsPool)
					require.Equal(t, value, row.MarkUtilized)
					prior.IsPool, prior.MarkUtilized, prior.LastUpdated = value, value, row.LastUpdated
					require.Equal(t, prior, row)
					requirePrefixPresenceParityRead(t, environment, row)
				}
			}
			for _, choice := range []string{"container", "active", "reserved", "deprecated"} {
				prior := row
				mutatePrefixPresenceParity(t, environment, transport, http.MethodPatch, id, &ipamv1.PrefixInput{Status: pointer(choice)})
				current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
				row = prefixPresenceParityRow(t, current, id)
				require.Equal(t, choice, row.Status)
				prior.Status, prior.LastUpdated = choice, row.LastUpdated
				require.Equal(t, prior, row)
				requirePrefixPresenceParityRead(t, environment, row)
			}
			mutatePrefixPresenceParity(t, environment, transport, http.MethodPut, id, &ipamv1.PrefixInput{
				Prefix: pointer(row.Prefix), IsPool: pointer(false), MarkUtilized: pointer(false),
				Description: pointer(" \t "), Comments: pointer(" \t "),
			})
			current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
			row = prefixPresenceParityRow(t, current, id)
			require.Empty(t, row.Description)
			require.Empty(t, row.Comments)
			require.False(t, row.IsPool)
			require.False(t, row.MarkUtilized)
			require.NotNil(t, row.VRFID)
			requirePrefixPresenceParityRead(t, environment, row)

			if transport == "REST" {
				prior := row
				environment.clock.advance()
				response := requestJSON(t, environment.router, http.MethodPut, prefixPresenceParityPath(id), map[string]any{"prefix": row.Prefix, "vrf": nil}, http.StatusOK)
				current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
				row = prefixPresenceParityRow(t, current, id)
				require.Nil(t, row.VRFID)
				prior.VRFID, prior.LastUpdated = nil, row.LastUpdated
				require.Equal(t, prior, row, "PUT null clears only nullable VRF; protobuf PUT cannot independently express that null")
				requirePrefixPresenceParityResponse(t, environment, id, response, nil)
				mutatePrefixPresenceParity(t, environment, transport, http.MethodPatch, id, &ipamv1.PrefixInput{Vrf: wrapperspb.Int64(environment.vrfs[index].ID)})
				current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
				row = prefixPresenceParityRow(t, current, id)
				require.Equal(t, environment.vrfs[index].ID, *row.VRFID)
			}

			prior := row
			environment.clock.advance()
			if transport == "REST" {
				response := requestJSON(t, environment.router, http.MethodPatch, prefixPresenceParityPath(id), map[string]any{"vrf": nil}, http.StatusOK)
				requirePrefixPresenceParityResponse(t, environment, id, response, nil)
			} else {
				response, err := environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: id, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"vrf"}}})
				require.NoError(t, err)
				require.NotNil(t, response.Prefix)
				requirePrefixPresenceParityResponse(t, environment, id, nil, response.Prefix)
			}
			current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
			row = prefixPresenceParityRow(t, current, id)
			require.Nil(t, row.VRFID)
			prior.VRFID, prior.LastUpdated = nil, row.LastUpdated
			require.Equal(t, prior, row)
			requirePrefixPresenceParityRead(t, environment, row)

			mutatePrefixPresenceParity(t, environment, transport, http.MethodPatch, id, &ipamv1.PrefixInput{
				Vrf:         wrapperspb.Int64(environment.vrfs[1-index].ID),
				Description: pointer("  " + strings.Repeat("界", 200) + "  "), Comments: pointer("  " + strings.Repeat("文", 240) + "  "),
			})
			current = requirePrefixPresenceParityMutation(t, environment, current, id, false)
			row = prefixPresenceParityRow(t, current, id)
			require.Equal(t, environment.vrfs[1-index].ID, *row.VRFID)
			require.Equal(t, strings.Repeat("界", 200), row.Description)
			require.Equal(t, strings.Repeat("文", 240), row.Comments)
			requirePrefixPresenceParityRead(t, environment, row)
		})
	}

	for _, field := range []string{"prefix", "status", "is_pool", "mark_utilized", "description", "comments"} {
		t.Run("shared explicit null/"+field, func(t *testing.T) {
			expected := []shared.FieldViolation{{Field: field, Reason: "null", Description: "This field may not be null."}}
			if field == "status" {
				expected[0].Reason, expected[0].Description = "invalid", "This field may not be blank."
			}
			for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
				before := loadPrefixPresenceParityState(t, environment)
				body := map[string]any{field: nil}
				if method != http.MethodPatch && field != "prefix" {
					body["prefix"] = "10.240.0.0/24"
				}
				environment.rest.reset()
				response := requestJSON(t, environment.router, method, prefixPresenceParityMutationPath(method, defaultID), body, http.StatusBadRequest)
				require.Equal(t, expected, environment.rest.violations())
				requirePrefixPresenceRESTViolations(t, response, expected)
				require.Equal(t, before, loadPrefixPresenceParityState(t, environment), "each rejected REST mutation is independently non-effecting")
			}
			before := loadPrefixPresenceParityState(t, environment)
			environment.grpc.reset()
			_, err := environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: defaultID, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{field}}})
			requirePrefixPresenceRPCInvalid(t, err)
			require.Equal(t, expected, environment.grpc.violations(), "known mask nil must reach the actual shared core, not mapper update_mask validation")
			require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
		})
	}

	for _, test := range []struct {
		name     string
		input    *ipamv1.PrefixInput
		expected []shared.FieldViolation
	}{
		{"blank prefix", &ipamv1.PrefixInput{Prefix: pointer(" \t ")}, []shared.FieldViolation{{Field: "prefix", Reason: "blank", Description: "This field may not be blank."}}},
		{"blank status", &ipamv1.PrefixInput{Status: pointer("")}, []shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "This field may not be blank."}}},
		{"whitespace status", &ipamv1.PrefixInput{Status: pointer(" ")}, []shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "  is not a valid choice."}}},
		{"padded status", &ipamv1.PrefixInput{Status: pointer(" reserved ")}, []shared.FieldViolation{{Field: "status", Reason: "invalid", Description: " reserved  is not a valid choice."}}},
		{"case changed status", &ipamv1.PrefixInput{Status: pointer("Reserved")}, []shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "Reserved is not a valid choice."}}},
		{"unknown status", &ipamv1.PrefixInput{Status: pointer("unknown")}, []shared.FieldViolation{{Field: "status", Reason: "invalid", Description: "unknown is not a valid choice."}}},
		{"VRF zero", &ipamv1.PrefixInput{Vrf: wrapperspb.Int64(0)}, []shared.FieldViolation{{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: 0"}}},
		{"VRF negative", &ipamv1.PrefixInput{Vrf: wrapperspb.Int64(-1)}, []shared.FieldViolation{{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: -1"}}},
		{"VRF absent positive ID", &ipamv1.PrefixInput{Vrf: wrapperspb.Int64(999999)}, []shared.FieldViolation{{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: 999999"}}},
		{"description NUL", &ipamv1.PrefixInput{Description: pointer("d\x00")}, []shared.FieldViolation{{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}}},
		{"comments NUL", &ipamv1.PrefixInput{Comments: pointer("c\x00")}, []shared.FieldViolation{{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."}}},
		{"description length and NUL", &ipamv1.PrefixInput{Description: pointer(strings.Repeat("界", 200) + "\x00")}, []shared.FieldViolation{
			{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
			{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
		}},
	} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			t.Run("shared validation/"+method+"/"+test.name, func(t *testing.T) {
				input := proto.Clone(test.input).(*ipamv1.PrefixInput)
				if method != http.MethodPatch && input.Prefix == nil {
					input.Prefix = pointer("10.241.0.0/24")
				}
				requirePrefixPresenceParityRejected(t, environment, method, defaultID, input, test.expected)
			})
		}
	}
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run("required prefix omission/"+method, func(t *testing.T) {
			requirePrefixPresenceParityRejected(t, environment, method, defaultID, &ipamv1.PrefixInput{}, []shared.FieldViolation{
				{Field: "prefix", Reason: "required", Description: "This field is required."},
			})
		})
	}

	t.Run("mixed null diagnostic parity", func(t *testing.T) {
		expected := []shared.FieldViolation{
			{Field: "prefix", Reason: "null", Description: "This field may not be null."},
			{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: 0"},
			{Field: "status", Reason: "invalid", Description: " unknown  is not a valid choice."},
			{Field: "is_pool", Reason: "null", Description: "This field may not be null."},
			{Field: "mark_utilized", Reason: "null", Description: "This field may not be null."},
			{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
			{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
			{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
		}
		body := map[string]any{
			"prefix": nil, "vrf": 0, "status": " unknown ", "is_pool": nil, "mark_utilized": nil,
			"description": strings.Repeat("界", 200) + "\x00", "comments": "c\x00",
		}
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			before := loadPrefixPresenceParityState(t, environment)
			environment.rest.reset()
			response := requestJSON(t, environment.router, method, prefixPresenceParityMutationPath(method, defaultID), body, http.StatusBadRequest)
			require.Equal(t, expected, environment.rest.violations())
			requirePrefixPresenceRESTViolations(t, response, expected)
			require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
		}
		before := loadPrefixPresenceParityState(t, environment)
		environment.grpc.reset()
		_, err := environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{
			Id: defaultID, Prefix: &ipamv1.PrefixInput{
				Vrf: wrapperspb.Int64(0), Status: pointer(" unknown "),
				Description: pointer(strings.Repeat("界", 200) + "\x00"), Comments: pointer("c\x00"),
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"prefix", "vrf", "status", "is_pool", "mark_utilized", "description", "comments"}},
		})
		requirePrefixPresenceRPCInvalid(t, err)
		require.Equal(t, expected, environment.grpc.violations(), "no fallback blank, duplicate field, or fabricated update_mask diagnostic")
		require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
	})

	t.Run("undeclared fields remain adapter fail closed", func(t *testing.T) {
		before := loadPrefixPresenceParityState(t, environment)
		environment.rest.reset()
		response := requestJSON(t, environment.router, http.MethodPatch, prefixPresenceParityPath(defaultID), map[string]any{"unsupported": true}, http.StatusBadRequest)
		require.Equal(t, map[string]any{"unsupported": []any{"This field is not supported by the active capability profile."}}, response)
		require.Zero(t, environment.rest.calls())
		require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
		environment.grpc.reset()
		_, err := environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: defaultID, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"unsupported"}}})
		requirePrefixPresenceRPCInvalid(t, err)
		require.Zero(t, environment.grpc.calls())
		require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
	})
	// This in-process boundary earns neither a pinned external differential nor a compatibility tier.
}

func TestPrefixNetworkRESTGRPCParity(t *testing.T) {
	cases := []struct {
		name, input, canonical, description string
	}{
		{"canonical IPv4", "198.51.100.0/24", "198.51.100.0/24", ""},
		{"bare IPv4 host", "203.0.113.17", "203.0.113.17/32", ""},
		{"canonical IPv6", "2001:db8:10::/64", "2001:db8:10::/64", ""},
		{"expanded IPv6 host", "2001:0DB8:0020:0000:0000:0000:0000:0017", "2001:db8:20::17/128", ""},
		{"expanded IPv6 network", "2001:0DB8:0030:0000:0000:0000:0000:0000/64", "2001:db8:30::/64", ""},
		{"IPv4 host bits", "203.0.113.17/24", "", "203.0.113.17/24 is not a valid prefix. Did you mean 203.0.113.0/24?"},
		{"IPv6 host bits", "2001:db8:40::17/64", "", "2001:db8:40::17/64 is not a valid prefix. Did you mean 2001:db8:40::/64?"},
		{"IPv4 slash zero", "0.0.0.0/0", "", "Cannot create prefix with /0 mask."},
		{"IPv6 slash zero", "::/0", "", "Cannot create prefix with /0 mask."},
	}
	for _, test := range cases {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			for _, transport := range []string{"REST", "gRPC"} {
				t.Run(test.name+"/"+method+"/"+transport, func(t *testing.T) {
					environment := newPrefixPresenceParityEnvironment(t)
					before := loadPrefixPresenceParityState(t, environment)
					require.Empty(t, before.prefixes)
					require.Empty(t, before.changes)
					expected := ipamrow.PrefixRow{
						Prefix: "192.0.2.0/24", VRFID: pointer(environment.vrfs[0].ID), Status: "reserved", IsPool: true, MarkUtilized: true,
						Description: "network proof description", Comments: "network proof comments",
					}
					var id int64
					if method != http.MethodPost {
						expected.Created, expected.LastUpdated = environment.clock.Now().Add(time.Minute), environment.clock.Now().Add(time.Minute)
						environment.rest.reset()
						environment.grpc.reset()
						id = createPrefixPresenceParity(t, environment, transport, prefixNetworkParityInput("192.0.2.0/24", environment.vrfs[0].ID))
						expected.ID = id
						before = requirePrefixPresenceParityMutation(t, environment, before, id, true)
						requirePrefixNetworkParityRow(t, prefixPresenceParityRow(t, before, id), expected)
					}
					input := &ipamv1.PrefixInput{Prefix: pointer(test.input)}
					if method == http.MethodPost {
						input = prefixNetworkParityInput(test.input, environment.vrfs[0].ID)
					}
					environment.rest.reset()
					environment.grpc.reset()
					if test.description != "" {
						violations := []shared.FieldViolation{{Field: "prefix", Reason: "invalid", Description: test.description}}
						if transport == "REST" {
							response := requestJSON(t, environment.router, method, prefixPresenceParityMutationPath(method, id), prefixPresenceParityBody(input), http.StatusBadRequest)
							require.Equal(t, 1, environment.rest.calls())
							require.Zero(t, environment.grpc.calls())
							require.Equal(t, violations, environment.rest.violations())
							requirePrefixPresenceRESTViolations(t, response, violations)
						} else {
							var err error
							switch method {
							case http.MethodPost:
								var response *ipamv1.CreatePrefixResponse
								response, err = environment.rpc.CreatePrefix(environment.ctx, &ipamv1.CreatePrefixRequest{Prefix: input})
								require.Nil(t, response)
							case http.MethodPut:
								var response *ipamv1.ReplacePrefixResponse
								response, err = environment.rpc.ReplacePrefix(environment.ctx, &ipamv1.ReplacePrefixRequest{Id: id, Prefix: input})
								require.Nil(t, response)
							case http.MethodPatch:
								var response *ipamv1.UpdatePrefixResponse
								response, err = environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: id, Prefix: input})
								require.Nil(t, response)
							}
							requirePrefixPresenceRPCInvalid(t, err)
							require.Equal(t, 1, environment.grpc.calls())
							require.Zero(t, environment.rest.calls())
							require.Equal(t, violations, environment.grpc.violations())
						}
						require.Equal(t, before, loadPrefixPresenceParityState(t, environment), "each rejected transport mutation preserves complete rows, timestamps, relations and change snapshots")
						return
					}
					at := environment.clock.Now().Add(time.Minute)
					environment.clock.advance()
					var rest map[string]any
					var rpc *ipamv1.Prefix
					if transport == "REST" {
						wantStatus := http.StatusOK
						if method == http.MethodPost {
							wantStatus = http.StatusCreated
						}
						rest = requestJSON(t, environment.router, method, prefixPresenceParityMutationPath(method, id), prefixPresenceParityBody(input), wantStatus)
						if method == http.MethodPost {
							id = jsonID(t, rest["id"])
						}
					} else {
						switch method {
						case http.MethodPost:
							response, err := environment.rpc.CreatePrefix(environment.ctx, &ipamv1.CreatePrefixRequest{Prefix: input})
							require.NoError(t, err)
							require.NotNil(t, response)
							rpc = response.Prefix
						case http.MethodPut:
							response, err := environment.rpc.ReplacePrefix(environment.ctx, &ipamv1.ReplacePrefixRequest{Id: id, Prefix: input})
							require.NoError(t, err)
							require.NotNil(t, response)
							rpc = response.Prefix
						case http.MethodPatch:
							response, err := environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: id, Prefix: input})
							require.NoError(t, err)
							require.NotNil(t, response)
							rpc = response.Prefix
						}
						require.NotNil(t, rpc)
						if method == http.MethodPost {
							id = rpc.Id
						}
					}
					if method == http.MethodPost {
						expected.ID, expected.Created = id, at
					}
					expected.Prefix, expected.LastUpdated = test.canonical, at
					after := requirePrefixPresenceParityMutation(t, environment, before, id, method == http.MethodPost)
					requirePrefixNetworkParityRow(t, prefixPresenceParityRow(t, after, id), expected)
					if transport == "REST" {
						require.Equal(t, 1, environment.rest.calls())
						require.Empty(t, environment.rest.violations())
						require.Zero(t, environment.grpc.calls())
					} else {
						require.Equal(t, 1, environment.grpc.calls())
						require.Empty(t, environment.grpc.violations())
						require.Zero(t, environment.rest.calls())
					}
					requirePrefixPresenceParityResponse(t, environment, id, rest, rpc)
				})
			}
		}
	}
}

func prefixNetworkParityInput(network string, vrfID int64) *ipamv1.PrefixInput {
	return &ipamv1.PrefixInput{
		Prefix: pointer(network), Vrf: wrapperspb.Int64(vrfID), Status: pointer("reserved"), IsPool: pointer(true), MarkUtilized: pointer(true),
		Description: pointer("network proof description"), Comments: pointer("network proof comments"),
	}
}

func requirePrefixNetworkParityRow(t *testing.T, actual, expected ipamrow.PrefixRow) {
	t.Helper()
	require.True(t, actual.Created.Equal(expected.Created))
	require.True(t, actual.LastUpdated.Equal(expected.LastUpdated))
	actual.Created, actual.LastUpdated = expected.Created, expected.LastUpdated
	require.Equal(t, expected, actual)
}

func createPrefixPresenceParity(t *testing.T, environment *prefixPresenceParityEnvironment, transport string, input *ipamv1.PrefixInput) int64 {
	t.Helper()
	environment.clock.advance()
	if transport == "REST" {
		response := requestJSON(t, environment.router, http.MethodPost, "/api/ipam/prefixes/", prefixPresenceParityBody(input), http.StatusCreated)
		id := jsonID(t, response["id"])
		requirePrefixPresenceParityResponse(t, environment, id, response, nil)
		return id
	}
	response, err := environment.rpc.CreatePrefix(environment.ctx, &ipamv1.CreatePrefixRequest{Prefix: input})
	require.NoError(t, err)
	require.NotNil(t, response.Prefix)
	requirePrefixPresenceParityResponse(t, environment, response.Prefix.Id, nil, response.Prefix)
	return response.Prefix.Id
}

func mutatePrefixPresenceParity(t *testing.T, environment *prefixPresenceParityEnvironment, transport, method string, id int64, input *ipamv1.PrefixInput) {
	t.Helper()
	environment.clock.advance()
	if transport == "REST" {
		response := requestJSON(t, environment.router, method, prefixPresenceParityPath(id), prefixPresenceParityBody(input), http.StatusOK)
		requirePrefixPresenceParityResponse(t, environment, id, response, nil)
		return
	}
	var err error
	var projection *ipamv1.Prefix
	if method == http.MethodPut {
		var response *ipamv1.ReplacePrefixResponse
		response, err = environment.rpc.ReplacePrefix(environment.ctx, &ipamv1.ReplacePrefixRequest{Id: id, Prefix: input})
		if response != nil {
			projection = response.Prefix
		}
	} else {
		var response *ipamv1.UpdatePrefixResponse
		response, err = environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: id, Prefix: input})
		if response != nil {
			projection = response.Prefix
		}
	}
	require.NoError(t, err)
	require.NotNil(t, projection, "the mutation itself must return its owned response projection")
	requirePrefixPresenceParityResponse(t, environment, id, nil, projection)
}

func requirePrefixPresenceParityRejected(t *testing.T, environment *prefixPresenceParityEnvironment, method string, id int64, input *ipamv1.PrefixInput, expected []shared.FieldViolation) {
	t.Helper()
	before := loadPrefixPresenceParityState(t, environment)
	environment.rest.reset()
	response := requestJSON(t, environment.router, method, prefixPresenceParityMutationPath(method, id), prefixPresenceParityBody(input), http.StatusBadRequest)
	require.Equal(t, expected, environment.rest.violations())
	requirePrefixPresenceRESTViolations(t, response, expected)
	require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
	environment.grpc.reset()
	var err error
	switch method {
	case http.MethodPost:
		_, err = environment.rpc.CreatePrefix(environment.ctx, &ipamv1.CreatePrefixRequest{Prefix: input})
	case http.MethodPut:
		_, err = environment.rpc.ReplacePrefix(environment.ctx, &ipamv1.ReplacePrefixRequest{Id: id, Prefix: input})
	default:
		_, err = environment.rpc.UpdatePrefix(environment.ctx, &ipamv1.UpdatePrefixRequest{Id: id, Prefix: input})
	}
	requirePrefixPresenceRPCInvalid(t, err)
	require.Equal(t, expected, environment.grpc.violations(), "compare exact shared typed diagnostics, not trimmed normalization or invented wire details")
	require.Equal(t, before, loadPrefixPresenceParityState(t, environment))
}

func prefixPresenceParityBody(input *ipamv1.PrefixInput) map[string]any {
	body := map[string]any{}
	if input.Prefix != nil {
		body["prefix"] = *input.Prefix
	}
	if input.Vrf != nil {
		body["vrf"] = input.Vrf.Value
	}
	if input.Status != nil {
		body["status"] = *input.Status
	}
	if input.IsPool != nil {
		body["is_pool"] = *input.IsPool
	}
	if input.MarkUtilized != nil {
		body["mark_utilized"] = *input.MarkUtilized
	}
	if input.Description != nil {
		body["description"] = *input.Description
	}
	if input.Comments != nil {
		body["comments"] = *input.Comments
	}
	return body
}

func prefixPresenceParityPath(id int64) string {
	return "/api/ipam/prefixes/" + strconv.FormatInt(id, 10) + "/"
}

func prefixPresenceParityMutationPath(method string, id int64) string {
	if method == http.MethodPost {
		return "/api/ipam/prefixes/"
	}
	return prefixPresenceParityPath(id)
}

type prefixPresenceParityState struct {
	prefixes []ipamrow.PrefixRow
	vrfs     []ipamrow.VRFRow
	changes  []postgreschangelog.ChangeRow
}

func loadPrefixPresenceParityState(t *testing.T, environment *prefixPresenceParityEnvironment) prefixPresenceParityState {
	t.Helper()
	var state prefixPresenceParityState
	require.NoError(t, environment.db.Order("id").Find(&state.prefixes).Error)
	require.NoError(t, environment.db.Order("id").Find(&state.vrfs).Error)
	require.NoError(t, environment.db.Order("id").Find(&state.changes).Error)
	var addresses int64
	require.NoError(t, environment.db.Model(&ipamrow.IPAddressRow{}).Count(&addresses).Error)
	require.Zero(t, addresses, "empty support rows earn no downstream IPAddress claim")
	return state
}

func prefixPresenceParityRow(t *testing.T, state prefixPresenceParityState, id int64) ipamrow.PrefixRow {
	t.Helper()
	for _, row := range state.prefixes {
		if row.ID == id {
			return row
		}
	}
	t.Fatalf("Prefix %d is absent", id)
	return ipamrow.PrefixRow{}
}

func requirePrefixPresenceParityMutation(t *testing.T, environment *prefixPresenceParityEnvironment, before prefixPresenceParityState, id int64, create bool) prefixPresenceParityState {
	t.Helper()
	after := loadPrefixPresenceParityState(t, environment)
	count := len(before.prefixes)
	if create {
		count++
	}
	require.Len(t, after.prefixes, count)
	require.Equal(t, before.vrfs, after.vrfs)
	for _, prior := range before.prefixes {
		if prior.ID != id {
			require.Equal(t, prior, prefixPresenceParityRow(t, after, prior.ID))
		}
	}
	current := prefixPresenceParityRow(t, after, id)
	require.False(t, current.Created.IsZero())
	require.True(t, current.LastUpdated.Equal(environment.clock.Now().Time))
	require.Len(t, after.changes, len(before.changes)+1)
	require.Equal(t, before.changes, after.changes[:len(before.changes)])
	change := after.changes[len(after.changes)-1]
	require.Positive(t, change.ID)
	require.Equal(t, environment.principal.ID, change.ActorID)
	require.Equal(t, domainipam.PrefixObjectType, change.Kind)
	require.Equal(t, id, change.ObjectID)
	require.True(t, change.OccurredAt.Equal(current.LastUpdated))
	if create {
		require.True(t, current.Created.Equal(current.LastUpdated))
		require.Equal(t, "create", change.Action)
		require.Empty(t, change.BeforeData)
	} else {
		require.Equal(t, "update", change.Action)
		prior := prefixPresenceParityRow(t, before, id)
		require.True(t, prior.Created.Equal(current.Created))
		requirePrefixPresenceParityPayload(t, change.BeforeData, prior)
	}
	requirePrefixPresenceParityPayload(t, change.AfterData, current)
	return after
}

func requirePrefixPresenceParityPayload(t *testing.T, data []byte, row ipamrow.PrefixRow) {
	t.Helper()
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(data, &payload))
	require.Len(t, payload, 7)
	for _, key := range []string{"prefix", "vrf", "status", "is_pool", "mark_utilized", "description", "comments"} {
		require.Contains(t, payload, key)
	}
	expected, err := json.Marshal(map[string]any{
		"prefix": row.Prefix, "vrf": row.VRFID, "status": row.Status, "is_pool": row.IsPool,
		"mark_utilized": row.MarkUtilized, "description": row.Description, "comments": row.Comments,
	})
	require.NoError(t, err)
	require.JSONEq(t, string(expected), string(data))
}

func requirePrefixPresenceParityRead(t *testing.T, environment *prefixPresenceParityEnvironment, row ipamrow.PrefixRow) {
	t.Helper()
	read, err := environment.rpc.GetPrefix(environment.ctx, &ipamv1.GetPrefixRequest{Id: row.ID})
	require.NoError(t, err)
	rest := requestJSON(t, environment.router, http.MethodGet, prefixPresenceParityPath(row.ID), nil, http.StatusOK)
	requirePrefixPresenceParityProjection(t, environment, rest, read.Prefix, row)
}

func requirePrefixPresenceParityResponse(t *testing.T, environment *prefixPresenceParityEnvironment, id int64, rest map[string]any, rpc *ipamv1.Prefix) {
	t.Helper()
	if rest == nil {
		rest = requestJSON(t, environment.router, http.MethodGet, prefixPresenceParityPath(id), nil, http.StatusOK)
	}
	if rpc == nil {
		read, err := environment.rpc.GetPrefix(environment.ctx, &ipamv1.GetPrefixRequest{Id: id})
		require.NoError(t, err)
		rpc = read.Prefix
	}
	row := prefixPresenceParityRow(t, loadPrefixPresenceParityState(t, environment), id)
	requirePrefixPresenceParityProjection(t, environment, rest, rpc, row)
}

func requirePrefixPresenceParityProjection(t *testing.T, environment *prefixPresenceParityEnvironment, rest map[string]any, rpc *ipamv1.Prefix, row ipamrow.PrefixRow) {
	t.Helper()
	require.NotNil(t, rpc)
	require.Equal(t, row.ID, rpc.Id)
	require.Equal(t, float64(row.ID), rest["id"])
	require.Equal(t, prefixPresenceParityPath(row.ID), rpc.Url)
	require.Equal(t, rpc.Url, rest["url"])
	require.Equal(t, row.Prefix, rpc.Display)
	require.Equal(t, row.Prefix, rest["display"])
	require.Equal(t, row.Prefix, rpc.Prefix)
	require.Equal(t, row.Prefix, rest["prefix"])
	family, familyLabel := uint32(4), "IPv4"
	if strings.Contains(row.Prefix, ":") {
		family, familyLabel = 6, "IPv6"
	}
	require.Equal(t, family, rpc.Family)
	require.Equal(t, map[string]any{"value": float64(family), "label": familyLabel}, rest["family"])
	statusLabel := map[string]string{"container": "Container", "active": "Active", "reserved": "Reserved", "deprecated": "Deprecated"}[row.Status]
	require.Equal(t, row.Status, rpc.Status)
	require.Equal(t, map[string]any{"value": row.Status, "label": statusLabel}, rest["status"])
	require.Equal(t, row.IsPool, rpc.IsPool)
	require.Equal(t, row.IsPool, rest["is_pool"])
	require.Equal(t, row.MarkUtilized, rpc.MarkUtilized)
	require.Equal(t, row.MarkUtilized, rest["mark_utilized"])
	require.Equal(t, row.Description, rpc.Description)
	require.Equal(t, row.Description, rest["description"])
	require.Equal(t, row.Comments, rpc.Comments)
	require.Equal(t, row.Comments, rest["comments"])
	if row.VRFID == nil {
		require.Nil(t, rpc.VrfId)
		require.Nil(t, rest["vrf"])
	} else {
		require.NotNil(t, rpc.VrfId)
		require.Equal(t, *row.VRFID, rpc.VrfId.Value)
		var display string
		for _, vrf := range environment.vrfs {
			if vrf.ID == *row.VRFID {
				display = vrf.Name
				if vrf.RD != nil && *vrf.RD != "" {
					display += " (" + *vrf.RD + ")"
				}
			}
		}
		require.NotEmpty(t, display)
		require.Equal(t, map[string]any{
			"id": float64(*row.VRFID), "url": "/api/ipam/vrfs/" + strconv.FormatInt(*row.VRFID, 10) + "/", "display": display,
		}, rest["vrf"])
	}
	require.NotNil(t, rpc.Created)
	require.NotNil(t, rpc.LastUpdated)
	require.NoError(t, rpc.Created.CheckValid())
	require.NoError(t, rpc.LastUpdated.CheckValid())
	require.True(t, row.Created.Equal(rpc.Created.AsTime()))
	require.True(t, row.LastUpdated.Equal(rpc.LastUpdated.AsTime()))
	for field, expected := range map[string]time.Time{"created": row.Created, "last_updated": row.LastUpdated} {
		text, ok := rest[field].(string)
		require.True(t, ok)
		actual, err := time.Parse(time.RFC3339Nano, text)
		require.NoError(t, err)
		require.False(t, actual.IsZero())
		require.True(t, expected.Equal(actual))
	}
	require.Zero(t, rpc.Children)
	require.Zero(t, rpc.Depth)
	require.Equal(t, float64(0), rest["children"])
	require.Equal(t, float64(0), rest["_depth"], "flat nonoverlapping fixtures earn no broader hierarchy contract")
}

func requirePrefixPresenceRESTViolations(t *testing.T, response map[string]any, violations []shared.FieldViolation) {
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

func requirePrefixPresenceRPCInvalid(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	require.Equal(t, codes.InvalidArgument, status.Code(err))
	require.Equal(t, "Invalid input.", status.Convert(err).Message())
}

// Each transport observes the same real core through its own resettable diagnostic recorder.
type prefixPresenceParityObserver struct {
	restadapter.PrefixService
	mutex sync.Mutex
	err   error
	count int
}

func (observer *prefixPresenceParityObserver) reset() {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.err, observer.count = nil, 0
}

func (observer *prefixPresenceParityObserver) remember(err error) {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	observer.err, observer.count = err, observer.count+1
}

func (observer *prefixPresenceParityObserver) violations() []shared.FieldViolation {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	return append([]shared.FieldViolation(nil), shared.ViolationsOf(observer.err)...)
}

func (observer *prefixPresenceParityObserver) calls() int {
	observer.mutex.Lock()
	defer observer.mutex.Unlock()
	return observer.count
}

func (observer *prefixPresenceParityObserver) CreatePrefix(ctx context.Context, principal identity.Principal, command applicationipam.CreatePrefixCommand) (*domainipam.Prefix, error) {
	result, err := observer.PrefixService.CreatePrefix(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *prefixPresenceParityObserver) ReplacePrefix(ctx context.Context, principal identity.Principal, command applicationipam.ReplacePrefixCommand) (*domainipam.Prefix, error) {
	result, err := observer.PrefixService.ReplacePrefix(ctx, principal, command)
	observer.remember(err)
	return result, err
}

func (observer *prefixPresenceParityObserver) UpdatePrefix(ctx context.Context, principal identity.Principal, command applicationipam.UpdatePrefixCommand) (*domainipam.Prefix, error) {
	result, err := observer.PrefixService.UpdatePrefix(ctx, principal, command)
	observer.remember(err)
	return result, err
}

type prefixPresenceParityClock struct {
	mutex sync.Mutex
	now   time.Time
}

func (clock *prefixPresenceParityClock) Now() shared.Timestamp {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	return shared.NewTimestamp(clock.now)
}

func (clock *prefixPresenceParityClock) advance() {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	clock.now = clock.now.Add(time.Minute)
}

type prefixPresenceParityEnvironment struct {
	*profileParityEnvironment
	clock *prefixPresenceParityClock
	rest  *prefixPresenceParityObserver
	grpc  *prefixPresenceParityObserver
	rpc   *grpcadapter.PrefixRPCHandler
	vrfs  []ipamrow.VRFRow
}

func newPrefixPresenceParityEnvironment(t *testing.T) *prefixPresenceParityEnvironment {
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
	for _, entry := range entries {
		require.False(t, db.Migrator().HasTable(entry.Model))
	}
	registry, err := bootstrap.NewRegistry(entries...)
	require.NoError(t, err)
	result, err := bootstrap.Run(t.Context(), db, registry)
	require.NoError(t, err)
	require.Len(t, result.Created, len(entries))
	require.Empty(t, result.Existing)
	now := time.Date(2026, time.October, 13, 8, 10, 0, 0, time.UTC)
	require.NoError(t, db.Create(&identityrow.UserRow{
		ID: 1, Username: "prefix-parity-test", PasswordHash: "not-used-by-tests",
		IsActive: true, IsSuperuser: true, Permissions: []byte("[]"), Created: now, Updated: now,
	}).Error)
	vrfs := []ipamrow.VRFRow{
		{RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Prefix parity VRF one", RD: pointer("64512:1331"), EnforceUnique: true},
		{RowMetadata: ipamrow.RowMetadata{Created: now, LastUpdated: now}, Name: "Prefix parity VRF two", RD: pointer("64512:1332"), EnforceUnique: true},
	}
	require.NoError(t, db.Create(&vrfs).Error)
	principal := identity.Principal{ID: 1, Username: "prefix-parity-test", IsSuperuser: true}
	clock := &prefixPresenceParityClock{now: now}
	core, err := applicationipam.NewPrefixService(
		postgresipam.NewPrefixRepository(db), postgresipam.NewVRFRepository(db),
		postgrestransaction.NewUnitOfWork(db), postgreschangelog.NewRecorder(db), authz.AllowAll{}, clock,
	)
	require.NoError(t, err)
	restObserver := &prefixPresenceParityObserver{PrefixService: core}
	grpcObserver := &prefixPresenceParityObserver{PrefixService: core}
	router := gin.New()
	restadapter.NewPrefixRESTHandler(restObserver).Register(router, func(c *gin.Context) {
		restadapter.SetPrincipal(c, principal)
		c.Next()
	})
	return &prefixPresenceParityEnvironment{
		profileParityEnvironment: &profileParityEnvironment{db: db, router: router, principal: principal, ctx: identity.WithPrincipal(t.Context(), principal)},
		clock:                    clock, rest: restObserver, grpc: grpcObserver, rpc: grpcadapter.NewPrefixRPCHandler(grpcObserver), vrfs: vrfs,
	}
}
