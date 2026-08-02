package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
)

func TestSiteRPCDispatchUsesTypedService(t *testing.T) {
	typed := &grpcTypedSiteCallSpy{}
	services := completeTypedDCIMTestServices()
	services.sites = typed
	server := services.server()

	response, err := server.ListSites(
		identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-site"}),
		&dcimv1.ListSitesRequest{},
	)
	require.NoError(t, err)
	require.Equal(t, uint64(0), response.Page.Count)
	require.Equal(t, 1, typed.listCalls)
}

func TestOrganizationRPCDispatchUsesTypedServices(t *testing.T) {
	organizations := &organizationGRPCServiceSpy{}
	rackTypes := &rackTypeGRPCServiceSpy{}
	services := completeTypedDCIMTestServices()
	services.manufacturers = organizations
	services.rackRoles = organizations
	services.rackTypes = rackTypes
	server := services.server()
	ctx := identity.WithPrincipal(t.Context(), identity.Principal{ID: 1, Username: "typed-organization"})

	manufacturerResponse, err := server.ListManufacturers(ctx, &dcimv1.ListManufacturersRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(0), manufacturerResponse.Page.Count)
	rackRoleResponse, err := server.ListRackRoles(ctx, &dcimv1.ListRackRolesRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(0), rackRoleResponse.Page.Count)
	rackTypeResponse, err := server.ListRackTypes(ctx, &dcimv1.ListRackTypesRequest{})
	require.NoError(t, err)
	require.Equal(t, uint64(0), rackTypeResponse.Page.Count)
	require.Equal(t, 1, organizations.manufacturerListCalls)
}

type grpcTypedSiteCallSpy struct{ listCalls int }

func (spy *grpcTypedSiteCallSpy) ListSites(
	context.Context,
	identity.Principal,
	applicationdcim.ListSitesQuery,
) (applicationdcim.SitePage, error) {
	spy.listCalls++
	return applicationdcim.SitePage{}, nil
}
func (*grpcTypedSiteCallSpy) GetSite(context.Context, identity.Principal, applicationdcim.GetSiteQuery) (*domaindcim.Site, error) {
	return nil, nil
}
func (*grpcTypedSiteCallSpy) CreateSite(context.Context, identity.Principal, applicationdcim.CreateSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*grpcTypedSiteCallSpy) ReplaceSite(context.Context, identity.Principal, applicationdcim.ReplaceSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*grpcTypedSiteCallSpy) UpdateSite(context.Context, identity.Principal, applicationdcim.UpdateSiteCommand) (*domaindcim.Site, error) {
	return nil, nil
}
func (*grpcTypedSiteCallSpy) DeleteSite(context.Context, identity.Principal, applicationdcim.DeleteSiteCommand) error {
	return nil
}
