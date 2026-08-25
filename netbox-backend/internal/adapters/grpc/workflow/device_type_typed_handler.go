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

// DCIMDeviceTypeService is the typed application contract for the six
// DeviceType RPCs.
type DCIMDeviceTypeService interface {
	ListDeviceTypes(context.Context, identity.Principal, applicationdcim.ListDeviceTypesQuery) (applicationdcim.DeviceTypePage, error)
	GetDeviceType(context.Context, identity.Principal, applicationdcim.GetDeviceTypeQuery) (*domaindcim.DeviceType, error)
	CreateDeviceType(context.Context, identity.Principal, applicationdcim.CreateDeviceTypeCommand) (*domaindcim.DeviceType, error)
	ReplaceDeviceType(context.Context, identity.Principal, applicationdcim.ReplaceDeviceTypeCommand) (*domaindcim.DeviceType, error)
	UpdateDeviceType(context.Context, identity.Principal, applicationdcim.UpdateDeviceTypeCommand) (*domaindcim.DeviceType, error)
	DeleteDeviceType(context.Context, identity.Principal, applicationdcim.DeleteDeviceTypeCommand) error
}

var _ DCIMDeviceTypeService = (*applicationdcim.DeviceTypeService)(nil)

type DeviceTypeRPCHandler struct{ service DCIMDeviceTypeService }

func NewDeviceTypeRPCHandler(service DCIMDeviceTypeService) *DeviceTypeRPCHandler {
	if service == nil {
		panic("DeviceType gRPC handler requires a typed DeviceType service")
	}
	return &DeviceTypeRPCHandler{service: service}
}

func (handler *DeviceTypeRPCHandler) ListDeviceTypes(
	ctx context.Context,
	request *dcimv1.ListDeviceTypesRequest,
) (*dcimv1.ListDeviceTypesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListDeviceTypes(ctx, p, typedDeviceTypeListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.DeviceType, 0, len(page.Results))
	for _, deviceType := range page.Results {
		results = append(results, typedDeviceTypeProto(deviceType))
	}
	return &dcimv1.ListDeviceTypesResponse{
		Page:    &typesv1.PageInfo{Count: page.Count},
		Results: results,
	}, nil
}

