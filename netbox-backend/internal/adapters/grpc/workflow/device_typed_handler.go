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

// DCIMDeviceService is the typed application contract for the six Device
// RPCs.
type DCIMDeviceService interface {
	ListDevices(
		context.Context,
		identity.Principal,
		applicationdcim.ListDevicesQuery,
	) (applicationdcim.DevicePage, error)
	GetDevice(
		context.Context,
		identity.Principal,
		applicationdcim.GetDeviceQuery,
	) (*domaindcim.Device, error)
	CreateDevice(
		context.Context,
		identity.Principal,
		applicationdcim.CreateDeviceCommand,
	) (*domaindcim.Device, error)
	ReplaceDevice(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceDeviceCommand,
	) (*domaindcim.Device, error)
	UpdateDevice(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateDeviceCommand,
	) (*domaindcim.Device, error)
	DeleteDevice(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteDeviceCommand,
	) error
}

var _ DCIMDeviceService = (*applicationdcim.DeviceService)(nil)

type DeviceRPCHandler struct {
	service DCIMDeviceService
}

func NewDeviceRPCHandler(service DCIMDeviceService) *DeviceRPCHandler {
	if service == nil {
		panic("Device gRPC handler requires a typed Device service")
	}
	return &DeviceRPCHandler{service: service}
}

