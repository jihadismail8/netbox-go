package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	"netbox-go/internal/domain/shared"
)

func TestDeviceRoleGRPCScalarPresenceMatrix(t *testing.T) {
	t.Parallel()

	falseValue := false
	name := "Concrete Role"
	slug := "concrete-role"
	color := "00ff00"
	description := "Concrete description"
	comments := "Concrete comments"
	allPresent := &dcimv1.DeviceRoleInput{
		Parent: wrapperspb.Int64(7), Name: &name, Slug: &slug, Color: &color,
		VmRole: &falseValue, Description: &description, Comments: &comments,
	}

	create := typedDeviceRoleCreateCommand(allPresent)
	requireDeviceRoleGRPCPresentFields(
		t, create.Parent, create.Name, create.Slug, create.Color, create.VMRole,
		create.Description, create.Comments,
		7, name, slug, color, false, description, comments,
	)
	omittedCreate := typedDeviceRoleCreateCommand(nil)
	requireDeviceRoleGRPCOmittedFields(
		t, omittedCreate.Parent, omittedCreate.Name, omittedCreate.Slug, omittedCreate.Color,
		omittedCreate.VMRole, omittedCreate.Description, omittedCreate.Comments,
	)

	replace := typedDeviceRoleReplaceCommand(8, allPresent)
	assert.Equal(t, shared.ID(8), replace.ID)
	requireDeviceRoleGRPCPresentFields(
		t, replace.Parent, replace.Name, replace.Slug, replace.Color, replace.VMRole,
		replace.Description, replace.Comments,
		7, name, slug, color, false, description, comments,
	)
	omittedReplace := typedDeviceRoleReplaceCommand(8, nil)
	assert.Equal(t, shared.ID(8), omittedReplace.ID)
	requireDeviceRoleGRPCOmittedFields(
		t, omittedReplace.Parent, omittedReplace.Name, omittedReplace.Slug, omittedReplace.Color,
		omittedReplace.VMRole, omittedReplace.Description, omittedReplace.Comments,
	)

	blank := ""
	withoutMask, err := typedDeviceRoleUpdateCommand(
		8,
		&dcimv1.DeviceRoleInput{VmRole: &falseValue, Description: &blank},
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, applicationdcim.FieldOmitted, withoutMask.Parent.State())
	assert.Equal(t, applicationdcim.FieldOmitted, withoutMask.Name.State())
	assert.Equal(t, applicationdcim.FieldOmitted, withoutMask.Slug.State())
	assert.Equal(t, applicationdcim.FieldOmitted, withoutMask.Color.State())
	vmRole, present := withoutMask.VMRole.Get()
	require.True(t, present)
	assert.False(t, vmRole)
	actualDescription, present := withoutMask.Description.Get()
	require.True(t, present)
	assert.Empty(t, actualDescription)
	assert.Equal(t, applicationdcim.FieldOmitted, withoutMask.Comments.State())

	concrete, err := typedDeviceRoleUpdateCommand(
		8,
		&dcimv1.DeviceRoleInput{
			Parent: wrapperspb.Int64(7), Name: &name, VmRole: &falseValue,
		},
		&fieldmaskpb.FieldMask{Paths: []string{"parent", "name", "vm_role"}},
	)
	require.NoError(t, err)
	parent, present := concrete.Parent.Get()
	require.True(t, present)
	assert.Equal(t, shared.ID(7), parent)
	actualName, present := concrete.Name.Get()
	require.True(t, present)
	assert.Equal(t, name, actualName)
	vmRole, present = concrete.VMRole.Get()
	require.True(t, present)
	assert.False(t, vmRole)
	assert.Equal(t, applicationdcim.FieldOmitted, concrete.Slug.State())
	assert.Equal(t, applicationdcim.FieldOmitted, concrete.Color.State())
	assert.Equal(t, applicationdcim.FieldOmitted, concrete.Description.State())
	assert.Equal(t, applicationdcim.FieldOmitted, concrete.Comments.State())

	for _, field := range []string{"parent", "name", "slug", "color", "vm_role", "description", "comments"} {
		field := field
		t.Run("masked absent "+field+" carries explicit null intent", func(t *testing.T) {
			command, commandErr := typedDeviceRoleUpdateCommand(
				8,
				&dcimv1.DeviceRoleInput{},
				&fieldmaskpb.FieldMask{Paths: []string{field}},
			)
			require.NoError(t, commandErr)
			switch field {
			case "parent":
				assert.Equal(t, applicationdcim.FieldNull, command.Parent.State())
			case "name":
				assert.Equal(t, applicationdcim.FieldNull, command.Name.State())
			case "slug":
				assert.Equal(t, applicationdcim.FieldNull, command.Slug.State())
			case "color":
				assert.Equal(t, applicationdcim.FieldNull, command.Color.State())
			case "vm_role":
				assert.Equal(t, applicationdcim.FieldNull, command.VMRole.State())
			case "description":
				assert.Equal(t, applicationdcim.FieldNull, command.Description.State())
			case "comments":
				assert.Equal(t, applicationdcim.FieldNull, command.Comments.State())
			}
		})
	}

	_, err = typedDeviceRoleUpdateCommand(
		8,
		&dcimv1.DeviceRoleInput{},
		&fieldmaskpb.FieldMask{Paths: []string{"unsupported"}},
	)
	require.Error(t, err)
	assert.Equal(t, []shared.FieldViolation{{
		Field:       "update_mask",
		Description: "Every update_mask path must name a supported field.",
	}}, shared.ViolationsOf(err))
}

