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

// DCIMRackTypeService is the typed application contract for the six RackType
// RPCs.
type DCIMRackTypeService interface {
	ListRackTypes(context.Context, identity.Principal, applicationdcim.ListRackTypesQuery) (applicationdcim.RackTypePage, error)
	GetRackType(context.Context, identity.Principal, applicationdcim.GetRackTypeQuery) (*domaindcim.RackType, error)
	CreateRackType(context.Context, identity.Principal, applicationdcim.CreateRackTypeCommand) (*domaindcim.RackType, error)
	ReplaceRackType(context.Context, identity.Principal, applicationdcim.ReplaceRackTypeCommand) (*domaindcim.RackType, error)
	UpdateRackType(context.Context, identity.Principal, applicationdcim.UpdateRackTypeCommand) (*domaindcim.RackType, error)
	DeleteRackType(context.Context, identity.Principal, applicationdcim.DeleteRackTypeCommand) error
}

var _ DCIMRackTypeService = (*applicationdcim.RackTypeService)(nil)

// RackTypeRPCHandler implements the six RackType RPCs delegated by DCIMServer.
type RackTypeRPCHandler struct{ service DCIMRackTypeService }

func NewRackTypeRPCHandler(service DCIMRackTypeService) *RackTypeRPCHandler {
	if service == nil {
		panic("RackType gRPC handler requires a typed RackType service")
	}
	return &RackTypeRPCHandler{service: service}
}

