package dcim_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestRackTypeNormalizesCompleteStateAndExposesPhysicalAttributes(t *testing.T) {
	t.Parallel()

	rackType, err := dcim.NewRackType(dcim.RackTypeValues{
		Manufacturer: manufacturerReference(t, 7, "  Acme  ", "acme"),
		Model:        "  R42  ", Slug: "  r42  ", FormFactor: "4-post-cabinet",
		Width: 19, UHeight: 42, StartingUnit: 1, DescUnits: true,
		Description: "  Standard rack  ", Comments: "  Notes  ",
	}, testTime)
	require.NoError(t, err)

	assert.Equal(t, "R42", rackType.Model())
	assert.Equal(t, "r42", rackType.Slug().String())
	assert.Equal(t, "Acme R42", rackType.FullName())
	assert.Equal(t, dcim.RackPhysicalAttributes{
		FormFactor: dcim.RackFormFactorFourPostCabinet,
		Width:      dcim.RackWidth19, UHeight: 42, StartingUnit: 1, DescUnits: true,
	}, rackType.PhysicalAttributes())
	assert.Equal(t, dcim.RackTypeSnapshot{
		ManufacturerID: 7, Model: "R42", Slug: "r42", FormFactor: "4-post-cabinet",
		Width: 19, UHeight: 42, StartingUnit: 1, DescUnits: true,
		Description: "Standard rack", Comments: "Notes",
	}, rackType.Snapshot())
}

func TestRackTypeScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	for _, formFactor := range []dcim.RackFormFactor{
		dcim.RackFormFactorTwoPostFrame,
		dcim.RackFormFactorFourPostFrame,
		dcim.RackFormFactorFourPostCabinet,
		dcim.RackFormFactorWallFrame,
		dcim.RackFormFactorWallFrameVertical,
		dcim.RackFormFactorWallCabinet,
		dcim.RackFormFactorWallCabinetVertical,
	} {
		formFactor := formFactor
		t.Run(formFactor.String(), func(t *testing.T) {
			t.Parallel()
			rackType, err := dcim.NewRackType(dcim.RackTypeValues{
				Manufacturer: manufacturerReference(t, 7, "Acme", "acme"),
				Model:        "  R42  ", Slug: "  r42  ", FormFactor: formFactor.String(),
				Width: 19, UHeight: 42, StartingUnit: 1,
				Description: "  Standard rack  ", Comments: "  Notes  ",
			}, testTime)
			require.NoError(t, err)
			assert.Equal(t, "R42", rackType.Model())
			assert.Equal(t, "r42", rackType.Slug().String())
			assert.Equal(t, formFactor, rackType.FormFactor())
			assert.Equal(t, "Standard rack", rackType.Description())
			assert.Equal(t, "Notes", rackType.Comments())
		})
	}

	for _, test := range []struct {
		name        string
		formFactor  string
		reason      string
		description string
	}{
		{
			name: "blank is rejected by the nonblank model contract", formFactor: "",
			reason: "required", description: "This field may not be blank.",
		},
		{
			name: "choice whitespace is not trimmed", formFactor: " 4-post-cabinet ",
			reason: "invalid_choice", description: "Select a valid choice.",
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.NewRackType(dcim.RackTypeValues{
				Manufacturer: manufacturerReference(t, 7, "Acme", "acme"),
				Model:        "R42", Slug: "r42", FormFactor: test.formFactor,
				Width: 19, UHeight: 42, StartingUnit: 1,
			}, testTime)
			require.Error(t, err)
			assert.Equal(t, []shared.FieldViolation{{
				Field: "form_factor", Reason: test.reason, Description: test.description,
			}}, shared.ViolationsOf(err))
		})
	}
}

func TestRackTypeReturnsEveryProfileValidationViolation(t *testing.T) {
	t.Parallel()

	_, err := dcim.NewRackType(dcim.RackTypeValues{
		Model: strings.Repeat("m", dcim.RackTypeModelMaxLength+1),
		Slug:  "invalid slug!", FormFactor: "unknown", Width: 20,
		UHeight:      dcim.RackTypeMaximumUHeight + 1,
		StartingUnit: dcim.RackTypeMaximumStartingUnit + 1,
		Description:  strings.Repeat("d", dcim.RackTypeDescriptionMaxLength+1),
	}, testTime)
	require.Error(t, err)
	assert.Equal(t, map[string]string{
		"manufacturer": "invalid_choice", "model": "max_length", "slug": "invalid",
		"form_factor": "invalid_choice", "width": "invalid_choice", "u_height": "range",
		"starting_unit": "range", "description": "max_length",
	}, violationReasons(err))
}

func TestRackTypePatchPreservesOmittedFieldsAndIsAtomic(t *testing.T) {
	t.Parallel()

	rackType := newRackType(t)
	require.NoError(t, rackType.AssignID(23))
	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	width := uint32(23)
	comments := ""
	require.NoError(t, rackType.ApplyPatch(dcim.RackTypePatch{
		Width: &width, Comments: &comments,
	}, updatedAt))
	assert.Equal(t, dcim.RackWidth23, rackType.Width())
	assert.Empty(t, rackType.Comments())
	assert.Equal(t, "Original Rack", rackType.Model())
	assert.Equal(t, testTime, rackType.Created())
	assert.Equal(t, updatedAt, rackType.LastUpdated())

	invalidHeight := uint32(0)
	err := rackType.ApplyPatch(dcim.RackTypePatch{UHeight: &invalidHeight}, updatedAt)
	require.Error(t, err)
	assert.Equal(t, uint32(42), rackType.UHeight(), "a rejected patch must not partially mutate state")
}

func TestRackTypeRejectsEmptyPatchAndInvalidRestore(t *testing.T) {
	t.Parallel()

	rackType := newRackType(t)
	err := rackType.ApplyPatch(dcim.RackTypePatch{}, testTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)

	_, err = dcim.RestoreRackType(dcim.RackTypeState{
		ID: 5, Manufacturer: manufacturerReference(t, 1, "Acme", "acme"),
		Model: "Invalid", Slug: "invalid", FormFactor: "bad", Width: 19,
		UHeight: 42, StartingUnit: 1, Created: testTime, LastUpdated: testTime,
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func newRackType(t *testing.T) *dcim.RackType {
	t.Helper()
	rackType, err := dcim.NewRackType(dcim.RackTypeValues{
		Manufacturer: manufacturerReference(t, 1, "Original Vendor", "original-vendor"),
		Model:        "Original Rack", Slug: "original-rack", FormFactor: "4-post-cabinet",
		Width: 19, UHeight: 42, StartingUnit: 1,
		Description: "Original description", Comments: "Original comments",
	}, testTime)
	require.NoError(t, err)
	return rackType
}

func manufacturerReference(t *testing.T, id shared.ID, name, slug string) dcim.ManufacturerReference {
	t.Helper()
	reference, err := dcim.NewManufacturerReference(id, name, slug)
	require.NoError(t, err)
	return reference
}
