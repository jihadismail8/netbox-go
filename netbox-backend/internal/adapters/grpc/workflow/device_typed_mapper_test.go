package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceGRPCScalarPresenceMatrix(t *testing.T) {
	t.Run("nil and empty inputs preserve omission", func(t *testing.T) {
		for _, command := range []applicationdcim.CreateDeviceCommand{
			typedDeviceCreateCommand(nil),
			typedDeviceCreateCommand(&dcimv1.DeviceInput{}),
			typedDeviceReplaceCommand(79, nil).CreateDeviceCommand,
		} {
			requireDeviceGRPCStates(t, deviceCreateGRPCStates(command), applicationdcim.FieldOmitted)
		}
	})

	t.Run("blank zero and concrete values remain present", func(t *testing.T) {
		zero := int64(0)
		status, serial, airflow, description, comments := "active", "", "", "  description  ", ""
		input := &dcimv1.DeviceInput{
			DeviceType:  &zero,
			Role:        &zero,
			Name:        wrapperspb.String(""),
			Site:        &zero,
			Rack:        wrapperspb.Int64(0),
			Position:    wrapperspb.String("0"),
			Face:        wrapperspb.String(""),
			Status:      &status,
			Serial:      &serial,
			AssetTag:    wrapperspb.String(""),
			Airflow:     &airflow,
			Description: &description,
			Comments:    &comments,
		}
		create := typedDeviceCreateCommand(input)
		requireDeviceGRPCStates(t, deviceCreateGRPCStates(create), applicationdcim.FieldPresent)
		requireDeviceGRPCFieldValue(t, create.DeviceType, shared.ID(0), "device_type")
		requireDeviceGRPCFieldValue(t, create.Role, shared.ID(0), "role")
		requireDeviceGRPCFieldValue(t, create.Name, "", "name")
		requireDeviceGRPCFieldValue(t, create.Site, shared.ID(0), "site")
		requireDeviceGRPCFieldValue(t, create.Rack, shared.ID(0), "rack")
		requireDeviceGRPCFieldValue(t, create.Position, "0", "position")
		requireDeviceGRPCFieldValue(t, create.Face, "", "face")
		requireDeviceGRPCFieldValue(t, create.Status, "active", "status")
		requireDeviceGRPCFieldValue(t, create.Serial, "", "serial")
		requireDeviceGRPCFieldValue(t, create.AssetTag, "", "asset_tag")
		requireDeviceGRPCFieldValue(t, create.Airflow, "", "airflow")
		requireDeviceGRPCFieldValue(t, create.Description, "  description  ", "description")
		requireDeviceGRPCFieldValue(t, create.Comments, "", "comments")

		requireDeviceGRPCStates(
			t,
			deviceCreateGRPCStates(typedDeviceReplaceCommand(79, input).CreateDeviceCommand),
			applicationdcim.FieldPresent,
		)
	})

	t.Run("one nullable masked absence at a time is an explicit clear", func(t *testing.T) {
		for _, field := range []string{"name", "rack", "position", "face", "asset_tag", "airflow"} {
			field := field
			t.Run(field, func(t *testing.T) {
				command, err := typedDeviceUpdateCommand(
					79,
					&dcimv1.DeviceInput{},
					&fieldmaskpb.FieldMask{Paths: []string{field}},
				)
				require.NoError(t, err)
				states := deviceUpdateGRPCStates(command)
				for name, state := range states {
					want := applicationdcim.FieldOmitted
					if name == field {
						want = applicationdcim.FieldNull
					}
					require.Equal(t, want, state, name)
				}
			})
		}
	})

	t.Run("nonnullable masked absence and unknown paths fail closed", func(t *testing.T) {
		for _, field := range []string{
			"device_type", "role", "site", "status", "serial", "description", "comments", "unknown",
		} {
			_, err := typedDeviceUpdateCommand(
				79,
				&dcimv1.DeviceInput{},
				&fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.Error(t, err, field)
		}
	})
}

func deviceCreateGRPCStates(command applicationdcim.CreateDeviceCommand) map[string]applicationdcim.FieldState {
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

func deviceUpdateGRPCStates(command applicationdcim.UpdateDeviceCommand) map[string]applicationdcim.FieldState {
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

func requireDeviceGRPCStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for field, state := range states {
		require.Equal(t, want, state, field)
	}
}

func requireDeviceGRPCFieldValue[T comparable](
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
