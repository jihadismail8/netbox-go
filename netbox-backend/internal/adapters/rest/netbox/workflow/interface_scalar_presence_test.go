package workflow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceRESTScalarPresenceMatrix(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name string
		body string
		want applicationdcim.FieldState
	}{
		{"omitted", `{}`, applicationdcim.FieldOmitted},
		{"null", `{"device":null,"name":null,"label":null,"type":null,"enabled":null,"mgmt_only":null,"mtu":null,"speed":null,"duplex":null,"description":null}`, applicationdcim.FieldNull},
		{"present", `{"device":0,"name":"","label":"","type":" other ","enabled":false,"mgmt_only":false,"mtu":0,"speed":0,"duplex":"","description":"  text  "}`, applicationdcim.FieldPresent},
	} {
		t.Run(test.name, func(t *testing.T) {
			input, err := decodeInterfacePresenceJSON(t, test.body)
			require.NoError(t, err)
			for _, states := range []map[string]applicationdcim.FieldState{
				interfaceCreateRESTStates(input.createCommand()),
				interfaceCreateRESTStates(input.replaceCommand(41).CreateInterfaceCommand),
				interfaceUpdateRESTStates(input.updateCommand(41)),
			} {
				for field, state := range states {
					require.Equal(t, test.want, state, field)
				}
			}
			if test.want == applicationdcim.FieldPresent {
				command := input.createCommand()
				requireInterfaceRESTValue(t, command.Device, shared.ID(0))
				requireInterfaceRESTValue(t, command.Name, "")
				requireInterfaceRESTValue(t, command.Label, "")
				requireInterfaceRESTValue(t, command.Type, " other ")
				requireInterfaceRESTValue(t, command.Enabled, false)
				requireInterfaceRESTValue(t, command.MgmtOnly, false)
				requireInterfaceRESTValue(t, command.Duplex, "")
				requireInterfaceRESTValue(t, command.Description, "  text  ")
			}
		})
	}

	for _, test := range []struct {
		name string
		raw  string
		want string
	}{
		{"negative", "-1", "-1"},
		{"zero", "0", "0"},
		{"MTU upper bound", "65536", "65536"},
		{"speed upper bound", "2147483647", "2147483647"},
		{"above semantic bound", "2147483648", "2147483648"},
		{"integral decimal", "1.0", "1"},
		{"integral exponent", "1e3", "1000"},
		{"minimum signed envelope", "-9223372036854775808", "-9223372036854775808"},
		{"maximum signed envelope", "9223372036854775807", "9223372036854775807"},
	} {
		for _, field := range []string{"mtu", "speed"} {
			t.Run("signed numeric intent/"+field+"/"+test.name, func(t *testing.T) {
				input, err := decodeInterfacePresenceJSON(t, `{"`+field+`":`+test.raw+`}`)
				require.NoError(t, err)
				command := input.createCommand()
				if field == "mtu" {
					value, present := command.MTU.Get()
					require.True(t, present)
					require.Equal(t, test.want, fmt.Sprint(value))
				} else {
					value, present := command.Speed.Get()
					require.True(t, present)
					require.Equal(t, test.want, fmt.Sprint(value))
				}
			})
		}
	}
	for _, field := range []string{"mtu", "speed"} {
		for _, raw := range []string{`"1"`, `true`, `[]`, `{}`, `1.5`, `1e-1`, `9223372036854775808`, `-9223372036854775809`} {
			t.Run("malformed numeric/"+field+"/"+raw, func(t *testing.T) {
				_, err := decodeInterfacePresenceJSON(t, `{"`+field+`":`+raw+`}`)
				require.Error(t, err)
				require.Equal(t, []shared.FieldViolation{{Field: field, Reason: "invalid", Description: "A valid integer is required."}}, shared.ViolationsOf(err))
			})
		}
	}
	for _, body := range []string{`{"device":{"id":7}}`, `{"device":"7"}`, `{"bridge":7}`, `{"count_ipaddresses":1}`} {
		_, err := decodeInterfacePresenceJSON(t, body)
		require.Error(t, err, body)
	}

	t.Run("HTTP mutation mapper and choice projection", func(t *testing.T) {
		spy := &interfaceServiceSpy{networkInterface: restInterfaceFixture(t)}
		router := interfaceRESTRouter(spy)
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
			path, wantStatus := "/api/dcim/interfaces/41/", http.StatusOK
			if method == http.MethodPost {
				path, wantStatus = "/api/dcim/interfaces/", http.StatusCreated
			}
			response := interfaceTemplateRESTRequest(t, router, method, path, `{"mtu":-1,"speed":0,"enabled":false,"duplex":""}`)
			require.Equal(t, wantStatus, response.Code, response.Body.String())
		}
		require.Equal(t, 1, spy.createCalls)
		require.Equal(t, 1, spy.replaceCalls)
		require.Equal(t, 1, spy.updateCalls)
		for _, command := range []applicationdcim.CreateInterfaceCommand{spy.createCommand, spy.replaceCommand.CreateInterfaceCommand} {
			mtu, present := command.MTU.Get()
			require.True(t, present)
			require.Equal(t, "-1", fmt.Sprint(mtu))
			speed, present := command.Speed.Get()
			require.True(t, present)
			require.Equal(t, "0", fmt.Sprint(speed))
		}
		dto := newInterfaceDTO(spy.networkInterface)
		require.Nil(t, dto.Duplex, "stored blank projects to REST null")
		require.Equal(t, "10gbase-sr", dto.Type.Value)
		require.Equal(t, "10GBASE-SR (10GE)", dto.Type.Label)
		for _, duplex := range []domaindcim.DeviceNullable[string]{domaindcim.NullDeviceValue[string](), domaindcim.NonNullDeviceValue("full")} {
			state := spy.networkInterface.State()
			state.Duplex = duplex
			item, err := domaindcim.RestoreInterface(state)
			require.NoError(t, err)
			projected := newInterfaceDTO(item)
			if value, present := duplex.Get(); present {
				require.Equal(t, &rackTypeStringChoiceDTO{Value: value, Label: "Full"}, projected.Duplex)
			} else {
				require.Nil(t, projected.Duplex)
			}
		}
	})
}

func decodeInterfacePresenceJSON(t *testing.T, body string) (decodedInterfaceInput, error) {
	t.Helper()
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Request = httptest.NewRequest(http.MethodPatch, "/api/dcim/interfaces/41/", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	return decodeInterfaceInput(context)
}

func interfaceCreateRESTStates(command applicationdcim.CreateInterfaceCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device": command.Device.State(), "name": command.Name.State(), "label": command.Label.State(),
		"type": command.Type.State(), "enabled": command.Enabled.State(), "mgmt_only": command.MgmtOnly.State(),
		"mtu": command.MTU.State(), "speed": command.Speed.State(), "duplex": command.Duplex.State(), "description": command.Description.State(),
	}
}

func interfaceUpdateRESTStates(command applicationdcim.UpdateInterfaceCommand) map[string]applicationdcim.FieldState {
	return interfaceCreateRESTStates(applicationdcim.CreateInterfaceCommand{
		Device: command.Device, Name: command.Name, Label: command.Label, Type: command.Type,
		Enabled: command.Enabled, MgmtOnly: command.MgmtOnly, MTU: command.MTU, Speed: command.Speed,
		Duplex: command.Duplex, Description: command.Description,
	})
}

func requireInterfaceRESTValue[T comparable](t *testing.T, field applicationdcim.Field[T], want T) {
	t.Helper()
	actual, present := field.Get()
	require.True(t, present)
	require.Equal(t, want, actual)
}
