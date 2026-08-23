package workflow

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestSiteRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name      string
		body      string
		wantState applicationdcim.FieldState
		wantValue string
	}{
		{name: "omitted", body: `{}`, wantState: applicationdcim.FieldOmitted},
		{
			name:      "explicit null",
			body:      `{"name":null,"slug":null,"status":null,"facility":null,"description":null,"comments":null}`,
			wantState: applicationdcim.FieldNull,
		},
		{
			name:      "present blank",
			body:      `{"name":"","slug":"","status":"","facility":"","description":"","comments":""}`,
			wantState: applicationdcim.FieldPresent,
		},
		{
			name:      "present concrete without transport normalization",
			body:      `{"name":"  value  ","slug":"  value  ","status":"  value  ","facility":"  value  ","description":"  value  ","comments":"  value  "}`,
			wantState: applicationdcim.FieldPresent,
			wantValue: "  value  ",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeSiteJSON(t, test.body)
			require.NoError(t, err)

			create := input.createCommand()
			requireSiteRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": create.Name, "slug": create.Slug, "status": create.Status,
				"facility": create.Facility, "description": create.Description,
				"comments": create.Comments,
			}, test.wantState, test.wantValue)

			replace := input.replaceCommand(73)
			require.Equal(t, shared.ID(73), replace.ID)
			requireSiteRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": replace.Name, "slug": replace.Slug, "status": replace.Status,
				"facility": replace.Facility, "description": replace.Description,
				"comments": replace.Comments,
			}, test.wantState, test.wantValue)

			update := input.updateCommand(73)
			require.Equal(t, shared.ID(73), update.ID)
			requireSiteRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": update.Name, "slug": update.Slug, "status": update.Status,
				"facility": update.Facility, "description": update.Description,
				"comments": update.Comments,
			}, test.wantState, test.wantValue)
		})
	}
}

func decodeSiteJSON(t *testing.T, body string) (decodedSiteInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPatch, "/api/dcim/sites/73/", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeSiteInput(context)
}

func requireSiteRESTFields(
	t *testing.T,
	fields map[string]applicationdcim.Field[string],
	wantState applicationdcim.FieldState,
	wantValue string,
) {
	t.Helper()
	for name, field := range fields {
		require.Equal(t, wantState, field.State(), name)
		if wantState != applicationdcim.FieldPresent {
			continue
		}
		got, present := field.Get()
		require.True(t, present, name)
		require.Equal(t, wantValue, got, name)
	}
}
