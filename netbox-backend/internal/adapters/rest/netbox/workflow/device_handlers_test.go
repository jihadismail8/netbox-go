package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRESTListMapsFiltersAndReturnsCanonicalDTO(t *testing.T) {
	spy := &deviceServiceSpy{device: restDeviceFixture(t)}
	router := deviceRESTRouter(spy)

	response := deviceRESTRequest(
		t, router, http.MethodGet,
		"/api/dcim/devices/?limit=0&offset=7&q=edge+core&id=-1&id=21"+
			"&site_id=-1&site_id=12&site_slug=primary&rack_id=13"+
			"&device_type_id=10&device_type_slug=router-9000&role_id=11"+
			"&role_slug=core-router&name=edge-01&status=active&ordering=-site,name",
		"",
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, uint32(0), spy.listQuery.Limit)
	assert.True(t, spy.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDevicePageLimit, spy.listQuery.EffectiveLimit())
	assert.Equal(t, uint32(7), spy.listQuery.Offset)
	assert.Equal(t, "edge core", spy.listQuery.Query)
	assert.Equal(t, []int64{-1, 21}, spy.listQuery.IDs)
	assert.Equal(t, []int64{-1, 12}, spy.listQuery.SiteIDs)
	assert.Equal(t, []string{"primary"}, spy.listQuery.SiteSlugs)
	assert.Equal(t, []int64{13}, spy.listQuery.RackIDs)
	assert.Equal(t, []int64{10}, spy.listQuery.DeviceTypeIDs)
	assert.Equal(t, []string{"router-9000"}, spy.listQuery.DeviceTypeSlugs)
	assert.Equal(t, []int64{11}, spy.listQuery.RoleIDs)
	assert.Equal(t, []string{"core-router"}, spy.listQuery.RoleSlugs)
	assert.Equal(t, []string{"edge-01"}, spy.listQuery.Names)
	assert.Equal(t, []string{"active"}, spy.listQuery.Statuses)
	assert.Equal(t, []string{"-site", "name"}, spy.listQuery.Ordering)

	var envelope map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &envelope))
	assert.Equal(t, float64(8), envelope["count"])
	assert.Nil(t, envelope["next"])
	previous, ok := envelope["previous"].(string)
	require.True(t, ok)
	assert.Contains(t, previous, "limit=1000")
	assert.NotContains(t, previous, "offset=")
	results, ok := envelope["results"].([]any)
	require.True(t, ok)
	require.Len(t, results, 1)
	assert.Equal(t, map[string]any{
		"id":      float64(21),
		"url":     "/api/dcim/devices/21/",
		"display": "edge-01",
		"device_type": map[string]any{
			"id": float64(10), "url": "/api/dcim/device-types/10/", "display": "Router 9000",
		},
		"role": map[string]any{
			"id": float64(11), "url": "/api/dcim/device-roles/11/", "display": "Core Router",
		},
		"name": "edge-01",
		"site": map[string]any{
			"id": float64(12), "url": "/api/dcim/sites/12/", "display": "Primary",
		},
		"rack": map[string]any{
			"id": float64(13), "url": "/api/dcim/racks/13/", "display": "Rack A",
		},
		"position":        float64(10.5),
		"face":            map[string]any{"value": "front", "label": "Front"},
		"status":          map[string]any{"value": "active", "label": "Active"},
		"serial":          "SN-1",
		"asset_tag":       nil,
		"airflow":         map[string]any{"value": "front-to-rear", "label": "Front to rear"},
		"description":     "",
		"comments":        "",
		"created":         "2026-07-25T08:00:00Z",
		"last_updated":    "2026-07-25T09:00:00Z",
		"interface_count": float64(2),
	}, results[0])
}

