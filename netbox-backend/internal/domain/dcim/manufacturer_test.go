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

func TestManufacturerNormalizesCompleteStateAndSnapshot(t *testing.T) {
	t.Parallel()
	manufacturer, err := dcim.NewManufacturer(dcim.ManufacturerValues{
		Name: "  Juniper Networks  ", Slug: "  Juniper_Networks  ", Description: "  Hardware vendor  ",
	}, testTime)
	require.NoError(t, err)
	assert.Equal(t, "Juniper Networks", manufacturer.Name())
	assert.Equal(t, "Juniper_Networks", manufacturer.Slug().String())
	assert.Equal(t, "Hardware vendor", manufacturer.Description())
	assert.Equal(t, testTime, manufacturer.Created())
	assert.Equal(t, dcim.ManufacturerSnapshot{
		Name: "Juniper Networks", Slug: "Juniper_Networks", Description: "Hardware vendor",
	}, manufacturer.Snapshot())
}

func TestManufacturerReturnsEveryLocalViolation(t *testing.T) {
	t.Parallel()
	_, err := dcim.NewManufacturer(dcim.ManufacturerValues{
		Name:        strings.Repeat("n", dcim.ManufacturerNameMaxLength+1),
		Slug:        "invalid slug!",
		Description: strings.Repeat("d", dcim.ManufacturerDescriptionMaxLength+1),
	}, testTime)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	fields := violationReasons(err)
	assert.Equal(t, map[string]string{
		"name": "max_length", "slug": "invalid", "description": "max_length",
	}, fields)
}

func TestManufacturerPatchIsAtomicAndPreservesOmittedFields(t *testing.T) {
	t.Parallel()
	manufacturer := newManufacturer(t)
	require.NoError(t, manufacturer.AssignID(31))
	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	description := ""
	require.NoError(t, manufacturer.ApplyPatch(dcim.ManufacturerPatch{Description: &description}, updatedAt))
	assert.Equal(t, "Original Manufacturer", manufacturer.Name())
	assert.Empty(t, manufacturer.Description())
	assert.Equal(t, testTime, manufacturer.Created())
	assert.Equal(t, updatedAt, manufacturer.LastUpdated())

	invalidSlug := "not valid"
	err := manufacturer.ApplyPatch(dcim.ManufacturerPatch{Slug: &invalidSlug}, updatedAt)
	require.Error(t, err)
	assert.Equal(t, "original-manufacturer", manufacturer.Slug().String())
}

func TestManufacturerRejectsEmptyPatchAndInvalidRestore(t *testing.T) {
	t.Parallel()
	manufacturer := newManufacturer(t)
	err := manufacturer.ApplyPatch(dcim.ManufacturerPatch{}, testTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)

	_, err = dcim.RestoreManufacturer(dcim.ManufacturerState{
		ID: 4, Name: "Persisted", Slug: "bad slug", Created: testTime, LastUpdated: testTime,
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func TestRestoreManufacturerRetainsReadOnlyCounter(t *testing.T) {
	t.Parallel()
	manufacturer, err := dcim.RestoreManufacturer(dcim.ManufacturerState{
		ID: 8, Name: "Acme", Slug: "acme", Description: "Vendor",
		Created: testTime, LastUpdated: testTime, DeviceTypeCount: 12,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(12), manufacturer.DeviceTypeCount())
	assert.Equal(t, shared.ID(8), manufacturer.ID())
}

func newManufacturer(t *testing.T) *dcim.Manufacturer {
	t.Helper()
	manufacturer, err := dcim.NewManufacturer(dcim.ManufacturerValues{
		Name: "Original Manufacturer", Slug: "original-manufacturer", Description: "Original description",
	}, testTime)
	require.NoError(t, err)
	return manufacturer
}

func violationReasons(err error) map[string]string {
	reasons := make(map[string]string)
	for _, violation := range shared.ViolationsOf(err) {
		reasons[violation.Field] = violation.Reason
	}
	return reasons
}
