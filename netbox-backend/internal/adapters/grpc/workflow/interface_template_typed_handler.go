package workflow

import (
	"context"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DCIMInterfaceTemplateService interface {
	ListInterfaceTemplates(
		context.Context,
		identity.Principal,
		applicationdcim.ListInterfaceTemplatesQuery,
	) (applicationdcim.InterfaceTemplatePage, error)
	GetInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.GetInterfaceTemplateQuery,
	) (*domaindcim.InterfaceTemplate, error)
	CreateInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.CreateInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	ReplaceInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	UpdateInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateInterfaceTemplateCommand,
	) (*domaindcim.InterfaceTemplate, error)
	DeleteInterfaceTemplate(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteInterfaceTemplateCommand,
	) error
}

var _ DCIMInterfaceTemplateService = (*applicationdcim.InterfaceTemplateService)(nil)

type InterfaceTemplateRPCHandler struct {
	service DCIMInterfaceTemplateService
}

func NewInterfaceTemplateRPCHandler(
	service DCIMInterfaceTemplateService,
) *InterfaceTemplateRPCHandler {
	if service == nil {
		panic("InterfaceTemplate gRPC handler requires a typed InterfaceTemplate service")
	}
	return &InterfaceTemplateRPCHandler{service: service}
}