func TestDeviceRESTCreateAndPatchPreserveOmittedNullAndBlankSemantics(t *testing.T) {
	spy := &deviceServiceSpy{device: restDeviceFixture(t)}
	router := deviceRESTRouter(spy)

	create := deviceRESTRequest(
		t, router, http.MethodPost, "/api/dcim/devices/",
		`{"device_type":10,"role":11,"name":null,"site":12,"rack":null,`+
			`"position":10.5,"face":null,"status":"active","serial":"",`+
			`"asset_tag":null,"airflow":null}`,
	)
	require.Equal(t, http.StatusCreated, create.Code, create.Body.String())
	assert.Equal(t, "/api/dcim/devices/21/", create.Header().Get("Location"))
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.DeviceType.State())
	deviceTypeID, present := spy.createCommand.DeviceType.Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(10), deviceTypeID)
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Role.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.createCommand.Name.State())
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Site.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.createCommand.Rack.State())
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Position.State())
	position, present := spy.createCommand.Position.Get()
	require.True(t, present)
	assert.Equal(t, "10.5", position)
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Face.State())
	face, present := spy.createCommand.Face.Get()
	require.True(t, present)
	assert.Empty(t, face)
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Serial.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.createCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldPresent, spy.createCommand.Airflow.State())
	airflow, present := spy.createCommand.Airflow.Get()
	require.True(t, present)
	assert.Empty(t, airflow)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.createCommand.Description.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.createCommand.Comments.State())

	var payload map[string]any
	require.NoError(t, json.Unmarshal(create.Body.Bytes(), &payload))
	assert.Equal(t, "edge-01", payload["display"])
	assert.Nil(t, payload["asset_tag"])
	assert.Equal(t, float64(2), payload["interface_count"])

	update := deviceRESTRequest(
		t, router, http.MethodPatch, "/api/dcim/devices/21/",
		`{"rack":null,"position":null,"face":null,"name":null,`+
			`"asset_tag":"","airflow":null,"comments":""}`,
	)
	require.Equal(t, http.StatusOK, update.Code, update.Body.String())
	assert.Equal(t, shared.ID(21), spy.updateCommand.ID)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.DeviceType.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Role.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.Rack.State())
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.Position.State())
	face, present = spy.updateCommand.Face.Get()
	require.True(t, present)
	assert.Empty(t, face)
	assert.Equal(t, applicationdcim.FieldNull, spy.updateCommand.Name.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Status.State())
	assetTag, present := spy.updateCommand.AssetTag.Get()
	require.True(t, present)
	assert.Empty(t, assetTag)
	airflow, present = spy.updateCommand.Airflow.Get()
	require.True(t, present)
	assert.Empty(t, airflow)
	comments, present := spy.updateCommand.Comments.Get()
	require.True(t, present)
	assert.Empty(t, comments)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Description.State())
}

