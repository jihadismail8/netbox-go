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

func TestRackTypeRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name      string
		body      string
		wantState applicationdcim.FieldState
		assert    func(*testing.T, applicationdcim.CreateRackTypeCommand)
	}{
		{name: "omitted", body: `{}`, wantState: applicationdcim.FieldOmitted},
		{
			name: "explicit null",
			body: `{"manufacturer":null,"model":null,"slug":null,"form_factor":null,` +
				`"width":null,"u_height":null,"starting_unit":null,"desc_units":null,` +
				`"description":null,"comments":null}`,
			wantState: applicationdcim.FieldNull,
		},
		{
			name: "present zero blank and false",
			body: `{"manufacturer":0,"model":"","slug":"","form_factor":"",` +
				`"width":0,"u_height":0,"starting_unit":0,"desc_units":false,` +
				`"description":"","comments":""}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateRackTypeCommand) {
				requireRackTypeRESTFieldValue(t, command.Manufacturer, shared.ID(0), "manufacturer")
				requireRackTypeRESTFieldValue(t, command.Model, "", "model")
				requireRackTypeRESTFieldValue(t, command.Slug, "", "slug")
				requireRackTypeRESTFieldValue(t, command.FormFactor, "", "form_factor")
				requireRackTypeRESTFieldValue(t, command.Width, uint32(0), "width")
				requireRackTypeRESTFieldValue(t, command.UHeight, uint32(0), "u_height")
				requireRackTypeRESTFieldValue(t, command.StartingUnit, uint32(0), "starting_unit")
				requireRackTypeRESTFieldValue(t, command.DescUnits, false, "desc_units")
				requireRackTypeRESTFieldValue(t, command.Description, "", "description")
				requireRackTypeRESTFieldValue(t, command.Comments, "", "comments")
			},
		},
		{
			name: "present concrete without transport normalization",
			body: `{"manufacturer":73,"model":"  model  ","slug":"  slug  ",` +
				`"form_factor":"  factor  ","width":23,"u_height":48,"starting_unit":2,` +
				`"desc_units":true,"description":"  description  ","comments":"  comments  "}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateRackTypeCommand) {
				requireRackTypeRESTFieldValue(t, command.Manufacturer, shared.ID(73), "manufacturer")
				requireRackTypeRESTFieldValue(t, command.Model, "  model  ", "model")
				requireRackTypeRESTFieldValue(t, command.Slug, "  slug  ", "slug")
				requireRackTypeRESTFieldValue(t, command.FormFactor, "  factor  ", "form_factor")
				requireRackTypeRESTFieldValue(t, command.Width, uint32(23), "width")
				requireRackTypeRESTFieldValue(t, command.UHeight, uint32(48), "u_height")
				requireRackTypeRESTFieldValue(t, command.StartingUnit, uint32(2), "starting_unit")
				requireRackTypeRESTFieldValue(t, command.DescUnits, true, "desc_units")
				requireRackTypeRESTFieldValue(t, command.Description, "  description  ", "description")
				requireRackTypeRESTFieldValue(t, command.Comments, "  comments  ", "comments")
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeRackTypeJSON(t, test.body)
			require.NoError(t, err)

			create := input.createCommand()
			requireRackTypeRESTStates(t, rackTypeCreateRESTStates(create), test.wantState)
			if test.assert != nil {
				test.assert(t, create)
			}

			replace := input.replaceCommand(79)
			require.Equal(t, shared.ID(79), replace.ID)
			requireRackTypeRESTStates(t, rackTypeCreateRESTStates(replace.CreateRackTypeCommand), test.wantState)

			update := input.updateCommand(79)
			require.Equal(t, shared.ID(79), update.ID)
			requireRackTypeRESTStates(t, rackTypeUpdateRESTStates(update), test.wantState)
		})
	}
}

func decodeRackTypeJSON(t *testing.T, body string) (decodedRackTypeInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/rack-types/79/",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeRackTypeInput(context)
}

func rackTypeCreateRESTStates(command applicationdcim.CreateRackTypeCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "form_factor": command.FormFactor.State(),
		"width": command.Width.State(), "u_height": command.UHeight.State(),
		"starting_unit": command.StartingUnit.State(), "desc_units": command.DescUnits.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func rackTypeUpdateRESTStates(command applicationdcim.UpdateRackTypeCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "form_factor": command.FormFactor.State(),
		"width": command.Width.State(), "u_height": command.UHeight.State(),
		"starting_unit": command.StartingUnit.State(), "desc_units": command.DescUnits.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func requireRackTypeRESTStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireRackTypeRESTFieldValue[T comparable](
	t *testing.T,
	field applicationdcim.Field[T],
	want T,
	name string,
) {
	t.Helper()
	got, present := field.Get()
	require.True(t, present, name)
	require.Equal(t, want, got, name)
}
