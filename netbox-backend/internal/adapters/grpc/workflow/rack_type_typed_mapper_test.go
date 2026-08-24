package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestRackTypeGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	manufacturer := int64(73)
	model := "  retained model  "
	slug := "  retained-slug  "
	formFactor := "  retained factor  "
	zero := uint32(0)
	descUnits := false
	description := ""
	comments := ""
	allPresent := &dcimv1.RackTypeInput{
		Manufacturer: &manufacturer, Model: &model, Slug: &slug, FormFactor: &formFactor,
		Width: &zero, UHeight: &zero, StartingUnit: &zero, DescUnits: &descUnits,
		Description: &description, Comments: &comments,
	}

	omitted := typedRackTypeCreateCommand(nil)
	requireRackTypeGRPCStates(t, rackTypeCreateGRPCStates(omitted), applicationdcim.FieldOmitted)

	create := typedRackTypeCreateCommand(allPresent)
	requireRackTypeGRPCStates(t, rackTypeCreateGRPCStates(create), applicationdcim.FieldPresent)
	requireRackTypeGRPCFieldValue(t, create.Manufacturer, shared.ID(73), "manufacturer")
	requireRackTypeGRPCFieldValue(t, create.Model, model, "model")
	requireRackTypeGRPCFieldValue(t, create.Slug, slug, "slug")
	requireRackTypeGRPCFieldValue(t, create.FormFactor, formFactor, "form_factor")
	requireRackTypeGRPCFieldValue(t, create.Width, zero, "width")
	requireRackTypeGRPCFieldValue(t, create.UHeight, zero, "u_height")
	requireRackTypeGRPCFieldValue(t, create.StartingUnit, zero, "starting_unit")
	requireRackTypeGRPCFieldValue(t, create.DescUnits, false, "desc_units")
	requireRackTypeGRPCFieldValue(t, create.Description, "", "description")
	requireRackTypeGRPCFieldValue(t, create.Comments, "", "comments")

	replace := typedRackTypeReplaceCommand(47, nil)
	require.Equal(t, shared.ID(47), replace.ID)
	requireRackTypeGRPCStates(
		t,
		rackTypeCreateGRPCStates(replace.CreateRackTypeCommand),
		applicationdcim.FieldOmitted,
	)

	paths := []string{
		"manufacturer", "model", "slug", "form_factor", "width", "u_height",
		"starting_unit", "desc_units", "description", "comments",
	}
	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			command, err := typedRackTypeUpdateCommand(
				47,
				allPresent,
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err)
			require.Equal(t, shared.ID(47), command.ID)
			for name, state := range rackTypeUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, name)
			}

			command, err = typedRackTypeUpdateCommand(
				47,
				&dcimv1.RackTypeInput{},
				&fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err, "a supported mask path without proto presence is explicit null")
			for name, state := range rackTypeUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if name == path {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, state, name)
			}
		})
	}

	withoutMask, err := typedRackTypeUpdateCommand(47, allPresent, nil)
	require.NoError(t, err)
	requireRackTypeGRPCStates(t, rackTypeUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent)
	requireRackTypeGRPCFieldValue(t, withoutMask.Width, uint32(0), "width")
	requireRackTypeGRPCFieldValue(t, withoutMask.DescUnits, false, "desc_units")
	requireRackTypeGRPCFieldValue(t, withoutMask.Description, "", "description")

	_, err = typedRackTypeUpdateCommand(
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

func rackTypeCreateGRPCStates(command applicationdcim.CreateRackTypeCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "form_factor": command.FormFactor.State(),
		"width": command.Width.State(), "u_height": command.UHeight.State(),
		"starting_unit": command.StartingUnit.State(), "desc_units": command.DescUnits.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func rackTypeUpdateGRPCStates(command applicationdcim.UpdateRackTypeCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"manufacturer": command.Manufacturer.State(), "model": command.Model.State(),
		"slug": command.Slug.State(), "form_factor": command.FormFactor.State(),
		"width": command.Width.State(), "u_height": command.UHeight.State(),
		"starting_unit": command.StartingUnit.State(), "desc_units": command.DescUnits.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func requireRackTypeGRPCStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for name, state := range states {
		require.Equal(t, want, state, name)
	}
}

func requireRackTypeGRPCFieldValue[T comparable](
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
