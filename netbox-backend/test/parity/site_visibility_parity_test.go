package parity

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/application/authz"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

func TestTypedSiteObjectVisibilityHasRESTGRPCCountAndPageParity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:typed_site_visibility_parity?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(dcimrow.Models()...))
	require.NoError(t, db.AutoMigrate(ipamrow.Models()...))
	require.NoError(t, db.AutoMigrate(&postgreschangelog.ChangeRow{}))

	admin := identity.Principal{ID: 1, Username: "admin", IsSuperuser: true}
	writerCore := composition.NewCoreWithAuthorizer(db, authz.AllowAll{})
	first, err := writerCore.Sites.CreateSite(t.Context(), admin, applicationdcim.CreateSiteCommand{
		Name: applicationdcim.FieldValue("First"), Slug: applicationdcim.FieldValue("first"),
	})
	require.NoError(t, err)
	second, err := writerCore.Sites.CreateSite(t.Context(), admin, applicationdcim.CreateSiteCommand{
		Name: applicationdcim.FieldValue("Second"), Slug: applicationdcim.FieldValue("second"),
	})
	require.NoError(t, err)

	permissionAuthorizer := authz.PermissionAuthorizer{}
	permissionCore := composition.NewCoreWithAuthorizer(db, permissionAuthorizer)
	principal := identity.Principal{
		ID: 2, Username: "restricted",
		Permissions: map[string]struct{}{"dcim.view_site": {}},
		ObjectVisibility: map[string]map[int64]struct{}{
			"dcim.view_site": {second.ID().Int64(): {}},
		},
	}

	router := newParityRESTRouter(permissionCore, principal)
	restResponse := httptest.NewRecorder()
	router.ServeHTTP(restResponse, httptest.NewRequest(http.MethodGet, "/api/dcim/sites/?ordering=id&limit=1", nil))
	require.Equal(t, http.StatusOK, restResponse.Code, restResponse.Body.String())
	var restPage struct {
		Count   uint64 `json:"count"`
		Results []struct {
			ID int64 `json:"id"`
		} `json:"results"`
	}
	require.NoError(t, json.Unmarshal(restResponse.Body.Bytes(), &restPage))
	require.Equal(t, uint64(1), restPage.Count)
	require.Len(t, restPage.Results, 1)
	require.Equal(t, second.ID().Int64(), restPage.Results[0].ID)
	require.NotEqual(t, first.ID(), second.ID())

	grpcResponse, err := newParityDCIMServer(permissionCore).ListSites(
		identity.WithPrincipal(t.Context(), principal),
		&dcimv1.ListSitesRequest{},
	)
	require.NoError(t, err)
	require.Equal(t, restPage.Count, grpcResponse.Page.Count)
	require.Len(t, grpcResponse.Results, 1)
	require.Equal(t, restPage.Results[0].ID, grpcResponse.Results[0].Id)

	denied := identity.Principal{ID: 3, Username: "denied"}
	deniedRouter := newParityRESTRouter(permissionCore, denied)
	deniedREST := httptest.NewRecorder()
	deniedRouter.ServeHTTP(deniedREST, httptest.NewRequest(http.MethodGet, "/api/dcim/sites/", nil))
	require.Equal(t, http.StatusForbidden, deniedREST.Code, deniedREST.Body.String())
	_, err = newParityDCIMServer(permissionCore).ListSites(
		identity.WithPrincipal(t.Context(), denied),
		&dcimv1.ListSitesRequest{},
	)
	require.Equal(t, codes.PermissionDenied, status.Code(err))
}
