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

func TestInterfaceTemplateRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name      string
		body      string
		wantState applicationdcim.FieldState
		assert    func(*testing.T, applicationdcim.CreateInterfaceTemplateCommand)
	}{
		{name: "omitted", body: `{}`, wantState: applicationdcim.FieldOmitted},
		{
			name: "explicit null",
			body: `{"device_type":null,"name":null,"label":null,"type":null,` +
				`"enabled":null,"mgmt_only":null,"description":null}`,
			wantState: applicationdcim.FieldNull,
		},
		{
			name: "present zero blank and false",
			body: `{"device_type":0,"name":"","label":"","type":"",` +
				`"enabled":false,"mgmt_only":false,"description":""}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateInterfaceTemplateCommand) {
				requireInterfaceTemplateRESTFieldValue(t, command.DeviceType, shared.ID(0), "device_type")
				requireInterfaceTemplateRESTFieldValue(t, command.Name, "", "name")
				requireInterfaceTemplateRESTFieldValue(t, command.Label, "", "label")
				requireInterfaceTemplateRESTFieldValue(t, command.Type, "", "type")
				requireInterfaceTemplateRESTFieldValue(t, command.Enabled, false, "enabled")
				requireInterfaceTemplateRESTFieldValue(t, command.MgmtOnly, false, "mgmt_only")
				requireInterfaceTemplateRESTFieldValue(t, command.Description, "", "description")
			},
		},
		{
			name: "present concrete without transport normalization",
			body: `{"device_type":73,"name":"  Ethernet1  ","label":"  WAN  ",` +
				`"type":"10gbase-sr","enabled":true,"mgmt_only":true,` +
				`"description":"  description  "}`,
			wantState: applicationdcim.FieldPresent,
			assert: func(t *testing.T, command applicationdcim.CreateInterfaceTemplateCommand) {
				requireInterfaceTemplateRESTFieldValue(t, command.DeviceType, shared.ID(73), "device_type")
				requireInterfaceTemplateRESTFieldValue(t, command.Name, "  Ethernet1  ", "name")
				requireInterfaceTemplateRESTFieldValue(t, command.Label, "  WAN  ", "label")
				requireInterfaceTemplateRESTFieldValue(t, command.Type, "10gbase-sr", "type")
				requireInterfaceTemplateRESTFieldValue(t, command.Enabled, true, "enabled")
				requireInterfaceTemplateRESTFieldValue(t, command.MgmtOnly, true, "mgmt_only")
				requireInterfaceTemplateRESTFieldValue(t, command.Description, "  description  ", "description")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeInterfaceTemplateJSON(t, test.body)
			require.NoError(t, err)

			create := input.createCommand()
			requireInterfaceTemplateRESTStates(
				t, interfaceTemplateCreateRESTStates(create), test.wantState,
			)
			if test.assert != nil {
				test.assert(t, create)
			}

			replace := input.replaceCommand(79)
			require.Equal(t, shared.ID(79), replace.ID)
			requireInterfaceTemplateRESTStates(
				t,
				interfaceTemplateCreateRESTStates(replace.CreateInterfaceTemplateCommand),
				test.wantState,
			)

			update := input.updateCommand(79)
			require.Equal(t, shared.ID(79), update.ID)
			requireInterfaceTemplateRESTStates(
				t, interfaceTemplateUpdateRESTStates(update), test.wantState,
			)
		})
	}

	for _, body := range []string{
		`{"device_type":{"id":73}}`,
		`{"device_type":"73"}`,
	} {
		_, err := decodeInterfaceTemplateJSON(t, body)
		require.Error(t, err, "alternate DeviceType forms must fail at REST decoding")
	}
}

func decodeInterfaceTemplateJSON(
	t *testing.T,
	body string,
) (decodedInterfaceTemplateInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(
		http.MethodPatch,
		"/api/dcim/interface-templates/79/",
		strings.NewReader(body),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeInterfaceTemplateInput(context)
}

func interfaceTemplateCreateRESTStates(
	command applicationdcim.CreateInterfaceTemplateCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(),
		"name":        command.Name.State(),
		"label":       command.Label.State(),
		"type":        command.Type.State(),
		"enabled":     command.Enabled.State(),
		"mgmt_only":   command.MgmtOnly.State(),
		"description": command.Description.State(),
	}
}

func interfaceTemplateUpdateRESTStates(
	command applicationdcim.UpdateInterfaceTemplateCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(),
		"name":        command.Name.State(),
		"label":       command.Label.State(),
		"type":        command.Type.State(),
		"enabled":     command.Enabled.State(),
		"mgmt_only":   command.MgmtOnly.State(),
		"description": command.Description.State(),
	}
}

func requireInterfaceTemplateRESTStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireInterfaceTemplateRESTFieldValue[T comparable](
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
