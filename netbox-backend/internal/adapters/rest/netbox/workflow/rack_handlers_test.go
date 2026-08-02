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

func TestRackRESTListPreservesRepeatedProfileFiltersAndExplicitZeroLimit(t *testing.T) {
	service := &rackRESTServiceStub{rack: restRackFixture(t)}
	router := rackRESTRouter(service)

	response := rackRESTRequest(
		t, router, http.MethodGet,
		"/api/dcim/racks/?limit=0&id=-1&id=41&site_id=-1&site_id=3"+
			"&site_slug=old&site_slug=moscow&name=A01&name=A02&status=active"+
			"&role_id=9&role_slug=production&rack_type_id=8&rack_type_slug=r24"+
			"&ordering=-site,name",
		"",
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, uint32(0), service.listQuery.Limit)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumRackPageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, []int64{-1, 41}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1, 3}, service.listQuery.SiteIDs)
	assert.Equal(t, []string{"old", "moscow"}, service.listQuery.SiteSlugs)
	assert.Equal(t, []string{"A01", "A02"}, service.listQuery.Names)
	assert.Equal(t, []string{"active"}, service.listQuery.Statuses)
	assert.Equal(t, []int64{9}, service.listQuery.RoleIDs)
	assert.Equal(t, []string{"production"}, service.listQuery.RoleSlugs)
	assert.Equal(t, []int64{8}, service.listQuery.RackTypeIDs)
	assert.Equal(t, []string{"r24"}, service.listQuery.RackTypeSlugs)
	assert.Equal(t, []string{"-site", "name"}, service.listQuery.Ordering)
}

func TestRackRESTCreatePreservesPresenceAndReturnsNetBoxShape(t *testing.T) {
	service := &rackRESTServiceStub{rack: restRackFixture(t)}
	router := rackRESTRouter(service)

	response := rackRESTRequest(
		t, router, http.MethodPost, "/api/dcim/racks/",
		`{"site":3,"name":"A01","rack_type":8,"role":9,"asset_tag":"","airflow":null}`,
	)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Equal(t, "/api/dcim/racks/41/", response.Header().Get("Location"))
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.AssetTag.State())
	assetTag, present := service.createCommand.AssetTag.Get()
	assert.True(t, present)
	assert.Empty(t, assetTag)
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.Airflow.State())
	airflow, present := service.createCommand.Airflow.Get()
	assert.True(t, present)
	assert.Empty(t, airflow, "REST null is coerced to the pinned blank airflow choice")
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.Width.State())

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "A01 (F01)", body["display"])
	assert.Equal(t, map[string]any{
		"id": float64(3), "url": "/api/dcim/sites/3/", "display": "Moscow",
	}, body["site"])
	assert.Equal(t, map[string]any{
		"id": float64(8), "url": "/api/dcim/rack-types/8/", "display": "R24",
	}, body["rack_type"])
	assert.Equal(t, map[string]any{
		"id": float64(9), "url": "/api/dcim/rack-roles/9/", "display": "Production",
	}, body["role"])
	assert.Equal(t, map[string]any{"value": "active", "label": "Active"}, body["status"])
	assert.Equal(t, map[string]any{"value": "wall-frame", "label": "Wall-mounted frame"}, body["form_factor"])
	assert.Equal(t, map[string]any{"value": float64(23), "label": "23 inches"}, body["width"])
	assert.Equal(t, map[string]any{"value": "front-to-rear", "label": "Front to rear"}, body["airflow"])
	_, hasCount := body["device_count"]
	assert.False(t, hasCount, "create responses omit queryset-only counters")
}

func TestRackRESTPatchDistinguishesNullableClearBlankAndOmission(t *testing.T) {
	service := &rackRESTServiceStub{rack: restRackFixture(t)}
	router := rackRESTRouter(service)

	response := rackRESTRequest(
		t, router, http.MethodPatch, "/api/dcim/racks/41/",
		`{"facility_id":null,"asset_tag":"","form_factor":null,"airflow":null}`,
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.FacilityID.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.AssetTag.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.FormFactor.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.Airflow.State())
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Name.State())
}

