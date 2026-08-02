package workflow

import (
	"context"
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
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestParseVRFListPreservesRepeatedFiltersSignedIDsAndExplicitZeroLimit(t *testing.T) {
	query, err := parseVRFList(url.Values{
		"limit":          {"0"},
		"id":             {"-7", "0,11"},
		"name":           {"Blue", "Green"},
		"rd":             {"65000:1", "65000:2"},
		"enforce_unique": {"false"},
		"ordering":       {"name,rd"},
	})
	require.NoError(t, err)
	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumVRFPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-7, 0, 11}, query.IDs)
	assert.Equal(t, []string{"Blue", "Green"}, query.Names)
	assert.Equal(t, []string{"65000:1", "65000:2"}, query.RDs)
	require.NotNil(t, query.EnforceUnique)
	assert.False(t, *query.EnforceUnique)
	assert.Equal(t, []string{"name", "rd"}, query.Ordering)
}

func TestVRFRESTListUsesTypedServiceAndPinnedPageLinks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restVRFServiceSpy{listPage: applicationipam.VRFPage{
		Count:   1001,
		Results: []*domainipam.VRF{restVRFFixture(t, 7, "Blue", "65000:7", 3, 4)},
	}}
	router := gin.New()
	NewVRFRESTHandler(spy).Register(router, testVRFPrincipalMiddleware())

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodGet,
		"/api/ipam/vrfs/?limit=0&id=-1&id=7&name=Blue&name=Green&rd=65000%3A7&rd=65000%3A8",
		nil,
	))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, spy.listCalls)
	assert.Equal(t, []int64{-1, 7}, spy.listQuery.IDs)
	assert.Equal(t, []string{"Blue", "Green"}, spy.listQuery.Names)
	assert.Contains(t, response.Body.String(), `"next":"/api/ipam/vrfs/?`)
	assert.Contains(t, response.Body.String(), `limit=1000`)
	assert.Contains(t, response.Body.String(), `"ipaddress_count":3`)
	assert.Contains(t, response.Body.String(), `"prefix_count":4`)
}

func TestVRFRESTCreatePreservesNullableRDAndOmitsAnnotatedCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restVRFServiceSpy{created: restVRFFixture(t, 9, "No RD", "", 0, 0)}
	spy.created = restVRFNullRDFixture(t, 9, "No RD")
	router := gin.New()
	NewVRFRESTHandler(spy).Register(router, testVRFPrincipalMiddleware())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ipam/vrfs/",
		strings.NewReader(`{"name":"No RD","rd":null,"enforce_unique":false}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.Equal(t, "/api/ipam/vrfs/9/", response.Header().Get("Location"))
	assert.Equal(t, applicationipam.FieldPresent, spy.createCommand.Name.State())
	assert.Equal(t, applicationipam.FieldNull, spy.createCommand.RD.State())
	assert.Equal(t, applicationipam.FieldPresent, spy.createCommand.EnforceUnique.State())
	assert.Contains(t, response.Body.String(), `"rd":null`)
	assert.NotContains(t, response.Body.String(), "ipaddress_count")
	assert.NotContains(t, response.Body.String(), "prefix_count")
}

func TestVRFRESTRejectsUnknownFieldsBeforeCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restVRFServiceSpy{}
	router := gin.New()
	NewVRFRESTHandler(spy).Register(router, testVRFPrincipalMiddleware())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ipam/vrfs/",
		strings.NewReader(`{"name":"Blue","tenant":1}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Zero(t, spy.createCalls)
	assert.Contains(t, response.Body.String(), "tenant")
}

func testVRFPrincipalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "vrf-transport"})
		c.Next()
	}
}

func restVRFFixture(
	t *testing.T,
	id shared.ID,
	name string,
	rdValue string,
	ipAddressCount uint64,
	prefixCount uint64,
) *domainipam.VRF {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	stamp := shared.NewTimestamp(time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC))
	vrf, err := domainipam.RestoreVRF(domainipam.VRFState{
		ID: id, Name: name, RD: domainipam.NonNullRouteDistinguisher(rd), EnforceUnique: true,
		Created: stamp, LastUpdated: stamp, IPAddressCount: ipAddressCount, PrefixCount: prefixCount,
	})
	require.NoError(t, err)
	return vrf
}

func restVRFNullRDFixture(t *testing.T, id shared.ID, name string) *domainipam.VRF {
	t.Helper()
	stamp := shared.NewTimestamp(time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC))
	vrf, err := domainipam.RestoreVRF(domainipam.VRFState{
		ID: id, Name: name, RD: domainipam.NullRouteDistinguisher(), EnforceUnique: false,
		Created: stamp, LastUpdated: stamp,
	})
	require.NoError(t, err)
	return vrf
}

type restVRFServiceSpy struct {
	listCalls     int
	createCalls   int
	listQuery     applicationipam.ListVRFsQuery
	createCommand applicationipam.CreateVRFCommand
	listPage      applicationipam.VRFPage
	created       *domainipam.VRF
}

func (spy *restVRFServiceSpy) ListVRFs(
	_ context.Context,
	_ identity.Principal,
	query applicationipam.ListVRFsQuery,
) (applicationipam.VRFPage, error) {
	spy.listCalls++
	spy.listQuery = query
	return spy.listPage, nil
}

func (*restVRFServiceSpy) GetVRF(
	context.Context,
	identity.Principal,
	applicationipam.GetVRFQuery,
) (*domainipam.VRF, error) {
	return nil, nil
}

func (spy *restVRFServiceSpy) CreateVRF(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.CreateVRFCommand,
) (*domainipam.VRF, error) {
	spy.createCalls++
	spy.createCommand = command
	return spy.created, nil
}

func (*restVRFServiceSpy) ReplaceVRF(
	context.Context,
	identity.Principal,
	applicationipam.ReplaceVRFCommand,
) (*domainipam.VRF, error) {
	return nil, nil
}

func (*restVRFServiceSpy) UpdateVRF(
	context.Context,
	identity.Principal,
	applicationipam.UpdateVRFCommand,
) (*domainipam.VRF, error) {
	return nil, nil
}

func (*restVRFServiceSpy) DeleteVRF(
	context.Context,
	identity.Principal,
	applicationipam.DeleteVRFCommand,
) error {
	return nil
}