func (handler *RackTypeRPCHandler) ListRackTypes(
	ctx context.Context,
	request *dcimv1.ListRackTypesRequest,
) (*dcimv1.ListRackTypesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListRackTypes(ctx, p, typedRackTypeListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.RackType, 0, len(page.Results))
	for _, rackType := range page.Results {
		results = append(results, typedRackTypeProto(rackType))
	}
	return &dcimv1.ListRackTypesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *RackTypeRPCHandler) GetRackType(
	ctx context.Context,
	request *dcimv1.GetRackTypeRequest,
) (*dcimv1.GetRackTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rackType, err := handler.service.GetRackType(
		ctx, p, applicationdcim.GetRackTypeQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetRackTypeResponse{RackType: typedRackTypeProto(rackType)}, nil
}

func (handler *RackTypeRPCHandler) CreateRackType(
	ctx context.Context,
	request *dcimv1.CreateRackTypeRequest,
) (*dcimv1.CreateRackTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rackType, err := handler.service.CreateRackType(
		ctx, p, typedRackTypeCreateCommand(request.GetRackType()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateRackTypeResponse{RackType: typedRackTypeProto(rackType)}, nil
}

func (handler *RackTypeRPCHandler) ReplaceRackType(
	ctx context.Context,
	request *dcimv1.ReplaceRackTypeRequest,
) (*dcimv1.ReplaceRackTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rackType, err := handler.service.ReplaceRackType(
		ctx, p, typedRackTypeReplaceCommand(shared.ID(request.GetId()), request.GetRackType()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceRackTypeResponse{RackType: typedRackTypeProto(rackType)}, nil
}

func (handler *RackTypeRPCHandler) UpdateRackType(
	ctx context.Context,
	request *dcimv1.UpdateRackTypeRequest,
) (*dcimv1.UpdateRackTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedRackTypeUpdateCommand(
		shared.ID(request.GetId()), request.GetRackType(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rackType, err := handler.service.UpdateRackType(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateRackTypeResponse{RackType: typedRackTypeProto(rackType)}, nil
}

func (handler *RackTypeRPCHandler) DeleteRackType(
	ctx context.Context,
	request *dcimv1.DeleteRackTypeRequest,
) (*dcimv1.DeleteRackTypeResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteRackType(
		ctx, p, applicationdcim.DeleteRackTypeCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteRackTypeResponse{}, nil
}

func typedRackTypeListQuery(request *dcimv1.ListRackTypesRequest) applicationdcim.ListRackTypesQuery {
	query := applicationdcim.ListRackTypesQuery{}
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

type typedRackTypeFields struct {
	manufacturer applicationdcim.Field[shared.ID]
	model        applicationdcim.Field[string]
	slug         applicationdcim.Field[string]
	formFactor   applicationdcim.Field[string]
	width        applicationdcim.Field[uint32]
	uHeight      applicationdcim.Field[uint32]
	startingUnit applicationdcim.Field[uint32]
	descUnits    applicationdcim.Field[bool]
	description  applicationdcim.Field[string]
	comments     applicationdcim.Field[string]
}

func typedRackTypeInputFields(input *dcimv1.RackTypeInput) typedRackTypeFields {
	if input == nil {
		return typedRackTypeFields{}
	}
	return typedRackTypeFields{
		manufacturer: typedRackTypeProtoID(input.Manufacturer), model: typedProtoString(input.Model),
		slug: typedProtoString(input.Slug), formFactor: typedProtoString(input.FormFactor),
		width: typedRackTypeProtoUint32(input.Width), uHeight: typedRackTypeProtoUint32(input.UHeight),
		startingUnit: typedRackTypeProtoUint32(input.StartingUnit), descUnits: typedRackTypeProtoBool(input.DescUnits),
		description: typedProtoString(input.Description), comments: typedProtoString(input.Comments),
	}
}

func typedRackTypeProtoID(value *int64) applicationdcim.Field[shared.ID] {
	if value == nil {
		return applicationdcim.OmittedField[shared.ID]()
	}
	return applicationdcim.FieldValue(shared.ID(*value))
}

func typedRackTypeProtoUint32(value *uint32) applicationdcim.Field[uint32] {
	if value == nil {
		return applicationdcim.OmittedField[uint32]()
	}
	return applicationdcim.FieldValue(*value)
}

func typedRackTypeProtoBool(value *bool) applicationdcim.Field[bool] {
	if value == nil {
		return applicationdcim.OmittedField[bool]()
	}
	return applicationdcim.FieldValue(*value)
}

func typedRackTypeCreateCommand(input *dcimv1.RackTypeInput) applicationdcim.CreateRackTypeCommand {
	fields := typedRackTypeInputFields(input)
	return applicationdcim.CreateRackTypeCommand{
		Manufacturer: fields.manufacturer, Model: fields.model, Slug: fields.slug,
		FormFactor: fields.formFactor, Width: fields.width, UHeight: fields.uHeight,
		StartingUnit: fields.startingUnit, DescUnits: fields.descUnits,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedRackTypeReplaceCommand(id shared.ID, input *dcimv1.RackTypeInput) applicationdcim.ReplaceRackTypeCommand {
	return applicationdcim.ReplaceRackTypeCommand{
		ID: id, CreateRackTypeCommand: typedRackTypeCreateCommand(input),
	}
}

func typedRackTypeUpdateCommand(
	id shared.ID,
	input *dcimv1.RackTypeInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateRackTypeCommand, error) {
	fields := typedRackTypeInputFields(input)
	command := applicationdcim.UpdateRackTypeCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Manufacturer = fields.manufacturer
		command.Model = fields.model
		command.Slug = fields.slug
		command.FormFactor = fields.formFactor
		command.Width = fields.width
		command.UHeight = fields.uHeight
		command.StartingUnit = fields.startingUnit
		command.DescUnits = fields.descUnits
		command.Description = fields.description
		command.Comments = fields.comments
		return command, nil
	}

	for _, path := range mask.Paths {
		switch path {
		case "manufacturer":
			field := fields.manufacturer
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[shared.ID]()
			}
			command.Manufacturer = field
		case "model":
			field := fields.model
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[string]()
			}
			command.Model = field
		case "slug":
			field := fields.slug
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[string]()
			}
			command.Slug = field
		case "form_factor":
			field := fields.formFactor
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[string]()
			}
			command.FormFactor = field
		case "width":
			field := fields.width
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[uint32]()
			}
			command.Width = field
		case "u_height":
			field := fields.uHeight
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[uint32]()
			}
			command.UHeight = field
		case "starting_unit":
			field := fields.startingUnit
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[uint32]()
			}
			command.StartingUnit = field
		case "desc_units":
			field := fields.descUnits
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[bool]()
			}
			command.DescUnits = field
		case "description":
			field := fields.description
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[string]()
			}
			command.Description = field
		case "comments":
			field := fields.comments
			if field.State() == applicationdcim.FieldOmitted {
				field = applicationdcim.NullField[string]()
			}
			command.Comments = field
		default:
			return applicationdcim.UpdateRackTypeCommand{}, invalidTypedRackTypeMask()
		}
	}
	return command, nil
}

func invalidTypedRackTypeMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field.",
		},
	)
}

func typedRackTypeProto(rackType *domaindcim.RackType) *dcimv1.RackType {
	if rackType == nil {
		return nil
	}
	manufacturer := rackType.Manufacturer()
	return &dcimv1.RackType{
		Id: rackType.ID().Int64(), Url: "/api/dcim/rack-types/" + rackType.ID().String() + "/",
		Display: rackType.Display(),
		Manufacturer: &typesv1.ObjectReference{
			Id:      manufacturer.ID().Int64(),
			Url:     "/api/dcim/manufacturers/" + manufacturer.ID().String() + "/",
			Display: manufacturer.Display(),
		},
		Model: rackType.Model(), Slug: rackType.Slug().String(), FormFactor: rackType.FormFactor().String(),
		Width: rackType.Width().Uint32(), UHeight: rackType.UHeight(), StartingUnit: rackType.StartingUnit(),
		DescUnits: rackType.DescUnits(), Description: rackType.Description(), Comments: rackType.Comments(),
		Created: timestamppb.New(rackType.Created().Time), LastUpdated: timestamppb.New(rackType.LastUpdated().Time),
	}
}
