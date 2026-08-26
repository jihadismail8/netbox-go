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

func TestRackGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	site, rackType, role := int64(3), int64(8), int64(9)
	name, facilityID := "  A01  ", ""
	status, serial, assetTag := " active ", "", ""
	formFactor := " wall-frame "
	width, uHeight, startingUnit := uint32(0), uint32(0), uint32(0)
	descUnits := false
	airflow, description, comments := " front-to-rear ", "  description  ", ""
	allPresent := &dcimv1.RackInput{
		Site: &site, Name: &name, FacilityId: &facilityID,
		RackType: wrapperspb.Int64(rackType), Status: &status, Role: wrapperspb.Int64(role),
		Serial: &serial, AssetTag: &assetTag, FormFactor: &formFactor,
		Width: &width, UHeight: &uHeight, StartingUnit: &startingUnit,
		DescUnits: &descUnits, Airflow: &airflow,
		Description: &description, Comments: &comments,
	}

	omitted := typedRackCreateCommand(nil)
	requireRackGRPCStates(t, rackCreateGRPCStates(omitted), applicationdcim.FieldOmitted)

	create := typedRackCreateCommand(allPresent)
	requireRackGRPCStates(t, rackCreateGRPCStates(create), applicationdcim.FieldPresent)
	requireRackGRPCFieldValue(t, create.Site, shared.ID(3), "site")
	requireRackGRPCFieldValue(t, create.Name, name, "name")
	requireRackGRPCFieldValue(t, create.FacilityID, "", "facility_id")
	requireRackGRPCFieldValue(t, create.RackType, shared.ID(8), "rack_type")
	requireRackGRPCFieldValue(t, create.Status, status, "status")
	requireRackGRPCFieldValue(t, create.Role, shared.ID(9), "role")
	requireRackGRPCFieldValue(t, create.FormFactor, formFactor, "form_factor")
	requireRackGRPCFieldValue(t, create.Width, uint32(0), "width")
	requireRackGRPCFieldValue(t, create.UHeight, uint32(0), "u_height")
	requireRackGRPCFieldValue(t, create.StartingUnit, uint32(0), "starting_unit")
	requireRackGRPCFieldValue(t, create.DescUnits, false, "desc_units")
	requireRackGRPCFieldValue(t, create.Airflow, airflow, "airflow")

	replace := typedRackReplaceCommand(47, allPresent)
	require.Equal(t, shared.ID(47), replace.ID)
	requireRackGRPCStates(
		t, rackCreateGRPCStates(replace.CreateRackCommand), applicationdcim.FieldPresent,
	)

	for _, path := range rackGRPCScalarFields() {
		path := path
		t.Run("single masked value/"+path, func(t *testing.T) {
			command, err := typedRackUpdateCommand(
				47, allPresent, &fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err)
			require.Equal(t, shared.ID(47), command.ID)
			for field, state := range rackUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if field == path {
					want = applicationdcim.FieldPresent
				}
				require.Equal(t, want, state, field)
			}
		})
	}

	for _, path := range rackGRPCScalarFields() {
		path := path
		t.Run("masked absent maps to null/"+path, func(t *testing.T) {
			command, err := typedRackUpdateCommand(
				47, &dcimv1.RackInput{}, &fieldmaskpb.FieldMask{Paths: []string{path}},
			)
			require.NoError(t, err)
			for field, state := range rackUpdateGRPCStates(command) {
				want := applicationdcim.FieldOmitted
				if field == path {
					want = applicationdcim.FieldNull
				}
				require.Equal(t, want, state, field)
			}
		})
	}

	blank := ""
	blankAirflow, err := typedRackUpdateCommand(
		47, &dcimv1.RackInput{Airflow: &blank},
		&fieldmaskpb.FieldMask{Paths: []string{"airflow"}},
	)
	require.NoError(t, err)
	require.Equal(t, applicationdcim.FieldPresent, blankAirflow.Airflow.State())
	requireRackGRPCFieldValue(t, blankAirflow.Airflow, "", "airflow")

	withoutMask, err := typedRackUpdateCommand(47, allPresent, nil)
	require.NoError(t, err)
	requireRackGRPCStates(t, rackUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent)
	requireRackGRPCFieldValue(t, withoutMask.DescUnits, false, "desc_units")
	requireRackGRPCFieldValue(t, withoutMask.Width, uint32(0), "width")

	withoutMask, err = typedRackUpdateCommand(47, allPresent, &fieldmaskpb.FieldMask{})
	require.NoError(t, err)
	requireRackGRPCStates(t, rackUpdateGRPCStates(withoutMask), applicationdcim.FieldPresent)

	_, err = typedRackUpdateCommand(
		47, allPresent, &fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
	require.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
}

func rackGRPCScalarFields() []string {
	return []string{
		"site", "name", "facility_id", "rack_type", "status", "role", "serial",
		"asset_tag", "form_factor", "width", "u_height", "starting_unit",
		"desc_units", "airflow", "description", "comments",
	}
}

func rackCreateGRPCStates(command applicationdcim.CreateRackCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"site": command.Site.State(), "name": command.Name.State(),
		"facility_id": command.FacilityID.State(), "rack_type": command.RackType.State(),
		"status": command.Status.State(), "role": command.Role.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"form_factor": command.FormFactor.State(), "width": command.Width.State(),
		"u_height": command.UHeight.State(), "starting_unit": command.StartingUnit.State(),
		"desc_units": command.DescUnits.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func rackUpdateGRPCStates(command applicationdcim.UpdateRackCommand) map[string]applicationdcim.FieldState {
	return map[string]applicationdcim.FieldState{
		"site": command.Site.State(), "name": command.Name.State(),
		"facility_id": command.FacilityID.State(), "rack_type": command.RackType.State(),
		"status": command.Status.State(), "role": command.Role.State(),
		"serial": command.Serial.State(), "asset_tag": command.AssetTag.State(),
		"form_factor": command.FormFactor.State(), "width": command.Width.State(),
		"u_height": command.UHeight.State(), "starting_unit": command.StartingUnit.State(),
		"desc_units": command.DescUnits.State(), "airflow": command.Airflow.State(),
		"description": command.Description.State(), "comments": command.Comments.State(),
	}
}

func requireRackGRPCStates(
	t *testing.T,
	states map[string]applicationdcim.FieldState,
	want applicationdcim.FieldState,
) {
	t.Helper()
	for field, state := range states {
		require.Equal(t, want, state, field)
	}
}

func requireRackGRPCFieldValue[T comparable](
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