func TestDeviceRESTRejectsInvalidAdapterInputBeforeCallingTypedService(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   string
	}{
		{
			name: "unsupported write field", method: http.MethodPost,
			path: "/api/dcim/devices/",
			body: `{"device_type":10,"role":11,"site":12,"platform":9}`,
			want: `{"platform":["This field is not supported by the active capability profile."]}`,
		},
		{
			name: "non-integer relation", method: http.MethodPost,
			path: "/api/dcim/devices/",
			body: `{"device_type":"ten","role":11,"site":12}`,
			want: `{"device_type":["A valid integer is required."]}`,
		},
		{
			name: "invalid position", method: http.MethodPatch,
			path: "/api/dcim/devices/21/",
			body: `{"position":"rack-unit"}`,
			want: `{"position":["A valid number is required."]}`,
		},
		{
			name: "unsupported list filter", method: http.MethodGet,
			path: "/api/dcim/devices/?platform_id=9",
			want: `{"platform_id":["Unsupported filter."]}`,
		},
		{
			name: "invalid path id", method: http.MethodDelete,
			path: "/api/dcim/devices/0/",
			want: `{"id":["A positive integer is required."]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spy := &deviceServiceSpy{device: restDeviceFixture(t)}
			response := deviceRESTRequest(
				t, deviceRESTRouter(spy), tt.method, tt.path, tt.body,
			)
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			assert.JSONEq(t, tt.want, response.Body.String())
			assert.Zero(t, spy.totalCalls())
		})
	}
}

func TestDeviceRESTRoutesRequireAuthenticatedPrincipal(t *testing.T) {
	spy := &deviceServiceSpy{device: restDeviceFixture(t)}
	router := newTestRouter()
	NewDeviceRESTHandler(spy).Register(router)
	tests := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/dcim/devices/", ""},
		{http.MethodGet, "/api/dcim/devices/21/", ""},
		{
			http.MethodPost, "/api/dcim/devices/",
			`{"device_type":10,"role":11,"site":12}`,
		},
		{
			http.MethodPut, "/api/dcim/devices/21/",
			`{"device_type":10,"role":11,"site":12}`,
		},
		{http.MethodPatch, "/api/dcim/devices/21/", `{"status":"offline"}`},
		{http.MethodDelete, "/api/dcim/devices/21/", ""},
	}

	for _, tt := range tests {
		response := deviceRESTRequest(t, router, tt.method, tt.path, tt.body)
		require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
		assert.JSONEq(
			t,
			`{"detail":"Authentication credentials were not provided."}`,
			response.Body.String(),
		)
	}
	assert.Zero(t, spy.totalCalls())
}

func TestDeviceRoutesUseTypedServiceForAllMethodsIncludingDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &deviceServiceSpy{device: restDeviceFixture(t)}
	router := gin.New()
	newCompleteTypedHandler(
		&typedSiteCallSpy{},
		WithDeviceService(typed),
	).Register(router, authenticatedRESTMiddleware())

	requests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{http.MethodGet, "/api/dcim/devices/", "", http.StatusOK},
		{http.MethodGet, "/api/dcim/devices/21/", "", http.StatusOK},
		{
			http.MethodPost, "/api/dcim/devices/",
			`{"device_type":10,"role":11,"site":12}`,
			http.StatusCreated,
		},
		{
			http.MethodPut, "/api/dcim/devices/21/",
			`{"device_type":10,"role":11,"site":12}`,
			http.StatusOK,
		},
		{
			http.MethodPatch, "/api/dcim/devices/21/",
			`{"status":"offline"}`,
			http.StatusOK,
		},
		{http.MethodDelete, "/api/dcim/devices/21/", "", http.StatusNoContent},
	}
	for _, request := range requests {
		response := deviceRESTRequest(
			t, router, request.method, request.path, request.body,
		)
		require.Equal(t, request.status, response.Code, response.Body.String())
	}

	assert.Equal(t, 1, typed.listCalls)
	assert.Equal(t, 1, typed.getCalls)
	assert.Equal(t, 1, typed.createCalls)
	assert.Equal(t, 1, typed.replaceCalls)
	assert.Equal(t, 1, typed.updateCalls)
	assert.Equal(t, 1, typed.deleteCalls)
	assert.Equal(t, shared.ID(21), typed.deleteCommand.ID)
}

func deviceRESTRouter(service DeviceService) http.Handler {
	engine := newTestRouter()
	NewDeviceRESTHandler(service).Register(engine, authenticatedRESTMiddleware())
	return engine
}

func deviceRESTRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

type deviceServiceSpy struct {
	listCalls      int
	getCalls       int
	createCalls    int
	updateCalls    int
	replaceCalls   int
	deleteCalls    int
	listQuery      applicationdcim.ListDevicesQuery
	getQuery       applicationdcim.GetDeviceQuery
	createCommand  applicationdcim.CreateDeviceCommand
	updateCommand  applicationdcim.UpdateDeviceCommand
	replaceCommand applicationdcim.ReplaceDeviceCommand
	deleteCommand  applicationdcim.DeleteDeviceCommand
	device         *domaindcim.Device
}

func (spy *deviceServiceSpy) ListDevices(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListDevicesQuery,
) (applicationdcim.DevicePage, error) {
	spy.listCalls++
	spy.listQuery = query
	return applicationdcim.DevicePage{
		Count: 8, Results: []*domaindcim.Device{spy.device},
	}, nil
}

func (spy *deviceServiceSpy) GetDevice(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetDeviceQuery,
) (*domaindcim.Device, error) {
	spy.getCalls++
	spy.getQuery = query
	return spy.device, nil
}

func (spy *deviceServiceSpy) CreateDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateDeviceCommand,
) (*domaindcim.Device, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.device, nil
}

func (spy *deviceServiceSpy) ReplaceDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceDeviceCommand,
) (*domaindcim.Device, error) {
	spy.replaceCalls++
	spy.replaceCommand = command
	return spy.device, nil
}

func (spy *deviceServiceSpy) UpdateDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateDeviceCommand,
) (*domaindcim.Device, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.device, nil
}

func (spy *deviceServiceSpy) DeleteDevice(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteDeviceCommand,
) error {
	spy.deleteCalls++
	spy.deleteCommand = command
	return nil
}

func (spy *deviceServiceSpy) totalCalls() int {
	return spy.listCalls + spy.getCalls + spy.createCalls +
		spy.replaceCalls + spy.updateCalls + spy.deleteCalls
}

func restDeviceFixture(t *testing.T) *domaindcim.Device {
	t.Helper()
	height, err := domaindcim.ParseDeviceHeight("1.5")
	require.NoError(t, err)
	deviceType, err := domaindcim.NewDeviceTypeInstanceReference(
		10, "Router 9000", "router-9000", "Acme", height, true,
		domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
	)
	require.NoError(t, err)
	site, err := domaindcim.NewSiteReference(12, "Primary", "primary")
	require.NoError(t, err)
	rack, err := domaindcim.NewRackReference(13, "Rack A", 12, 1, 42)
	require.NoError(t, err)
	position, err := domaindcim.ParseRackPosition("10.5")
	require.NoError(t, err)
	device, err := domaindcim.RestoreDevice(domaindcim.DeviceState{
		ID: 21, DeviceType: deviceType,
		Role: domaindcim.DeviceRoleReference{ID: 11, Display: "Core Router"},
		Name: domaindcim.NonNullDeviceValue("edge-01"),
		Site: site, Rack: domaindcim.NonNullDeviceValue(rack),
		Position: domaindcim.NonNullDeviceValue(position),
		Face:     "front", Status: "active", Serial: "SN-1",
		AssetTag: domaindcim.NullDeviceValue[string](),
		Airflow:  domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
		Created: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
		),
		LastUpdated: shared.NewTimestamp(
			time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC),
		),
		InterfaceCount: 2,
	})
	require.NoError(t, err)
	return device
}
