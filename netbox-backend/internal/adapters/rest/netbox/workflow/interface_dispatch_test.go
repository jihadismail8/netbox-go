package workflow

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/identity"
)

func TestInterfaceRoutesUseTypedService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	typed := &interfaceServiceSpy{}
	router := gin.New()
	newCompleteTypedHandler(
		&typedSiteCallSpy{},
		WithInterfaceService(typed),
	).Register(router, func(c *gin.Context) {
		SetPrincipal(c, identity.Principal{ID: 1, Username: "typed-interface"})
		c.Next()
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(
		response,
		httptest.NewRequest(http.MethodGet, "/api/dcim/interfaces/", nil),
	)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Equal(t, 1, typed.listCalls)
}
