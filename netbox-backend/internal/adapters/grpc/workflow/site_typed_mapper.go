package workflow

import (
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/shared"
)

func typedSiteListQuery(request *dcimv1.ListSitesRequest) applicationdcim.ListSitesQuery {
	query := applicationdcim.ListSitesQuery{}
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
	query.Statuses = oneString(request.Status)
	return query
}

func oneString(value *string) []string {
	if value == nil {
		return nil
	}
	return []string{*value}
}

func typedSiteCreateCommand(input *dcimv1.SiteInput) applicationdcim.CreateSiteCommand {
	fields := typedSiteInputFields(input)
	return applicationdcim.CreateSiteCommand{
		Name: fields.name, Slug: fields.slug, Status: fields.status,
		Facility: fields.facility, Description: fields.description, Comments: fields.comments,
	}
}

func typedSiteReplaceCommand(id shared.ID, input *dcimv1.SiteInput) applicationdcim.ReplaceSiteCommand {
	fields := typedSiteInputFields(input)
	return applicationdcim.ReplaceSiteCommand{
		ID: id, Name: fields.name, Slug: fields.slug, Status: fields.status,
		Facility: fields.facility, Description: fields.description, Comments: fields.comments,
	}
}

func typedSiteUpdateCommand(
	id shared.ID,
	input *dcimv1.SiteInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateSiteCommand, error) {
	fields := typedSiteInputFields(input)
	command := applicationdcim.UpdateSiteCommand{ID: id}
	available := map[string]applicationdcim.Field[string]{
		"name": fields.name, "slug": fields.slug, "status": fields.status,
		"facility": fields.facility, "description": fields.description, "comments": fields.comments,
	}
	if mask == nil || len(mask.Paths) == 0 {
		command.Name = fields.name
		command.Slug = fields.slug
		command.Status = fields.status
		command.Facility = fields.facility
		command.Description = fields.description
		command.Comments = fields.comments
		return command, nil
	}
	for _, path := range mask.Paths {
		field, supported := available[path]
		if !supported {
			return applicationdcim.UpdateSiteCommand{}, shared.NewValidationError(
				shared.FieldViolation{
					Field:       "update_mask",
					Description: "Every update_mask path must name a supported field.",
				},
			)
		}
		if field.State() == applicationdcim.FieldOmitted {
			field = applicationdcim.NullField[string]()
		}
		switch path {
		case "name":
			command.Name = field
		case "slug":
			command.Slug = field
		case "status":
			command.Status = field
		case "facility":
			command.Facility = field
		case "description":
			command.Description = field
		case "comments":
			command.Comments = field
		}
	}
	return command, nil
}

type typedSiteFields struct {
	name        applicationdcim.Field[string]
	slug        applicationdcim.Field[string]
	status      applicationdcim.Field[string]
	facility    applicationdcim.Field[string]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func typedSiteInputFields(input *dcimv1.SiteInput) typedSiteFields {
	if input == nil {
		return typedSiteFields{}
	}
	return typedSiteFields{
		name:        typedProtoString(input.Name),
		slug:        typedProtoString(input.Slug),
		status:      typedProtoString(input.Status),
		facility:    typedProtoString(input.Facility),
		description: typedProtoString(input.Description),
		comments:    typedProtoString(input.Comments),
	}
}

func typedProtoString(value *string) applicationdcim.Field[string] {
	if value == nil {
		return applicationdcim.OmittedField[string]()
	}
	return applicationdcim.FieldValue(*value)
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func typedSiteProto(site *domaindcim.Site) *dcimv1.Site {
	if site == nil {
		return nil
	}
	return &dcimv1.Site{
		Id:          site.ID().Int64(),
		Url:         "/api/dcim/sites/" + site.ID().String() + "/",
		Display:     site.Display(),
		Name:        site.Name(),
		Slug:        site.Slug().String(),
		Status:      site.Status().String(),
		Facility:    site.Facility(),
		Description: site.Description(),
		Comments:    site.Comments(),
		Created:     timestamppb.New(site.Created().Time),
		LastUpdated: timestamppb.New(site.LastUpdated().Time),
		DeviceCount: site.DeviceCount(),
		PrefixCount: site.PrefixCount(),
		RackCount:   site.RackCount(),
	}
}
