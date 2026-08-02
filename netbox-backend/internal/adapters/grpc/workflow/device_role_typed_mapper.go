package workflow

import (
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func typedDeviceRoleListQuery(request *dcimv1.ListDeviceRolesRequest) applicationdcim.ListDeviceRolesQuery {
	query := applicationdcim.ListDeviceRolesQuery{}
	if request == nil {
		return query
	}
	if request.Page != nil {
		if request.Page.Limit != nil {
			query.Limit = *request.Page.Limit
			query.LimitPresent = true
		}
		if request.Page.Offset != nil {
			query.Offset = *request.Page.Offset
		}
		query.Query = request.Page.GetQuery()
		query.Ordering = append([]string(nil), request.Page.Ordering...)
		query.IDs = append([]int64(nil), request.Page.Id...)
	}
	query.Names = oneString(request.Name)
	query.Slugs = oneString(request.Slug)
	return query
}

type typedDeviceRoleFields struct {
	parent      applicationdcim.Field[shared.ID]
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	color       applicationdcim.Field[string]
	vmRole      applicationdcim.Field[bool]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func typedDeviceRoleInputFields(input *dcimv1.DeviceRoleInput) typedDeviceRoleFields {
	if input == nil {
		return typedDeviceRoleFields{}
	}
	parent := applicationdcim.OmittedField[shared.ID]()
	if input.Parent != nil {
		parent = applicationdcim.FieldValue(shared.ID(input.Parent.Value))
	}
	return typedDeviceRoleFields{
		parent: parent, name: typedProtoString(input.Name), slug: typedProtoString(input.Slug),
		color: typedProtoString(input.Color), vmRole: typedDeviceRoleProtoBool(input.VmRole),
		description: typedProtoString(input.Description), comments: typedProtoString(input.Comments),
	}
}

func typedDeviceRoleProtoBool(value *bool) applicationdcim.Field[bool] {
	if value == nil {
		return applicationdcim.OmittedField[bool]()
	}
	return applicationdcim.FieldValue(*value)
}

func typedDeviceRoleCreateCommand(input *dcimv1.DeviceRoleInput) applicationdcim.CreateDeviceRoleCommand {
	fields := typedDeviceRoleInputFields(input)
	return applicationdcim.CreateDeviceRoleCommand{
		Parent: fields.parent, Name: fields.name, Slug: fields.slug, Color: fields.color,
		VMRole: fields.vmRole, Description: fields.description, Comments: fields.comments,
	}
}

func typedDeviceRoleReplaceCommand(
	id shared.ID,
	input *dcimv1.DeviceRoleInput,
) applicationdcim.ReplaceDeviceRoleCommand {
	fields := typedDeviceRoleInputFields(input)
	return applicationdcim.ReplaceDeviceRoleCommand{
		ID: id, Parent: fields.parent, Name: fields.name, Slug: fields.slug, Color: fields.color,
		VMRole: fields.vmRole, Description: fields.description, Comments: fields.comments,
	}
}

func typedDeviceRoleUpdateCommand(
	id shared.ID,
	input *dcimv1.DeviceRoleInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateDeviceRoleCommand, error) {
	fields := typedDeviceRoleInputFields(input)
	command := applicationdcim.UpdateDeviceRoleCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Parent = fields.parent
		command.Name = fields.name
		command.Slug = fields.slug
		command.Color = fields.color
		command.VMRole = fields.vmRole
		command.Description = fields.description
		command.Comments = fields.comments
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "parent":
			// A wrapper cannot carry JSON null. With an explicit mask, an absent
			// parent wrapper is the protocol's clear-to-root operation.
			if fields.parent.State() == applicationdcim.FieldOmitted {
				command.Parent = applicationdcim.NullField[shared.ID]()
			} else {
				command.Parent = fields.parent
			}
		case "name":
			if fields.name.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.Name = fields.name
		case "slug":
			if fields.slug.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.Slug = fields.slug
		case "color":
			if fields.color.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.Color = fields.color
		case "vm_role":
			if fields.vmRole.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.VMRole = fields.vmRole
		case "description":
			if fields.description.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.Description = fields.description
		case "comments":
			if fields.comments.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
			}
			command.Comments = fields.comments
		default:
			return applicationdcim.UpdateDeviceRoleCommand{}, invalidTypedDeviceRoleMask()
		}
	}
	return command, nil
}

func invalidTypedDeviceRoleMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value.",
		},
	)
}

func typedDeviceRoleProto(role *domaindcim.DeviceRole) *dcimv1.DeviceRole {
	if role == nil {
		return nil
	}
	result := &dcimv1.DeviceRole{
		Id: role.ID().Int64(), Url: "/api/dcim/device-roles/" + role.ID().String() + "/",
		Display: role.Display(), Name: role.Name(), Slug: role.Slug().String(),
		Color: role.Color().String(), VmRole: role.VMRole(), Description: role.Description(),
		Comments: role.Comments(), Created: timestamppb.New(role.Created().Time),
		LastUpdated: timestamppb.New(role.LastUpdated().Time), DeviceCount: role.DeviceCount(),
		Depth: role.Depth(),
	}
	if parent, present := role.Parent().Get(); present {
		result.ParentId = wrapperspb.Int64(parent.Int64())
	}
	return result
}
