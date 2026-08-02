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
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

func TestParsePrefixListPreservesRepeatedFiltersSignedIDsAndExplicitZeroLimit(t *testing.T) {
	query, err := parsePrefixList(url.Values{
		"limit":          {"0"},
		"id":             {"-7", "0,11"},
		"vrf_id":         {"-9", "0,13"},
		"vrf_rd":         {"65000:1", "65000:2"},
		"prefix":         {"10.0.0.0/8", "2001:db8::/32"},
		"family":         {"-4"},
		"status":         {"active", "reserved"},
		"within":         {"10.0.0.0/8"},
		"within_include": {"10.1.0.0/16"},
		"contains":       {"10.1.2.3"},
		"ordering":       {"vrf,-prefix"},
	})
	require.NoError(t, err)
	assert.True(t, query.LimitPresent)
	assert.Equal(t, applicationipam.MaximumPrefixPageLimit, query.EffectiveLimit())
	assert.Equal(t, []int64{-7, 0, 11}, query.IDs)
	assert.Equal(t, []int64{-9, 0, 13}, query.VRFIDs)
	assert.Equal(t, []string{"65000:1", "65000:2"}, query.VRFRDs)
	assert.Equal(t, []string{"10.0.0.0/8", "2001:db8::/32"}, query.Prefixes)
	assert.Equal(t, []string{"active", "reserved"}, query.Statuses)
	require.NotNil(t, query.Family)
	assert.Equal(t, int64(-4), *query.Family)
	assert.Equal(t, "10.0.0.0/8", *query.Within)
	assert.Equal(t, "10.1.0.0/16", *query.WithinInclude)
	assert.Equal(t, "10.1.2.3", *query.Contains)
	assert.Equal(t, []string{"vrf", "-prefix"}, query.Ordering)
}

func TestPrefixRESTListUsesTypedServiceAndProjectsRelationshipsAndHierarchy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	vrf := restPrefixVRFReference(t, 7, "Blue", "65000:7", true)
	prefix := restPrefixFixture(t, 17, "10.0.0.0/8", domainipam.NonNullVRFReference(vrf), 2, 3)
	spy := &restPrefixServiceSpy{listPage: applicationipam.PrefixPage{
		Count: 1001, Results: []*domainipam.Prefix{prefix},
	}}
	router := gin.New()
	NewPrefixRESTHandler(spy).Register(router, testPrefixPrincipalMiddleware())

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(
		http.MethodGet,
		"/api/ipam/prefixes/?limit=0&id=-1&id=17&vrf_id=-2&vrf_id=7&prefix=10.0.0.0%2F8&status=reserved",
		nil,
	))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, spy.listCalls)
	assert.Equal(t, []int64{-1, 17}, spy.listQuery.IDs)
	assert.Equal(t, []int64{-2, 7}, spy.listQuery.VRFIDs)
	assert.Equal(t, []string{"10.0.0.0/8"}, spy.listQuery.Prefixes)

	var body prefixListDTO
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	require.Len(t, body.Results, 1)
	projected := body.Results[0]
	assert.Equal(t, prefixChoiceUint32DTO{Value: 4, Label: "IPv4"}, projected.Family)
	assert.Equal(t, prefixChoiceStringDTO{Value: "reserved", Label: "Reserved"}, projected.Status)
	require.NotNil(t, projected.VRF)
	assert.Equal(t, int64(7), projected.VRF.ID)
	assert.Equal(t, "Blue (65000:7)", projected.VRF.Display)
	assert.Equal(t, uint64(2), projected.Children)
	assert.Equal(t, uint32(3), projected.Depth)
	require.NotNil(t, body.Next)
	assert.Contains(t, *body.Next, "limit=1000")
}

func TestPrefixRESTPatchCanClearNullableVRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restPrefixServiceSpy{prefix: restPrefixFixture(
		t, 17, "10.0.0.0/8", domainipam.NullVRFReference(), 0, 0,
	)}
	router := gin.New()
	NewPrefixRESTHandler(spy).Register(router, testPrefixPrincipalMiddleware())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPatch, "/api/ipam/prefixes/17/", strings.NewReader(`{"vrf":null}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, 1, spy.updateCalls)
	assert.Equal(t, applicationipam.FieldNull, spy.updateCommand.VRF.State())
	assert.Equal(t, applicationipam.FieldOmitted, spy.updateCommand.Prefix.State())
	assert.Contains(t, response.Body.String(), `"vrf":null`)
}

func TestPrefixRESTRejectsUnknownFieldsBeforeCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	spy := &restPrefixServiceSpy{}
	router := gin.New()
	NewPrefixRESTHandler(spy).Register(router, testPrefixPrincipalMiddleware())

	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ipam/prefixes/",
		strings.NewReader(`{"prefix":"10.0.0.0/8","tenant":1}`),
	)
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	assert.Zero(t, spy.createCalls)
	assert.Contains(t, response.Body.String(), "tenant")
}

func testPrefixPrincipalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "prefix-transport"})
		c.Next()
	}
}

func restPrefixFixture(
	t *testing.T,
	id shared.ID,
	network string,
	vrf domainipam.NullableVRFReference,
	children uint64,
	depth uint32,
) *domainipam.Prefix {
	t.Helper()
	stamp := shared.NewTimestamp(time.Date(2026, time.July, 22, 14, 0, 0, 0, time.UTC))
	prefix, err := domainipam.RestorePrefix(domainipam.PrefixState{
		ID: id, Prefix: network, VRF: vrf, Status: domainipam.PrefixStatusReserved.String(),
		IsPool: true, MarkUtilized: true, Description: "description", Comments: "comments",
		Created: stamp, LastUpdated: stamp, Children: children, Depth: depth,
	})
	require.NoError(t, err)
	return prefix
}

func restPrefixVRFReference(
	t *testing.T,
	id shared.ID,
	name string,
	rdValue string,
	enforceUnique bool,
) domainipam.VRFReference {
	t.Helper()
	rd, err := domainipam.ParseRouteDistinguisher(rdValue)
	require.NoError(t, err)
	reference, err := domainipam.NewVRFReference(
		id, name, domainipam.NonNullRouteDistinguisher(rd), enforceUnique,
	)
	require.NoError(t, err)
	return reference
}

type restPrefixServiceSpy struct {
	listCalls     int
	createCalls   int
	updateCalls   int
	listQuery     applicationipam.ListPrefixesQuery
	updateCommand applicationipam.UpdatePrefixCommand
	listPage      applicationipam.PrefixPage
	prefix        *domainipam.Prefix
}

func (spy *restPrefixServiceSpy) ListPrefixes(
	_ context.Context,
	_ identity.Principal,
	query applicationipam.ListPrefixesQuery,
) (applicationipam.PrefixPage, error) {
	spy.listCalls++
	spy.listQuery = query
	return spy.listPage, nil
}

func (spy *restPrefixServiceSpy) GetPrefix(
	context.Context,
	identity.Principal,
	applicationipam.GetPrefixQuery,
) (*domainipam.Prefix, error) {
	return spy.prefix, nil
}

func (spy *restPrefixServiceSpy) CreatePrefix(
	context.Context,
	identity.Principal,
	applicationipam.CreatePrefixCommand,
) (*domainipam.Prefix, error) {
	spy.createCalls++
	return spy.prefix, nil
}

func (spy *restPrefixServiceSpy) ReplacePrefix(
	context.Context,
	identity.Principal,
	applicationipam.ReplacePrefixCommand,
) (*domainipam.Prefix, error) {
	return spy.prefix, nil
}

func (spy *restPrefixServiceSpy) UpdatePrefix(
	_ context.Context,
	_ identity.Principal,
	command applicationipam.UpdatePrefixCommand,
) (*domainipam.Prefix, error) {
	spy.updateCalls++
	spy.updateCommand = command
	return spy.prefix, nil
}

func (*restPrefixServiceSpy) DeletePrefix(
	context.Context,
	identity.Principal,
	applicationipam.DeletePrefixCommand,
) error {
	return nil
}
