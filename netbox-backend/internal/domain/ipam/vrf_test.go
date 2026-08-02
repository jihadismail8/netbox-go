package ipam_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

var vrfTestTime = shared.NewTimestamp(time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC))

func TestRouteDistinguisherPreservesNullPresentBlankAndText(t *testing.T) {
	t.Parallel()

	nullable := ipam.NullRouteDistinguisher()
	assert.True(t, nullable.IsNull())
	_, present := nullable.Get()
	assert.False(t, present)

	blank, err := ipam.ParseRouteDistinguisher("   ")
	require.NoError(t, err)
	assert.Empty(t, blank.String())
	nonNullBlank := ipam.NonNullRouteDistinguisher(blank)
	loadedBlank, present := nonNullBlank.Get()
	assert.True(t, present)
	assert.Empty(t, loadedBlank.String())

	rd, err := ipam.ParseRouteDistinguisher("  65000:100  ")
	require.NoError(t, err)
	assert.Equal(t, "65000:100", rd.String())

	// The pinned NetBox CharField has a length constraint but no RFC-shape
	// validator. Rejecting this would be stricter than the oracle.
	arbitrary, err := ipam.ParseRouteDistinguisher("plain-text")
	require.NoError(t, err)
	assert.Equal(t, "plain-text", arbitrary.String())
}

func TestRouteDistinguisherRejectsMoreThanTwentyOneCharacters(t *testing.T) {
	t.Parallel()

	_, err := ipam.ParseRouteDistinguisher(strings.Repeat("r", ipam.VRFRouteDistinguisherMaxLength+1))
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))
	require.Len(t, shared.ViolationsOf(err), 1)
	assert.Equal(t, "rd", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "max_length", shared.ViolationsOf(err)[0].Reason)
}

func TestNewVRFNormalizesFieldsAndBuildsDisplay(t *testing.T) {
	t.Parallel()

	vrf, err := ipam.NewVRF(ipam.VRFValues{
		Name:          "  Tenant Blue  ",
		RD:            nullableRD(t, " 65000:100 "),
		EnforceUnique: true,
		Description:   "  Production table  ",
		Comments:      "  managed  ",
	}, vrfTestTime)
	require.NoError(t, err)

	assert.Equal(t, "Tenant Blue", vrf.Name())
	assert.Equal(t, "65000:100", mustRDValue(t, vrf.RD()))
	assert.True(t, vrf.EnforceUnique())
	assert.Equal(t, "Production table", vrf.Description())
	assert.Equal(t, "managed", vrf.Comments())
	assert.Equal(t, "Tenant Blue (65000:100)", vrf.Display())
	assert.Equal(t, vrfTestTime, vrf.Created())
	assert.Equal(t, vrfTestTime, vrf.LastUpdated())
}

func TestVRFDisplayOmitsNullAndPresentBlankRD(t *testing.T) {
	t.Parallel()

	for name, rd := range map[string]ipam.NullableRouteDistinguisher{
		"null":  ipam.NullRouteDistinguisher(),
		"blank": nullableRD(t, ""),
	} {
		t.Run(name, func(t *testing.T) {
			vrf, err := ipam.NewVRF(ipam.VRFValues{Name: "Global", RD: rd}, vrfTestTime)
			require.NoError(t, err)
			assert.Equal(t, "Global", vrf.Display())
		})
	}
}

func TestNewVRFReturnsAllLocalFieldViolations(t *testing.T) {
	t.Parallel()

	_, err := ipam.NewVRF(ipam.VRFValues{
		Name:        strings.Repeat("n", ipam.VRFNameMaxLength+1),
		Description: strings.Repeat("d", ipam.VRFDescriptionMaxLength+1),
	}, vrfTestTime)
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{
		{Field: "name", Reason: "max_length", Description: "Ensure this field has no more than 100 characters."},
		{Field: "description", Reason: "max_length", Description: "Ensure this field has no more than 200 characters."},
	}, shared.ViolationsOf(err))

	_, err = ipam.NewVRF(ipam.VRFValues{Name: "  "}, vrfTestTime)
	require.Error(t, err)
	assert.Equal(t, "required", shared.ViolationsOf(err)[0].Reason)
}

