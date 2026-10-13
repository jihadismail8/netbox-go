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

// This spy proves transport decoding/projection, not application validation or persistence.
func TestVRFRESTScalarPresenceMatrix(t *testing.T) {
	for _, operation := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/api/ipam/vrfs/", http.StatusCreated},
		{http.MethodPut, "/api/ipam/vrfs/41/", http.StatusOK},
		{http.MethodPatch, "/api/ipam/vrfs/41/", http.StatusOK},
	} {
		t.Run(operation.method, func(t *testing.T) {
			request := func(spy *vrfRESTPresenceSpy, body string) *httptest.ResponseRecorder {
				router := newTestRouter()
				NewVRFRESTHandler(spy).Register(router, testVRFPrincipalMiddleware())
				response := httptest.NewRecorder()
				input := httptest.NewRequest(operation.method, operation.path, strings.NewReader(body))
				input.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(response, input)
				return response
			}
			fixture := restVRFFixture(t, 41, "Projection", "65000:41", 3, 4)
			for _, field := range []string{"name", "rd", "description", "comments", "enforce_unique"} {
				for _, presence := range []string{"omitted", "null", "blank-or-false", "concrete"} {
					t.Run("decoded intent/"+field+"/"+presence, func(t *testing.T) {
						body := "{}"
						want := applicationipam.CreateVRFCommand{}
						stringField := applicationipam.OmittedField[string]()
						boolField := applicationipam.OmittedField[bool]()
						var rawValue any
						switch presence {
						case "null":
							stringField, boolField = applicationipam.NullField[string](), applicationipam.NullField[bool]()
						case "blank-or-false":
							rawValue = ""
							if field == "enforce_unique" {
								rawValue = false
							}
							stringField, boolField = applicationipam.FieldValue(""), applicationipam.FieldValue(false)
						case "concrete":
							// Trimming and NUL validation belong to the shared core, not this decoder.
							text := "  raw " + field + "\x00  "
							rawValue = text
							if field == "enforce_unique" {
								rawValue = true
							}
							stringField, boolField = applicationipam.FieldValue(text), applicationipam.FieldValue(true)
						}
						if presence != "omitted" {
							encoded, err := json.Marshal(map[string]any{field: rawValue})
							require.NoError(t, err)
							body = string(encoded)
						}
						switch field {
						case "name":
							want.Name = stringField
						case "rd":
							want.RD = stringField
						case "description":
							want.Description = stringField
						case "comments":
							want.Comments = stringField
						case "enforce_unique":
							want.EnforceUnique = boolField
						}
						spy := &vrfRESTPresenceSpy{result: fixture}
						response := request(spy, body)
						require.Equal(t, operation.status, response.Code, response.Body.String())
						require.Equal(t, 1, spy.calls)
						require.Equal(t, operation.method, spy.method)
						require.Equal(t, identity.Principal{ID: 1, Username: "vrf-transport"}, spy.principal)
						require.Equal(t, want, spy.command, "unselected fields must remain omitted")
						if operation.method == http.MethodPost {
							require.Zero(t, spy.id)
						} else {
							require.Equal(t, shared.ID(41), spy.id)
						}
					})
				}
				wrongTypes := []string{`0`, `false`, `[]`, `{}`}
				description := "Must be a string."
				if field == "enforce_unique" {
					wrongTypes = []string{`0`, `"false"`, `""`, `[]`, `{}`}
					description = "Must be a valid boolean."
				}
				for _, raw := range wrongTypes {
					t.Run("wrong scalar type/"+field+"/"+raw, func(t *testing.T) {
						spy := &vrfRESTPresenceSpy{result: fixture}
						response := request(spy, `{"`+field+`":`+raw+`}`)
						require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
						require.Zero(t, spy.calls, "malformed scalars must fail before the core")
						expected, err := json.Marshal(map[string][]string{field: {description}})
						require.NoError(t, err)
						require.JSONEq(t, string(expected), response.Body.String())
					})
				}
			}

			for _, field := range []string{"tenant", "tags", "custom_fields", "id", "url", "display", "created", "last_updated", "ipaddress_count", "prefix_count"} {
				t.Run("undeclared field/"+field, func(t *testing.T) {
					spy := &vrfRESTPresenceSpy{result: fixture}
					response := request(spy, `{"name":"valid shape","`+field+`":null}`)
					require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
					require.Zero(t, spy.calls)
					expected, err := json.Marshal(map[string][]string{field: {"This field is not supported by the active capability profile."}})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), response.Body.String())
				})
			}
			for _, envelope := range []struct {
				name, body, description string
			}{
				{"empty", "", "Expected a JSON object."},
				{"null", `null`, "Expected a JSON object."},
				{"array", `[]`, "Expected a JSON object."},
				{"string", `"VRF"`, "Expected a JSON object."},
				{"malformed", `{"name":}`, "Expected a JSON object."},
				{"second object", `{} {}`, "Request body must contain one JSON object."},
				{"trailing null", `{} null`, "Request body must contain one JSON object."},
				{"trailing malformed data", `{} junk`, "Request body must contain one JSON object."},
			} {
				t.Run("strict envelope/"+envelope.name, func(t *testing.T) {
					spy := &vrfRESTPresenceSpy{result: fixture}
					response := request(spy, envelope.body)
					require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
					require.Zero(t, spy.calls)
					expected, err := json.Marshal(map[string][]string{"non_field_errors": {envelope.description}})
					require.NoError(t, err)
					require.JSONEq(t, string(expected), response.Body.String())
				})
			}

			for _, rd := range []string{"null", "blank", "concrete"} {
				t.Run("response projection/"+rd, func(t *testing.T) {
					state := fixture.State()
					state.EnforceUnique = false
					state.Description, state.Comments = "stored description", "stored comments"
					state.LastUpdated = shared.NewTimestamp(state.Created.Add(time.Minute))
					var expectedRD any
					if rd == "null" {
						state.RD = domainipam.NullRouteDistinguisher()
					} else {
						value := "65000:41"
						if rd == "blank" {
							value = ""
						}
						parsed, err := domainipam.ParseRouteDistinguisher(value)
						require.NoError(t, err)
						state.RD, expectedRD = domainipam.NonNullRouteDistinguisher(parsed), value
					}
					result, err := domainipam.RestoreVRF(state)
					require.NoError(t, err)
					spy := &vrfRESTPresenceSpy{result: result}
					response := request(spy, `{"name":"transport input"}`)
					require.Equal(t, operation.status, response.Code, response.Body.String())
					var payload map[string]any
					require.NoError(t, json.Unmarshal(response.Body.Bytes(), &payload))
					want := map[string]any{
						"id": float64(41), "url": "/api/ipam/vrfs/41/", "display": result.Display(),
						"name": state.Name, "rd": expectedRD, "enforce_unique": false,
						"description": state.Description, "comments": state.Comments,
						"created": state.Created.UTC().Format(time.RFC3339Nano), "last_updated": state.LastUpdated.UTC().Format(time.RFC3339Nano),
					}
					if operation.method == http.MethodPost {
						require.Equal(t, "/api/ipam/vrfs/41/", response.Header().Get("Location"))
					} else {
						want["ipaddress_count"], want["prefix_count"] = float64(3), float64(4)
						require.Empty(t, response.Header().Get("Location"))
					}
					// Exact keys catch both response-field leakage and lost null/blank/false values.
					require.Equal(t, want, payload)
				})
			}

			t.Run("canonical shared field-error presentation", func(t *testing.T) {
				spy := &vrfRESTPresenceSpy{result: fixture, err: shared.NewValidationError(
					shared.FieldViolation{Field: "name", Reason: "blank", Description: "This field may not be blank."},
					shared.FieldViolation{Field: "rd", Reason: "max_length", Description: "Ensure this field has no more than 21 characters."},
					shared.FieldViolation{Field: "rd", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
					shared.FieldViolation{Field: "enforce_unique", Reason: "null", Description: "This field may not be null."},
					shared.FieldViolation{Field: "description", Reason: "null", Description: "This field may not be null."},
					shared.FieldViolation{Field: "comments", Reason: "null_characters_not_allowed", Description: "Null characters are not allowed."},
				)}
				response := request(spy, `{"name":" ","rd":"\u0000","enforce_unique":null,"description":null,"comments":"\u0000"}`)
				require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
				require.Equal(t, 1, spy.calls, "the transport must present the injected shared error without owning validation")
				require.JSONEq(t, `{"name":["This field may not be blank."],"rd":["Ensure this field has no more than 21 characters.","Null characters are not allowed."],"enforce_unique":["This field may not be null."],"description":["This field may not be null."],"comments":["Null characters are not allowed."]}`, response.Body.String())
			})
			t.Run("untyped infrastructure error is redacted", func(t *testing.T) {
				spy := &vrfRESTPresenceSpy{result: fixture, err: errors.New("private storage failure")}
				response := request(spy, `{"name":"valid shape"}`)
				require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
				require.Equal(t, 1, spy.calls)
				require.JSONEq(t, `{"detail":"An internal error occurred."}`, response.Body.String())
			})
		})
	}
}

