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

func TestRackRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all fields omitted", func(t *testing.T) {
		input, err := decodeRackPresenceJSON(t, `{}`)
		require.NoError(t, err)
		requireRackRESTStates(t, rackCreateRESTStates(input.createCommand()), applicationdcim.FieldOmitted)
		requireRackRESTStates(
			t, rackCreateRESTStates(input.replaceCommand(79).CreateRackCommand),
			applicationdcim.FieldOmitted,
		)
		requireRackRESTStates(t, rackUpdateRESTStates(input.updateCommand(79)), applicationdcim.FieldOmitted)
	})

	t.Run("explicit null preserves nullability and coerces airflow to blank", func(t *testing.T) {
		input, err := decodeRackPresenceJSON(t, `{
			"site":null,"name":null,"facility_id":null,"rack_type":null,
			"status":null,"role":null,"serial":null,"asset_tag":null,
			"form_factor":null,"width":null,"u_height":null,"starting_unit":null,
			"desc_units":null,"airflow":null,"description":null,"comments":null
		}`)
		require.NoError(t, err)
		for _, states := range []map[string]applicationdcim.FieldState{
			rackCreateRESTStates(input.createCommand()),
			rackCreateRESTStates(input.replaceCommand(79).CreateRackCommand),
			rackUpdateRESTStates(input.updateCommand(79)),
		} {
			for field, state := range states {
				want := applicationdcim.FieldNull
				if field == "airflow" {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, field)
			}
		}
		requireRackRESTFieldValue(t, input.createCommand().Airflow, "", "airflow")
	})

	t.Run("zero blank false and concrete values reach the application unchanged", func(t *testing.T) {
		input, err := decodeRackPresenceJSON(t, `{
			"site":0,"name":"","facility_id":"","rack_type":0,
			"status":" active ","role":0,"serial":"","asset_tag":"",
			"form_factor":" wall-frame ","width":0,"u_height":0,"starting_unit":0,
			"desc_units":false,"airflow":" front-to-rear ",
			"description":"  description  ","comments":""
		}`)
		require.NoError(t, err)
		create := input.createCommand()
		requireRackRESTStates(t, rackCreateRESTStates(create), applicationdcim.FieldPresent)
		requireRackRESTFieldValue(t, create.Site, shared.ID(0), "site")
		requireRackRESTFieldValue(t, create.Name, "", "name")
		requireRackRESTFieldValue(t, create.FacilityID, "", "facility_id")
		requireRackRESTFieldValue(t, create.RackType, shared.ID(0), "rack_type")
		requireRackRESTFieldValue(t, create.Status, " active ", "status")
		requireRackRESTFieldValue(t, create.Role, shared.ID(0), "role")
		requireRackRESTFieldValue(t, create.Serial, "", "serial")
		requireRackRESTFieldValue(t, create.AssetTag, "", "asset_tag")
		requireRackRESTFieldValue(t, create.FormFactor, " wall-frame ", "form_factor")
		requireRackRESTFieldValue(t, create.Width, uint32(0), "width")
		requireRackRESTFieldValue(t, create.UHeight, uint32(0), "u_height")
		requireRackRESTFieldValue(t, create.StartingUnit, uint32(0), "starting_unit")
		requireRackRESTFieldValue(t, create.DescUnits, false, "desc_units")
		requireRackRESTFieldValue(t, create.Airflow, " front-to-rear ", "airflow")
		requireRackRESTFieldValue(t, create.Description, "  description  ", "description")
		requireRackRESTFieldValue(t, create.Comments, "", "comments")

		requireRackRESTStates(
			t, rackCreateRESTStates(input.replaceCommand(79).CreateRackCommand),
			applicationdcim.FieldPresent,
		)
		requireRackRESTStates(t, rackUpdateRESTStates(input.updateCommand(79)), applicationdcim.FieldPresent)
	})

	for _, test := range []struct {
		name string
		body string
	}{
		{name: "nested Site", body: `{"site":{"id":3}}`},
		{name: "string Site", body: `{"site":"3"}`},
		{name: "nested RackType", body: `{"rack_type":{"id":8}}`},
		{name: "string RackType", body: `{"rack_type":"8"}`},
		{name: "nested role", body: `{"role":{"id":9}}`},
		{name: "string role", body: `{"role":"9"}`},
	} {
		test := test
		t.Run("alternate relationship rejected/"+test.name, func(t *testing.T) {
			_, err := decodeRackPresenceJSON(t, test.body)
			require.Error(t, err)
		})
	}
}

func decodeRackPresenceJSON(t *testing.T, body string) (decodedRackInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch, "/api/dcim/racks/79/", strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeRackInput(context)
}

func rackCreateRESTStates(command applicationdcim.CreateRackCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"site": command.Site.State(), "name": command.Name.State(),
		"facility_id": command.FacilityID.State(), "rack_type": command.RackType.State(),
		"status": command.Status.State(), "role": command.Role.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"form_factor": command.FormFactor.State(), "width": command.Width.State(),
		"u_height": command.UHeight.State(), "starting_unit": command.StartingUnit.State(),
		"desc_units": command.DescUnits.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func rackUpdateRESTStates(command applicationdcim.UpdateRackCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"site": command.Site.State(), "name": command.Name.State(),
		"facility_id": command.FacilityID.State(), "rack_type": command.RackType.State(),
		"status": command.Status.State(), "role": command.Role.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"form_factor": command.FormFactor.State(), "width": command.Width.State(),
		"u_height": command.UHeight.State(), "starting_unit": command.StartingUnit.State(),
		"desc_units": command.DescUnits.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func requireRackRESTStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for field, state := range states {
		require.Equal(t, want, state, field)
	}
}

func requireRackRESTFieldValue[T comparable](
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
