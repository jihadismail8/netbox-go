package identity

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	application "netbox-go/internal/application/identity"
	domain "netbox-go/internal/domain/identity"
)

type restCredentialClock struct{ now time.Time }

func (clock restCredentialClock) Now() time.Time { return clock.now }

type restCredentialStore struct {
	application.Store
	record    application.TokenRecord
	user      domain.User
	lookupErr error
	touchErr  error
}

func (store *restCredentialStore) TokenByHash(context.Context, []byte) (application.TokenRecord, domain.User, error) {
	return store.record, store.user, store.lookupErr
}

func (store *restCredentialStore) TouchToken(context.Context, int64, time.Time) error {
	return store.touchErr
}

func TestRESTTokenCredentialMatrixInfrastructureFailures(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)
	infrastructureFailure := errors.New("credential storage unavailable")

	tests := []struct {
		name  string
		store *restCredentialStore
	}{
		{
			name: "lookup",
			store: &restCredentialStore{
				lookupErr: infrastructureFailure,
			},
		},
		{
			name: "touch",
			store: &restCredentialStore{
				record: application.TokenRecord{Token: domain.APIToken{
					ID: 17, UserID: 41, WriteEnabled: true,
				}},
				user:     domain.User{ID: 41, Username: "rest-user", IsActive: true},
				touchErr: infrastructureFailure,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := application.NewService(test.store, restCredentialClock{now: now})
			handler := NewHandler(service, false)
			handlerCalled := false
			router := gin.New()
			router.GET("/protected", handler.BaselineMiddleware(), func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.RemoteAddr = "192.0.2.1:443"
			request.Header.Set("Authorization", "Token present")
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			require.Equal(t, http.StatusInternalServerError, response.Code)
			require.JSONEq(t, `{"detail":"An internal error occurred."}`, response.Body.String())
			require.NotContains(t, response.Body.String(), infrastructureFailure.Error())
			require.False(t, handlerCalled)
		})
	}
}

func TestRESTTokenCredentialMatrixSourceDenialPrecedesWriteDenial(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2026, 8, 17, 7, 0, 0, 0, time.UTC)
	store := &restCredentialStore{
		record: application.TokenRecord{Token: domain.APIToken{
			ID: 17, UserID: 41, WriteEnabled: false,
			AllowedIPs: []string{"192.0.2.0/24"}, LastUsed: &now,
		}},
		user: domain.User{ID: 41, Username: "rest-user", IsActive: true},
	}
	service := application.NewService(store, restCredentialClock{now: now})
	handler := NewHandler(service, false)
	handlerCalled := false
	router := gin.New()
	router.POST("/protected", handler.BaselineMiddleware(), func(c *gin.Context) {
		handlerCalled = true
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodPost, "/protected", nil)
	request.RemoteAddr = "198.51.100.1:443"
	request.Header.Set("Authorization", "Token present")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusForbidden, response.Code)
	require.JSONEq(t, `{"detail":"Source IP 198.51.100.1 is not permitted to authenticate using this token."}`, response.Body.String())
	require.NotContains(t, response.Body.String(), "permission")
	require.False(t, handlerCalled)
}
