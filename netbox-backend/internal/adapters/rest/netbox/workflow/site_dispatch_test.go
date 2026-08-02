package workflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
)

func TestProfileHandlerRejectsMissingTypedServices(t *testing.T) {
	require.PanicsWithValue(
		t,
		"REST profile handler requires typed services for: "+
			"Site, Manufacturer, RackRole, RackType, Rack, DeviceRole, "+
			"DeviceType, InterfaceTemplate, Device, Interface, VRF, Prefix, IPAddress",
		func() {
			NewHandler(nil)
		},
	)
}

func TestSiteRouteDispatchUsesTypedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &typedSiteCallSpy{}
	router := gin.New()
	newCompleteTypedHandler(typed).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-site"})
		c.Next()
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/dcim/sites/", nil))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, typed.listCalls)
}

func TestTypedOrganizationAndVRFRoutesUseTypedServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sites := &typedSiteCallSpy{}
	organizations := &organizationHTTPServiceSpy{}
	rackTypes := &rackTypeRESTServiceStub{}
	vrfs := &restVRFServiceSpy{}
	prefixes := &restPrefixServiceSpy{}
	router := gin.New()
	newCompleteTypedHandler(
		sites,
		WithOrganizationServices(organizations, organizations),
		WithRackTypeService(rackTypes),
		WithVRFService(vrfs),
		WithPrefixService(prefixes),
	).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-cutover"})
		c.Next()
	})

	for _, path := range []string{
		"/api/dcim/manufacturers/",
		"/api/dcim/rack-roles/",
		"/api/dcim/rack-types/",
		"/api/ipam/vrfs/",
		"/api/ipam/prefixes/",
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, response.Code, path+": "+response.Body.String())
	}

	require.Equal(t, 1, organizations.manufacturerListCalls)
	require.Equal(t, 1, vrfs.listCalls)
	require.Equal(t, 1, prefixes.listCalls)
}

type typedSiteCallSpy struct{ listCalls int }

func (spy *typedSiteCallSpy) ListSites(
	context.Context,
	identity.Principal,
	applicationdcim.ListSitesQuery,
) (applicationdcim.SitePage, error) {
	spy.listCalls++
	return applicationdcim.SitePage{}, nil
}
func (*typedSiteCallSpy) GetSite(context.Context, identity.Principal, applicationdcim.GetSiteQuery) (*domaindcim.Site, error) {
	return nil, nil
}
func (*typedSiteCallSpy) CreateSite(context.Context, identity.Principal, applicationdcim.CreateSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*typedSiteCallSpy) ReplaceSite(context.Context, identity.Principal, applicationdcim.ReplaceSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*typedSiteCallSpy) UpdateSite(context.Context, identity.Principal, applicationdcim.UpdateSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*typedSiteCallSpy) DeleteSite(context.Context, identity.Principal, applicationdcim.DeleteSiteCommand) error {
	return nil
}

// typedProfileServiceFallback satisfies every non-Site profile service
// contract for focused dispatch tests. Each test overrides the service it
// exercises; invoking an embedded nil fallback would fail the test loudly.
type typedProfileServiceFallback struct {
	ManufacturerService
	RackRoleService
	RackTypeService
	RackService
	DeviceRoleService
	DeviceTypeService
	InterfaceTemplateService
	DeviceService
	InterfaceService
	VRFService
	PrefixService
	IPAddressService
}

func newCompleteTypedHandler(
	sites SiteService,
	overrides ...HandlerOption,
) *Handler {
	fallback := &typedProfileServiceFallback{}
	options := []HandlerOption{
		WithOrganizationServices(fallback, fallback),
		WithRackTypeService(fallback),
		WithRackService(fallback),
		WithDeviceRoleService(fallback),
		WithDeviceTypeService(fallback),
		WithInterfaceTemplateService(fallback),
		WithDeviceService(fallback),
		WithInterfaceService(fallback),
		WithVRFService(fallback),
		WithPrefixService(fallback),
		WithIPAddressService(fallback),
	}
	options = append(options, overrides...)
	return NewHandler(sites, options...)
}
