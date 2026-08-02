package parity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/application/authz"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

func TestRESTAndGRPCReachSameSiteState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:site_parity?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(dcimrow.Models()...))
	require.NoError(t, db.AutoMigrate(ipamrow.Models()...))
	require.NoError(t, db.AutoMigrate(&postgreschangelog.ChangeRow{}))
	core := composition.NewCoreWithAuthorizer(db, authz.AllowAll{})
	principal := identity.Principal{ID: 1, Username: "admin", IsSuperuser: true}
	router := newParityRESTRouter(core, principal)
	request := httptest.NewRequest(http.MethodPost, "/api/dcim/sites/", bytes.NewBufferString(`{"name":"Parity Site","slug":"parity-site"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	var rest map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &rest))
	id := int64(rest["id"].(float64))
	grpcServer := newParityDCIMServer(core)
	ctx := identity.WithPrincipal(t.Context(), principal)
	got, err := grpcServer.GetSite(ctx, &dcimv1.GetSiteRequest{Id: id})
	require.NoError(t, err)
	require.Equal(t, "Parity Site", got.Site.Name)
	updated, err := grpcServer.UpdateSite(ctx, &dcimv1.UpdateSiteRequest{Id: id, Site: &dcimv1.SiteInput{Description: ptr("from grpc")}, UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"description"}}})
	require.NoError(t, err)
	require.Equal(t, "from grpc", updated.Site.Description)
	detail := httptest.NewRecorder()
	router.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/dcim/sites/"+strconv.FormatInt(id, 10)+"/", nil))
	require.Equal(t, http.StatusOK, detail.Code)
	require.Contains(t, detail.Body.String(), "from grpc")
}
func ptr(value string) *string { return &value }