func requireDeviceRoleGRPCPresentFields(
	t *testing.T,
	parent applicationdcim.Field[shared.ID],
	name applicationdcim.Field[string],
	slug applicationdcim.Field[string],
	color applicationdcim.Field[string],
	vmRole applicationdcim.Field[bool],
	description applicationdcim.Field[string],
	comments applicationdcim.Field[string],
	wantParent shared.ID,
	wantName string,
	wantSlug string,
	wantColor string,
	wantVMRole bool,
	wantDescription string,
	wantComments string,
) {
	t.Helper()
	actualParent, present := parent.Get()
	require.True(t, present)
	assert.Equal(t, wantParent, actualParent)
	actualName, present := name.Get()
	require.True(t, present)
	assert.Equal(t, wantName, actualName)
	actualSlug, present := slug.Get()
	require.True(t, present)
	assert.Equal(t, wantSlug, actualSlug)
	actualColor, present := color.Get()
	require.True(t, present)
	assert.Equal(t, wantColor, actualColor)
	actualVMRole, present := vmRole.Get()
	require.True(t, present)
	assert.Equal(t, wantVMRole, actualVMRole)
	actualDescription, present := description.Get()
	require.True(t, present)
	assert.Equal(t, wantDescription, actualDescription)
	actualComments, present := comments.Get()
	require.True(t, present)
	assert.Equal(t, wantComments, actualComments)
}

func requireDeviceRoleGRPCOmittedFields(
	t *testing.T,
	parent applicationdcim.Field[shared.ID],
	name applicationdcim.Field[string],
	slug applicationdcim.Field[string],
	color applicationdcim.Field[string],
	vmRole applicationdcim.Field[bool],
	description applicationdcim.Field[string],
	comments applicationdcim.Field[string],
) {
	t.Helper()
	assert.Equal(t, applicationdcim.FieldOmitted, parent.State())
	assert.Equal(t, applicationdcim.FieldOmitted, name.State())
	assert.Equal(t, applicationdcim.FieldOmitted, slug.State())
	assert.Equal(t, applicationdcim.FieldOmitted, color.State())
	assert.Equal(t, applicationdcim.FieldOmitted, vmRole.State())
	assert.Equal(t, applicationdcim.FieldOmitted, description.State())
	assert.Equal(t, applicationdcim.FieldOmitted, comments.State())
}
