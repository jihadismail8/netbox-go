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

var testTime = shared.NewTimestamp(time.Date(2026, time.July, 16, 12, 0, 0, 0, time.FixedZone("test", 3*60*60)))

func TestNewSiteNormalizesAndAcceptsEveryBaselineStatus(t *testing.T) {
	t.Parallel()

	statuses := []dcim.SiteStatus{
		dcim.SiteStatusPlanned,
		dcim.SiteStatusStaging,
		dcim.SiteStatusActive,
		dcim.SiteStatusDecommissioning,
		dcim.SiteStatusRetired,
	}
	for _, status := range statuses {
		status := status
		t.Run(status.String(), func(t *testing.T) {
			t.Parallel()
			site, err := dcim.NewSite(dcim.SiteValues{
				Name:        "  Moscow DC  ",
				Slug:        "  moscow-dc  ",
				Status:      status.String(),
				Facility:    "  M9  ",
				Description: "  Core site  ",
				Comments:    "  maintained  ",
			}, testTime)
			require.NoError(t, err)
			assert.Equal(t, "Moscow DC", site.Name())
			assert.Equal(t, "moscow-dc", site.Slug().String())
			assert.Equal(t, status, site.Status())
			assert.Equal(t, "M9", site.Facility())
			assert.Equal(t, testTime, site.Created())
			assert.Equal(t, testTime, site.LastUpdated())
		})
	}
}

func TestSiteScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	site, err := dcim.NewSite(dcim.SiteValues{
		Name:        "  Moscow DC  ",
		Slug:        "  moscow-dc  ",
		Status:      dcim.SiteStatusActive.String(),
		Facility:    "  M9  ",
		Description: "  Core site  ",
		Comments:    "  maintained  ",
	}, testTime)
	require.NoError(t, err)
	assert.Equal(t, "Moscow DC", site.Name())
	assert.Equal(t, "moscow-dc", site.Slug().String())
	assert.Equal(t, "M9", site.Facility())
	assert.Equal(t, "Core site", site.Description())
	assert.Equal(t, "maintained", site.Comments())

	for _, test := range []struct {
		name        string
		status      string
		description string
	}{
		{
			name:        "blank",
			status:      "",
			description: "This field may not be blank.",
		},
		{
			name:        "surrounding whitespace",
			status:      " active ",
			description: " active  is not a valid choice.",
		},
		{
			name:        "unknown",
			status:      "unknown",
			description: "unknown is not a valid choice.",
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.NewSite(dcim.SiteValues{
				Name:   "Moscow DC",
				Slug:   "moscow-dc",
				Status: test.status,
			}, testTime)
			require.Error(t, err)
			assert.Equal(t, []shared.FieldViolation{{
				Field:       "status",
				Reason:      map[bool]string{true: "blank", false: "invalid_choice"}[test.status == ""],
				Description: test.description,
			}}, shared.ViolationsOf(err))
		})
	}
}

func TestNewSiteReturnsAllFieldViolations(t *testing.T) {
	t.Parallel()

	_, err := dcim.NewSite(dcim.SiteValues{
		Name:        " ",
		Slug:        "not valid!",
		Status:      "unknown",
		Facility:    strings.Repeat("f", dcim.SiteFacilityMaxLength+1),
		Description: strings.Repeat("d", dcim.SiteDescriptionMaxLength+1),
	}, testTime)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonValidation))

	fields := make(map[string]string)
	for _, violation := range shared.ViolationsOf(err) {
		fields[violation.Field] = violation.Reason
	}
	assert.Equal(t, map[string]string{
		"name":        "required",
		"slug":        "invalid",
		"status":      "invalid_choice",
		"facility":    "max_length",
		"description": "max_length",
	}, fields)
}

func TestSitePatchPreservesPresenceAndTimestamps(t *testing.T) {
	t.Parallel()

	site := newTestSite(t)
	require.NoError(t, site.AssignID(41))
	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	empty := ""
	name := "Replacement"
	require.NoError(t, site.ApplyPatch(dcim.SitePatch{
		Name:     &name,
		Facility: &empty,
	}, updatedAt))

	assert.Equal(t, shared.ID(41), site.ID())
	assert.Equal(t, "Replacement", site.Name())
	assert.Empty(t, site.Facility())
	assert.Equal(t, "Original description", site.Description())
	assert.Equal(t, testTime, site.Created())
	assert.Equal(t, updatedAt, site.LastUpdated())
}

func TestSitePatchRejectsEmptyMaskAndBlankRequiredField(t *testing.T) {
	t.Parallel()

	site := newTestSite(t)
	err := site.ApplyPatch(dcim.SitePatch{}, testTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)

	empty := ""
	err = site.ApplyPatch(dcim.SitePatch{Name: &empty}, testTime)
	require.Error(t, err)
	assert.Equal(t, "name", shared.ViolationsOf(err)[0].Field)
	assert.Equal(t, "Original", site.Name(), "failed mutation must leave the aggregate unchanged")
}

func TestSiteIDCanOnlyBeAssignedOnce(t *testing.T) {
	t.Parallel()

	site := newTestSite(t)
	require.Error(t, site.AssignID(0))
	require.NoError(t, site.AssignID(7))
	err := site.AssignID(8)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
	assert.Equal(t, shared.ID(7), site.ID())
}

func TestRestoreSiteRejectsInvalidPersistedStateAsInternal(t *testing.T) {
	t.Parallel()

	_, err := dcim.RestoreSite(dcim.SiteState{
		ID:          1,
		Name:        "Invalid",
		Slug:        "invalid slug",
		Status:      dcim.SiteStatusActive.String(),
		Created:     testTime,
		LastUpdated: testTime,
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func newTestSite(t *testing.T) *dcim.Site {
	t.Helper()
	site, err := dcim.NewSite(dcim.SiteValues{
		Name:        "Original",
		Slug:        "original",
		Status:      dcim.SiteStatusActive.String(),
		Facility:    "M9",
		Description: "Original description",
		Comments:    "Original comments",
	}, testTime)
	require.NoError(t, err)
	return site
}