func (handler *InterfaceTemplateRPCHandler) ListInterfaceTemplates(
	ctx context.Context,
	request *dcimv1.ListInterfaceTemplatesRequest,
) (*dcimv1.ListInterfaceTemplatesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListInterfaceTemplates(
		ctx, p, typedInterfaceTemplateListQuery(request),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.InterfaceTemplate, 0, len(page.Results))
	for _, template := range page.Results {
		results = append(results, typedInterfaceTemplateProto(template))
	}
	return &dcimv1.ListInterfaceTemplatesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *InterfaceTemplateRPCHandler) GetInterfaceTemplate(
	ctx context.Context,
	request *dcimv1.GetInterfaceTemplateRequest,
) (*dcimv1.GetInterfaceTemplateResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	template, err := handler.service.GetInterfaceTemplate(
		ctx, p, applicationdcim.GetInterfaceTemplateQuery{
			ID: shared.ID(request.GetId()),
		},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetInterfaceTemplateResponse{
		InterfaceTemplate: typedInterfaceTemplateProto(template),
	}, nil
}

func (handler *InterfaceTemplateRPCHandler) CreateInterfaceTemplate(
	ctx context.Context,
	request *dcimv1.CreateInterfaceTemplateRequest,
) (*dcimv1.CreateInterfaceTemplateResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	template, err := handler.service.CreateInterfaceTemplate(
		ctx, p, typedInterfaceTemplateCreateCommand(request.GetInterfaceTemplate()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateInterfaceTemplateResponse{
		InterfaceTemplate: typedInterfaceTemplateProto(template),
	}, nil
}

func (handler *InterfaceTemplateRPCHandler) ReplaceInterfaceTemplate(
	ctx context.Context,
	request *dcimv1.ReplaceInterfaceTemplateRequest,
) (*dcimv1.ReplaceInterfaceTemplateResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	template, err := handler.service.ReplaceInterfaceTemplate(
		ctx, p, applicationdcim.ReplaceInterfaceTemplateCommand{
			ID: shared.ID(request.GetId()),
			CreateInterfaceTemplateCommand: typedInterfaceTemplateCreateCommand(
				request.GetInterfaceTemplate(),
			),
		},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceInterfaceTemplateResponse{
		InterfaceTemplate: typedInterfaceTemplateProto(template),
	}, nil
}

func (handler *InterfaceTemplateRPCHandler) UpdateInterfaceTemplate(
	ctx context.Context,
	request *dcimv1.UpdateInterfaceTemplateRequest,
) (*dcimv1.UpdateInterfaceTemplateResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedInterfaceTemplateUpdateCommand(
		shared.ID(request.GetId()),
		request.GetInterfaceTemplate(),
		request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	template, err := handler.service.UpdateInterfaceTemplate(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateInterfaceTemplateResponse{
		InterfaceTemplate: typedInterfaceTemplateProto(template),
	}, nil
}

func (handler *InterfaceTemplateRPCHandler) DeleteInterfaceTemplate(
	ctx context.Context,
	request *dcimv1.DeleteInterfaceTemplateRequest,
) (*dcimv1.DeleteInterfaceTemplateResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteInterfaceTemplate(
		ctx, p, applicationdcim.DeleteInterfaceTemplateCommand{
			ID: shared.ID(request.GetId()),
		},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteInterfaceTemplateResponse{}, nil
}

func typedInterfaceTemplateListQuery(
	request *dcimv1.ListInterfaceTemplatesRequest,
) applicationdcim.ListInterfaceTemplatesQuery {
	query := applicationdcim.ListInterfaceTemplatesQuery{}
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
	if request.DeviceTypeId != nil {
		query.DeviceTypeIDs = []int64{*request.DeviceTypeId}
	}
	query.Names = oneString(request.Name)
	query.Types = oneString(request.Type)
	query.Enabled = copyProtoBool(request.Enabled)
	query.MgmtOnly = copyProtoBool(request.MgmtOnly)
	return query
}

func copyProtoBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

type typedInterfaceTemplateFields struct {
	deviceType    applicationdcim.Field[shared.ID]
	name          applicationdcim.Field[string]
	label         applicationdcim.Field[string]
	interfaceType applicationdcim.Field[string]
	enabled       applicationdcim.Field[bool]
	mgmtOnly      applicationdcim.Field[bool]
	description   applicationdcim.Field[string]
}

func typedInterfaceTemplateInputFields(
	input *dcimv1.InterfaceTemplateInput,
) typedInterfaceTemplateFields {
	if input == nil {
		return typedInterfaceTemplateFields{}
	}
	return typedInterfaceTemplateFields{
		deviceType:    typedRackTypeProtoID(input.DeviceType),
		name:          typedProtoString(input.Name),
		label:         typedProtoString(input.Label),
		interfaceType: typedProtoString(input.Type),
		enabled:       typedRackTypeProtoBool(input.Enabled),
		mgmtOnly:      typedRackTypeProtoBool(input.MgmtOnly),
		description:   typedProtoString(input.Description),
	}
}

func typedInterfaceTemplateCreateCommand(
	input *dcimv1.InterfaceTemplateInput,
) applicationdcim.CreateInterfaceTemplateCommand {
	fields := typedInterfaceTemplateInputFields(input)
	return applicationdcim.CreateInterfaceTemplateCommand{
		DeviceType: fields.deviceType, Name: fields.name, Label: fields.label,
		Type: fields.interfaceType, Enabled: fields.enabled, MgmtOnly: fields.mgmtOnly,
		Description: fields.description,
	}
}

func typedInterfaceTemplateUpdateCommand(
	id shared.ID,
	input *dcimv1.InterfaceTemplateInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateInterfaceTemplateCommand, error) {
	fields := typedInterfaceTemplateInputFields(input)
	command := applicationdcim.UpdateInterfaceTemplateCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.DeviceType = fields.deviceType
		command.Name = fields.name
		command.Label = fields.label
		command.Type = fields.interfaceType
		command.Enabled = fields.enabled
		command.MgmtOnly = fields.mgmtOnly
		command.Description = fields.description
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "device_type":
			command.DeviceType = maskedInterfaceTemplateField(fields.deviceType)
		case "name":
			command.Name = maskedInterfaceTemplateField(fields.name)
		case "label":
			command.Label = maskedInterfaceTemplateField(fields.label)
		case "type":
			command.Type = maskedInterfaceTemplateField(fields.interfaceType)
		case "enabled":
			command.Enabled = maskedInterfaceTemplateField(fields.enabled)
		case "mgmt_only":
			command.MgmtOnly = maskedInterfaceTemplateField(fields.mgmtOnly)
		case "description":
			command.Description = maskedInterfaceTemplateField(fields.description)
		default:
			return applicationdcim.UpdateInterfaceTemplateCommand{},
				invalidTypedInterfaceTemplateMask()
		}
	}
	return command, nil
}

func maskedInterfaceTemplateField[T any](
	field applicationdcim.Field[T],
) applicationdcim.Field[T] {
	if field.State() == applicationdcim.FieldOmitted {
		return applicationdcim.NullField[T]()
	}
	return field
}

func invalidTypedInterfaceTemplateMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field.",
		},
	)
}

func typedInterfaceTemplateProto(
	template *domaindcim.InterfaceTemplate,
) *dcimv1.InterfaceTemplate {
	if template == nil {
		return nil
	}
	deviceType := template.DeviceType()
	return &dcimv1.InterfaceTemplate{
		Id: template.ID().Int64(), Url: interfaceTemplateRPCURL(template.ID()),
		Display: template.Display(),
		DeviceType: &typesv1.ObjectReference{
			Id: deviceType.ID().Int64(), Url: deviceTypeRPCURL(deviceType.ID()),
			Display: deviceType.Display(),
		},
		Name: template.Name(), Label: template.Label(), Type: template.Type().String(),
		Enabled: template.Enabled(), MgmtOnly: template.MgmtOnly(),
		Description: template.Description(),
		Created:     timestamppb.New(template.Created().Time),
		LastUpdated: timestamppb.New(template.LastUpdated().Time),
	}
}

func interfaceTemplateRPCURL(id shared.ID) string {
	return "/api/dcim/interface-templates/" + id.String() + "/"
}

func deviceTypeRPCURL(id shared.ID) string {
	return "/api/dcim/device-types/" + id.String() + "/"
}
