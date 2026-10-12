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

func TestDeviceRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all fields omitted", func(t *testing.T) {
		input, err := decodeDevicePresenceJSON(t, `{}`)
		require.NoError(t, err)
		requireDeviceRESTStates(t, deviceCreateRESTStates(input.createCommand()), applicationdcim.FieldOmitted)
		requireDeviceRESTStates(
			t,
			deviceCreateRESTStates(input.replaceCommand(79).CreateDeviceCommand),
			applicationdcim.FieldOmitted,
		)
		requireDeviceRESTStates(t, deviceUpdateRESTStates(input.updateCommand(79)), applicationdcim.FieldOmitted)
	})

	t.Run("explicit null preserves nullability and coerces face and airflow to blank", func(t *testing.T) {
		input, err := decodeDevicePresenceJSON(t, `{
			"device_type":null,"role":null,"name":null,"site":null,"rack":null,
			"position":null,"face":null,"status":null,"serial":null,
			"asset_tag":null,"airflow":null,"description":null,"comments":null
		}`)
		require.NoError(t, err)
		for _, states := range []map[string]applicationdcim.FieldState{
			deviceCreateRESTStates(input.createCommand()),
			deviceCreateRESTStates(input.replaceCommand(79).CreateDeviceCommand),
			deviceUpdateRESTStates(input.updateCommand(79)),
		} {
			for field, state := range states {
				want := applicationdcim.FieldNull
				if field == "face" || field == "airflow" {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, field)
			}
		}
		requireDeviceRESTFieldValue(t, input.createCommand().Face, "", "face")
		requireDeviceRESTFieldValue(t, input.createCommand().Airflow, "", "airflow")
	})

	t.Run("zero blank and concrete values reach the application unchanged", func(t *testing.T) {
		input, err := decodeDevicePresenceJSON(t, `{
			"device_type":0,"role":0,"name":"","site":0,"rack":0,
			"position":0,"face":" front ","status":" active ","serial":"",
			"asset_tag":"","airflow":" front-to-rear ",
			"description":"  description  ","comments":""
		}`)
		require.NoError(t, err)
		create := input.createCommand()
		requireDeviceRESTStates(t, deviceCreateRESTStates(create), applicationdcim.FieldPresent)
		requireDeviceRESTFieldValue(t, create.DeviceType, shared.ID(0), "device_type")
		requireDeviceRESTFieldValue(t, create.Role, shared.ID(0), "role")
		requireDeviceRESTFieldValue(t, create.Name, "", "name")
		requireDeviceRESTFieldValue(t, create.Site, shared.ID(0), "site")
		requireDeviceRESTFieldValue(t, create.Rack, shared.ID(0), "rack")
		requireDeviceRESTFieldValue(t, create.Position, "0", "position")
		requireDeviceRESTFieldValue(t, create.Face, " front ", "face")
		requireDeviceRESTFieldValue(t, create.Status, " active ", "status")
		requireDeviceRESTFieldValue(t, create.Serial, "", "serial")
		requireDeviceRESTFieldValue(t, create.AssetTag, "", "asset_tag")
		requireDeviceRESTFieldValue(t, create.Airflow, " front-to-rear ", "airflow")
		requireDeviceRESTFieldValue(t, create.Description, "  description  ", "description")
		requireDeviceRESTFieldValue(t, create.Comments, "", "comments")

		requireDeviceRESTStates(
			t,
			deviceCreateRESTStates(input.replaceCommand(79).CreateDeviceCommand),
			applicationdcim.FieldPresent,
		)
		requireDeviceRESTStates(t, deviceUpdateRESTStates(input.updateCommand(79)), applicationdcim.FieldPresent)
	})

	for _, test := range []struct {
		name string
		body string
	}{
		{name: "nested DeviceType", body: `{"device_type":{"id":10}}`},
		{name: "string DeviceType", body: `{"device_type":"10"}`},
		{name: "nested role", body: `{"role":{"id":11}}`},
		{name: "string role", body: `{"role":"11"}`},
		{name: "nested Site", body: `{"site":{"id":12}}`},
		{name: "string Site", body: `{"site":"12"}`},
		{name: "nested Rack", body: `{"rack":{"id":13}}`},
		{name: "string Rack", body: `{"rack":"13"}`},
	} {
		t.Run("alternate relationship rejected/"+test.name, func(t *testing.T) {
			_, err := decodeDevicePresenceJSON(t, test.body)
			require.Error(t, err)
		})
	}
}

func decodeDevicePresenceJSON(t *testing.T, body string) (decodedDeviceInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/devices/79/",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeDeviceInput(context)
}

func deviceCreateRESTStates(command applicationdcim.CreateDeviceCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(), "role": command.Role.State(),
		"name": command.Name.State(), "site": command.Site.State(),
		"rack": command.Rack.State(), "position": command.Position.State(),
		"face": command.Face.State(), "status": command.Status.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"airflow": command.Airflow.State(), "description": command.Description.State(),
		"comments": command.Comments.State(),
	}
}

func deviceUpdateRESTStates(command applicationdcim.UpdateDeviceCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(), "role": command.Role.State(),
		"name": command.Name.State(), "site": command.Site.State(),
		"rack": command.Rack.State(), "position": command.Position.State(),
		"face": command.Face.State(), "status": command.Status.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"airflow": command.Airflow.State(), "description": command.Description.State(),
		"comments": command.Comments.State(),
	}
}

func requireDeviceRESTStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for field, state := range states {
		require.Equal(t, want, state, field)
	}
}

func requireDeviceRESTFieldValue[T comparable](
	t *testing.T,
	field applicationdcim.Field[T],
	want T,
	name string,
) {
	t.Helper()
	actual, present := field.Get()
	require.True(t, present, name)
	require.Equal(t, want, actual, name)
}
