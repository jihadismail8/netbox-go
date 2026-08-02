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

type DCIMInterfaceService interface {
	ListInterfaces(
		context.Context,
		identity.Principal,
		applicationdcim.ListInterfacesQuery,
	) (applicationdcim.InterfacePage, error)
	GetInterface(
		context.Context,
		identity.Principal,
		applicationdcim.GetInterfaceQuery,
	) (*domaindcim.Interface, error)
	CreateInterface(
		context.Context,
		identity.Principal,
		applicationdcim.CreateInterfaceCommand,
	) (*domaindcim.Interface, error)
	ReplaceInterface(
		context.Context,
		identity.Principal,
		applicationdcim.ReplaceInterfaceCommand,
	) (*domaindcim.Interface, error)
	UpdateInterface(
		context.Context,
		identity.Principal,
		applicationdcim.UpdateInterfaceCommand,
	) (*domaindcim.Interface, error)
	DeleteInterface(
		context.Context,
		identity.Principal,
		applicationdcim.DeleteInterfaceCommand,
	) error
}

var _ DCIMInterfaceService = (*applicationdcim.InterfaceService)(nil)

type InterfaceRPCHandler struct {
	service DCIMInterfaceService
}

func NewInterfaceRPCHandler(service DCIMInterfaceService) *InterfaceRPCHandler {
	if service == nil {
		panic("Interface gRPC handler requires a typed Interface service")
	}
	return &InterfaceRPCHandler{service: service}
}

