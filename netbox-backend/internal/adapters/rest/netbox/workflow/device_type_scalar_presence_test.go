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

func TestDeviceTypeRESTScalarPresenceMatrix(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantState applicationdcim.FieldState
		assert    func(*testing.T, applicationdcim.CreateDeviceTypeCommand)
	}{
		{name: "omitted", body: `{}`, wantState: applicationdcim.FieldOmitted},
		{
			name: "explicit null",
			body: `{"manufacturer":null,"model":null,"slug":null,"part_number":null,` +
				`"u_height":null,"exclude_from_utilization":null,"is_full_depth":null,` +
				`"airflow":null,"description":null,"comments":null}`,
			wantState: applicationdcim.FieldNull,
		},
		{
			name: "present zero blank and false",
			body: `{"manufacturer":0,"model":"","slug":"","part_number":"",` +
				`"u_height":0,"exclude_from_utilization":false,"is_full_depth":false,` +
				`"airflow":"","description":"","comments":""}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateDeviceTypeCommand) {
				requireDeviceTypeRESTFieldValue(t, command.Manufacturer, shared.ID(0), "manufacturer")
				requireDeviceTypeRESTFieldValue(t, command.Model, "", "model")
				requireDeviceTypeRESTFieldValue(t, command.Slug, "", "slug")
				requireDeviceTypeRESTFieldValue(t, command.PartNumber, "", "part_number")
				requireDeviceTypeRESTFieldValue(t, command.UHeight, "0", "u_height")
				requireDeviceTypeRESTFieldValue(t, command.ExcludeFromUtilization, false, "exclude_from_utilization")
				requireDeviceTypeRESTFieldValue(t, command.IsFullDepth, false, "is_full_depth")
				requireDeviceTypeRESTFieldValue(t, command.Airflow, "", "airflow")
				requireDeviceTypeRESTFieldValue(t, command.Description, "", "description")
				requireDeviceTypeRESTFieldValue(t, command.Comments, "", "comments")
			},
		},
		{
			name: "present concrete without transport normalization",
			body: `{"manufacturer":73,"model":"  model  ","slug":"  slug  ",` +
				`"part_number":"  part  ","u_height":"2.5",` +
				`"exclude_from_utilization":true,"is_full_depth":true,` +
				`"airflow":"front-to-rear","description":"  description  ",` +
				`"comments":"  comments  "}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateDeviceTypeCommand) {
				requireDeviceTypeRESTFieldValue(t, command.Manufacturer, shared.ID(73), "manufacturer")
				requireDeviceTypeRESTFieldValue(t, command.Model, "  model  ", "model")
				requireDeviceTypeRESTFieldValue(t, command.Slug, "  slug  ", "slug")
				requireDeviceTypeRESTFieldValue(t, command.PartNumber, "  part  ", "part_number")
				requireDeviceTypeRESTFieldValue(t, command.UHeight, "2.5", "u_height")
				requireDeviceTypeRESTFieldValue(t, command.ExcludeFromUtilization, true, "exclude_from_utilization")
				requireDeviceTypeRESTFieldValue(t, command.IsFullDepth, true, "is_full_depth")
				requireDeviceTypeRESTFieldValue(t, command.Airflow, "front-to-rear", "airflow")
				requireDeviceTypeRESTFieldValue(t, command.Description, "  description  ", "description")
				requireDeviceTypeRESTFieldValue(t, command.Comments, "  comments  ", "comments")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeDeviceTypeJSON(t, test.body)
			require.NoError(t, err)

			create := input.createCommand()
			requireDeviceTypeRESTStates(t, deviceTypeCreateRESTStates(create), test.wantState)
			if test.assert != nil {
				test.assert(t, create)
			}

			replace := input.replaceCommand(79)
			require.Equal(t, shared.ID(79), replace.ID)
			requireDeviceTypeRESTStates(
				t, deviceTypeCreateRESTStates(replace.CreateDeviceTypeCommand), test.wantState,
			)

			update := input.updateCommand(79)
			require.Equal(t, shared.ID(79), update.ID)
			requireDeviceTypeRESTStates(t, deviceTypeUpdateRESTStates(update), test.wantState)
		})
	}

	_, err := decodeDeviceTypeJSON(t, `{"manufacturer":{"id":73}}`)
	require.Error(t, err, "alternate nested Manufacturer forms must fail at REST decoding")
}

func decodeDeviceTypeJSON(t *testing.T, body string) (decodedDeviceTypeInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/device-types/79/",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeDeviceTypeInput(context)
}

func deviceTypeCreateRESTStates(
	command applicationdcim.CreateDeviceTypeCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer":             command.Manufacturer.State(),
		"model":                    command.Model.State(),
		"slug":                     command.Slug.State(),
		"part_number":              command.PartNumber.State(),
		"u_height":                 command.UHeight.State(),
		"exclude_from_utilization": command.ExcludeFromUtilization.State(),
		"is_full_depth":            command.IsFullDepth.State(),
		"airflow":                  command.Airflow.State(),
		"description":              command.Description.State(),
		"comments":                 command.Comments.State(),
	}
}

func deviceTypeUpdateRESTStates(
	command applicationdcim.UpdateDeviceTypeCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer":             command.Manufacturer.State(),
		"model":                    command.Model.State(),
		"slug":                     command.Slug.State(),
		"part_number":              command.PartNumber.State(),
		"u_height":                 command.UHeight.State(),
		"exclude_from_utilization": command.ExcludeFromUtilization.State(),
		"is_full_depth":            command.IsFullDepth.State(),
		"airflow":                  command.Airflow.State(),
		"description":              command.Description.State(),
		"comments":                 command.Comments.State(),
	}
}

func requireDeviceTypeRESTStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireDeviceTypeRESTFieldValue[T comparable](
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
