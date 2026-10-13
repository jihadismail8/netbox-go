package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

// The spy checks decoding/projection/error presentation, not business semantics;
// the actual shared service is exercised in the separate parity matrix.
func TestPrefixRESTScalarPresenceMatrix(t *testing.T) {
	for _, operation := range []struct {
		method, path string
		status       int
	}{
		{http.MethodPost, "/api/ipam/prefixes/", http.StatusCreated},
		{http.MethodPut, "/api/ipam/prefixes/17/", http.StatusOK},
		{http.MethodPatch, "/api/ipam/prefixes/17/", http.StatusOK},
	} {
		t.Run(operation.method, func(t *testing.T) {
			request := func(spy *prefixRESTPresenceSpy, body string) *httptest.ResponseRecorder {
				router := newTestRouter()
				NewPrefixRESTHandler(spy).Register(router, testPrefixPrincipalMiddleware())
				response := httptest.NewRecorder()
				input := httptest.NewRequest(operation.method, operation.path, strings.NewReader(body))
				input.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(response, input)
				return response
			}
			fixture := restPrefixFixture(t, 17, "192.0.2.0/24", domainipam.NullVRFReference(), 2, 3)
			for _, field := range []string{"prefix", "vrf", "status", "is_pool", "mark_utilized", "description", "comments"} {
				for _, presence := range []string{"omitted", "null", "blank-zero-false", "concrete"} {
					t.Run("decoded intent/"+field+"/"+presence, func(t *testing.T) {
						want := applicationipam.CreatePrefixCommand{}
						stringField := applicationipam.OmittedField[string]()
						boolField := applicationipam.OmittedField[bool]()
						idField := applicationipam.OmittedField[int64]()
						var value any
						switch presence {
						case "null":
							stringField, boolField, idField = applicationipam.NullField[string](), applicationipam.NullField[bool](), applicationipam.NullField[int64]()
						case "blank-zero-false", "concrete":
							text, flag, id := "", false, int64(0)
							if presence == "concrete" {
								text, flag, id = " raw "+field+"\x00 ", true, 7
							}
							// Raw lexical strings reach the core unchanged; this spy validates none.
							value = text
							switch field {
							case "vrf":
								value = id
							case "is_pool", "mark_utilized":
								value = flag
							}
							stringField, boolField, idField = applicationipam.FieldValue(text), applicationipam.FieldValue(flag), applicationipam.FieldValue(id)
						}
						switch field {
						case "prefix":
							want.Prefix = stringField
						case "vrf":
							want.VRF = idField
						case "status":
							want.Status = stringField
						case "is_pool":
							want.IsPool = boolField
						case "mark_utilized":
							want.MarkUtilized = boolField
						case "description":
							want.Description = stringField
						case "comments":
							want.Comments = stringField
						}
						body := "{}"
						if presence != "omitted" {
							encoded, err := json.Marshal(map[string]any{field: value})
							require.NoError(t, err)
							body = string(encoded)
						}
						spy := &prefixRESTPresenceSpy{result: fixture}
						response := request(spy, body)
						require.Equal(t, operation.status, response.Code, response.Body.String())
						require.Equal(t, 1, spy.calls)
						require.Equal(t, operation.method, spy.method)
						require.Equal(t, identity.Principal{ID: 1, Username: "prefix-transport"}, spy.principal)
						require.Equal(t, want, spy.command, "all unselected fields remain omitted")
						id := shared.ID(17)
						if operation.method == http.MethodPost {
							id = 0
						}
						require.Equal(t, id, spy.id)
					})
				}
				wrongTypes, message := []string{`0`, `false`, `[]`, `{}`}, "Must be a string."
				switch field {
				case "vrf":
					wrongTypes, message = []string{`"7"`, `true`, `1.5`, `[]`, `{}`, `9223372036854775808`, `-9223372036854775809`}, "A valid object ID is required."
				case "is_pool", "mark_utilized":
					wrongTypes, message = []string{`0`, `"false"`, `""`, `[]`, `{}`}, "Must be a valid boolean."
				}
				for _, raw := range wrongTypes {
					t.Run("wrong scalar type/"+field+"/"+raw, func(t *testing.T) {
						spy := &prefixRESTPresenceSpy{result: fixture}
						response := request(spy, `{"`+field+`":`+raw+`}`)
						require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
						require.Zero(t, spy.calls)
						expected, err := json.Marshal(map[string][]string{field: {message}})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), response.Body.String())
					})
				}
			}
			for _, id := range []int64{-9, -9223372036854775808, 9223372036854775807} {
				body, err := json.Marshal(map[string]any{"vrf": id})
				require.NoError(t, err)
				spy := &prefixRESTPresenceSpy{result: fixture}
				response := request(spy, string(body))
				require.Equal(t, operation.status, response.Code, response.Body.String())
				require.Equal(t, 1, spy.calls)
				require.Equal(t, applicationipam.FieldValue(id), spy.command.VRF)
			}
			for _, field := range []string{"tenant", "role", "scope", "vlan", "tags", "custom_fields", "id", "url", "display", "family", "created", "last_updated", "children", "_depth"} {
				t.Run("undeclared field/"+field, func(t *testing.T) {
					spy := &prefixRESTPresenceSpy{result: fixture}
					response := request(spy, `{"prefix":"192.0.2.0/24","`+field+`":null}`)
					require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
					require.Zero(t, spy.calls)
					expected, err := json.Marshal(map[string][]string{field: {"This field is not supported by the active capability profile."}})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), response.Body.String())
				})
			}
			for _, body := range []string{"", `null`, `[]`, `"Prefix"`, `{"prefix":}`} {
				spy := &prefixRESTPresenceSpy{result: fixture}
				response := request(spy, body)
				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				require.Zero(t, spy.calls)
				require.JSONEq(t, `{"non_field_errors":["Expected a JSON object."]}`, response.Body.String())
			}
			for _, body := range []string{`{} {}`, `{} null`, `{} junk`} {
				spy := &prefixRESTPresenceSpy{result: fixture}
				response := request(spy, body)
				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				require.Zero(t, spy.calls)
				require.JSONEq(t, `{"non_field_errors":["Request body must contain one JSON object."]}`, response.Body.String())
			}
			t.Run("complete response envelope", func(t *testing.T) {
				for _, network := range []string{"192.0.2.0/24", "2001:db8::/48"} {
					for _, nullable := range []domainipam.NullableVRFReference{
						domainipam.NullVRFReference(),
						domainipam.NonNullVRFReference(restPrefixVRFReference(t, 7, "Blue", "65000:7", true)),
					} {
						state := restPrefixFixture(t, 17, network, nullable, 2, 3).State()
						state.IsPool, state.MarkUtilized, state.Description, state.Comments = false, false, "", ""
						state.LastUpdated = shared.NewTimestamp(state.Created.Add(time.Minute))
						item, err := domainipam.RestorePrefix(state)
						require.NoError(t, err)
						spy := &prefixRESTPresenceSpy{result: item}
						response := request(spy, `{"prefix":"198.51.100.0/24"}`)
						require.Equal(t, operation.status, response.Code, response.Body.String())
						var payload map[string]any
						require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
						var vrf any
						if !nullable.IsNull() {
							vrf = map[string]any{"id": float64(7), "url": "/api/ipam/vrfs/7/", "display": "Blue (65000:7)"}
						}
						family, label := float64(4), "IPv4"
						if item.Family() == 6 {
							family, label = 6, "IPv6"
						}
						require.Equal(t, map[string]any{
							"id": float64(17), "url": "/api/ipam/prefixes/17/", "display": network,
							"prefix": network, "vrf": vrf, "family": map[string]any{"value": family, "label": label},
							"status":  map[string]any{"value": "reserved", "label": "Reserved"},
							"is_pool": false, "mark_utilized": false, "description": "", "comments": "",
							"created": state.Created.UTC().Format(time.RFC3339Nano), "last_updated": state.LastUpdated.UTC().Format(time.RFC3339Nano),
							"children": float64(2), "_depth": float64(3),
						}, payload)
						if operation.method == http.MethodPost {
							require.Equal(t, "/api/ipam/prefixes/17/", response.Header().Get("Location"))
						} else {
							require.Empty(t, response.Header().Get("Location"))
						}
					}
				}
			})
			t.Run("canonical shared field errors", func(t *testing.T) {
				spy := &prefixRESTPresenceSpy{result: fixture, err: shared.NewValidationError(
					shared.FieldViolation{Field: "prefix", Reason: "null", Description: "This field may not be null."},
					shared.FieldViolation{Field: "vrf", Reason: "invalid", Description: "Related object not found using the provided numeric ID: -9"},
					shared.FieldViolation{Field: "status", Reason: "invalid", Description: "This field may not be blank."},
					shared.FieldViolation{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
					shared.FieldViolation{Field: "description", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
				)}
				response := request(spy, `{"prefix":null,"vrf":-9,"status":null,"description":"\u0000"}`)
				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				require.Equal(t, 1, spy.calls)
				require.JSONEq(t, `{"prefix":["This field may not be null."],"vrf":["Related object not found using the provided numeric ID: -9"],"status":["This field may not be blank."],"description":["Ensure this field has no more than 200 characters.","Null characters are not allowed."]}`, response.Body.String())
			})
			t.Run("infrastructure errors are redacted", func(t *testing.T) {
				spy := &prefixRESTPresenceSpy{result: fixture, err: errors.New("private storage failure")}
				response := request(spy, `{"prefix":"192.0.2.0/24"}`)
				require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
				require.Equal(t, 1, spy.calls)
				require.JSONEq(t, `{"detail":"An internal error occurred."}`, response.Body.String())
			})
		})
	}
}

