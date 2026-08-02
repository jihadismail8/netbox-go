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

func TestRackTypeRESTListPreservesRepeatedFiltersAndExplicitZeroLimit(t *testing.T) {
	service := &rackTypeRESTServiceStub{rackType: restRackTypeFixture(t)}
	router := rackTypeRESTRouter(service)

	response := rackTypeRESTRequest(t, router, http.MethodGet,
		"/api/dcim/rack-types/?limit=0&id=-7&id=42&manufacturer_id=-1&manufacturer_id=9"+
			"&manufacturer_slug=alpha&manufacturer_slug=beta&model=A&model=B&slug=a&slug=b"+
			"&ordering=-manufacturer,model", "")
	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, uint32(0), service.listQuery.Limit)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumRackTypePageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, []int64{-7, 42}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1, 9}, service.listQuery.ManufacturerIDs)
	assert.Equal(t, []string{"alpha", "beta"}, service.listQuery.ManufacturerSlugs)
	assert.Equal(t, []string{"A", "B"}, service.listQuery.Models)
	assert.Equal(t, []string{"a", "b"}, service.listQuery.Slugs)
	assert.Equal(t, []string{"-manufacturer", "model"}, service.listQuery.Ordering)
}

func TestRackTypeRESTCreatePreservesPresenceAndReturnsNetBoxShape(t *testing.T) {
	service := &rackTypeRESTServiceStub{rackType: restRackTypeFixture(t)}
	router := rackTypeRESTRouter(service)
	response := rackTypeRESTRequest(t, router, http.MethodPost, "/api/dcim/rack-types/",
		`{"manufacturer":9,"model":"R42","slug":"r42","form_factor":"4-post-cabinet","desc_units":false}`)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Equal(t, "/api/dcim/rack-types/41/", response.Header().Get("Location"))
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Width.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.UHeight.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.StartingUnit.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.DescUnits.State())
	descUnits, present := service.createCommand.DescUnits.Get()
	assert.True(t, present)
	assert.False(t, descUnits)
	manufacturer, present := service.createCommand.Manufacturer.Get()
	assert.True(t, present)
	assert.Equal(t, shared.ID(9), manufacturer)

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "R42", body["display"])
	assert.Equal(t, map[string]any{
		"id": float64(9), "url": "/api/dcim/manufacturers/9/", "display": "Acme",
	}, body["manufacturer"])
	assert.Equal(t, map[string]any{"value": "4-post-cabinet", "label": "4-post cabinet"}, body["form_factor"])
	assert.Equal(t, map[string]any{"value": float64(19), "label": "19 inches"}, body["width"])
}

func TestRackTypeRESTPatchDistinguishesOmittedNullAndBlank(t *testing.T) {
	service := &rackTypeRESTServiceStub{rackType: restRackTypeFixture(t)}
	router := rackTypeRESTRouter(service)
	response := rackTypeRESTRequest(t, router, http.MethodPatch, "/api/dcim/rack-types/41/",
		`{"width":23,"description":"","comments":null}`)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, shared.ID(41), service.updateCommand.ID)
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Manufacturer.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Width.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Description.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Comments.State())
	description, present := service.updateCommand.Description.Get()
	assert.True(t, present)
	assert.Empty(t, description)
}

func TestRackTypeRESTRejectsUnknownAndNonIntegralFieldsBeforeService(t *testing.T) {
	service := &rackTypeRESTServiceStub{rackType: restRackTypeFixture(t)}
	router := rackTypeRESTRouter(service)

	unknown := rackTypeRESTRequest(t, router, http.MethodPost, "/api/dcim/rack-types/", `{"tags":[]}`)
	assert.Equal(t, http.StatusBadRequest, unknown.Code)
	assert.JSONEq(t, `{"tags":["This field is not supported by the active capability profile."]}`, unknown.Body.String())

	nonIntegral := rackTypeRESTRequest(t, router, http.MethodPost, "/api/dcim/rack-types/", `{"manufacturer":1.5}`)
	assert.Equal(t, http.StatusBadRequest, nonIntegral.Code)
	assert.JSONEq(t, `{"manufacturer":["A valid integer is required."]}`, nonIntegral.Body.String())
	assert.Zero(t, service.createCalls)
}

