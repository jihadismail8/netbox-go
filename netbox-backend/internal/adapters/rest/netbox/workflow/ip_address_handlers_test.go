package workflow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationipam "netbox-go/internal/application/ipam"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestParseIPAddressListPreservesRepeatedAndAssignmentFilters(t *testing.T) {
	query, err := parseIPAddressList(url.Values{
		"limit": {"0"}, "id": {"-1,0", "11"}, "vrf_id": {"-2", "7"},
		"address": {"192.0.2.1", "2001:db8::1/64"}, "family": {"-4"},
		"parent": {"192.0.2.99/24"}, "status": {"active", "reserved"},
		"assigned": {"true"}, "interface_id": {"-3,9"},
		"device_id": {"-4", "8"}, "ordering": {"vrf,-address"},
	})
	require.NoError(t, err)
	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumIPAddressPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 0, 11}, query.IDs)
	assert.Equal(t, []int64{-2, 7}, query.VRFIDs)
	assert.Equal(t, []string{"192.0.2.1", "2001:db8::1/64"}, query.Addresses)
	assert.Equal(t, int64(-4), *query.Family)
	assert.Equal(t, "192.0.2.99/24", *query.Parent)
	assert.True(t, *query.Assigned)
	assert.Equal(t, []int64{-3, 9}, query.InterfaceIDs)
	assert.Equal(t, []int64{-4, 8}, query.DeviceIDs)
	assert.Equal(t, []string{"vrf", "-address"}, query.Ordering)
}

func TestIPAddressRESTPatchPreservesAssignmentNullPresence(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restIPAddressServiceSpy{address: restIPAddressFixture(t, true)}
	router := gin.New()
	NewIPAddressRESTHandler(spy).Register(
		router, testPrefixPrincipalMiddleware(),
	)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/ipam/ip-addresses/17/",
		strings.NewReader(
			`{"dns_name":"EDGE.EXAMPLE","assigned_object_type":null,"assigned_object_id":null}`,
		),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, applicationipam.FieldPresent, spy.updateCommand.DNSName.State())
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.AssignedObjectType.State())
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.AssignedObjectID.State())
	assert.Equal(t, applicationipam.FieldOmitted, spy.updateCommand.Address.State())

	var projected ipAddressDTO
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &projected))
	require.NotNil(t, projected.AssignedObject)
	assert.Equal(t, int64(11), projected.AssignedObject.ID)
	assert.Equal(t, "Ethernet1 (uplink)", projected.AssignedObject.Display)
	assert.Equal(t, "dcim.interface", *projected.AssignedObjectType)
	assert.Equal(t, int64(11), *projected.AssignedObjectID)
}