type prefixRESTPresenceSpy struct {
	PrefixService
	calls     int
	method    string
	principal identity.Principal
	id        shared.ID
	command   applicationipam.CreatePrefixCommand
	result    *domainipam.Prefix
	err       error
}

func (spy *prefixRESTPresenceSpy) CreatePrefix(_ context.Context, principal identity.Principal, command applicationipam.CreatePrefixCommand) (*domainipam.Prefix, error) {
	spy.calls++
	spy.method, spy.principal, spy.command = http.MethodPost, principal, command
	return spy.result, spy.err
}

func (spy *prefixRESTPresenceSpy) ReplacePrefix(_ context.Context, principal identity.Principal, command applicationipam.ReplacePrefixCommand) (*domainipam.Prefix, error) {
	spy.calls++
	spy.method, spy.principal, spy.id, spy.command = http.MethodPut, principal, command.ID, command.CreatePrefixCommand
	return spy.result, spy.err
}

func (spy *prefixRESTPresenceSpy) UpdatePrefix(_ context.Context, principal identity.Principal, command applicationipam.UpdatePrefixCommand) (*domainipam.Prefix, error) {
	spy.calls++
	spy.method, spy.principal, spy.id = http.MethodPatch, principal, command.ID
	spy.command = applicationipam.CreatePrefixCommand{
		Prefix: command.Prefix, VRF: command.VRF, Status: command.Status,
		IsPool: command.IsPool, MarkUtilized: command.MarkUtilized,
		Description: command.Description, Comments: command.Comments,
	}
	return spy.result, spy.err
}

var _ PrefixService = (*prefixRESTPresenceSpy)(nil)
