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

func TestRackScalarNormalizationContract(t *testing.T) {
	testRackScalarNormalizationContract(t)
}

func TestRackNormalizesStateAndPreservesNullableBlankValues(t *testing.T) {
	testRackScalarNormalizationContract(t)
}

func testRackScalarNormalizationContract(t *testing.T) {
	t.Helper()
	t.Parallel()

	rack, err := dcim.NewRack(dcim.RackValues{
		Site:         rackSiteReference(t, 3, "  Moscow  ", "moscow"),
		Name:         "  A01  ",
		FacilityID:   dcim.NullRackValue[string](),
		Status:       "active",
		Serial:       "  serial  ",
		AssetTag:     dcim.NonNullRackValue("  "),
		FormFactor:   dcim.NonNullRackValue(""),
		Width:        19,
		UHeight:      42,
		StartingUnit: 1,
		Airflow:      dcim.NonNullRackValue(""),
		Description:  "  description  ",
		Comments:     "  comments  ",
	}, testTime)
	require.NoError(t, err)

	assert.Equal(t, "A01", rack.Name())
	assert.Equal(t, "serial", rack.Serial())
	assert.Equal(t, "description", rack.Description())
	assert.Equal(t, "comments", rack.Comments())
	assert.True(t, rack.FacilityID().IsNull())
	assert.Equal(t, "", rackNullableValue(t, rack.AssetTag()))
	assert.Equal(t, dcim.RackFormFactor(""), rackNullableValue(t, rack.FormFactor()))
	assert.Equal(t, dcim.RackAirflow(""), rackNullableValue(t, rack.Airflow()))
	assert.Equal(t, "A01", rack.Display())

	atLimit, err := dcim.NewRack(dcim.RackValues{
		Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A02",
		Status: "active", Width: 19, UHeight: 100,
		StartingUnit: dcim.RackTypeMaximumStartingUnit,
	}, testTime)
	require.NoError(t, err)
	assert.Equal(t, dcim.RackTypeMaximumStartingUnit, atLimit.StartingUnit())

	for _, test := range []struct {
		name   string
		values dcim.RackValues
		field  string
	}{
		{
			name: "status choice whitespace is not trimmed",
			values: dcim.RackValues{
				Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A03",
				Status: " active ", Width: 19, UHeight: 42, StartingUnit: 1,
			},
			field: "status",
		},
		{
			name: "form factor choice whitespace is not trimmed",
			values: dcim.RackValues{
				Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A04",
				Status: "active", FormFactor: dcim.NonNullRackValue(" wall-frame "),
				Width: 19, UHeight: 42, StartingUnit: 1,
			},
			field: "form_factor",
		},
		{
			name: "airflow choice whitespace is not trimmed",
			values: dcim.RackValues{
				Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A05",
				Status: "active", Airflow: dcim.NonNullRackValue(" front-to-rear "),
				Width: 19, UHeight: 42, StartingUnit: 1,
			},
			field: "airflow",
		},
		{
			name: "starting unit exceeds small integer storage",
			values: dcim.RackValues{
				Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A06",
				Status: "active", Width: 19, UHeight: 42,
				StartingUnit: dcim.RackTypeMaximumStartingUnit + 1,
			},
			field: "starting_unit",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, createErr := dcim.NewRack(test.values, testTime)
			require.Error(t, createErr)
			require.Equal(t, test.field, shared.ViolationsOf(createErr)[0].Field)
		})
	}
}

func TestRackTypeOwnsAllPhysicalFieldsOnEverySave(t *testing.T) {
	t.Parallel()

	reference := rackTypeReference(t, 8, dcim.RackPhysicalAttributes{
		FormFactor:   dcim.RackFormFactorWallFrame,
		Width:        dcim.RackWidth23,
		UHeight:      24,
		StartingUnit: 3,
		DescUnits:    true,
	})
	rack, err := dcim.NewRack(dcim.RackValues{
		Site:         rackSiteReference(t, 3, "Moscow", "moscow"),
		Name:         "A01",
		RackType:     dcim.NonNullRackValue(reference),
		Status:       "active",
		FormFactor:   dcim.NonNullRackValue("4-post-cabinet"),
		Width:        19,
		UHeight:      42,
		StartingUnit: 1,
	}, testTime)
	require.NoError(t, err)

	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	require.NoError(t, rack.ApplyRackTypeOwnership(updatedAt))
	assert.Equal(t, dcim.RackFormFactorWallFrame, rackNullableValue(t, rack.FormFactor()))
	assert.Equal(t, dcim.RackWidth23, rack.Width())
	assert.Equal(t, uint32(24), rack.UHeight())
	assert.Equal(t, uint32(3), rack.StartingUnit())
	assert.True(t, rack.DescUnits())

	// A caller may submit conflicting physical fields, but the RackType projection
	// wins again before persistence on every application save.
	require.NoError(t, rack.ApplyPatch(dcim.RackPatch{
		Width:   rackUint32Pointer(19),
		UHeight: rackUint32Pointer(48),
	}, shared.NewTimestamp(testTime.Add(2*time.Minute))))
	require.NoError(t, rack.ApplyRackTypeOwnership(shared.NewTimestamp(testTime.Add(3*time.Minute))))
	assert.Equal(t, dcim.RackWidth23, rack.Width())
	assert.Equal(t, uint32(24), rack.UHeight())
}

func TestRackReturnsCompleteValidationAndRejectedPatchIsAtomic(t *testing.T) {
	t.Parallel()

	rack, err := dcim.NewRack(dcim.RackValues{
		Site:         dcim.SiteReference{},
		Name:         strings.Repeat("n", dcim.RackNameMaxLength+1),
		Status:       "unknown",
		FormFactor:   dcim.NonNullRackValue("unknown"),
		Width:        20,
		UHeight:      0,
		StartingUnit: 0,
		Airflow:      dcim.NonNullRackValue("sideways"),
	}, testTime)
	require.Error(t, err)
	assert.Nil(t, rack)
	assert.Equal(t, map[string]string{
		"site": "invalid_choice", "name": "max_length", "status": "invalid_choice",
		"form_factor": "invalid_choice", "width": "invalid_choice",
		"u_height": "range", "starting_unit": "range", "airflow": "invalid_choice",
	}, violationReasons(err))

	valid, err := dcim.NewRack(dcim.RackValues{
		Site: rackSiteReference(t, 3, "Moscow", "moscow"), Name: "A01",
		Status: "active", Width: 19, UHeight: 42, StartingUnit: 1,
	}, testTime)
	require.NoError(t, err)
	invalidHeight := uint32(0)
	err = valid.ApplyPatch(dcim.RackPatch{UHeight: &invalidHeight}, testTime)
	require.Error(t, err)
	assert.Equal(t, uint32(42), valid.UHeight())

	err = valid.ApplyPatch(dcim.RackPatch{}, testTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)
}

func rackSiteReference(t *testing.T, id shared.ID, name, slug string) dcim.SiteReference {
	t.Helper()
	reference, err := dcim.NewSiteReference(id, name, slug)
	require.NoError(t, err)
	return reference
}

func rackTypeReference(
	t *testing.T,
	id shared.ID,
	attributes dcim.RackPhysicalAttributes,
) dcim.RackTypeReference {
	t.Helper()
	reference, err := dcim.NewRackTypeReference(id, "R24", "r24", attributes)
	require.NoError(t, err)
	return reference
}

func rackNullableValue[T any](t *testing.T, value dcim.RackNullable[T]) T {
	t.Helper()
	out, present := value.Get()
	require.True(t, present)
	return out
}

func rackUint32Pointer(value uint32) *uint32 { return &value }