func (handler *DeviceRPCHandler) ListDevices(
	ctx context.Context,
	request *dcimv1.ListDevicesRequest,
) (*dcimv1.ListDevicesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListDevices(ctx, p, typedDeviceListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.Device, 0, len(page.Results))
	for _, device := range page.Results {
		results = append(results, typedDeviceProto(device))
	}
	return &dcimv1.ListDevicesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *DeviceRPCHandler) GetDevice(
	ctx context.Context,
	request *dcimv1.GetDeviceRequest,
) (*dcimv1.GetDeviceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	device, err := handler.service.GetDevice(
		ctx, p, applicationdcim.GetDeviceQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetDeviceResponse{Device: typedDeviceProto(device)}, nil
}

func (handler *DeviceRPCHandler) CreateDevice(
	ctx context.Context,
	request *dcimv1.CreateDeviceRequest,
) (*dcimv1.CreateDeviceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	device, err := handler.service.CreateDevice(
		ctx, p, typedDeviceCreateCommand(request.GetDevice()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateDeviceResponse{Device: typedDeviceProto(device)}, nil
}

func (handler *DeviceRPCHandler) ReplaceDevice(
	ctx context.Context,
	request *dcimv1.ReplaceDeviceRequest,
) (*dcimv1.ReplaceDeviceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	device, err := handler.service.ReplaceDevice(
		ctx, p, typedDeviceReplaceCommand(shared.ID(request.GetId()), request.GetDevice()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceDeviceResponse{Device: typedDeviceProto(device)}, nil
}

func (handler *DeviceRPCHandler) UpdateDevice(
	ctx context.Context,
	request *dcimv1.UpdateDeviceRequest,
) (*dcimv1.UpdateDeviceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedDeviceUpdateCommand(
		shared.ID(request.GetId()), request.GetDevice(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	device, err := handler.service.UpdateDevice(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateDeviceResponse{Device: typedDeviceProto(device)}, nil
}

func (handler *DeviceRPCHandler) DeleteDevice(
	ctx context.Context,
	request *dcimv1.DeleteDeviceRequest,
) (*dcimv1.DeleteDeviceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteDevice(
		ctx, p, applicationdcim.DeleteDeviceCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteDeviceResponse{}, nil
}

func typedDeviceListQuery(
	request *dcimv1.ListDevicesRequest,
) applicationdcim.ListDevicesQuery {
	query := applicationdcim.ListDevicesQuery{}
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
	if request.SiteId != nil {
		query.SiteIDs = []int64{*request.SiteId}
	}
	query.SiteSlugs = oneString(request.SiteSlug)
	if request.RackId != nil {
		query.RackIDs = []int64{*request.RackId}
	}
	if request.DeviceTypeId != nil {
		query.DeviceTypeIDs = []int64{*request.DeviceTypeId}
	}
	query.DeviceTypeSlugs = oneString(request.DeviceTypeSlug)
	if request.RoleId != nil {
		query.RoleIDs = []int64{*request.RoleId}
	}
	query.RoleSlugs = oneString(request.RoleSlug)
	query.Names = oneString(request.Name)
	query.Statuses = oneString(request.Status)
	return query
}

type typedDeviceFields struct {
	deviceType  applicationdcim.Field[shared.ID]
	role        applicationdcim.Field[shared.ID]
	name        applicationdcim.Field[string]
	site        applicationdcim.Field[shared.ID]
	rack        applicationdcim.Field[shared.ID]
	position    applicationdcim.Field[string]
	face        applicationdcim.Field[string]
	status      applicationdcim.Field[string]
	serial      applicationdcim.Field[string]
	assetTag    applicationdcim.Field[string]
	airflow     applicationdcim.Field[string]
	description applicationdcim.Field[string]
	comments    applicationdcim.Field[string]
}

func typedDeviceInputFields(input *dcimv1.DeviceInput) typedDeviceFields {
	if input == nil {
		return typedDeviceFields{}
	}
	return typedDeviceFields{
		deviceType:  typedRackTypeProtoID(input.DeviceType),
		role:        typedRackTypeProtoID(input.Role),
		name:        typedDeviceWrappedString(input.Name),
		site:        typedRackTypeProtoID(input.Site),
		rack:        typedDeviceWrappedID(input.Rack),
		position:    typedDeviceWrappedString(input.Position),
		face:        typedDeviceWrappedString(input.Face),
		status:      typedProtoString(input.Status),
		serial:      typedProtoString(input.Serial),
		assetTag:    typedDeviceWrappedString(input.AssetTag),
		airflow:     typedProtoString(input.Airflow),
		description: typedProtoString(input.Description),
		comments:    typedProtoString(input.Comments),
	}
}

func typedDeviceWrappedString(
	value *wrapperspb.StringValue,
) applicationdcim.Field[string] {
	if value == nil {
		return applicationdcim.OmittedField[string]()
	}
	return applicationdcim.FieldValue(value.Value)
}

func typedDeviceWrappedID(
	value *wrapperspb.Int64Value,
) applicationdcim.Field[shared.ID] {
	if value == nil {
		return applicationdcim.OmittedField[shared.ID]()
	}
	return applicationdcim.FieldValue(shared.ID(value.Value))
}

func typedDeviceCreateCommand(
	input *dcimv1.DeviceInput,
) applicationdcim.CreateDeviceCommand {
	fields := typedDeviceInputFields(input)
	return applicationdcim.CreateDeviceCommand{
		DeviceType: fields.deviceType, Role: fields.role, Name: fields.name,
		Site: fields.site, Rack: fields.rack, Position: fields.position,
		Face: fields.face, Status: fields.status, Serial: fields.serial,
		AssetTag: fields.assetTag, Airflow: fields.airflow,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedDeviceReplaceCommand(
	id shared.ID,
	input *dcimv1.DeviceInput,
) applicationdcim.ReplaceDeviceCommand {
	return applicationdcim.ReplaceDeviceCommand{
		ID: id, CreateDeviceCommand: typedDeviceCreateCommand(input),
	}
}

func typedDeviceUpdateCommand(
	id shared.ID,
	input *dcimv1.DeviceInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateDeviceCommand, error) {
	fields := typedDeviceInputFields(input)
	command := applicationdcim.UpdateDeviceCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.DeviceType, command.Role, command.Name = fields.deviceType, fields.role, fields.name
		command.Site, command.Rack, command.Position = fields.site, fields.rack, fields.position
		command.Face, command.Status, command.Serial = fields.face, fields.status, fields.serial
		command.AssetTag, command.Airflow = fields.assetTag, fields.airflow
		command.Description, command.Comments = fields.description, fields.comments
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "device_type":
			if fields.deviceType.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.DeviceType = fields.deviceType
		case "role":
			if fields.role.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Role = fields.role
		case "name":
			command.Name = nullableTypedDeviceField(fields.name)
		case "site":
			if fields.site.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Site = fields.site
		case "rack":
			command.Rack = nullableTypedDeviceField(fields.rack)
		case "position":
			command.Position = nullableTypedDeviceField(fields.position)
		case "face":
			command.Face = nullableTypedDeviceField(fields.face)
		case "status":
			if fields.status.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Status = fields.status
		case "serial":
			if fields.serial.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Serial = fields.serial
		case "asset_tag":
			command.AssetTag = nullableTypedDeviceField(fields.assetTag)
		case "airflow":
			command.Airflow = nullableTypedDeviceField(fields.airflow)
		case "description":
			if fields.description.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Description = fields.description
		case "comments":
			if fields.comments.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
			}
			command.Comments = fields.comments
		default:
			return applicationdcim.UpdateDeviceCommand{}, invalidTypedDeviceMask()
		}
	}
	return command, nil
}

func nullableTypedDeviceField[T any](
	field applicationdcim.Field[T],
) applicationdcim.Field[T] {
	if field.State() == applicationdcim.FieldPresent {
		return field
	}
	return applicationdcim.NullField[T]()
}

func invalidTypedDeviceMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value, or a nullable field to clear.",
		},
	)
}

func typedDeviceProto(device *domaindcim.Device) *dcimv1.Device {
	if device == nil {
		return nil
	}
	deviceType := device.DeviceType()
	role := device.Role()
	site := device.Site()
	out := &dcimv1.Device{
		Id: device.ID().Int64(), Url: deviceRPCURL(device.ID()),
		Display: device.Display(),
		DeviceType: &typesv1.ObjectReference{
			Id:      deviceType.ID().Int64(),
			Url:     "/api/dcim/device-types/" + deviceType.ID().String() + "/",
			Display: deviceType.Display(),
		},
		Role: &typesv1.ObjectReference{
			Id:      role.ID.Int64(),
			Url:     "/api/dcim/device-roles/" + role.ID.String() + "/",
			Display: role.Display,
		},
		Site: &typesv1.ObjectReference{
			Id:      site.ID().Int64(),
			Url:     "/api/dcim/sites/" + site.ID().String() + "/",
			Display: site.Display(),
		},
		Face: device.Face().String(), Status: device.Status().String(),
		Serial: device.Serial(), Description: device.Description(), Comments: device.Comments(),
		Created:        timestamppb.New(device.Created().Time),
		LastUpdated:    timestamppb.New(device.LastUpdated().Time),
		InterfaceCount: device.InterfaceCount(),
	}
	if value, present := device.Name().Get(); present {
		out.Name = value
	}
	if value, present := device.Rack().Get(); present {
		out.RackId = wrapperspb.Int64(value.ID().Int64())
	}
	if value, present := device.Position().Get(); present {
		out.Position = value.String()
	}
	if value, present := device.AssetTag().Get(); present {
		out.AssetTag = value
	}
	if value, present := device.Airflow().Get(); present {
		out.Airflow = value.String()
	}
	return out
}