func (handler *DeviceTypeRPCHandler) GetDeviceType(
	ctx context.Context,
	request *dcimv1.GetDeviceTypeRequest,
) (*dcimv1.GetDeviceTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	deviceType, err := handler.service.GetDeviceType(
		ctx, p, applicationdcim.GetDeviceTypeQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetDeviceTypeResponse{DeviceType: typedDeviceTypeProto(deviceType)}, nil
}

func (handler *DeviceTypeRPCHandler) CreateDeviceType(
	ctx context.Context,
	request *dcimv1.CreateDeviceTypeRequest,
) (*dcimv1.CreateDeviceTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	deviceType, err := handler.service.CreateDeviceType(
		ctx, p, typedDeviceTypeCreateCommand(request.GetDeviceType()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateDeviceTypeResponse{DeviceType: typedDeviceTypeProto(deviceType)}, nil
}

func (handler *DeviceTypeRPCHandler) ReplaceDeviceType(
	ctx context.Context,
	request *dcimv1.ReplaceDeviceTypeRequest,
) (*dcimv1.ReplaceDeviceTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	deviceType, err := handler.service.ReplaceDeviceType(
		ctx,
		p,
		typedDeviceTypeReplaceCommand(shared.ID(request.GetId()), request.GetDeviceType()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceDeviceTypeResponse{DeviceType: typedDeviceTypeProto(deviceType)}, nil
}

func (handler *DeviceTypeRPCHandler) UpdateDeviceType(
	ctx context.Context,
	request *dcimv1.UpdateDeviceTypeRequest,
) (*dcimv1.UpdateDeviceTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedDeviceTypeUpdateCommand(
		shared.ID(request.GetId()), request.GetDeviceType(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	deviceType, err := handler.service.UpdateDeviceType(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateDeviceTypeResponse{DeviceType: typedDeviceTypeProto(deviceType)}, nil
}

func (handler *DeviceTypeRPCHandler) DeleteDeviceType(
	ctx context.Context,
	request *dcimv1.DeleteDeviceTypeRequest,
) (*dcimv1.DeleteDeviceTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteDeviceType(
		ctx, p, applicationdcim.DeleteDeviceTypeCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteDeviceTypeResponse{}, nil
}

func typedDeviceTypeListQuery(
	request *dcimv1.ListDeviceTypesRequest,
) applicationdcim.ListDeviceTypesQuery {
	query := applicationdcim.ListDeviceTypesQuery{}
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
	if request.ManufacturerId != nil {
		query.ManufacturerIDs = []int64{*request.ManufacturerId}
	}
	query.ManufacturerSlugs = oneString(request.ManufacturerSlug)
	query.Models = oneString(request.Model)
	query.Slugs = oneString(request.Slug)
	return query
}

type typedDeviceTypeFields struct {
	manufacturer           applicationdcim.Field[shared.ID]
	model                  applicationdcim.Field[string]
	slug                   applicationdcim.Field[string]
	partNumber             applicationdcim.Field[string]
	uHeight                applicationdcim.Field[string]
	excludeFromUtilization applicationdcim.Field[bool]
	isFullDepth            applicationdcim.Field[bool]
	airflow                applicationdcim.Field[string]
	description            applicationdcim.Field[string]
	comments               applicationdcim.Field[string]
}

func typedDeviceTypeInputFields(input *dcimv1.DeviceTypeInput) typedDeviceTypeFields {
	if input == nil {
		return typedDeviceTypeFields{}
	}
	return typedDeviceTypeFields{
		manufacturer:           typedRackTypeProtoID(input.Manufacturer),
		model:                  typedProtoString(input.Model),
		slug:                   typedProtoString(input.Slug),
		partNumber:             typedProtoString(input.PartNumber),
		uHeight:                typedProtoString(input.UHeight),
		excludeFromUtilization: typedRackTypeProtoBool(input.ExcludeFromUtilization),
		isFullDepth:            typedRackTypeProtoBool(input.IsFullDepth),
		airflow:                typedProtoString(input.Airflow),
		description:            typedProtoString(input.Description),
		comments:               typedProtoString(input.Comments),
	}
}

func typedDeviceTypeCreateCommand(
	input *dcimv1.DeviceTypeInput,
) applicationdcim.CreateDeviceTypeCommand {
	fields := typedDeviceTypeInputFields(input)
	return applicationdcim.CreateDeviceTypeCommand{
		Manufacturer: fields.manufacturer,
		Model:        fields.model, Slug: fields.slug, PartNumber: fields.partNumber,
		UHeight: fields.uHeight, ExcludeFromUtilization: fields.excludeFromUtilization,
		IsFullDepth: fields.isFullDepth, Airflow: fields.airflow,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedDeviceTypeReplaceCommand(
	id shared.ID,
	input *dcimv1.DeviceTypeInput,
) applicationdcim.ReplaceDeviceTypeCommand {
	return applicationdcim.ReplaceDeviceTypeCommand{
		ID: id, CreateDeviceTypeCommand: typedDeviceTypeCreateCommand(input),
	}
}

func typedDeviceTypeUpdateCommand(
	id shared.ID,
	input *dcimv1.DeviceTypeInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateDeviceTypeCommand, error) {
	fields := typedDeviceTypeInputFields(input)
	command := applicationdcim.UpdateDeviceTypeCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Manufacturer = fields.manufacturer
		command.Model = fields.model
		command.Slug = fields.slug
		command.PartNumber = fields.partNumber
		command.UHeight = fields.uHeight
		command.ExcludeFromUtilization = fields.excludeFromUtilization
		command.IsFullDepth = fields.isFullDepth
		command.Airflow = fields.airflow
		command.Description = fields.description
		command.Comments = fields.comments
		return command, nil
	}

	for _, path := range mask.Paths {
		switch path {
		case "manufacturer":
			command.Manufacturer = maskedDeviceTypeField(fields.manufacturer)
		case "model":
			command.Model = maskedDeviceTypeField(fields.model)
		case "slug":
			command.Slug = maskedDeviceTypeField(fields.slug)
		case "part_number":
			command.PartNumber = maskedDeviceTypeField(fields.partNumber)
		case "u_height":
			command.UHeight = maskedDeviceTypeField(fields.uHeight)
		case "exclude_from_utilization":
			command.ExcludeFromUtilization = maskedDeviceTypeField(fields.excludeFromUtilization)
		case "is_full_depth":
			command.IsFullDepth = maskedDeviceTypeField(fields.isFullDepth)
		case "airflow":
			command.Airflow = maskedDeviceTypeField(fields.airflow)
		case "description":
			command.Description = maskedDeviceTypeField(fields.description)
		case "comments":
			command.Comments = maskedDeviceTypeField(fields.comments)
		default:
			return applicationdcim.UpdateDeviceTypeCommand{}, invalidTypedDeviceTypeMask()
		}
	}
	return command, nil
}

func maskedDeviceTypeField[T any](field applicationdcim.Field[T]) applicationdcim.Field[T] {
	if field.State() == applicationdcim.FieldOmitted {
		return applicationdcim.NullField[T]()
	}
	return field
}

func invalidTypedDeviceTypeMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field.",
		},
	)
}

func typedDeviceTypeProto(deviceType *domaindcim.DeviceType) *dcimv1.DeviceType {
	if deviceType == nil {
		return nil
	}
	manufacturer := deviceType.Manufacturer()
	airflow := ""
	if value, present := deviceType.Airflow().Get(); present {
		airflow = value.String()
	}
	return &dcimv1.DeviceType{
		Id:      deviceType.ID().Int64(),
		Url:     "/api/dcim/device-types/" + deviceType.ID().String() + "/",
		Display: deviceType.Display(),
		Manufacturer: &typesv1.ObjectReference{
			Id:      manufacturer.ID().Int64(),
			Url:     "/api/dcim/manufacturers/" + manufacturer.ID().String() + "/",
			Display: manufacturer.Display(),
		},
		Model: deviceType.Model(), Slug: deviceType.Slug().String(),
		PartNumber: deviceType.PartNumber(), UHeight: deviceType.UHeight().String(),
		ExcludeFromUtilization: deviceType.ExcludeFromUtilization(),
		IsFullDepth:            deviceType.IsFullDepth(), Airflow: airflow,
		Description: deviceType.Description(), Comments: deviceType.Comments(),
		Created:                timestamppb.New(deviceType.Created().Time),
		LastUpdated:            timestamppb.New(deviceType.LastUpdated().Time),
		DeviceCount:            deviceType.DeviceCount(),
		InterfaceTemplateCount: deviceType.InterfaceTemplateCount(),
	}
}
