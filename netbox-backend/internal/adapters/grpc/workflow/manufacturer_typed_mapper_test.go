package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestManufacturerGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	value := "  retained by the transport  "
	allPresent := &dcimv1.ManufacturerInput{
		Name: &value, Slug: &value, Description: &value,
	}
	create := typedManufacturerCreateCommand(allPresent)
	requireManufacturerGRPCFields(t, map[string]applicationdcim.Field[string]{
		"name": create.Name, "slug": create.Slug, "description": create.Description,
	}, applicationdcim.FieldPresent, value)

	replace := typedManufacturerReplaceCommand(43, nil)
	require.Equal(t, shared.ID(43), replace.ID)
	requireManufacturerGRPCFields(t, map[string]applicationdcim.Field[string]{
		"name": replace.Name, "slug": replace.Slug, "description": replace.Description,
	}, applicationdcim.FieldOmitted, "")

	setters := map[string]func(*dcimv1.ManufacturerInput, *string){
		"name":        func(input *dcimv1.ManufacturerInput, value *string) { input.Name = value },
		"slug":        func(input *dcimv1.ManufacturerInput, value *string) { input.Slug = value },
		"description": func(input *dcimv1.ManufacturerInput, value *string) { input.Description = value },
	}
	for field, set := range setters {
		field, set := field, set
		t.Run(field, func(t *testing.T) {
			input := &dcimv1.ManufacturerInput{}
			set(input, &value)
			command, err := typedManufacturerUpdateCommand(
				43, input, &fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.NoError(t, err)
			for name, commandField := range manufacturerGRPCUpdateFields(command) {
				if name == field {
					require.Equal(t, applicationdcim.FieldPresent, commandField.State(), name)
					got, present := commandField.Get()
					require.True(t, present, name)
					require.Equal(t, value, got, name)
					continue
				}
				require.Equal(t, applicationdcim.FieldOmitted, commandField.State(), name)
			}

			command, err = typedManufacturerUpdateCommand(
				43,
				&dcimv1.ManufacturerInput{},
				&fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.NoError(t, err, "a supported mask path without proto presence is explicit null")
			for name, commandField := range manufacturerGRPCUpdateFields(command) {
				want := applicationdcim.FieldOmitted
				if name == field {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, commandField.State(), name)
			}
		})
	}

	blank := ""
	withoutMask, err := typedManufacturerUpdateCommand(
		43,
		&dcimv1.ManufacturerInput{Description: &blank},
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, applicationdcim.FieldPresent, withoutMask.Description.State())
	got, present := withoutMask.Description.Get()
	require.True(t, present)
	require.Empty(t, got)
	require.Equal(t, applicationdcim.FieldOmitted, withoutMask.Name.State())

	_, err = typedManufacturerUpdateCommand(
		43,
		&dcimv1.ManufacturerInput{Description: &value},
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	require.Equal(t, []shared.FieldViolation{{
		Field: "update_mask", Description: "Every update_mask path must name a supported field.",
	}}, shared.ViolationsOf(err))
}

func manufacturerGRPCUpdateFields(
	command applicationdcim.UpdateManufacturerCommand,
) map[string]applicationdcim.Field[string] {
	return map[string]applicationdcim.Field[string]{
		"name": command.Name, "slug": command.Slug, "description": command.Description,
	}
}

func requireManufacturerGRPCFields(
	t *testing.T,
	fields map[string]applicationdcim.Field[string],
	wantState applicationdcim.FieldState,
	wantValue string,
) {
	t.Helper()
	for name, field := range fields {
		require.Equal(t, wantState, field.State(), name)
		if wantState != applicationdcim.FieldPresent {
			continue
		}
		got, present := field.Get()
		require.True(t, present, name)
		require.Equal(t, wantValue, got, name)
	}
}
