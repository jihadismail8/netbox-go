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

func TestDeviceTypeRESTListPreservesRepeatedFiltersAndExplicitZeroLimit(t *testing.T) {
	service := &deviceTypeRESTServiceStub{deviceType: restDeviceTypeFixture(t)}
	router := deviceTypeRESTRouter(service)

	response := deviceTypeRESTRequest(
		t, router, http.MethodGet,
		"/api/dcim/device-types/?limit=0&id=-7&id=41"+
			"&manufacturer_id=-1&manufacturer_id=9"+
			"&manufacturer_slug=alpha&manufacturer_slug=beta"+
			"&model=A&model=B&slug=a&slug=b&ordering=-manufacturer,model",
		"",
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, uint32(0), service.listQuery.Limit)
	assert.True(t, service.listQuery.LimitPresent)
	assert.Equal(t, applicationdcim.MaximumDeviceTypePageLimit, service.listQuery.EffectiveLimit())
	assert.Equal(t, []int64{-7, 41}, service.listQuery.IDs)
	assert.Equal(t, []int64{-1, 9}, service.listQuery.ManufacturerIDs)
	assert.Equal(t, []string{"alpha", "beta"}, service.listQuery.ManufacturerSlugs)
	assert.Equal(t, []string{"A", "B"}, service.listQuery.Models)
	assert.Equal(t, []string{"a", "b"}, service.listQuery.Slugs)
	assert.Equal(t, []string{"-manufacturer", "model"}, service.listQuery.Ordering)
}

func TestDeviceTypeRESTCreatePreservesPresenceAndResponseCounters(t *testing.T) {
	service := &deviceTypeRESTServiceStub{deviceType: restDeviceTypeFixture(t)}
	router := deviceTypeRESTRouter(service)

	response := deviceTypeRESTRequest(
		t, router, http.MethodPost, "/api/dcim/device-types/",
		`{"manufacturer":9,"model":"Router 9000","slug":"router-9000",`+
			`"u_height":1.5,"exclude_from_utilization":false,"airflow":null}`,
	)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Equal(t, "/api/dcim/device-types/41/", response.Header().Get("Location"))
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.UHeight.State())
	height, present := service.createCommand.UHeight.Get()
	assert.True(t, present)
	assert.Equal(t, "1.5", height)
	assert.Equal(t, applicationdcim.FieldPresent, service.createCommand.ExcludeFromUtilization.State())
	excluded, present := service.createCommand.ExcludeFromUtilization.Get()
	assert.True(t, present)
	assert.False(t, excluded)
	assert.Equal(t, applicationdcim.FieldOmitted, service.createCommand.IsFullDepth.State())
	assert.Equal(t, applicationdcim.FieldNull, service.createCommand.Airflow.State())

	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.NotContains(t, body, "device_count")
	assert.Equal(t, float64(6), body["interface_template_count"])
	assert.Equal(t, float64(1.5), body["u_height"])
	assert.Equal(t, map[string]any{
		"value": "front-to-rear", "label": "Front to rear",
	}, body["airflow"])
}

func TestDeviceTypeRESTPatchDistinguishesOmittedNullBlankAndQuotedHeight(t *testing.T) {
	service := &deviceTypeRESTServiceStub{deviceType: restDeviceTypeFixture(t)}
	router := deviceTypeRESTRouter(service)

	response := deviceTypeRESTRequest(
		t, router, http.MethodPatch, "/api/dcim/device-types/41/",
		`{"u_height":"2.5","part_number":"","airflow":null}`,
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, shared.ID(41), service.updateCommand.ID)
	assert.Equal(t, applicationdcim.FieldOmitted, service.updateCommand.Model.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.UHeight.State())
	assert.Equal(t, applicationdcim.FieldPresent, service.updateCommand.PartNumber.State())
	assert.Equal(t, applicationdcim.FieldNull, service.updateCommand.Airflow.State())
	height, present := service.updateCommand.UHeight.Get()
	assert.True(t, present)
	assert.Equal(t, "2.5", height)
	partNumber, present := service.updateCommand.PartNumber.Get()
	assert.True(t, present)
	assert.Empty(t, partNumber)
}

