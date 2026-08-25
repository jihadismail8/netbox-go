package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestInterfaceTemplateGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	deviceType := int64(73)
	name := "  Ethernet1  "
	label := ""
	interfaceType := " bridge "
	falseValue := false
	description := ""
	allPresent := &dcimv1.InterfaceTemplateInput{
		DeviceType: &deviceType, Name: &name, Label: &label, Type: &interfaceType,
		Enabled: &falseValue, MgmtOnly: &falseValue, Description: &description,
	}

	omitted := typedInterfaceTemplateCreateCommand(nil)
	requireInterfaceTemplateGRPCStates(
		t, interfaceTemplateCreateGRPCStates(omitted), applicationdcim.FieldOmitted,
	)

	create := typedInterfaceTemplateCreateCommand(allPresent)
	requireInterfaceTemplateGRPCStates(
		t, interfaceTemplateCreateGRPCStates(create), applicationdcim.FieldPresent,
	)
	requireInterfaceTemplateGRPCFieldValue(t, create.DeviceType, shared.ID(73), "device_type")
	requireInterfaceTemplateGRPCFieldValue(t, create.Name, name, "name")
	requireInterfaceTemplateGRPCFieldValue(t, create.Label, "", "label")
	requireInterfaceTemplateGRPCFieldValue(t, create.Type, interfaceType, "type")
	requireInterfaceTemplateGRPCFieldValue(t, create.Enabled, false, "enabled")
	requireInterfaceTemplateGRPCFieldValue(t, create.MgmtOnly, false, "mgmt_only")
	requireInterfaceTemplateGRPCFieldValue(t, create.Description, "", "description")

	replace := applicationdcim.ReplaceInterfaceTemplateCommand{
		ID: 47, CreateInterfaceTemplateCommand: typedInterfaceTemplateCreateCommand(allPresent),
	}
	require.Equal(t, shared.ID(47), replace.ID)
	requireInterfaceTemplateGRPCStates(
		t,
		interfaceTemplateCreateGRPCStates(replace.CreateInterfaceTemplateCommand),
		applicationdcim.FieldPresent,
	)

	paths := []string{
		"device_type", "name", "label", "type", "enabled", "mgmt_only", "description",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			command, err := typedInterfaceTemplateUpdateCommand(
				47,
				allPresent,
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err)
			require.Equal(t, shared.ID(47), command.ID)
			for name, state := range interfaceTemplateUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, name)
			}

			command, err = typedInterfaceTemplateUpdateCommand(
				47,
				&dcimv1.InterfaceTemplateInput{},
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err, "a supported mask path without proto presence is explicit null")
			for name, state := range interfaceTemplateUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, state, name)
			}
		})
	}

	withoutMask, err := typedInterfaceTemplateUpdateCommand(47, allPresent, nil)
	require.NoError(t, err)
	requireInterfaceTemplateGRPCStates(
		t, interfaceTemplateUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent,
	)
	requireInterfaceTemplateGRPCFieldValue(t, withoutMask.Enabled, false, "enabled")
	requireInterfaceTemplateGRPCFieldValue(t, withoutMask.MgmtOnly, false, "mgmt_only")

	withoutMask, err = typedInterfaceTemplateUpdateCommand(
		47, allPresent, &fieldmaskpb.FieldMask{},
	)
	require.NoError(t, err)
	requireInterfaceTemplateGRPCStates(
		t, interfaceTemplateUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent,
	)

	_, err = typedInterfaceTemplateUpdateCommand(
		47,
		allPresent,
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	require.Equal(t, []shared.FieldViolation{{
		Field: "update_mask", Description: "Every update_mask path must name a supported field.",
	}}, shared.ViolationsOf(err))
}

func interfaceTemplateCreateGRPCStates(
	command applicationdcim.CreateInterfaceTemplateCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(), "name": command.Name.State(),
		"label": command.Label.State(), "type": command.Type.State(),
		"enabled": command.Enabled.State(), "mgmt_only": command.MgmtOnly.State(),
		"description": command.Description.State(),
	}
}

func interfaceTemplateUpdateGRPCStates(
	command applicationdcim.UpdateInterfaceTemplateCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"device_type": command.DeviceType.State(), "name": command.Name.State(),
		"label": command.Label.State(), "type": command.Type.State(),
		"enabled": command.Enabled.State(), "mgmt_only": command.MgmtOnly.State(),
		"description": command.Description.State(),
	}
}

func requireInterfaceTemplateGRPCStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireInterfaceTemplateGRPCFieldValue[T comparable](
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
