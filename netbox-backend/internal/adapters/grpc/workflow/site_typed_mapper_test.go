package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestTypedSiteFieldMaskPreservesExplicitPresence(t *testing.T) {
	description := "updated"
	command, err := typedSiteUpdateCommand(
		17,
		&dcimv1.SiteInput{Description: &description},
		&fieldmaskpb.FieldMask{Paths: []string{"description"}},
	)
	require.NoError(t, err)
	require.Equal(t, applicationdcim.FieldPresent, command.Description.State())
	require.Equal(t, applicationdcim.FieldOmitted, command.Name.State())

	_, err = typedSiteUpdateCommand(
		17,
		&dcimv1.SiteInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"description"}},
	)
	require.Error(t, err, "a protobuf FieldMask path without scalar presence must fail closed")
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	require.Equal(t, []shared.FieldViolation{{
		Field:       "update_mask",
		Description: "Every update_mask path must name a supported field with an explicit value.",
	}}, shared.ViolationsOf(err))

	_, err = typedSiteUpdateCommand(
		17,
		&dcimv1.SiteInput{Description: &description},
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
}
