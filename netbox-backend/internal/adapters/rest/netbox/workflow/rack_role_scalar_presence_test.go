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

func TestRackRoleRESTScalarPresenceMatrix(t *testing.T) {
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
			body:      `{"name":null,"slug":null,"color":null,"description":null}`,
			wantState: applicationdcim.FieldNull,
		},
		{
			name:      "present blank",
			body:      `{"name":"","slug":"","color":"","description":""}`,
			wantState: applicationdcim.FieldPresent,
		},
		{
			name:      "present concrete without transport normalization",
			body:      `{"name":"  value  ","slug":"  value  ","color":"  value  ","description":"  value  "}`,
			wantState: applicationdcim.FieldPresent,
			wantValue: "  value  ",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeRackRoleJSON(t, test.body)
			require.NoError(t, err)

			create := input.createCommand()
			requireRackRoleRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": create.Name, "slug": create.Slug, "color": create.Color,
				"description": create.Description,
			}, test.wantState, test.wantValue)

			replace := input.replaceCommand(79)
			require.Equal(t, shared.ID(79), replace.ID)
			requireRackRoleRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": replace.Name, "slug": replace.Slug, "color": replace.Color,
				"description": replace.Description,
			}, test.wantState, test.wantValue)

			update := input.updateCommand(79)
			require.Equal(t, shared.ID(79), update.ID)
			requireRackRoleRESTFields(t, map[string]applicationdcim.Field[string]{
				"name": update.Name, "slug": update.Slug, "color": update.Color,
				"description": update.Description,
			}, test.wantState, test.wantValue)
		})
	}
}

func decodeRackRoleJSON(t *testing.T, body string) (decodedRackRoleInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/rack-roles/79/",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeRackRoleInput(context)
}

func requireRackRoleRESTFields(
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
