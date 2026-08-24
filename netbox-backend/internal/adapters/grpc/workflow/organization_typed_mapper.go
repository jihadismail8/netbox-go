package workflow

import (
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func typedManufacturerListQuery(request *dcimv1.ListManufacturersRequest) applicationdcim.ListManufacturersQuery {
	query := applicationdcim.ListManufacturersQuery{}
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

func typedRackRoleListQuery(request *dcimv1.ListRackRolesRequest) applicationdcim.ListRackRolesQuery {
	query := applicationdcim.ListRackRolesQuery{}
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

type typedManufacturerFields struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	description applicationdcim.Field[string]
}

func typedManufacturerInputFields(input *dcimv1.ManufacturerInput) typedManufacturerFields {
	if input == nil {
		return typedManufacturerFields{}
	}
	return typedManufacturerFields{
		name: typedProtoString(input.Name), slug: typedProtoString(input.Slug),
		description: typedProtoString(input.Description),
	}
}

func typedManufacturerCreateCommand(input *dcimv1.ManufacturerInput) applicationdcim.CreateManufacturerCommand {
	fields := typedManufacturerInputFields(input)
	return applicationdcim.CreateManufacturerCommand{
		Name: fields.name, Slug: fields.slug, Description: fields.description,
	}
}

func typedManufacturerReplaceCommand(
	id shared.ID,
	input *dcimv1.ManufacturerInput,
) applicationdcim.ReplaceManufacturerCommand {
	fields := typedManufacturerInputFields(input)
	return applicationdcim.ReplaceManufacturerCommand{
		ID: id, Name: fields.name, Slug: fields.slug, Description: fields.description,
	}
}

func typedManufacturerUpdateCommand(
	id shared.ID,
	input *dcimv1.ManufacturerInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateManufacturerCommand, error) {
	fields := typedManufacturerInputFields(input)
	command := applicationdcim.UpdateManufacturerCommand{ID: id}
	available := map[string]applicationdcim.Field[string]{
		"name": fields.name, "slug": fields.slug, "description": fields.description,
	}
	if mask == nil || len(mask.Paths) == 0 {
		command.Name = fields.name
		command.Slug = fields.slug
		command.Description = fields.description
		return command, nil
	}
	for _, path := range mask.Paths {
		field, supported := available[path]
		if !supported {
			return applicationdcim.UpdateManufacturerCommand{}, invalidTypedManufacturerMask()
		}
		if field.State() == applicationdcim.FieldOmitted {
			field = applicationdcim.NullField[string]()
		}
		switch path {
		case "name":
			command.Name = field
		case "slug":
			command.Slug = field
		case "description":
			command.Description = field
		}
	}
	return command, nil
}

func invalidTypedManufacturerMask() error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "update_mask", Description: "Every update_mask path must name a supported field.",
	})
}

type typedRackRoleFields struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	color       applicationdcim.Field[string]
	description applicationdcim.Field[string]
}

func typedRackRoleInputFields(input *dcimv1.RackRoleInput) typedRackRoleFields {
	if input == nil {
		return typedRackRoleFields{}
	}
	return typedRackRoleFields{
		name: typedProtoString(input.Name), slug: typedProtoString(input.Slug),
		color: typedProtoString(input.Color), description: typedProtoString(input.Description),
	}
}

func typedRackRoleCreateCommand(input *dcimv1.RackRoleInput) applicationdcim.CreateRackRoleCommand {
	fields := typedRackRoleInputFields(input)
	return applicationdcim.CreateRackRoleCommand{
		Name: fields.name, Slug: fields.slug, Color: fields.color, Description: fields.description,
	}
}

func typedRackRoleReplaceCommand(
	id shared.ID,
	input *dcimv1.RackRoleInput,
) applicationdcim.ReplaceRackRoleCommand {
	fields := typedRackRoleInputFields(input)
	return applicationdcim.ReplaceRackRoleCommand{
		ID: id, Name: fields.name, Slug: fields.slug, Color: fields.color,
		Description: fields.description,
	}
}

func typedRackRoleUpdateCommand(
	id shared.ID,
	input *dcimv1.RackRoleInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateRackRoleCommand, error) {
	fields := typedRackRoleInputFields(input)
	command := applicationdcim.UpdateRackRoleCommand{ID: id}
	available := map[string]applicationdcim.Field[string]{
		"name": fields.name, "slug": fields.slug, "color": fields.color,
		"description": fields.description,
	}
	if mask == nil || len(mask.Paths) == 0 {
		command.Name = fields.name
		command.Slug = fields.slug
		command.Color = fields.color
		command.Description = fields.description
		return command, nil
	}
	for _, path := range mask.Paths {
		field, supported := available[path]
		if !supported {
			return applicationdcim.UpdateRackRoleCommand{}, invalidTypedRackRoleMask()
		}
		if field.State() == applicationdcim.FieldOmitted {
			field = applicationdcim.NullField[string]()
		}
		switch path {
		case "name":
			command.Name = field
		case "slug":
			command.Slug = field
		case "color":
			command.Color = field
		case "description":
			command.Description = field
		}
	}
	return command, nil
}

func invalidTypedRackRoleMask() error {
	return shared.NewValidationError(shared.FieldViolation{
		Field: "update_mask", Description: "Every update_mask path must name a supported field.",
	})
}

func typedManufacturerProto(manufacturer *domaindcim.Manufacturer) *dcimv1.Manufacturer {
	if manufacturer == nil {
		return nil
	}
	return &dcimv1.Manufacturer{
		Id:      manufacturer.ID().Int64(),
		Url:     "/api/dcim/manufacturers/" + manufacturer.ID().String() + "/",
		Display: manufacturer.Display(), Name: manufacturer.Name(), Slug: manufacturer.Slug().String(),
		Description: manufacturer.Description(), Created: timestamppb.New(manufacturer.Created().Time),
		LastUpdated:     timestamppb.New(manufacturer.LastUpdated().Time),
		DevicetypeCount: manufacturer.DeviceTypeCount(),
	}
}

func typedRackRoleProto(role *domaindcim.RackRole) *dcimv1.RackRole {
	if role == nil {
		return nil
	}
	return &dcimv1.RackRole{
		Id: role.ID().Int64(), Url: "/api/dcim/rack-roles/" + role.ID().String() + "/",
		Display: role.Display(), Name: role.Name(), Slug: role.Slug().String(),
		Color: role.Color().String(), Description: role.Description(),
		Created: timestamppb.New(role.Created().Time), LastUpdated: timestamppb.New(role.LastUpdated().Time),
		RackCount: role.RackCount(),
	}
}