func TestRackTypeRESTRegistersGetReplaceAndDelete(t *testing.T) {
	service := &rackTypeRESTServiceStub{rackType: restRackTypeFixture(t)}
	router := rackTypeRESTRouter(service)

	get := rackTypeRESTRequest(t, router, http.MethodGet, "/api/dcim/rack-types/41/", "")
	assert.Equal(t, http.StatusOK, get.Code)
	replace := rackTypeRESTRequest(t, router, http.MethodPut, "/api/dcim/rack-types/41/",
		`{"manufacturer":9,"model":"R42","slug":"r42","form_factor":"4-post-cabinet"}`)
	assert.Equal(t, http.StatusOK, replace.Code, replace.Body.String())
	deleted := rackTypeRESTRequest(t, router, http.MethodDelete, "/api/dcim/rack-types/41/", "")
	assert.Equal(t, http.StatusNoContent, deleted.Code)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)
}

func rackTypeRESTRouter(service RackTypeService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 17, Username: "rack-type-rest", IsSuperuser: true})
		c.Next()
	})
	NewRackTypeRESTHandler(service).Register(router)
	return router
}

func rackTypeRESTRequest(
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

func restRackTypeFixture(t *testing.T) *domaindcim.RackType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	rackType, err := domaindcim.RestoreRackType(domaindcim.RackTypeState{
		ID: 41, Manufacturer: reference, Model: "R42", Slug: "r42",
		FormFactor: "4-post-cabinet", Width: 19, UHeight: 42, StartingUnit: 1,
		Description: "Rack type", Comments: "Notes", Created: now, LastUpdated: now,
	})
	require.NoError(t, err)
	return rackType
}

type rackTypeRESTServiceStub struct {
	rackType       *domaindcim.RackType
	listQuery      applicationdcim.ListRackTypesQuery
	getQuery       applicationdcim.GetRackTypeQuery
	createCommand  applicationdcim.CreateRackTypeCommand
	replaceCommand applicationdcim.ReplaceRackTypeCommand
	updateCommand  applicationdcim.UpdateRackTypeCommand
	deleteCommand  applicationdcim.DeleteRackTypeCommand
	createCalls    int
}

func (stub *rackTypeRESTServiceStub) ListRackTypes(_ context.Context, _ identity.Principal, query applicationdcim.ListRackTypesQuery) (applicationdcim.RackTypePage, error) {
	stub.listQuery = query
	return applicationdcim.RackTypePage{Results: []*domaindcim.RackType{}}, nil
}
func (stub *rackTypeRESTServiceStub) GetRackType(_ context.Context, _ identity.Principal, query applicationdcim.GetRackTypeQuery) (*domaindcim.RackType, error) {
	stub.getQuery = query
	return stub.rackType, nil
}
func (stub *rackTypeRESTServiceStub) CreateRackType(_ context.Context, _ identity.Principal, command applicationdcim.CreateRackTypeCommand) (*domaindcim.RackType, error) {
	stub.createCalls++
	stub.createCommand = command
	return stub.rackType, nil
}
func (stub *rackTypeRESTServiceStub) ReplaceRackType(_ context.Context, _ identity.Principal, command applicationdcim.ReplaceRackTypeCommand) (*domaindcim.RackType, error) {
	stub.replaceCommand = command
	return stub.rackType, nil
}
func (stub *rackTypeRESTServiceStub) UpdateRackType(_ context.Context, _ identity.Principal, command applicationdcim.UpdateRackTypeCommand) (*domaindcim.RackType, error) {
	stub.updateCommand = command
	return stub.rackType, nil
}
func (stub *rackTypeRESTServiceStub) DeleteRackType(_ context.Context, _ identity.Principal, command applicationdcim.DeleteRackTypeCommand) error {
	stub.deleteCommand = command
	return nil
}
