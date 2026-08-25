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

func TestDeviceHeightUsesExactHalfUnitRepresentation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input     string
		halfUnits uint16
		rendered  string
	}{
		{input: "0", halfUnits: 0, rendered: "0"},
		{input: "0.5", halfUnits: 1, rendered: "0.5"},
		{input: ".5", halfUnits: 1, rendered: "0.5"},
		{input: "5e-1", halfUnits: 1, rendered: "0.5"},
		{input: "1", halfUnits: 2, rendered: "1"},
		{input: "1.", halfUnits: 2, rendered: "1"},
		{input: "1.5", halfUnits: 3, rendered: "1.5"},
		{input: "١.٥", halfUnits: 3, rendered: "1.5"},
		{input: "１.５", halfUnits: 3, rendered: "1.5"},
		{input: "𝟙.𝟝", halfUnits: 3, rendered: "1.5"},
		{input: "1_0", halfUnits: 20, rendered: "10"},
		{input: "-0", halfUnits: 0, rendered: "0"},
		{input: "-0.0", halfUnits: 0, rendered: "0"},
		{input: "999.5", halfUnits: 1999, rendered: "999.5"},
	} {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			height, err := dcim.ParseDeviceHeight(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.halfUnits, height.HalfUnits())
			assert.Equal(t, test.rendered, height.String())
			assert.Equal(t, float64(test.halfUnits)/2, height.Float64())
		})
	}

	for _, invalid := range []string{
		"-0.5", "0.1", "1e-1", "1.00", "1.25", "50e-2", "999.9", "1000",
		"1e3", "NaN", "Inf",
	} {
		invalid := invalid
		t.Run("invalid_"+invalid, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.ParseDeviceHeight(invalid)
			require.Error(t, err)
			assert.Equal(t, "u_height", shared.ViolationsOf(err)[0].Field)
		})
	}
}

func TestDeviceTypeScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	deviceType, err := dcim.NewDeviceType(dcim.DeviceTypeValues{
		Manufacturer: manufacturerReference(t, 7, "  Acme  ", "acme"),
		Model:        "  Router 9000  ", Slug: " router-9000 ", PartNumber: " PN-9 ",
		UHeight: " 1.5 ", IsFullDepth: true,
		Airflow:     dcim.NonNullDeviceAirflow(""),
		Description: " Core router ", Comments: " Notes ",
	}, testTime)
	require.NoError(t, err)
	require.NoError(t, deviceType.AssignID(23))

	assert.Equal(t, "Router 9000", deviceType.Model())
	assert.Equal(t, "router-9000", deviceType.Slug().String())
	assert.Equal(t, uint16(3), deviceType.UHeight().HalfUnits())
	airflow, present := deviceType.Airflow().Get()
	assert.True(t, present)
	assert.Empty(t, airflow)
	require.NotNil(t, deviceType.Snapshot().Airflow)
	assert.Empty(t, *deviceType.Snapshot().Airflow)

	null := dcim.NullDeviceAirflow()
	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	require.NoError(t, deviceType.ApplyPatch(dcim.DeviceTypePatch{
		Airflow: &null,
	}, updatedAt))
	assert.True(t, deviceType.Airflow().IsNull())
	assert.Nil(t, deviceType.Snapshot().Airflow)
	assert.Equal(t, updatedAt, deviceType.LastUpdated())

	t.Run("combined violations retain deterministic field order", func(t *testing.T) {
		t.Parallel()
		_, validationErr := dcim.NewDeviceType(dcim.DeviceTypeValues{
			Model: " ", Slug: "invalid slug!",
			PartNumber:  strings.Repeat("p", dcim.DeviceTypePartNumberMaxLength+1),
			UHeight:     "1.00",
			Airflow:     dcim.NonNullDeviceAirflow(" front-to-rear "),
			Description: strings.Repeat("d", dcim.DeviceTypeDescriptionMaxLength+1),
		}, testTime)
		require.Error(t, validationErr)
		violations := shared.ViolationsOf(validationErr)
		fields := make([]string, len(violations))
		for index, violation := range violations {
			fields[index] = violation.Field
		}
		assert.Equal(t, []string{
			"manufacturer", "model", "slug", "part_number", "u_height", "airflow",
			"description",
		}, fields)
	})
}

func TestDeviceTypePatchIsAtomicAndPreservesCounters(t *testing.T) {
	t.Parallel()

	deviceType, err := dcim.RestoreDeviceType(dcim.DeviceTypeState{
		ID: 29, Manufacturer: manufacturerReference(t, 7, "Acme", "acme"),
		Model: "Router", Slug: "router", UHeight: "1",
		IsFullDepth: true, Created: testTime, LastUpdated: testTime,
		DeviceCount: 4, InterfaceTemplateCount: 6,
	})
	require.NoError(t, err)
	model := "Changed"
	invalidHeight := "1.25"
	err = deviceType.ApplyPatch(dcim.DeviceTypePatch{
		Model: &model, UHeight: &invalidHeight,
	}, shared.NewTimestamp(testTime.Add(time.Minute)))
	require.Error(t, err)

	assert.Equal(t, "Router", deviceType.Model())
	assert.Equal(t, "1", deviceType.UHeight().String())
	assert.Equal(t, uint64(4), deviceType.DeviceCount())
	assert.Equal(t, uint64(6), deviceType.InterfaceTemplateCount())
	assert.Equal(t, testTime, deviceType.LastUpdated())
}