func TestIPAddressRESTScalarPresenceMatrix(t *testing.T) {
	type scalarCase struct {
		name   string
		body   string
		state  applicationipam.FieldState
		values map[string]string
	}
	tests := []scalarCase{
		{
			name:  "omitted",
			body:  `{}`,
			state: applicationipam.FieldOmitted,
		},
		{
			name: "explicit null",
			body: `{"address":null,"status":null,"role":null,` +
				`"dns_name":null,"description":null,"comments":null}`,
			state: applicationipam.FieldNull,
		},
		{
			name: "blank",
			body: `{"address":"","status":"","role":"",` +
				`"dns_name":"","description":"","comments":""}`,
			state: applicationipam.FieldPresent,
			values: map[string]string{
				"address": "", "status": "", "role": "", "dns_name": "",
				"description": "", "comments": "",
			},
		},
		{
			name: "concrete raw values",
			body: `{"address":"192.0.2.17","status":" reserved ",` +
				`"role":" loopback ","dns_name":" EDGE.EXAMPLE ",` +
				`"description":" description ","comments":" comments "}`,
			state: applicationipam.FieldPresent,
			values: map[string]string{
				"address": "192.0.2.17", "status": " reserved ",
				"role": " loopback ", "dns_name": " EDGE.EXAMPLE ",
				"description": " description ", "comments": " comments ",
			},
		},
	}
	type operation struct {
		name       string
		method     string
		path       string
		wantStatus int
		fields     func(*restIPAddressServiceSpy) map[string]applicationipam.Field[string]
	}
	operations := []operation{
		{
			name: "POST", method: http.MethodPost,
			path: "/api/ipam/ip-addresses/", wantStatus: http.StatusCreated,
			fields: func(spy *restIPAddressServiceSpy) map[string]applicationipam.Field[string] {
				return ipAddressCreateScalarFields(spy.createCommand)
			},
		},
		{
			name: "PUT", method: http.MethodPut,
			path: "/api/ipam/ip-addresses/17/", wantStatus: http.StatusOK,
			fields: func(spy *restIPAddressServiceSpy) map[string]applicationipam.Field[string] {
				return ipAddressCreateScalarFields(spy.replaceCommand.CreateIPAddressCommand)
			},
		},
		{
			name: "PATCH", method: http.MethodPatch,
			path: "/api/ipam/ip-addresses/17/", wantStatus: http.StatusOK,
			fields: func(spy *restIPAddressServiceSpy) map[string]applicationipam.Field[string] {
				return ipAddressUpdateScalarFields(spy.updateCommand)
			},
		},
	}

	for _, operation := range operations {
		for _, test := range tests {
			t.Run(operation.name+"/"+test.name, func(t *testing.T) {
				gin.SetMode(gin.TestMode)
				spy := &restIPAddressServiceSpy{address: restIPAddressFixture(t, false)}
				router := gin.New()
				NewIPAddressRESTHandler(spy).Register(
					router, testPrefixPrincipalMiddleware(),
				)

				response := httptest.NewRecorder()
				request := httptest.NewRequest(
					operation.method, operation.path, strings.NewReader(test.body),
				)
				request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(response, request)

				require.Equal(t, operation.wantStatus, response.Code, response.Body.String())
				for name, field := range operation.fields(spy) {
					assert.Equal(t, test.state, field.State(), name)
					if test.state != applicationipam.FieldPresent {
						continue
					}
					value, present := field.Get()
					require.True(t, present, name)
					assert.Equal(t, test.values[name], value, name)
				}
			})
		}
	}

	serveCreate := func(t *testing.T, body string) (*restIPAddressServiceSpy, *httptest.ResponseRecorder) {
		t.Helper()
		gin.SetMode(gin.TestMode)
		spy := &restIPAddressServiceSpy{address: restIPAddressFixture(t, false)}
		router := gin.New()
		NewIPAddressRESTHandler(spy).Register(router, testPrefixPrincipalMiddleware())
		response := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost, "/api/ipam/ip-addresses/", strings.NewReader(body),
		)
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, request)
		return spy, response
	}

	for _, test := range []struct {
		name  string
		body  string
		field string
		want  string
	}{
		{name: "status JSON boolean", body: `{"address":"192.0.2.17","status":true}`, field: "status", want: "true"},
		{name: "role JSON integer", body: `{"address":"192.0.2.17","role":1}`, field: "role", want: "1"},
		{name: "integer negative zero", body: `{"address":"192.0.2.17","role":-0}`, field: "role", want: "0"},
		{name: "DNS JSON integer", body: `{"address":"192.0.2.17","dns_name":123}`, field: "dns_name", want: "123"},
		{name: "description JSON float", body: `{"address":"192.0.2.17","description":1.5}`, field: "description", want: "1.5"},
		{name: "comments JSON exponent", body: `{"address":"192.0.2.17","comments":1e3}`, field: "comments", want: "1000.0"},
		{name: "fixed exponent threshold", body: `{"address":"192.0.2.17","description":1e15}`, field: "description", want: "1000000000000000.0"},
		{name: "scientific exponent threshold", body: `{"address":"192.0.2.17","description":1e16}`, field: "description", want: "1e+16"},
		{name: "fixed negative exponent threshold", body: `{"address":"192.0.2.17","comments":1e-4}`, field: "comments", want: "0.0001"},
		{name: "scientific negative exponent threshold", body: `{"address":"192.0.2.17","comments":1e-5}`, field: "comments", want: "1e-05"},
		{name: "overflow exponent", body: `{"address":"192.0.2.17","comments":1e400}`, field: "comments", want: "inf"},
		{name: "negative overflow exponent", body: `{"address":"192.0.2.17","comments":-1e400}`, field: "comments", want: "-inf"},
		{name: "underflow exponent", body: `{"address":"192.0.2.17","comments":1e-4000}`, field: "comments", want: "0.0"},
		{name: "negative underflow exponent", body: `{"address":"192.0.2.17","comments":-1e-4000}`, field: "comments", want: "-0.0"},
		{name: "negative zero", body: `{"address":"192.0.2.17","comments":-0.0}`, field: "comments", want: "-0.0"},
		{name: "description null character", body: `{"address":"192.0.2.17","description":"contains\u0000null"}`, field: "description", want: "contains\x00null"},
		{name: "paired surrogate", body: `{"address":"192.0.2.17","description":"\uD83D\uDE00"}`, field: "description", want: "😀"},
		{name: "escaped surrogate text", body: `{"address":"192.0.2.17","description":"\\uD800"}`, field: "description", want: `\uD800`},
	} {
		t.Run("decoder/"+test.name, func(t *testing.T) {
			spy, response := serveCreate(t, test.body)
			require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
			field := ipAddressCreateScalarFields(spy.createCommand)[test.field]
			value, present := field.Get()
			require.True(t, present)
			assert.Equal(t, test.want, value)
		})
	}

	for _, test := range []struct {
		name  string
		body  string
		field string
		want  string
	}{
		{
			name: "status dictionary", body: `{"address":"192.0.2.17","status":{"value":"active"}}`,
			field: "status", want: `Value must be passed directly (e.g. "foo": 123); do not use a dictionary or list.`,
		},
		{
			name: "role list", body: `{"address":"192.0.2.17","role":["loopback"]}`,
			field: "role", want: `Value must be passed directly (e.g. "foo": 123); do not use a dictionary or list.`,
		},
		{name: "address integer", body: `{"address":1}`, field: "address", want: "unexpected type <class 'int'> for addr arg"},
		{name: "DNS boolean", body: `{"address":"192.0.2.17","dns_name":true}`, field: "dns_name", want: "Not a valid string."},
		{name: "description dictionary", body: `{"address":"192.0.2.17","description":{}}`, field: "description", want: "Not a valid string."},
		{name: "nested surrogate dictionary", body: `{"address":"192.0.2.17","description":{"nested":"\uD800"}}`, field: "description", want: "Not a valid string."},
		{name: "high surrogate", body: `{"address":"192.0.2.17","description":"\uD800"}`, field: "description", want: "Surrogate characters are not allowed: U+D800."},
		{name: "low surrogate", body: `{"address":"192.0.2.17","comments":"\uDC00"}`, field: "comments", want: "Surrogate characters are not allowed: U+DC00."},
	} {
		t.Run("decoder rejects/"+test.name, func(t *testing.T) {
			_, response := serveCreate(t, test.body)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			var envelope map[string][]string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
			assert.Equal(t, []string{test.want}, envelope[test.field])
		})
	}

	t.Run("decoder rejects/DNS surrogate accumulates validators", func(t *testing.T) {
		_, response := serveCreate(
			t, `{"address":"192.0.2.17","dns_name":"\uD800"}`,
		)
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		var envelope map[string][]string
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		assert.Equal(t, []string{
			"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
			"Surrogate characters are not allowed: U+D800.",
		}, envelope["dns_name"])
	})

	for _, test := range []struct {
		name  string
		body  string
		field string
		want  []string
	}{
		{
			name:  "description NUL and surrogate",
			body:  `{"address":"192.0.2.17","description":"bad\u0000\uD800"}`,
			field: "description",
			want: []string{
				"Null characters are not allowed.",
				"Surrogate characters are not allowed: U+D800.",
			},
		},
		{
			name:  "DNS regex NUL and surrogate",
			body:  `{"address":"192.0.2.17","dns_name":"bad\u0000\uD800"}`,
			field: "dns_name",
			want: []string{
				"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
				"Null characters are not allowed.",
				"Surrogate characters are not allowed: U+D800.",
			},
		},
		{
			name:  "description max length and surrogate",
			body:  `{"address":"192.0.2.17","description":"` + strings.Repeat("a", 200) + `\uD800"}`,
			field: "description",
			want: []string{
				"Ensure this field has no more than 200 characters.",
				"Surrogate characters are not allowed: U+D800.",
			},
		},
		{
			name:  "DNS regex max length and surrogate",
			body:  `{"address":"192.0.2.17","dns_name":"` + strings.Repeat("a", 255) + `\uD800"}`,
			field: "dns_name",
			want: []string{
				"Only alphanumeric characters, asterisks, hyphens, periods, and underscores are allowed in DNS names",
				"Ensure this field has no more than 255 characters.",
				"Surrogate characters are not allowed: U+D800.",
			},
		},
	} {
		t.Run("decoder rejects/"+test.name, func(t *testing.T) {
			_, response := serveCreate(t, test.body)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			var envelope map[string][]string
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
			assert.Equal(t, test.want, envelope[test.field])
		})
	}

	t.Run("decoder rejects/malformed UTF-8 before service", func(t *testing.T) {
		body := `{"address":"192.0.2.17","description":"` +
			string([]byte{0xED, 0xA0, 0x80}) + `"}`
		spy, response := serveCreate(t, body)
		require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		assert.Zero(t, spy.createCalls)
		var envelope map[string][]string
		require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
		assert.Equal(t, []string{"Expected a JSON object."}, envelope["non_field_errors"])
	})
}

