package workflow

import (
	"context"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DCIMRackService interface {
	ListRacks(context.Context, identity.Principal, applicationdcim.ListRacksQuery) (applicationdcim.RackPage, error)
	GetRack(context.Context, identity.Principal, applicationdcim.GetRackQuery) (*domaindcim.Rack, error)
	CreateRack(context.Context, identity.Principal, applicationdcim.CreateRackCommand) (*domaindcim.Rack, error)
	ReplaceRack(context.Context, identity.Principal, applicationdcim.ReplaceRackCommand) (*domaindcim.Rack, error)
	UpdateRack(context.Context, identity.Principal, applicationdcim.UpdateRackCommand) (*domaindcim.Rack, error)
	DeleteRack(context.Context, identity.Principal, applicationdcim.DeleteRackCommand) error
}

var _ DCIMRackService = (*applicationdcim.RackService)(nil)

type RackRPCHandler struct{ service DCIMRackService }

func NewRackRPCHandler(service DCIMRackService) *RackRPCHandler {
	if service == nil {
		panic("Rack gRPC handler requires a typed Rack service")
	}
	return &RackRPCHandler{service: service}
}

func (handler *RackRPCHandler) ListRacks(
	ctx context.Context,
	request *dcimv1.ListRacksRequest,
) (*dcimv1.ListRacksResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListRacks(ctx, p, typedRackListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.Rack, 0, len(page.Results))
	for _, rack := range page.Results {
		results = append(results, typedRackProto(rack))
	}
	return &dcimv1.ListRacksResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *RackRPCHandler) GetRack(
	ctx context.Context,
	request *dcimv1.GetRackRequest,
) (*dcimv1.GetRackResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rack, err := handler.service.GetRack(
		ctx, p, applicationdcim.GetRackQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetRackResponse{Rack: typedRackProto(rack)}, nil
}

func (handler *RackRPCHandler) CreateRack(
	ctx context.Context,
	request *dcimv1.CreateRackRequest,
) (*dcimv1.CreateRackResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rack, err := handler.service.CreateRack(ctx, p, typedRackCreateCommand(request.GetRack()))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateRackResponse{Rack: typedRackProto(rack)}, nil
}

func (handler *RackRPCHandler) ReplaceRack(
	ctx context.Context,
	request *dcimv1.ReplaceRackRequest,
) (*dcimv1.ReplaceRackResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rack, err := handler.service.ReplaceRack(
		ctx, p, typedRackReplaceCommand(shared.ID(request.GetId()), request.GetRack()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceRackResponse{Rack: typedRackProto(rack)}, nil
}

func (handler *RackRPCHandler) UpdateRack(
	ctx context.Context,
	request *dcimv1.UpdateRackRequest,
) (*dcimv1.UpdateRackResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedRackUpdateCommand(
		shared.ID(request.GetId()), request.GetRack(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	rack, err := handler.service.UpdateRack(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateRackResponse{Rack: typedRackProto(rack)}, nil
}

func (handler *RackRPCHandler) DeleteRack(
	ctx context.Context,
	request *dcimv1.DeleteRackRequest,
) (*dcimv1.DeleteRackResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteRack(
		ctx, p, applicationdcim.DeleteRackCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteRackResponse{}, nil
}

func typedRackListQuery(request *dcimv1.ListRacksRequest) applicationdcim.ListRacksQuery {
	query := applicationdcim.ListRacksQuery{}
	if request == nil {
		return query
	}
	if request.Page != nil {
		if request.Page.Limit != nil {
			query.Limit, query.LimitPresent = *request.Page.Limit, true
		}
		if request.Page.Offset != nil {
			query.Offset = *request.Page.Offset
		}
		query.Query = request.Page.GetQuery()
		query.Ordering = append([]string(nil), request.Page.Ordering...)
		query.IDs = append([]int64(nil), request.Page.Id...)
	}
	if request.SiteId != nil {
		query.SiteIDs = []int64{*request.SiteId}
	}
	query.SiteSlugs = oneString(request.SiteSlug)
	query.Names = oneString(request.Name)
	query.Statuses = oneString(request.Status)
	if request.RoleId != nil {
		query.RoleIDs = []int64{*request.RoleId}
	}
	query.RoleSlugs = oneString(request.RoleSlug)
	if request.RackTypeId != nil {
		query.RackTypeIDs = []int64{*request.RackTypeId}
	}
	query.RackTypeSlugs = oneString(request.RackTypeSlug)
	return query
}

type typedRackFields struct {
	site         applicationdcim.Field[shared.ID]
	name         applicationdcim.Field[string]
	facilityID   applicationdcim.Field[string]
	rackType     applicationdcim.Field[shared.ID]
	status       applicationdcim.Field[string]
	role         applicationdcim.Field[shared.ID]
	serial       applicationdcim.Field[string]
	assetTag     applicationdcim.Field[string]
	formFactor   applicationdcim.Field[string]
	width        applicationdcim.Field[uint32]
	uHeight      applicationdcim.Field[uint32]
	startingUnit applicationdcim.Field[uint32]
	descUnits    applicationdcim.Field[bool]
	airflow      applicationdcim.Field[string]
	description  applicationdcim.Field[string]
	comments     applicationdcim.Field[string]
}

func typedRackInputFields(input *dcimv1.RackInput) typedRackFields {
	if input == nil {
		return typedRackFields{}
	}
	return typedRackFields{
		site: typedRackTypeProtoID(input.Site), name: typedProtoString(input.Name),
		facilityID: typedProtoString(input.FacilityId),
		rackType:   typedRackWrappedID(input.RackType), status: typedProtoString(input.Status),
		role: typedRackWrappedID(input.Role), serial: typedProtoString(input.Serial),
		assetTag: typedProtoString(input.AssetTag), formFactor: typedProtoString(input.FormFactor),
		width: typedRackTypeProtoUint32(input.Width), uHeight: typedRackTypeProtoUint32(input.UHeight),
		startingUnit: typedRackTypeProtoUint32(input.StartingUnit),
		descUnits:    typedRackTypeProtoBool(input.DescUnits), airflow: typedProtoString(input.Airflow),
		description: typedProtoString(input.Description), comments: typedProtoString(input.Comments),
	}
}

func typedRackWrappedID(value *wrapperspb.Int64Value) applicationdcim.Field[shared.ID] {
	if value == nil {
		return applicationdcim.OmittedField[shared.ID]()
	}
	return applicationdcim.FieldValue(shared.ID(value.Value))
}

func typedRackCreateCommand(input *dcimv1.RackInput) applicationdcim.CreateRackCommand {
	fields := typedRackInputFields(input)
	return applicationdcim.CreateRackCommand{
		Site: fields.site, Name: fields.name, FacilityID: fields.facilityID,
		RackType: fields.rackType, Status: fields.status, Role: fields.role,
		Serial: fields.serial, AssetTag: fields.assetTag, FormFactor: fields.formFactor,
		Width: fields.width, UHeight: fields.uHeight, StartingUnit: fields.startingUnit,
		DescUnits: fields.descUnits, Airflow: fields.airflow,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedRackReplaceCommand(id shared.ID, input *dcimv1.RackInput) applicationdcim.ReplaceRackCommand {
	return applicationdcim.ReplaceRackCommand{ID: id, CreateRackCommand: typedRackCreateCommand(input)}
}

func typedRackUpdateCommand(
	id shared.ID,
	input *dcimv1.RackInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateRackCommand, error) {
	fields := typedRackInputFields(input)
	command := applicationdcim.UpdateRackCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Site, command.Name, command.FacilityID = fields.site, fields.name, fields.facilityID
		command.RackType, command.Status, command.Role = fields.rackType, fields.status, fields.role
		command.Serial, command.AssetTag, command.FormFactor = fields.serial, fields.assetTag, fields.formFactor
		command.Width, command.UHeight, command.StartingUnit = fields.width, fields.uHeight, fields.startingUnit
		command.DescUnits, command.Airflow = fields.descUnits, fields.airflow
		command.Description, command.Comments = fields.description, fields.comments
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "site":
			command.Site = nullableTypedRackField(fields.site)
		case "name":
			command.Name = nullableTypedRackField(fields.name)
		case "facility_id":
			command.FacilityID = nullableTypedRackField(fields.facilityID)
		case "rack_type":
			command.RackType = nullableTypedRackField(fields.rackType)
		case "status":
			command.Status = nullableTypedRackField(fields.status)
		case "role":
			command.Role = nullableTypedRackField(fields.role)
		case "serial":
			command.Serial = nullableTypedRackField(fields.serial)
		case "asset_tag":
			command.AssetTag = nullableTypedRackField(fields.assetTag)
		case "form_factor":
			command.FormFactor = nullableTypedRackField(fields.formFactor)
		case "width":
			command.Width = nullableTypedRackField(fields.width)
		case "u_height":
			command.UHeight = nullableTypedRackField(fields.uHeight)
		case "starting_unit":
			command.StartingUnit = nullableTypedRackField(fields.startingUnit)
		case "desc_units":
			command.DescUnits = nullableTypedRackField(fields.descUnits)
		case "airflow":
			command.Airflow = nullableTypedRackField(fields.airflow)
		case "description":
			command.Description = nullableTypedRackField(fields.description)
		case "comments":
			command.Comments = nullableTypedRackField(fields.comments)
		default:
			return applicationdcim.UpdateRackCommand{}, invalidTypedRackMask()
		}
	}
	return command, nil
}

func nullableTypedRackField[T any](field applicationdcim.Field[T]) applicationdcim.Field[T] {
	if field.State() == applicationdcim.FieldPresent {
		return field
	}
	return applicationdcim.NullField[T]()
}

func invalidTypedRackMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field.",
		},
	)
}

func typedRackProto(rack *domaindcim.Rack) *dcimv1.Rack {
	if rack == nil {
		return nil
	}
	site := rack.Site()
	out := &dcimv1.Rack{
		Id: rack.ID().Int64(), Url: "/api/dcim/racks/" + rack.ID().String() + "/",
		Display: rack.Display(),
		Site: &typesv1.ObjectReference{
			Id: site.ID().Int64(), Url: "/api/dcim/sites/" + site.ID().String() + "/",
			Display: site.Display(),
		},
		Name: rack.Name(), Status: rack.Status().String(), Serial: rack.Serial(),
		Width: rack.Width().Uint32(), UHeight: rack.UHeight(), StartingUnit: rack.StartingUnit(),
		DescUnits: rack.DescUnits(), Description: rack.Description(), Comments: rack.Comments(),
		Created: timestamppb.New(rack.Created().Time), LastUpdated: timestamppb.New(rack.LastUpdated().Time),
		DeviceCount: rack.DeviceCount(),
	}
	if value, present := rack.FacilityID().Get(); present {
		out.FacilityId = value
	}
	if value, present := rack.RackType().Get(); present {
		out.RackTypeId = wrapperspb.Int64(value.ID().Int64())
	}
	if value, present := rack.Role().Get(); present {
		out.RoleId = wrapperspb.Int64(value.ID().Int64())
	}
	if value, present := rack.AssetTag().Get(); present {
		out.AssetTag = value
	}
	if value, present := rack.FormFactor().Get(); present {
		out.FormFactor = value.String()
	}
	if value, present := rack.Airflow().Get(); present {
		out.Airflow = value.String()
	}
	return out
}
