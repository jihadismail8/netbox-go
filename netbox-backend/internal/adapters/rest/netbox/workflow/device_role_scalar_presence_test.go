package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRoleRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &deviceRoleScalarPresenceService{role: restoredHTTPDeviceRole(t, false)}
	router := gin.New()
	NewDeviceRoleHandler(service).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "device-role-presence"})
		c.Next()
	})

	request := func(method, path, body string, status int) *httptest.ResponseRecorder {
		t.Helper()
		response := httptest.NewRecorder()
		httpRequest := httptest.NewRequest(method, path, strings.NewReader(body))
		httpRequest.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, httpRequest)
		require.Equal(t, status, response.Code, response.Body.String())
		return response
	}

	request(http.MethodPost, "/api/dcim/device-roles/", `{"name":"Root","slug":"root"}`, http.StatusCreated)
	require.Equal(t, 1, service.createCalls)
	assert.Equal(t, applicationdcim.FieldOmitted, service.create.Parent.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.create.Color.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.create.VMRole.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.create.Description.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.create.Comments.State())

	request(
		http.MethodPost,
		"/api/dcim/device-roles/",
		`{"parent":7,"name":" Child ","slug":" child ","color":" 00ff00 ","vm_role":false,"description":"","comments":""}`,
		http.StatusCreated,
	)
	parent, present := service.create.Parent.Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(7), parent)
	vmRole, present := service.create.VMRole.Get()
	require.True(t, present)
	assert.False(t, vmRole)
	description, present := service.create.Description.Get()
	require.True(t, present)
	assert.Empty(t, description)
	comments, present := service.create.Comments.Get()
	require.True(t, present)
	assert.Empty(t, comments)

	request(http.MethodPut, "/api/dcim/device-roles/8/", `{"name":"Replacement","slug":"replacement"}`, http.StatusOK)
	require.Equal(t, 1, service.replaceCalls)
	assert.Equal(t, applicationdcim.FieldOmitted, service.replace.Parent.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.replace.Color.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.replace.VMRole.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.replace.Description.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.replace.Comments.State())

	request(http.MethodPatch, "/api/dcim/device-roles/8/", `{"parent":null,"vm_role":false,"description":"","comments":""}`, http.StatusOK)
	require.Equal(t, 1, service.updateCalls)
	assert.Equal(t, applicationdcim.FieldNull, service.update.Parent.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.update.Name.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.update.Slug.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.update.Color.State())
	vmRole, present = service.update.VMRole.Get()
	require.True(t, present)
	assert.False(t, vmRole)

	beforeCalls := service.totalMutations()
	response := request(http.MethodPatch, "/api/dcim/device-roles/8/", `{"vm_role":null}`, http.StatusBadRequest)
	assert.JSONEq(t, `{"vm_role":["This field may not be null."]}`, response.Body.String())
	assert.Equal(t, beforeCalls+1, service.totalMutations())
	assert.Equal(t, applicationdcim.FieldNull, service.update.VMRole.State())

	beforeCalls = service.totalMutations()
	request(http.MethodPatch, "/api/dcim/device-roles/8/", `{"parent":{"id":7}}`, http.StatusBadRequest)
	assert.Equal(t, beforeCalls, service.totalMutations(), "alternate nested parent forms must fail before the service")
}

type deviceRoleScalarPresenceService struct {
	role         *domaindcim.DeviceRole
	create       applicationdcim.CreateDeviceRoleCommand
	replace      applicationdcim.ReplaceDeviceRoleCommand
	update       applicationdcim.UpdateDeviceRoleCommand
	createCalls  int
	replaceCalls int
	updateCalls  int
}

func (service *deviceRoleScalarPresenceService) totalMutations() int {
	return service.createCalls + service.replaceCalls + service.updateCalls
}

func (*deviceRoleScalarPresenceService) ListDeviceRoles(
	context.Context,
	identity.Principal,
	applicationdcim.ListDeviceRolesQuery,
) (applicationdcim.DeviceRolePage, error) {
	return applicationdcim.DeviceRolePage{}, nil
}

func (service *deviceRoleScalarPresenceService) GetDeviceRole(
	context.Context,
	identity.Principal,
	applicationdcim.GetDeviceRoleQuery,
) (*domaindcim.DeviceRole, error) {
	return service.role, nil
}

func (service *deviceRoleScalarPresenceService) CreateDeviceRole(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.createCalls++
	service.create = command
	return service.role, nil
}

func (service *deviceRoleScalarPresenceService) ReplaceDeviceRole(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.replaceCalls++
	service.replace = command
	return service.role, nil
}

func (service *deviceRoleScalarPresenceService) UpdateDeviceRole(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateDeviceRoleCommand,
) (*domaindcim.DeviceRole, error) {
	service.updateCalls++
	service.update = command
	if command.VMRole.State() == applicationdcim.FieldNull {
		return nil, shared.NewValidationError(shared.FieldViolation{
			Field: "vm_role", Reason: "null", Description: "This field may not be null.",
		})
	}
	return service.role, nil
}

func (*deviceRoleScalarPresenceService) DeleteDeviceRole(
	context.Context,
	identity.Principal,
	applicationdcim.DeleteDeviceRoleCommand,
) error {
	return nil
}
