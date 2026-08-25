package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceTypeGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	manufacturer := int64(73)
	model := "  retained model  "
	slug := "  retained-slug  "
	partNumber := ""
	zeroHeight := "0"
	falseValue := false
	airflow := ""
	description := ""
	comments := ""
	allPresent := &dcimv1.DeviceTypeInput{
		Manufacturer: &manufacturer, Model: &model, Slug: &slug,
		PartNumber: &partNumber, UHeight: &zeroHeight,
		ExcludeFromUtilization: &falseValue, IsFullDepth: &falseValue,
		Airflow: &airflow, Description: &description, Comments: &comments,
	}

	omitted := typedDeviceTypeCreateCommand(nil)
	requireDeviceTypeGRPCStates(
		t, deviceTypeCreateGRPCStates(omitted), applicationdcim.FieldOmitted,
	)

	create := typedDeviceTypeCreateCommand(allPresent)
	requireDeviceTypeGRPCStates(
		t, deviceTypeCreateGRPCStates(create), applicationdcim.FieldPresent,
	)
	requireDeviceTypeGRPCFieldValue(t, create.Manufacturer, shared.ID(73), "manufacturer")
	requireDeviceTypeGRPCFieldValue(t, create.Model, model, "model")
	requireDeviceTypeGRPCFieldValue(t, create.Slug, slug, "slug")
	requireDeviceTypeGRPCFieldValue(t, create.PartNumber, "", "part_number")
	requireDeviceTypeGRPCFieldValue(t, create.UHeight, "0", "u_height")
	requireDeviceTypeGRPCFieldValue(
		t, create.ExcludeFromUtilization, false, "exclude_from_utilization",
	)
	requireDeviceTypeGRPCFieldValue(t, create.IsFullDepth, false, "is_full_depth")
	requireDeviceTypeGRPCFieldValue(t, create.Airflow, "", "airflow")
	requireDeviceTypeGRPCFieldValue(t, create.Description, "", "description")
	requireDeviceTypeGRPCFieldValue(t, create.Comments, "", "comments")

	replace := typedDeviceTypeReplaceCommand(47, nil)
	require.Equal(t, shared.ID(47), replace.ID)
	requireDeviceTypeGRPCStates(
		t,
		deviceTypeCreateGRPCStates(replace.CreateDeviceTypeCommand),
		applicationdcim.FieldOmitted,
	)

	paths := []string{
		"manufacturer", "model", "slug", "part_number", "u_height",
		"exclude_from_utilization", "is_full_depth", "airflow", "description", "comments",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			command, err := typedDeviceTypeUpdateCommand(
				47,
				allPresent,
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err)
			require.Equal(t, shared.ID(47), command.ID)
			for name, state := range deviceTypeUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, name)
			}

			command, err = typedDeviceTypeUpdateCommand(
				47,
				&dcimv1.DeviceTypeInput{},
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err, "a supported mask path without proto presence is explicit null")
			for name, state := range deviceTypeUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, state, name)
			}
		})
	}

	withoutMask, err := typedDeviceTypeUpdateCommand(47, allPresent, nil)
	require.NoError(t, err)
	requireDeviceTypeGRPCStates(
		t, deviceTypeUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent,
	)
	requireDeviceTypeGRPCFieldValue(t, withoutMask.UHeight, "0", "u_height")
	requireDeviceTypeGRPCFieldValue(
		t, withoutMask.ExcludeFromUtilization, false, "exclude_from_utilization",
	)
	requireDeviceTypeGRPCFieldValue(t, withoutMask.IsFullDepth, false, "is_full_depth")
	requireDeviceTypeGRPCFieldValue(t, withoutMask.Airflow, "", "airflow")

	_, err = typedDeviceTypeUpdateCommand(
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

func deviceTypeCreateGRPCStates(
	command applicationdcim.CreateDeviceTypeCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "part_number": command.PartNumber.State(),
		"u_height":                 command.UHeight.State(),
		"exclude_from_utilization": command.ExcludeFromUtilization.State(),
		"is_full_depth":            command.IsFullDepth.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func deviceTypeUpdateGRPCStates(
	command applicationdcim.UpdateDeviceTypeCommand,
) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "part_number": command.PartNumber.State(),
		"u_height":                 command.UHeight.State(),
		"exclude_from_utilization": command.ExcludeFromUtilization.State(),
		"is_full_depth":            command.IsFullDepth.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func requireDeviceTypeGRPCStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireDeviceTypeGRPCFieldValue[T comparable](
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
