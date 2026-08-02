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

func TestOrganizationRESTListPreservesRepeatedFiltersSignedIDsAndExplicitZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &organizationHTTPServiceSpy{manufacturerPage: applicationdcim.ManufacturerPage{Count: 1001}}
	router := organizationTestRouter(service)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodGet,
		"/api/dcim/manufacturers/?limit=0&id=-1&id=2,3&name=Alpha&name=Beta&slug=alpha&slug=beta&ordering=-name,id",
		nil,
	))

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, service.manufacturerListCalls)
	query := service.manufacturerQuery
	assert.True(t, query.LimitPresent)
	assert.Zero(t, query.Limit)
	assert.Equal(t, applicationdcim.MaximumManufacturerPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-1, 2, 3}, query.IDs)
	assert.Equal(t, []string{"Alpha", "Beta"}, query.Names)
	assert.Equal(t, []string{"alpha", "beta"}, query.Slugs)
	assert.Equal(t, []string{"-name", "id"}, query.Ordering)

	var body struct {
		Count    uint64  `json:"count"`
		Next     *string `json:"next"`
		Previous *string `json:"previous"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, uint64(1001), body.Count)
	require.NotNil(t, body.Next)
	assert.Contains(t, *body.Next, "limit=1000")
	assert.Contains(t, *body.Next, "offset=1000")
	assert.Nil(t, body.Previous)
}

func TestOrganizationRESTCreateUsesTypedFieldsAndOmitsAnnotatedCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &organizationHTTPServiceSpy{rackRole: restoredHTTPRackRole(t)}
	router := organizationTestRouter(service)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/dcim/rack-roles/",
		strings.NewReader(`{"name":"Core","slug":"core","color":"00ff00","description":"Core racks"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	require.Equal(t, 1, service.rackRoleCreateCalls)
	name, present := service.rackRoleCreate.Name.Get()
	require.True(t, present)
	assert.Equal(t, "Core", name)
	color, present := service.rackRoleCreate.Color.Get()
	require.True(t, present)
	assert.Equal(t, "00ff00", color)
	assert.Equal(t, "/api/dcim/rack-roles/7/", response.Header().Get("Location"))
	assert.NotContains(t, response.Body.String(), `"rack_count"`)
	assert.NotContains(t, response.Body.String(), `"data"`)
}

func TestOrganizationRESTRejectsUnsupportedFieldsBeforeCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &organizationHTTPServiceSpy{}
	router := organizationTestRouter(service)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodPost,
		"/api/dcim/manufacturers/",
		strings.NewReader(`{"name":"Acme","slug":"acme","comments":"not in profile"}`),
	))

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Zero(t, service.manufacturerCreateCalls)
	assert.JSONEq(t, `{"comments":["This field is not supported by the active capability profile."]}`, response.Body.String())
}

func organizationTestRouter(service *organizationHTTPServiceSpy) *gin.Engine {
	router := gin.New()
	NewOrganizationHandler(service, service).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-organization"})
		c.Next()
	})
	return router
}

func restoredHTTPRackRole(t *testing.T) *domaindcim.RackRole {
	t.Helper()
	now := shared.NewTimestamp(time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC))
	role, err := domaindcim.RestoreRackRole(domaindcim.RackRoleState{
		ID: 7, Name: "Core", Slug: "core", Color: "00ff00", Description: "Core racks",
		Created: now, LastUpdated: now, RackCount: 4,
	})
	require.NoError(t, err)
	return role
}

type organizationHTTPServiceSpy struct {
	manufacturerPage        applicationdcim.ManufacturerPage
	manufacturerQuery       applicationdcim.ListManufacturersQuery
	manufacturerListCalls   int
	manufacturerCreateCalls int
	rackRole                *domaindcim.RackRole
	rackRoleCreate          applicationdcim.CreateRackRoleCommand
	rackRoleCreateCalls     int
}

func (service *organizationHTTPServiceSpy) ListManufacturers(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListManufacturersQuery,
) (applicationdcim.ManufacturerPage, error) {
	service.manufacturerListCalls++
	service.manufacturerQuery = query
	return service.manufacturerPage, nil
}

func (*organizationHTTPServiceSpy) GetManufacturer(
	context.Context, identity.Principal, applicationdcim.GetManufacturerQuery,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (service *organizationHTTPServiceSpy) CreateManufacturer(
	context.Context, identity.Principal, applicationdcim.CreateManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	service.manufacturerCreateCalls++
	return nil, nil
}

func (*organizationHTTPServiceSpy) ReplaceManufacturer(
	context.Context, identity.Principal, applicationdcim.ReplaceManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationHTTPServiceSpy) UpdateManufacturer(
	context.Context, identity.Principal, applicationdcim.UpdateManufacturerCommand,
) (*domaindcim.Manufacturer, error) {
	return nil, nil
}

func (*organizationHTTPServiceSpy) DeleteManufacturer(
	context.Context, identity.Principal, applicationdcim.DeleteManufacturerCommand,
) error {
	return nil
}

func (*organizationHTTPServiceSpy) ListRackRoles(
	context.Context, identity.Principal, applicationdcim.ListRackRolesQuery,
) (applicationdcim.RackRolePage, error) {
	return applicationdcim.RackRolePage{}, nil
}

func (*organizationHTTPServiceSpy) GetRackRole(
	context.Context, identity.Principal, applicationdcim.GetRackRoleQuery,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (service *organizationHTTPServiceSpy) CreateRackRole(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateRackRoleCommand,
) (*domaindcim.RackRole, error) {
	service.rackRoleCreateCalls++
	service.rackRoleCreate = command
	return service.rackRole, nil
}

func (*organizationHTTPServiceSpy) ReplaceRackRole(
	context.Context, identity.Principal, applicationdcim.ReplaceRackRoleCommand,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (*organizationHTTPServiceSpy) UpdateRackRole(
	context.Context, identity.Principal, applicationdcim.UpdateRackRoleCommand,
) (*domaindcim.RackRole, error) {
	return nil, nil
}

func (*organizationHTTPServiceSpy) DeleteRackRole(
	context.Context, identity.Principal, applicationdcim.DeleteRackRoleCommand,
) error {
	return nil
}