func ipAddressCreateScalarFields(
	command applicationipam.CreateIPAddressCommand,
) map[string]applicationipam.Field[string] {
	return map[string]applicationipam.Field[string]{
		"address": command.Address, "status": command.Status,
		"role": command.Role, "dns_name": command.DNSName,
		"description": command.Description, "comments": command.Comments,
	}
}

func ipAddressUpdateScalarFields(
	command applicationipam.UpdateIPAddressCommand,
) map[string]applicationipam.Field[string] {
	return map[string]applicationipam.Field[string]{
		"address": command.Address, "status": command.Status,
		"role": command.Role, "dns_name": command.DNSName,
		"description": command.Description, "comments": command.Comments,
	}
}

func restIPAddressFixture(
	t *testing.T,
	assigned bool,
) *domainipam.IPAddress {
	t.Helper()
	stamp := shared.NewTimestamp(
		time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
	)
	assignment := domainipam.NullInterfaceAssignment()
	if assigned {
		device, err := domaindcim.NewDeviceReference(
			7, domaindcim.NonNullDeviceValue("edge-01"), "edge-01",
		)
		require.NoError(t, err)
		networkInterface, err := domaindcim.RestoreInterface(
			domaindcim.InterfaceState{
				ID: 11, Device: device, Name: "Ethernet1", Label: "uplink",
				Type: "1000base-t", Enabled: true, Created: stamp,
				LastUpdated: stamp,
			},
		)
		require.NoError(t, err)
		value, err := domainipam.NewInterfaceAssignment(networkInterface)
		require.NoError(t, err)
		assignment = domainipam.NonNullInterfaceAssignment(value)
	}
	address, err := domainipam.RestoreIPAddress(domainipam.IPAddressState{
		ID: 17, Address: "192.0.2.1/24",
		VRF:        domainipam.NullVRFReference(),
		Status:     domainipam.IPAddressStatusActive.String(),
		Assignment: assignment, Created: stamp, LastUpdated: stamp,
	})
	require.NoError(t, err)
	return address
}

type restIPAddressServiceSpy struct {
	createCalls    int
	createCommand  applicationipam.CreateIPAddressCommand
	replaceCommand applicationipam.ReplaceIPAddressCommand
	updateCommand  applicationipam.UpdateIPAddressCommand
	address        *domainipam.IPAddress
}

func (*restIPAddressServiceSpy) ListIPAddresses(
	context.Context,
	identity.Principal,
	applicationipam.ListIPAddressesQuery,
) (applicationipam.IPAddressPage, error) {
	return applicationipam.IPAddressPage{}, nil
}

func (spy *restIPAddressServiceSpy) GetIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.GetIPAddressQuery,
) (*domainipam.IPAddress, error) {
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) CreateIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.CreateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) ReplaceIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.ReplaceIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.replaceCommand = command
	return spy.address, nil
}

func (spy *restIPAddressServiceSpy) UpdateIPAddress(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UpdateIPAddressCommand,
) (*domainipam.IPAddress, error) {
	spy.updateCommand = command
	return spy.address, nil
}

func (*restIPAddressServiceSpy) DeleteIPAddress(
	context.Context,
	identity.Principal,
	applicationipam.DeleteIPAddressCommand,
) error {
	return nil
}