func (handler *InterfaceRPCHandler) ListInterfaces(
	ctx context.Context,
	request *dcimv1.ListInterfacesRequest,
) (*dcimv1.ListInterfacesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListInterfaces(ctx, p, typedInterfaceListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.Interface, 0, len(page.Results))
	for _, networkInterface := range page.Results {
		results = append(results, typedInterfaceProto(networkInterface))
	}
	return &dcimv1.ListInterfacesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *InterfaceRPCHandler) GetInterface(
	ctx context.Context,
	request *dcimv1.GetInterfaceRequest,
) (*dcimv1.GetInterfaceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	networkInterface, err := handler.service.GetInterface(
		ctx, p, applicationdcim.GetInterfaceQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetInterfaceResponse{
		Interface: typedInterfaceProto(networkInterface),
	}, nil
}

func (handler *InterfaceRPCHandler) CreateInterface(
	ctx context.Context,
	request *dcimv1.CreateInterfaceRequest,
) (*dcimv1.CreateInterfaceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	networkInterface, err := handler.service.CreateInterface(
		ctx, p, typedInterfaceCreateCommand(request.GetInterface()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateInterfaceResponse{
		Interface: typedInterfaceProto(networkInterface),
	}, nil
}

func (handler *InterfaceRPCHandler) ReplaceInterface(
	ctx context.Context,
	request *dcimv1.ReplaceInterfaceRequest,
) (*dcimv1.ReplaceInterfaceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	networkInterface, err := handler.service.ReplaceInterface(
		ctx, p, applicationdcim.ReplaceInterfaceCommand{
			ID:                     shared.ID(request.GetId()),
			CreateInterfaceCommand: typedInterfaceCreateCommand(request.GetInterface()),
		},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceInterfaceResponse{
		Interface: typedInterfaceProto(networkInterface),
	}, nil
}

func (handler *InterfaceRPCHandler) UpdateInterface(
	ctx context.Context,
	request *dcimv1.UpdateInterfaceRequest,
) (*dcimv1.UpdateInterfaceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedInterfaceUpdateCommand(
		shared.ID(request.GetId()), request.GetInterface(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	networkInterface, err := handler.service.UpdateInterface(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateInterfaceResponse{
		Interface: typedInterfaceProto(networkInterface),
	}, nil
}

func (handler *InterfaceRPCHandler) DeleteInterface(
	ctx context.Context,
	request *dcimv1.DeleteInterfaceRequest,
) (*dcimv1.DeleteInterfaceResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteInterface(
		ctx, p, applicationdcim.DeleteInterfaceCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteInterfaceResponse{}, nil
}

func typedInterfaceListQuery(
	request *dcimv1.ListInterfacesRequest,
) applicationdcim.ListInterfacesQuery {
	query := applicationdcim.ListInterfacesQuery{}
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
	if request.DeviceId != nil {
		query.DeviceIDs = []int64{*request.DeviceId}
	}
	query.DeviceNames = oneString(request.DeviceName)
	query.Names = oneString(request.Name)
	query.Types = oneString(request.Type)
	query.Enabled = copyProtoBool(request.Enabled)
	query.MgmtOnly = copyProtoBool(request.MgmtOnly)
	return query
}

type typedInterfaceFields struct {
	device        applicationdcim.Field[shared.ID]
	name          applicationdcim.Field[string]
	label         applicationdcim.Field[string]
	interfaceType applicationdcim.Field[string]
	enabled       applicationdcim.Field[bool]
	mgmtOnly      applicationdcim.Field[bool]
	mtu           applicationdcim.Field[uint32]
	speed         applicationdcim.Field[uint64]
	duplex        applicationdcim.Field[string]
	description   applicationdcim.Field[string]
}

func typedInterfaceInputFields(input *dcimv1.InterfaceInput) typedInterfaceFields {
	if input == nil {
		return typedInterfaceFields{}
	}
	return typedInterfaceFields{
		device: typedRackTypeProtoID(input.Device), name: typedProtoString(input.Name),
		label: typedProtoString(input.Label), interfaceType: typedProtoString(input.Type),
		enabled:  typedRackTypeProtoBool(input.Enabled),
		mgmtOnly: typedRackTypeProtoBool(input.MgmtOnly),
		mtu:      typedInterfaceProtoMTU(input.Mtu), speed: typedInterfaceProtoSpeed(input.Speed),
		duplex:      typedInterfaceProtoDuplex(input.Duplex),
		description: typedProtoString(input.Description),
	}
}

func typedInterfaceProtoMTU(
	value *wrapperspb.Int32Value,
) applicationdcim.Field[uint32] {
	if value == nil {
		return applicationdcim.OmittedField[uint32]()
	}
	return applicationdcim.FieldValue(uint32(value.Value))
}

func typedInterfaceProtoSpeed(
	value *wrapperspb.Int64Value,
) applicationdcim.Field[uint64] {
	if value == nil {
		return applicationdcim.OmittedField[uint64]()
	}
	return applicationdcim.FieldValue(uint64(value.Value))
}

func typedInterfaceProtoDuplex(
	value *wrapperspb.StringValue,
) applicationdcim.Field[string] {
	if value == nil {
		return applicationdcim.OmittedField[string]()
	}
	return applicationdcim.FieldValue(value.Value)
}

func typedInterfaceCreateCommand(
	input *dcimv1.InterfaceInput,
) applicationdcim.CreateInterfaceCommand {
	fields := typedInterfaceInputFields(input)
	return applicationdcim.CreateInterfaceCommand{
		Device: fields.device, Name: fields.name, Label: fields.label,
		Type: fields.interfaceType, Enabled: fields.enabled, MgmtOnly: fields.mgmtOnly,
		MTU: fields.mtu, Speed: fields.speed, Duplex: fields.duplex,
		Description: fields.description,
	}
}

func typedInterfaceUpdateCommand(
	id shared.ID,
	input *dcimv1.InterfaceInput,
	mask *fieldmaskpb.FieldMask,
) (applicationdcim.UpdateInterfaceCommand, error) {
	fields := typedInterfaceInputFields(input)
	command := applicationdcim.UpdateInterfaceCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Device = fields.device
		command.Name = fields.name
		command.Label = fields.label
		command.Type = fields.interfaceType
		command.Enabled = fields.enabled
		command.MgmtOnly = fields.mgmtOnly
		command.MTU = fields.mtu
		command.Speed = fields.speed
		command.Duplex = fields.duplex
		command.Description = fields.description
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "device":
			if fields.device.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Device = fields.device
		case "name":
			if fields.name.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Name = fields.name
		case "label":
			if fields.label.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Label = fields.label
		case "type":
			if fields.interfaceType.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Type = fields.interfaceType
		case "enabled":
			if fields.enabled.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Enabled = fields.enabled
		case "mgmt_only":
			if fields.mgmtOnly.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.MgmtOnly = fields.mgmtOnly
		case "mtu":
			command.MTU = nullableMaskedInterfaceField(fields.mtu)
		case "speed":
			command.Speed = nullableMaskedInterfaceField(fields.speed)
		case "duplex":
			command.Duplex = nullableMaskedInterfaceField(fields.duplex)
		case "description":
			if fields.description.State() != applicationdcim.FieldPresent {
				return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
			}
			command.Description = fields.description
		default:
			return applicationdcim.UpdateInterfaceCommand{}, invalidTypedInterfaceMask()
		}
	}
	return command, nil
}

func nullableMaskedInterfaceField[T any](
	field applicationdcim.Field[T],
) applicationdcim.Field[T] {
	if field.State() == applicationdcim.FieldOmitted {
		return applicationdcim.NullField[T]()
	}
	return field
}

func invalidTypedInterfaceMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value.",
		},
	)
}

func typedInterfaceProto(
	networkInterface *domaindcim.Interface,
) *dcimv1.Interface {
	if networkInterface == nil {
		return nil
	}
	device := networkInterface.Device()
	var mtu *wrapperspb.Int32Value
	if value, present := networkInterface.MTU().Get(); present {
		mtu = wrapperspb.Int32(int32(value))
	}
	var speed *wrapperspb.Int64Value
	if value, present := networkInterface.Speed().Get(); present {
		speed = wrapperspb.Int64(int64(value))
	}
	var duplex *wrapperspb.StringValue
	if value, present := networkInterface.Duplex().Get(); present {
		duplex = wrapperspb.String(value.String())
	}
	return &dcimv1.Interface{
		Id: networkInterface.ID().Int64(), Url: interfaceRPCURL(networkInterface.ID()),
		Display: networkInterface.Display(),
		Device: &typesv1.ObjectReference{
			Id: device.ID().Int64(), Url: deviceRPCURL(device.ID()),
			Display: device.Display(),
		},
		Name: networkInterface.Name(), Label: networkInterface.Label(),
		Type: networkInterface.Type().String(), Enabled: networkInterface.Enabled(),
		MgmtOnly: networkInterface.MgmtOnly(), Mtu: mtu, Speed: speed, Duplex: duplex,
		Description:      networkInterface.Description(),
		Created:          timestamppb.New(networkInterface.Created().Time),
		LastUpdated:      timestamppb.New(networkInterface.LastUpdated().Time),
		CountIpaddresses: networkInterface.IPAddressCount(),
	}
}

func interfaceRPCURL(id shared.ID) string {
	return "/api/dcim/interfaces/" + id.String() + "/"
}

func deviceRPCURL(id shared.ID) string {
	return "/api/dcim/devices/" + id.String() + "/"
}
