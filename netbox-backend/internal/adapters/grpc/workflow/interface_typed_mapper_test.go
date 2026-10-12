package workflow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceGRPCScalarPresenceMatrix(t *testing.T) {
	for _, input := range []*dcimv1.InterfaceInput{nil, {}} {
		command := typedInterfaceCreateCommand(input)
		for field, state := range interfaceCreateGRPCStates(command) {
			require.Equal(t, applicationdcim.FieldOmitted, state, field)
		}
	}
	t.Run("concrete false zero and blank remain present", func(t *testing.T) {
		zero, blank, interfaceType, description, disabled := int64(0), "", " other ", "  text  ", false
		input := &dcimv1.InterfaceInput{
			Device: &zero, Name: &blank, Label: &blank, Type: &interfaceType,
			Enabled: &disabled, MgmtOnly: &disabled, Mtu: wrapperspb.Int32(0), Speed: wrapperspb.Int64(0),
			Duplex: wrapperspb.String(""), Description: &description,
		}
		create := typedInterfaceCreateCommand(input)
		for field, state := range interfaceCreateGRPCStates(create) {
			require.Equal(t, applicationdcim.FieldPresent, state, field)
		}
		requireInterfaceGRPCValue(t, create.Device, shared.ID(0))
		requireInterfaceGRPCValue(t, create.Name, "")
		requireInterfaceGRPCValue(t, create.Label, "")
		requireInterfaceGRPCValue(t, create.Type, " other ")
		requireInterfaceGRPCValue(t, create.Enabled, false)
		requireInterfaceGRPCValue(t, create.MgmtOnly, false)
		requireInterfaceGRPCValue(t, create.Duplex, "")
		requireInterfaceGRPCValue(t, create.Description, "  text  ")
		update, err := typedInterfaceUpdateCommand(41, input, nil)
		require.NoError(t, err)
		require.Equal(t, interfaceCreateGRPCStates(create), interfaceUpdateGRPCStates(update))
	})
	t.Run("signed MTU wrapper retains negative intent", func(t *testing.T) {
		command := typedInterfaceCreateCommand(&dcimv1.InterfaceInput{Mtu: wrapperspb.Int32(-1), Speed: wrapperspb.Int64(-1)})
		mtu, present := command.MTU.Get()
		require.True(t, present)
		require.Equal(t, "-1", fmt.Sprint(mtu))
	})
	t.Run("signed speed wrapper retains negative intent", func(t *testing.T) {
		command := typedInterfaceCreateCommand(&dcimv1.InterfaceInput{Speed: wrapperspb.Int64(-1)})
		speed, present := command.Speed.Get()
		require.True(t, present)
		require.Equal(t, "-1", fmt.Sprint(speed))
	})
	t.Run("numeric wrappers retain both sides of semantic bounds", func(t *testing.T) {
		for _, value := range []int32{0, 1, 65536, 65537} {
			command := typedInterfaceCreateCommand(&dcimv1.InterfaceInput{Mtu: wrapperspb.Int32(value)})
			requireInterfaceGRPCValue(t, command.MTU, int64(value))
		}
		for _, value := range []int64{0, 2147483647, 2147483648, -9223372036854775808, 9223372036854775807} {
			command := typedInterfaceCreateCommand(&dcimv1.InterfaceInput{Speed: wrapperspb.Int64(value)})
			requireInterfaceGRPCValue(t, command.Speed, value)
		}
	})
	for _, field := range []string{"device", "name", "label", "type", "enabled", "mgmt_only", "mtu", "speed", "duplex", "description"} {
		t.Run("known masked absent becomes null/"+field, func(t *testing.T) {
			command, err := typedInterfaceUpdateCommand(41, nil, &fieldmaskpb.FieldMask{Paths: []string{field}})
			require.NoError(t, err)
			for name, state := range interfaceUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == field {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, state, name)
			}
		})
	}
	t.Run("mask excludes unselected values and unknown paths fail closed", func(t *testing.T) {
		text := "keep outside mask"
		command, err := typedInterfaceUpdateCommand(41, &dcimv1.InterfaceInput{Name: &text, Mtu: wrapperspb.Int32(1)}, &fieldmaskpb.FieldMask{Paths: []string{"mtu"}})
		require.NoError(t, err)
		require.Equal(t, applicationdcim.FieldOmitted, command.Name.State())
		for _, path := range []string{"unknown", "device.id", "count_ipaddresses"} {
			_, err := typedInterfaceUpdateCommand(41, nil, &fieldmaskpb.FieldMask{Paths: []string{path}})
			require.Error(t, err)
			require.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)
		}
	})
	t.Run("projection preserves present blank duplex", func(t *testing.T) {
		result := typedInterfaceProto(grpcInterfaceFixture(t))
		require.NotNil(t, result.Duplex)
		require.Equal(t, "", result.Duplex.Value)
		require.Nil(t, result.Mtu)
		require.Equal(t, int64(2147483647), result.Speed.Value)
		for _, duplex := range []domaindcim.DeviceNullable[string]{domaindcim.NullDeviceValue[string](), domaindcim.NonNullDeviceValue("full")} {
			state := grpcInterfaceFixture(t).State()
			state.Duplex = duplex
			item, err := domaindcim.RestoreInterface(state)
			require.NoError(t, err)
			projected := typedInterfaceProto(item)
			if value, present := duplex.Get(); present {
				require.NotNil(t, projected.Duplex)
				require.Equal(t, value, projected.Duplex.Value)
			} else {
				require.Nil(t, projected.Duplex)
			}
		}
	})
}

func interfaceCreateGRPCStates(command applicationdcim.CreateInterfaceCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device": command.Device.State(), "name": command.Name.State(), "label": command.Label.State(),
		"type": command.Type.State(), "enabled": command.Enabled.State(), "mgmt_only": command.MgmtOnly.State(),
		"mtu": command.MTU.State(), "speed": command.Speed.State(), "duplex": command.Duplex.State(), "description": command.Description.State(),
	}
}

func interfaceUpdateGRPCStates(command applicationdcim.UpdateInterfaceCommand) map[string]applicationdcim.FieldState {
	return interfaceCreateGRPCStates(applicationdcim.CreateInterfaceCommand{
		Device: command.Device, Name: command.Name, Label: command.Label, Type: command.Type,
		Enabled: command.Enabled, MgmtOnly: command.MgmtOnly, MTU: command.MTU, Speed: command.Speed,
		Duplex: command.Duplex, Description: command.Description,
	})
}

func requireInterfaceGRPCValue[T comparable](t *testing.T, field applicationdcim.Field[T], want T) {
	t.Helper()
	actual, present := field.Get()
	require.True(t, present)
	require.Equal(t, want, actual)
}
