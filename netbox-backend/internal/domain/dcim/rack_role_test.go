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

func TestRackRoleNormalizesAndValidatesLowercaseRGBColor(t *testing.T) {
	t.Parallel()
	role, err := dcim.NewRackRole(dcim.RackRoleValues{
		Name: "  Distribution  ", Slug: "  distribution  ", Color: "  ff00aa  ",
		Description: "  Distribution racks  ",
	}, testTime)
	require.NoError(t, err)
	assert.Equal(t, "Distribution", role.Name())
	assert.Equal(t, "distribution", role.Slug().String())
	assert.Equal(t, "ff00aa", role.Color().String())
	assert.Equal(t, "Distribution racks", role.Description())

	for _, invalid := range []string{"FF00AA", "fff", "gg00aa", "#ff00aa", ""} {
		_, err := dcim.ParseRackRoleColor(invalid)
		require.Error(t, err, invalid)
		assert.Equal(t, "color", shared.ViolationsOf(err)[0].Field)
	}
}

func TestRackRoleScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	nameAtLimit := strings.Repeat("é", dcim.RackRoleNameMaxLength)
	slugAtLimit := strings.Repeat("s", dcim.RackRoleSlugMaxLength)
	descriptionAtLimit := strings.Repeat("界", dcim.RackRoleDescriptionMaxLength)
	role, err := dcim.NewRackRole(dcim.RackRoleValues{
		Name: "  " + nameAtLimit + "  ", Slug: "  " + slugAtLimit + "  ",
		Color: "  0a1b2c  ", Description: "  " + descriptionAtLimit + "  ",
	}, testTime)
	require.NoError(t, err)
	assert.Equal(t, nameAtLimit, role.Name())
	assert.Equal(t, slugAtLimit, role.Slug().String())
	assert.Equal(t, "0a1b2c", role.Color().String())
	assert.Equal(t, descriptionAtLimit, role.Description())

	for _, test := range []struct {
		name        string
		color       string
		reason      string
		description string
	}{
		{
			name: "blank", color: "  ", reason: "required",
			description: "This field may not be blank.",
		},
		{
			name: "uppercase", color: "A0b1c2", reason: "invalid",
			description: "Enter a valid hexadecimal RGB color code.",
		},
		{
			name: "leading hash", color: "#0a1b2c", reason: "invalid",
			description: "Enter a valid hexadecimal RGB color code.",
		},
		{
			name: "wrong length", color: "0a1b2", reason: "invalid",
			description: "Enter a valid hexadecimal RGB color code.",
		},
		{
			name: "non hexadecimal", color: "0a1b2g", reason: "invalid",
			description: "Enter a valid hexadecimal RGB color code.",
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := dcim.ParseRackRoleColor(test.color)
			require.Error(t, err)
			assert.Equal(t, []shared.FieldViolation{{
				Field: "color", Reason: test.reason, Description: test.description,
			}}, shared.ViolationsOf(err))
		})
	}

	_, err = dcim.NewRackRole(dcim.RackRoleValues{
		Name: strings.Repeat("é", dcim.RackRoleNameMaxLength+1),
		Slug: "not valid!", Color: "ABCDEF",
		Description: strings.Repeat("界", dcim.RackRoleDescriptionMaxLength+1),
	}, testTime)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 4)
	assert.Equal(t, []string{"name", "slug", "color", "description"}, []string{
		violations[0].Field, violations[1].Field, violations[2].Field, violations[3].Field,
	})
	assert.Equal(t, []string{"max_length", "invalid", "invalid", "max_length"}, []string{
		violations[0].Reason, violations[1].Reason, violations[2].Reason, violations[3].Reason,
	})
}

func TestRackRoleReturnsEveryLocalViolation(t *testing.T) {
	t.Parallel()
	_, err := dcim.NewRackRole(dcim.RackRoleValues{
		Name: strings.Repeat("n", dcim.RackRoleNameMaxLength+1),
		Slug: "invalid slug!", Color: "ABCDEF",
		Description: strings.Repeat("d", dcim.RackRoleDescriptionMaxLength+1),
	}, testTime)
	require.Error(t, err)
	assert.Equal(t, map[string]string{
		"name": "max_length", "slug": "invalid", "color": "invalid", "description": "max_length",
	}, violationReasons(err))
}

func TestRackRolePatchPreservesPresenceAndIsAtomic(t *testing.T) {
	t.Parallel()
	role := newRackRole(t)
	require.NoError(t, role.AssignID(42))
	updatedAt := shared.NewTimestamp(testTime.Add(time.Minute))
	color := "00ff00"
	description := ""
	require.NoError(t, role.ApplyPatch(dcim.RackRolePatch{
		Color: &color, Description: &description,
	}, updatedAt))
	assert.Equal(t, "00ff00", role.Color().String())
	assert.Empty(t, role.Description())
	assert.Equal(t, "Original Role", role.Name())
	assert.Equal(t, updatedAt, role.LastUpdated())

	invalid := "00FF00"
	err := role.ApplyPatch(dcim.RackRolePatch{Color: &invalid}, updatedAt)
	require.Error(t, err)
	assert.Equal(t, "00ff00", role.Color().String())
}

func TestRackRoleRejectsEmptyPatchAndInvalidPersistence(t *testing.T) {
	t.Parallel()
	role := newRackRole(t)
	err := role.ApplyPatch(dcim.RackRolePatch{}, testTime)
	require.Error(t, err)
	assert.Equal(t, "update_mask", shared.ViolationsOf(err)[0].Field)

	_, err = dcim.RestoreRackRole(dcim.RackRoleState{
		ID: 5, Name: "Invalid", Slug: "invalid", Color: "ABCDEF",
		Created: testTime, LastUpdated: testTime,
	})
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}

func TestRestoreRackRoleRetainsReadOnlyCounterAndSnapshot(t *testing.T) {
	t.Parallel()
	role, err := dcim.RestoreRackRole(dcim.RackRoleState{
		ID: 9, Name: "Access", Slug: "access", Color: "123abc", Description: "Edge",
		Created: testTime, LastUpdated: testTime, RackCount: 7,
	})
	require.NoError(t, err)
	assert.Equal(t, uint64(7), role.RackCount())
	assert.Equal(t, dcim.RackRoleSnapshot{
		Name: "Access", Slug: "access", Color: "123abc", Description: "Edge",
	}, role.Snapshot())
}

func newRackRole(t *testing.T) *dcim.RackRole {
	t.Helper()
	role, err := dcim.NewRackRole(dcim.RackRoleValues{
		Name: "Original Role", Slug: "original-role", Color: dcim.RackRoleDefaultColor,
		Description: "Original description",
	}, testTime)
	require.NoError(t, err)
	return role
}