type vrfRESTPresenceSpy struct {
	VRFService
	calls     int
	method    string
	principal identity.Principal
	id        shared.ID
	command   applicationipam.CreateVRFCommand
	result    *domainipam.VRF
	err       error
}

func (spy *vrfRESTPresenceSpy) CreateVRF(_ context.Context, principal identity.Principal, command applicationipam.CreateVRFCommand) (*domainipam.VRF, error) {
	spy.calls++
	spy.method, spy.principal, spy.command = http.MethodPost, principal, command
	return spy.result, spy.err
}

func (spy *vrfRESTPresenceSpy) ReplaceVRF(_ context.Context, principal identity.Principal, command applicationipam.ReplaceVRFCommand) (*domainipam.VRF, error) {
	spy.calls++
	spy.method, spy.principal, spy.id = http.MethodPut, principal, command.ID
	spy.command = applicationipam.CreateVRFCommand{
		Name: command.Name, RD: command.RD, EnforceUnique: command.EnforceUnique,
		Description: command.Description, Comments: command.Comments,
	}
	return spy.result, spy.err
}

func (spy *vrfRESTPresenceSpy) UpdateVRF(_ context.Context, principal identity.Principal, command applicationipam.UpdateVRFCommand) (*domainipam.VRF, error) {
	spy.calls++
	spy.method, spy.principal, spy.id = http.MethodPatch, principal, command.ID
	spy.command = applicationipam.CreateVRFCommand{
		Name: command.Name, RD: command.RD, EnforceUnique: command.EnforceUnique,
		Description: command.Description, Comments: command.Comments,
	}
	return spy.result, spy.err
}

var _ VRFService = (*vrfRESTPresenceSpy)(nil)
