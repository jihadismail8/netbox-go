package dcim_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRoleScalarNormalizationContract(t *testing.T) {
	t.Parallel()

	now := shared.NewTimestamp(time.Date(2026, time.August, 25, 6, 0, 0, 0, time.UTC))
	nameAtLimit := strings.Repeat("é", dcim.DeviceRoleNameMaxLength)
	slugAtLimit := strings.Repeat("s", dcim.DeviceRoleSlugMaxLength)
	descriptionAtLimit := strings.Repeat("界", dcim.DeviceRoleDescriptionMaxLength)
	role, err := dcim.NewDeviceRole(dcim.DeviceRoleValues{
		Parent: dcim.NonRootDeviceRoleParent(7), Name: "  " + nameAtLimit + "  ",
		Slug: "  " + slugAtLimit + "  ", Color: "  0a1b2c  ", VMRole: false,
		Description: "  " + descriptionAtLimit + "  ", Comments: "  retained note  ",
	}, now)
	require.NoError(t, err)
	assert.Equal(t, nameAtLimit, role.Name())
	assert.Equal(t, slugAtLimit, role.Slug().String())
	assert.Equal(t, "0a1b2c", role.Color().String())
	assert.False(t, role.VMRole())
	assert.Equal(t, descriptionAtLimit, role.Description())
	assert.Equal(t, "retained note", role.Comments())

	_, err = dcim.NewDeviceRole(dcim.DeviceRoleValues{
		Parent: dcim.NonRootDeviceRoleParent(0),
		Name:   strings.Repeat("é", dcim.DeviceRoleNameMaxLength+1),
		Slug:   "invalid slug!", Color: "ABCDEF",
		Description: strings.Repeat("界", dcim.DeviceRoleDescriptionMaxLength+1),
	}, now)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	require.Len(t, violations, 5)
	assert.Equal(t, []string{"parent", "name", "slug", "color", "description"}, []string{
		violations[0].Field, violations[1].Field, violations[2].Field,
		violations[3].Field, violations[4].Field,
	})
	assert.Equal(t, []string{"invalid", "max_length", "invalid", "invalid", "max_length"}, []string{
		violations[0].Reason, violations[1].Reason, violations[2].Reason,
		violations[3].Reason, violations[4].Reason,
	})

	before := role.Values()
	invalidColor := "ABCDEF"
	err = role.ValidatePatch(dcim.DeviceRolePatch{Color: &invalidColor})
	require.Error(t, err)
	assert.Equal(t, before, role.Values(), "validation preview must not mutate the aggregate")
}

func TestDeviceRoleNormalizesDefaultsAndPreservesNullableParent(t *testing.T) {
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 1, 0, 0, 0, time.UTC))
	role, err := dcim.NewDeviceRole(dcim.DeviceRoleValues{
		Parent: dcim.NonRootDeviceRoleParent(4), Name: "  Core  ", Slug: "core",
		Color: dcim.DeviceRoleDefaultColor, VMRole: true, Description: " role ", Comments: " note ",
	}, now)
	require.NoError(t, err)
	assert.Equal(t, "Core", role.Name())
	assert.True(t, role.VMRole())
	parentID, present := role.Parent().Get()
	assert.True(t, present)
	assert.Equal(t, shared.ID(4), parentID)

	root := dcim.RootDeviceRoleParent()
	require.NoError(t, role.ApplyPatch(dcim.DeviceRolePatch{Parent: &root}, now))
	assert.True(t, role.Parent().IsRoot())
}

func TestDeviceRoleRejectsInvalidCoreFields(t *testing.T) {
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 1, 0, 0, 0, time.UTC))
	_, err := dcim.NewDeviceRole(dcim.DeviceRoleValues{
		Parent: dcim.NonRootDeviceRoleParent(0), Slug: "Not Valid", Color: "ABCDEF",
	}, now)
	require.Error(t, err)
	violations := shared.ViolationsOf(err)
	assert.Contains(t, violations, shared.FieldViolation{
		Field: "parent", Reason: "invalid", Description: "A valid object ID is required.",
	})
	assert.Contains(t, violations, shared.FieldViolation{
		Field: "name", Reason: "required", Description: "This field may not be blank.",
	})
	assert.Contains(t, violations, shared.FieldViolation{
		Field: "color", Reason: "invalid", Description: "Enter a valid hexadecimal RGB color code.",
	})
}

func TestRestoreDeviceRoleRequiresConsistentHierarchyProjection(t *testing.T) {
	now := shared.NewTimestamp(time.Date(2026, time.July, 22, 1, 0, 0, 0, time.UTC))
	base := dcim.DeviceRoleState{
		ID: 2, Parent: dcim.NonRootDeviceRoleParent(1), ParentDisplay: "Root",
		Name: "Child", Slug: "child", Color: dcim.DeviceRoleDefaultColor, VMRole: true,
		Created: now, LastUpdated: now, Depth: 1,
	}
	role, err := dcim.RestoreDeviceRole(base)
	require.NoError(t, err)
	reference, present := role.ParentReference()
	assert.True(t, present)
	assert.Equal(t, "Root", reference.Display)

	base.ParentDisplay = ""
	_, err = dcim.RestoreDeviceRole(base)
	require.Error(t, err)
	assert.True(t, shared.HasReason(err, shared.ErrorReasonInternal))
}
