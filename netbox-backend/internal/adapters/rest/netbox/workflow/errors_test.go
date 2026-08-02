package workflow

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"netbox-go/internal/domain/shared"
)

func TestWriteErrorMapsOnlyCanonicalTypedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "validation",
			err:        shared.Invalid("limit", "A valid integer is required."),
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"limit":["A valid integer is required."]}`,
		},
		{
			name:       "unauthenticated",
			err:        shared.Unauthenticated(),
			wantStatus: http.StatusForbidden,
			wantBody:   `{"detail":"Authentication credentials were not provided."}`,
		},
		{
			name:       "unknown is sanitized",
			err:        errors.New("database details must stay private"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"detail":"An internal error occurred."}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)

			writeError(context, test.err)

			assert.Equal(t, test.wantStatus, response.Code)
			assert.JSONEq(t, test.wantBody, response.Body.String())
		})
	}
}