func TestDeviceTypeRESTRegistersGetReplaceAndDelete(t *testing.T) {
	service := &deviceTypeRESTServiceStub{deviceType: restDeviceTypeFixture(t)}
	router := deviceTypeRESTRouter(service)

	get := deviceTypeRESTRequest(t, router, http.MethodGet, "/api/dcim/device-types/41/", "")
	require.Equal(t, http.StatusOK, get.Code, get.Body.String())
	var body map[string]any
	require.NoError(t, json.Unmarshal(get.Body.Bytes(), &body))
	assert.Equal(t, float64(4), body["device_count"])
	assert.Equal(t, float64(6), body["interface_template_count"])

	replace := deviceTypeRESTRequest(
		t, router, http.MethodPut, "/api/dcim/device-types/41/",
		`{"manufacturer":9,"model":"Router 9000","slug":"router-9000"}`,
	)
	require.Equal(t, http.StatusOK, replace.Code, replace.Body.String())
	deleted := deviceTypeRESTRequest(
		t, router, http.MethodDelete, "/api/dcim/device-types/41/", "",
	)
	require.Equal(t, http.StatusNoContent, deleted.Code, deleted.Body.String())
	assert.Equal(t, shared.ID(41), service.getQuery.ID)
	assert.Equal(t, shared.ID(41), service.replaceCommand.ID)
	assert.Equal(t, shared.ID(41), service.deleteCommand.ID)
}

func TestDeviceTypeRoutesUseTypedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &deviceTypeRESTServiceStub{deviceType: restDeviceTypeFixture(t)}
	router := gin.New()
	newCompleteTypedHandler(
		&typedSiteCallSpy{},
		WithDeviceTypeService(typed),
	).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 17, Username: "typed-device-type"})
		c.Next()
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/dcim/device-types/", nil),
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, typed.listCalls)
}

func deviceTypeRESTRouter(service DeviceTypeService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{
			ID: 17, Username: "device-type-rest", IsSuperuser: true,
		})
		c.Next()
	})
	NewDeviceTypeRESTHandler(service).Register(router)
	return router
}

func deviceTypeRESTRequest(
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

func restDeviceTypeFixture(t *testing.T) *domaindcim.DeviceType {
	t.Helper()
	reference, err := domaindcim.NewManufacturerReference(9, "Acme", "acme")
	require.NoError(t, err)
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 8, 0, 0, 0, time.UTC))
	deviceType, err := domaindcim.RestoreDeviceType(domaindcim.DeviceTypeState{
		ID: 41, Manufacturer: reference,
		Model: "Router 9000", Slug: "router-9000", PartNumber: "PN-9",
		UHeight: "1.5", IsFullDepth: true,
		Airflow:     domaindcim.NonNullDeviceAirflow(domaindcim.DeviceAirflowFrontToRear),
		Description: "Core router", Comments: "Notes",
		Created: now, LastUpdated: now,
		DeviceCount: 4, InterfaceTemplateCount: 6,
	})
	require.NoError(t, err)
	return deviceType
}

type deviceTypeRESTServiceStub struct {
	deviceType     *domaindcim.DeviceType
	listQuery      applicationdcim.ListDeviceTypesQuery
	getQuery       applicationdcim.GetDeviceTypeQuery
	createCommand  applicationdcim.CreateDeviceTypeCommand
	replaceCommand applicationdcim.ReplaceDeviceTypeCommand
	updateCommand  applicationdcim.UpdateDeviceTypeCommand
	deleteCommand  applicationdcim.DeleteDeviceTypeCommand
	listCalls      int
}

func (stub *deviceTypeRESTServiceStub) ListDeviceTypes(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.ListDeviceTypesQuery,
) (applicationdcim.DeviceTypePage, error) {
	stub.listCalls++
	stub.listQuery = query
	return applicationdcim.DeviceTypePage{}, nil
}

func (stub *deviceTypeRESTServiceStub) GetDeviceType(
	_ context.Context,
	_ identity.Principal,
	query applicationdcim.GetDeviceTypeQuery,
) (*domaindcim.DeviceType, error) {
	stub.getQuery = query
	return stub.deviceType, nil
}

func (stub *deviceTypeRESTServiceStub) CreateDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.CreateDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	stub.createCommand = command
	return stub.deviceType, nil
}

func (stub *deviceTypeRESTServiceStub) ReplaceDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.ReplaceDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	stub.replaceCommand = command
	return stub.deviceType, nil
}

func (stub *deviceTypeRESTServiceStub) UpdateDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.UpdateDeviceTypeCommand,
) (*domaindcim.DeviceType, error) {
	stub.updateCommand = command
	return stub.deviceType, nil
}

func (stub *deviceTypeRESTServiceStub) DeleteDeviceType(
	_ context.Context,
	_ identity.Principal,
	command applicationdcim.DeleteDeviceTypeCommand,
) error {
	stub.deleteCommand = command
	return nil
}
