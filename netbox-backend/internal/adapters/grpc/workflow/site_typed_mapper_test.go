package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestSiteGRPCScalarPresenceMatrix(t *testing.T) {
	value := "  retained by the transport  "
	allPresent := &dcimv1.SiteInput{
		Name: &value, Slug: &value, Status: &value,
		Facility: &value, Description: &value, Comments: &value,
	}
	create := typedSiteCreateCommand(allPresent)
	requireSiteGRPCFields(t, map[string]applicationdcim.Field[string]{
		"name": create.Name, "slug": create.Slug, "status": create.Status,
		"facility": create.Facility, "description": create.Description,
		"comments": create.Comments,
	}, applicationdcim.FieldPresent, value)

	replace := typedSiteReplaceCommand(41, nil)
	require.Equal(t, shared.ID(41), replace.ID)
	requireSiteGRPCFields(t, map[string]applicationdcim.Field[string]{
		"name": replace.Name, "slug": replace.Slug, "status": replace.Status,
		"facility": replace.Facility, "description": replace.Description,
		"comments": replace.Comments,
	}, applicationdcim.FieldOmitted, "")

	setters := map[string]func(*dcimv1.SiteInput, *string){
		"name":        func(input *dcimv1.SiteInput, value *string) { input.Name = value },
		"slug":        func(input *dcimv1.SiteInput, value *string) { input.Slug = value },
		"status":      func(input *dcimv1.SiteInput, value *string) { input.Status = value },
		"facility":    func(input *dcimv1.SiteInput, value *string) { input.Facility = value },
		"description": func(input *dcimv1.SiteInput, value *string) { input.Description = value },
		"comments":    func(input *dcimv1.SiteInput, value *string) { input.Comments = value },
	}
	for field, set := range setters {
		field, set := field, set
		t.Run(field, func(t *testing.T) {
			input := &dcimv1.SiteInput{}
			set(input, &value)
			command, err := typedSiteUpdateCommand(
				41,
				input,
				&fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.NoError(t, err)
			for name, commandField := range siteGRPCUpdateFields(command) {
				if name == field {
					require.Equal(t, applicationdcim.FieldPresent, commandField.State(), name)
					got, present := commandField.Get()
					require.True(t, present, name)
					require.Equal(t, value, got, name)
					continue
				}
				require.Equal(t, applicationdcim.FieldOmitted, commandField.State(), name)
			}

			command, err = typedSiteUpdateCommand(
				41,
				&dcimv1.SiteInput{},
				&fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.NoError(t, err, "a supported mask path without proto presence is explicit null")
			for name, commandField := range siteGRPCUpdateFields(command) {
				want := applicationdcim.FieldOmitted
				if name == field {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, commandField.State(), name)
			}
		})
	}

	blank := ""
	withoutMask, err := typedSiteUpdateCommand(
		41,
		&dcimv1.SiteInput{Status: &blank},
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, applicationdcim.FieldPresent, withoutMask.Status.State())
	got, present := withoutMask.Status.Get()
	require.True(t, present)
	require.Empty(t, got)
	require.Equal(t, applicationdcim.FieldOmitted, withoutMask.Name.State())

	_, err = typedSiteUpdateCommand(
		41,
		&dcimv1.SiteInput{Description: &value},
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	require.Equal(t, []shared.FieldViolation{{
		Field:       "update_mask",
		Description: "Every update_mask path must name a supported field.",
	}}, shared.ViolationsOf(err))
}

func siteGRPCUpdateFields(command applicationdcim.UpdateSiteCommand) map[string]applicationdcim.Field[string] {
	return map[string]applicationdcim.Field[string]{
		"name": command.Name, "slug": command.Slug, "status": command.Status,
		"facility": command.Facility, "description": command.Description,
		"comments": command.Comments,
	}
}

func requireSiteGRPCFields(
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
