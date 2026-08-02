package workflow

import (
	"context"
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

func TestDeviceRoleRESTListPreservesFiltersAndMapsHierarchyProjection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	child := restoredHTTPDeviceRole(t, true)
	service := &deviceRoleHTTPServiceSpy{page: applicationdcim.DeviceRolePage{
		Count: 1001, Results: []*domaindcim.DeviceRole{child},
	}}
	router := deviceRoleTestRouter(service)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodGet,
		"/api/dcim/device-roles/?limit=0&id=-1&id=2,3&name=Leaf&name=Other&slug=leaf&slug=other",
		nil,
	))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.True(t, service.query.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDeviceRolePageLimit, service.query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 2, 3}, service.query.IDs)
	assert.Equal(t, []string{"Leaf", "Other"}, service.query.Names)
	assert.Equal(t, []string{"leaf", "other"}, service.query.Slugs)
	var body struct {
		Count   uint64 `json:"count"`
		Results []struct {
			Parent      *deviceRoleReferenceDTO `json:"parent"`
			DeviceCount uint64                  `json:"device_count"`
			Depth       uint32                  `json:"_depth"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Results, 1)
	require.NotNil(t, body.Results[0].Parent)
	assert.Equal(t, int64(7), body.Results[0].Parent.ID)
	assert.Equal(t, "Root", body.Results[0].Parent.Display)
	assert.Equal(t, uint64(4), body.Results[0].DeviceCount)
	assert.Equal(t, uint32(1), body.Results[0].Depth)
}

func TestDeviceRoleRESTPatchPreservesNullParentAndFalseBoolean(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deviceRoleHTTPServiceSpy{role: restoredHTTPDeviceRole(t, false)}
	router := deviceRoleTestRouter(service)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/device-roles/8/",
		strings.NewReader(`{"parent":null,"vm_role":false}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, 1, service.updateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.update.Parent.State())
	vmRole, present := service.update.VMRole.Get()
	require.True(t, present)
	assert.False(t, vmRole)
	assert.Contains(t, response.Body.String(), `"parent":null`)
	assert.Contains(t, response.Body.String(), `"device_count":4`)
}

func TestDeviceRoleRESTRejectsUnknownFieldBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deviceRoleHTTPServiceSpy{}
	router := deviceRoleTestRouter(service)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodPost,
		"/api/dcim/device-roles/",
		strings.NewReader(`{"name":"Leaf","slug":"leaf","tags":[]}`),
	))

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Zero(t, service.createCalls)
	assert.JSONEq(t, `{"tags":["This field is not supported by the active capability profile."]}`, response.Body.String())
}

func deviceRoleTestRouter(service *deviceRoleHTTPServiceSpy) *gin.Engine {
	router := gin.New()
	NewDeviceRoleHandler(service).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-device-role"})
		c.Next()
	})
	return router
}

func restoredHTTPDeviceRole(t *testing.T, child bool) *domaindcim.DeviceRole {
	t.Helper()
	now := shared.NewTimestamp(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	state := domaindcim.DeviceRoleState{
		ID: 8, Parent: domaindcim.RootDeviceRoleParent(), Name: "Leaf", Slug: "leaf",
		Color: "9e9e9e", VMRole: false, Created: now, LastUpdated: now, DeviceCount: 4,
	}
	if child {
		state.Parent = domaindcim.NonRootDeviceRoleParent(7)
		state.ParentDisplay = "Root"
		state.Depth = 1
	}
	role, err := domaindcim.RestoreDeviceRole(state)
	require.NoError(t, err)
	return role
}

type deviceRoleHTTPServiceSpy struct {
	page        applicationdcim.DeviceRolePage
	query       applicationdcim.ListDeviceRolesQuery
	role        *domaindcim.DeviceRole
	createCalls int
	update      applicationdcim.UpdateDeviceRoleCommand
	updateCalls int
}

func (service *deviceRoleHTTPServiceSpy) ListDeviceRoles(
	_ context.Context, _ identity.Principal, query applicationdcim.ListDeviceRolesQuery,
) (applicationdcim.DeviceRolePage, error) {
	service.query = query
	return service.page, nil
}

func (*deviceRoleHTTPServiceSpy) GetDeviceRole(
	context.Context, identity.Principal, applicationdcim.GetDeviceRoleQuery,
) (*domaindcim.DeviceRole, error) {
	return nil, nil
}

func (service *deviceRoleHTTPServiceSpy) CreateDeviceRole(
	context.Context, identity.Principal, applicationdcim.CreateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.createCalls++
	return service.role, nil
}

func (*deviceRoleHTTPServiceSpy) ReplaceDeviceRole(
	context.Context, identity.Principal, applicationdcim.ReplaceDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	return nil, nil
}

func (service *deviceRoleHTTPServiceSpy) UpdateDeviceRole(
	_ context.Context, _ identity.Principal, command applicationdcim.UpdateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.updateCalls++
	service.update = command
	return service.role, nil
}

func (*deviceRoleHTTPServiceSpy) DeleteDeviceRole(
	context.Context, identity.Principal, applicationdcim.DeleteDeviceRoleCommand,
) error {
	return nil
}
