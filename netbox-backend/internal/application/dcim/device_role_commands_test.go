package dcim

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dcimdomain "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

var deviceRoleCommandTestTime = shared.NewTimestamp(
	time.Date(2026, time.August, 25, 6, 10, 0, 0, time.UTC),
)

func TestDeviceRoleScalarCommandPresenceMatrix(t *testing.T) {
	t.Parallel()

	t.Run("POST defaults and required fields", func(t *testing.T) {
		values, err := (CreateDeviceRoleCommand{}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		assert.True(t, values.Parent.IsRoot())
		assert.Equal(t, dcimdomain.DeviceRoleDefaultColor, values.Color)
		assert.True(t, values.VMRole)
		assert.Empty(t, values.Description)
		assert.Empty(t, values.Comments)
	})

	t.Run("POST null and invalid relationship aggregate deterministically", func(t *testing.T) {
		values, err := (CreateDeviceRoleCommand{
			Parent: FieldValue(shared.ID(0)), Name: NullField[string](),
			Slug: NullField[string](), Color: NullField[string](),
			VMRole: NullField[bool](), Description: NullField[string](), Comments: NullField[string](),
		}).values()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "parent", Reason: "invalid", Description: "A valid object ID is required."},
			{Field: "name", Reason: "null", Description: "This field may not be null."},
			{Field: "slug", Reason: "null", Description: "This field may not be null."},
			{Field: "color", Reason: "null", Description: "This field may not be null."},
			{Field: "vm_role", Reason: "null", Description: "This field may not be null."},
			{Field: "description", Reason: "null", Description: "This field may not be null."},
			{Field: "comments", Reason: "null", Description: "This field may not be null."},
		}, shared.ViolationsOf(err))
		assert.False(t, values.Parent.IsRoot())
	})

	t.Run("POST concrete false and blanks remain present", func(t *testing.T) {
		values, err := (CreateDeviceRoleCommand{
			Parent: FieldValue(shared.ID(9)), Name: FieldValue("  Access  "),
			Slug: FieldValue("  access  "), Color: FieldValue("  00ff00  "),
			VMRole: FieldValue(false), Description: FieldValue(""), Comments: FieldValue(""),
		}).values()
		require.NoError(t, err)
		parentID, present := values.Parent.Get()
		assert.True(t, present)
		assert.Equal(t, shared.ID(9), parentID)
		role, err := dcimdomain.NewDeviceRole(values, deviceRoleCommandTestTime)
		require.NoError(t, err)
		assert.False(t, role.VMRole())
		assert.Empty(t, role.Description())
		assert.Empty(t, role.Comments())
	})

	t.Run("PUT resets omitted parent but preserves other omitted optionals", func(t *testing.T) {
		patch, err := (ReplaceDeviceRoleCommand{
			ID: 1, Name: FieldValue("Replacement"), Slug: FieldValue("replacement"),
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.Parent)
		assert.True(t, patch.Parent.IsRoot())
		require.NotNil(t, patch.Name)
		require.NotNil(t, patch.Slug)
		assert.Nil(t, patch.Color)
		assert.Nil(t, patch.VMRole)
		assert.Nil(t, patch.Description)
		assert.Nil(t, patch.Comments)

		role := newDeviceRoleCommandTestAggregate(t)
		require.NoError(t, role.ApplyPatch(patch, deviceRoleCommandTestTimeAfter()))
		assert.True(t, role.Parent().IsRoot())
		assert.Equal(t, "123abc", role.Color().String())
		assert.False(t, role.VMRole())
		assert.Equal(t, "Original description", role.Description())
		assert.Equal(t, "Original comments", role.Comments())
	})

	t.Run("PUT requires identity and accepts explicit parent null", func(t *testing.T) {
		patch, err := (ReplaceDeviceRoleCommand{ID: 1, Parent: NullField[shared.ID]()}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "required", Description: "This field is required."},
			{Field: "slug", Reason: "required", Description: "This field is required."},
		}, shared.ViolationsOf(err))
		require.NotNil(t, patch.Parent)
		assert.True(t, patch.Parent.IsRoot())
	})

	t.Run("PATCH omission preserves and explicit null clears only parent", func(t *testing.T) {
		patch, err := (UpdateDeviceRoleCommand{
			ID: 1, Parent: NullField[shared.ID](), Description: FieldValue("changed"),
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.Parent)
		assert.True(t, patch.Parent.IsRoot())
		assert.Nil(t, patch.Name)
		assert.Nil(t, patch.Slug)
		assert.Nil(t, patch.Color)
		assert.Nil(t, patch.VMRole)
		require.NotNil(t, patch.Description)
		assert.Nil(t, patch.Comments)
	})

	t.Run("PATCH preserves valid siblings when null fields fail", func(t *testing.T) {
		patch, err := (UpdateDeviceRoleCommand{
			ID: 1, Name: NullField[string](), Slug: NullField[string](),
			Color: NullField[string](), VMRole: NullField[bool](),
			Description: NullField[string](), Comments: FieldValue("retained sibling"),
		}).patch()
		assert.Equal(t, []shared.FieldViolation{
			{Field: "name", Reason: "null", Description: "This field may not be null."},
			{Field: "slug", Reason: "null", Description: "This field may not be null."},
			{Field: "color", Reason: "null", Description: "This field may not be null."},
			{Field: "vm_role", Reason: "null", Description: "This field may not be null."},
			{Field: "description", Reason: "null", Description: "This field may not be null."},
		}, shared.ViolationsOf(err))
		require.NotNil(t, patch.Comments)
		assert.Equal(t, "retained sibling", *patch.Comments)
	})

	t.Run("parent zero is invalid for every operation", func(t *testing.T) {
		_, createErr := (CreateDeviceRoleCommand{
			Parent: FieldValue(shared.ID(0)), Name: FieldValue("Role"), Slug: FieldValue("role"),
		}).values()
		_, replaceErr := (ReplaceDeviceRoleCommand{
			ID: 1, Parent: FieldValue(shared.ID(0)), Name: FieldValue("Role"), Slug: FieldValue("role"),
		}).patch()
		_, updateErr := (UpdateDeviceRoleCommand{
			ID: 1, Parent: FieldValue(shared.ID(0)), Comments: FieldValue("sibling"),
		}).patch()
		for _, err := range []error{createErr, replaceErr, updateErr} {
			assert.Equal(t, []shared.FieldViolation{{
				Field: "parent", Reason: "invalid", Description: "A valid object ID is required.",
			}}, shared.ViolationsOf(err))
		}
	})

	t.Run("PATCH concrete false remains present", func(t *testing.T) {
		patch, err := (UpdateDeviceRoleCommand{
			ID: 1, VMRole: FieldValue(false),
		}).patch()
		require.NoError(t, err)
		require.NotNil(t, patch.VMRole)
		assert.False(t, *patch.VMRole)
	})
}

func newDeviceRoleCommandTestAggregate(t *testing.T) *dcimdomain.DeviceRole {
	t.Helper()
	role, err := dcimdomain.NewDeviceRole(dcimdomain.DeviceRoleValues{
		Parent: dcimdomain.NonRootDeviceRoleParent(7), Name: "Original Role",
		Slug: "original-role", Color: "123abc", VMRole: false,
		Description: "Original description", Comments: "Original comments",
	}, deviceRoleCommandTestTime)
	require.NoError(t, err)
	return role
}

func deviceRoleCommandTestTimeAfter() shared.Timestamp {
	return shared.NewTimestamp(deviceRoleCommandTestTime.Add(time.Minute))
}