func TestRackRESTRegistersRetrieveReplaceDeleteAndRejectsUnknownFields(t *testing.T) {
	service := &rackRESTServiceStub{rack: restRackFixture(t)}
	router := rackRESTRouter(service)

	get := rackRESTRequest(t, router, http.MethodGet, "/api/dcim/racks/41/", "")
	assert.Equal(t, http.StatusOK, get.Code)
	replace := rackRESTRequest(
		t, router, http.MethodPut, "/api/dcim/racks/41/", `{"site":3,"name":"A01"}`,
	)
	assert.Equal(t, http.StatusOK, replace.Code, replace.Body.String())
	deleted := rackRESTRequest(t, router, http.MethodDelete, "/api/dcim/racks/41/", "")
	assert.Equal(t, http.StatusNoContent, deleted.Code)
	unknown := rackRESTRequest(
		t, router, http.MethodPost, "/api/dcim/racks/", `{"site":3,"name":"A01","location":7}`,
	)
	assert.Equal(t, http.StatusBadRequest, unknown.Code)
	assert.JSONEq(
		t, `{"location":["This field is not supported by the active capability profile."]}`,
		unknown.Body.String(),
	)
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)
}

func TestRackRouteDispatchUsesTypedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &rackRESTServiceStub{rack: restRackFixture(t)}
	router := gin.New()
	newCompleteTypedHandler(
		&typedSiteCallSpy{},
		WithRackService(typed),
	).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 17, Username: "typed-rack", IsSuperuser: true})
		c.Next()
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dcim/racks/", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, typed.listCalls)
}

func rackRESTRouter(service RackService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 17, Username: "rack-rest", IsSuperuser: true})
		c.Next()
	})
	NewRackRESTHandler(service).Register(router)
	return router
}

func rackRESTRequest(
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

func restRackFixture(t *testing.T) *domaindcim.Rack {
	t.Helper()
	site, err := domaindcim.NewSiteReference(3, "Moscow", "moscow")
	require.NoError(t, err)
	rackType, err := domaindcim.NewRackTypeReference(
		8, "R24", "r24", domaindcim.RackPhysicalAttributes{
			FormFactor: domaindcim.RackFormFactorWallFrame,
			Width:      domaindcim.RackWidth23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		},
	)
	require.NoError(t, err)
	role, err := domaindcim.NewRackRoleReference(9, "Production", "production")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	rack, err := domaindcim.RestoreRack(domaindcim.RackState{
		ID: 41, Site: site, Name: "A01",
		FacilityID: domaindcim.NonNullRackValue("F01"),
		RackType:   domaindcim.NonNullRackValue(rackType),
		Status:     "active", Role: domaindcim.NonNullRackValue(role),
		Serial: "serial", AssetTag: domaindcim.NonNullRackValue(""),
		FormFactor: domaindcim.NonNullRackValue("wall-frame"),
		Width:      23, UHeight: 24, StartingUnit: 3, DescUnits: true,
		Airflow:     domaindcim.NonNullRackValue("front-to-rear"),
		Description: "Rack", Comments: "Notes",
		Created: now, LastUpdated: now, DeviceCount: 2,
	})
	require.NoError(t, err)
	return rack
}

type rackRESTServiceStub struct {
	rack           *domaindcim.Rack
	listQuery      applicationdcim.ListRacksQuery
	getQuery       applicationdcim.GetRackQuery
	createCommand  applicationdcim.CreateRackCommand
	replaceCommand applicationdcim.ReplaceRackCommand
	updateCommand  applicationdcim.UpdateRackCommand
	deleteCommand  applicationdcim.DeleteRackCommand
	listCalls      int
}

func (stub *rackRESTServiceStub) ListRacks(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListRacksQuery,
) (applicationdcim.RackPage, error) {
	stub.listCalls++
	stub.listQuery = query
	return applicationdcim.RackPage{Count: 1, Results: []*domaindcim.Rack{stub.rack}}, nil
}

func (stub *rackRESTServiceStub) GetRack(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetRackQuery,
) (*domaindcim.Rack, error) {
	stub.getQuery = query
	return stub.rack, nil
}

func (stub *rackRESTServiceStub) CreateRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateRackCommand,
) (*domaindcim.Rack, error) {
	stub.createCommand = command
	return stub.rack, nil
}

func (stub *rackRESTServiceStub) ReplaceRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceRackCommand,
) (*domaindcim.Rack, error) {
	stub.replaceCommand = command
	return stub.rack, nil
}

func (stub *rackRESTServiceStub) UpdateRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateRackCommand,
) (*domaindcim.Rack, error) {
	stub.updateCommand = command
	return stub.rack, nil
}

func (stub *rackRESTServiceStub) DeleteRack(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteRackCommand,
) error {
	stub.deleteCommand = command
	return nil
}
