package workflow

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestInterfaceTemplateRESTPreservesChoicesFiltersAndWritePresence(t *testing.T) {
	spy := &interfaceTemplateServiceSpy{
		template: restInterfaceTemplateFixture(t),
	}
	router := interfaceTemplateRESTRouter(spy)

	create := interfaceTemplateRESTRequest(
		t, router, http.MethodPost, "/api/dcim/interface-templates/",
		`{"device_type":7,"name":"Ethernet1","type":"10gbase-sr","enabled":false}`,
	)
	require.Equal(t, http.StatusCreated, create.Code, create.Body.String())
	var payload map[string]any
	require.NoError(t, json.Unmarshal(create.Body.Bytes(), &payload))
	assert.Equal(t, "Ethernet1 (WAN)", payload["display"])
	assert.Equal(t, map[string]any{
		"value": "10gbase-sr", "label": "10GBASE-SR (10GE)",
	}, payload["type"])
	assert.Equal(t, map[string]any{
		"id": float64(7), "url": "/api/dcim/device-types/7/", "display": "Router",
	}, payload["device_type"])
	require.Equal(t, 1, spy.createCalls)
	deviceTypeID, present := spy.createCommand.DeviceType.Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(7), deviceTypeID)
	enabled, present := spy.createCommand.Enabled.Get()
	require.True(t, present)
	assert.False(t, enabled)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.createCommand.MgmtOnly.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.createCommand.Description.State())

	update := interfaceTemplateRESTRequest(
		t, router, http.MethodPatch, "/api/dcim/interface-templates/41/",
		`{"label":"","mgmt_only":false}`,
	)
	require.Equal(t, http.StatusOK, update.Code, update.Body.String())
	require.Equal(t, 1, spy.updateCalls)
	assert.Equal(t, applicationdcim.FieldPresent, spy.updateCommand.Label.State())
	mgmtOnly, present := spy.updateCommand.MgmtOnly.Get()
	require.True(t, present)
	assert.False(t, mgmtOnly)
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.Enabled.State())
	assert.Equal(t, applicationdcim.FieldOmitted, spy.updateCommand.DeviceType.State())

	replace := interfaceTemplateRESTRequest(
		t, router, http.MethodPut, "/api/dcim/interface-templates/41/",
		`{"device_type":7,"name":"Ethernet2","type":"other"}`,
	)
	require.Equal(t, http.StatusOK, replace.Code, replace.Body.String())
	require.Equal(t, 1, spy.replaceCalls)
	assert.Equal(
		t, applicationdcim.FieldOmitted,
		spy.replaceCommand.Enabled.State(),
	)

	list := interfaceTemplateRESTRequest(
		t, router, http.MethodGet,
		"/api/dcim/interface-templates/?device_type_id=-1&device_type_id=7"+
			"&name=Ethernet1&type=10gbase-sr&enabled=false&mgmt_only=true"+
			"&ordering=-type,name",
		"",
	)
	require.Equal(t, http.StatusOK, list.Code, list.Body.String())
	assert.Equal(t, []int64{-1, 7}, spy.listQuery.DeviceTypeIDs)
	assert.Equal(t, []string{"Ethernet1"}, spy.listQuery.Names)
	assert.Equal(t, []string{"10gbase-sr"}, spy.listQuery.Types)
	require.NotNil(t, spy.listQuery.Enabled)
	assert.False(t, *spy.listQuery.Enabled)
	require.NotNil(t, spy.listQuery.MgmtOnly)
	assert.True(t, *spy.listQuery.MgmtOnly)
	assert.Equal(t, []string{"-type", "name"}, spy.listQuery.Ordering)

	unknown := interfaceTemplateRESTRequest(
		t, router, http.MethodPost, "/api/dcim/interface-templates/",
		`{"device_type":7,"name":"Ethernet1","type":"other","bridge":9}`,
	)
	require.Equal(t, http.StatusBadRequest, unknown.Code, unknown.Body.String())
	assert.Equal(t, 1, spy.createCalls)
}

func interfaceTemplateRESTRouter(service InterfaceTemplateService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewInterfaceTemplateRESTHandler(service).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{
			ID: 1, Username: "interface-template-rest", IsSuperuser: true,
		})
		c.Next()
	})
	return router
}

func interfaceTemplateRESTRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func restInterfaceTemplateFixture(t *testing.T) *domaindcim.InterfaceTemplate {
	t.Helper()
	reference, err := domaindcim.NewDeviceTypeReference(7, "Router", "router")
	require.NoError(t, err)
	template, err := domaindcim.RestoreInterfaceTemplate(
		domaindcim.InterfaceTemplateState{
			ID: 41, DeviceType: reference, Name: "Ethernet1", Label: "WAN",
			Type: "10gbase-sr", Enabled: false, MgmtOnly: true,
			Description: "Template description",
			Created: shared.NewTimestamp(
				time.Date(2026, time.July, 25, 8, 0, 0, 0, time.UTC),
			),
			LastUpdated: shared.NewTimestamp(
				time.Date(2026, time.July, 25, 9, 0, 0, 0, time.UTC),
			),
		},
	)
	require.NoError(t, err)
	return template
}