func TestVRFPatchPreservesPresenceAndCanClearRDToNull(t *testing.T) {
	t.Parallel()

	vrf := newDomainVRF(t)
	require.NoError(t, vrf.AssignID(41))
	updatedAt := shared.NewTimestamp(vrfTestTime.Add(time.Minute))
	empty := ""
	enforceUnique := false
	nullRD := ipam.NullRouteDistinguisher()
	require.NoError(t, vrf.ApplyPatch(ipam.VRFPatch{
		RD:            &nullRD,
		EnforceUnique: &enforceUnique,
		Description:   &empty,
	}, updatedAt))

	assert.Equal(t, shared.ID(41), vrf.ID())
	assert.True(t, vrf.RD().IsNull())
	assert.False(t, vrf.EnforceUnique())
	assert.Empty(t, vrf.Description())
	assert.Equal(t, "Original comments", vrf.Comments())
	assert.Equal(t, vrfTestTime, vrf.Created())
	assert.Equal(t, updatedAt, vrf.LastUpdated())
}

func TestVRFPatchRejectsEmptyMaskAndLeavesFailedMutationUnchanged(t *testing.T) {
	t.Parallel()

	vrf := newDomainVRF(t)
	err := vrf.ApplyPatch(ipam.VRFPatch{}, vrfTestTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)

	empty := ""
	err = vrf.ApplyPatch(ipam.VRFPatch{Name: &empty}, vrfTestTime)
	require.Error(t, err)
	assert.Equal(t, "Original", vrf.Name())
	assert.Equal(t, "65000:10", mustRDValue(t, vrf.RD()))
}

func TestRestoreVRFPreservesCountsAndRejectsCorruptState(t *testing.T) {
	t.Parallel()

	vrf, err := ipam.RestoreVRF(ipam.VRFState{
		ID:             9,
		Name:           "Restored",
		RD:             nullableRD(t, "64512:9"),
		EnforceUnique:  true,
		Description:    "description",
		Comments:       "comments",
		Created:        vrfTestTime,
		LastUpdated:    vrfTestTime,
		IPAddressCount: 3,
		PrefixCount:    4,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(3), vrf.IPAddressCount())
	assert.Equal(t, uint64(4), vrf.PrefixCount())

	_, err = ipam.RestoreVRF(ipam.VRFState{
		ID:          9,
		Name:        " ",
		Created:     vrfTestTime,
		LastUpdated: vrfTestTime,
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func TestVRFIDCanOnlyBeAssignedOnce(t *testing.T) {
	t.Parallel()

	vrf := newDomainVRF(t)
	require.Error(t, vrf.AssignID(0))
	require.NoError(t, vrf.AssignID(7))
	err := vrf.AssignID(8)
	require.Error(t, err)
	assert.Equal(t, shared.ID(7), vrf.ID())
}

func newDomainVRF(t *testing.T) *ipam.VRF {
	t.Helper()
	vrf, err := ipam.NewVRF(ipam.VRFValues{
		Name:          "Original",
		RD:            nullableRD(t, "65000:10"),
		EnforceUnique: true,
		Description:   "Original description",
		Comments:      "Original comments",
	}, vrfTestTime)
	require.NoError(t, err)
	return vrf
}

func nullableRD(t *testing.T, value string) ipam.NullableRouteDistinguisher {
	t.Helper()
	rd, err := ipam.ParseRouteDistinguisher(value)
	require.NoError(t, err)
	return ipam.NonNullRouteDistinguisher(rd)
}

func mustRDValue(t *testing.T, nullable ipam.NullableRouteDistinguisher) string {
	t.Helper()
	rd, present := nullable.Get()
	require.True(t, present)
	return rd.String()
}
