package dcim_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

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
