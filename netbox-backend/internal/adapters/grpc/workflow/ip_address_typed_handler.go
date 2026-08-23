package workflow

import (
	"context"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	ipamv1 "netbox-go/gen/go/netbox/ipam/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	applicationipam "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/identity"
	domainipam "netbox-go/internal/domain/ipam"
	"netbox-go/internal/domain/shared"
)

type IPAMIPAddressService interface {
	ListIPAddresses(
		context.Context,
		identity.Principal,
		applicationipam.ListIPAddressesQuery,
	) (applicationipam.IPAddressPage, error)
	GetIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.GetIPAddressQuery,
	) (*domainipam.IPAddress, error)
	CreateIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.CreateIPAddressCommand,
	) (*domainipam.IPAddress, error)
	ReplaceIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.ReplaceIPAddressCommand,
	) (*domainipam.IPAddress, error)
	UpdateIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.UpdateIPAddressCommand,
	) (*domainipam.IPAddress, error)
	DeleteIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.DeleteIPAddressCommand,
	) error
	AssignIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.AssignIPAddressCommand,
	) (*domainipam.IPAddress, error)
	UnassignIPAddress(
		context.Context,
		identity.Principal,
		applicationipam.UnassignIPAddressCommand,
	) (*domainipam.IPAddress, error)
}

type IPAddressRPCHandler struct{ service IPAMIPAddressService }

func NewIPAddressRPCHandler(
	service IPAMIPAddressService,
) *IPAddressRPCHandler {
	if service == nil {
		panic("IPAddress gRPC handler requires a typed IPAddress service")
	}
	return &IPAddressRPCHandler{service: service}
}

func (handler *IPAddressRPCHandler) ListIPAddresses(
	ctx context.Context,
	request *ipamv1.ListIPAddressesRequest,
) (*ipamv1.ListIPAddressesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListIPAddresses(
		ctx, p, typedIPAddressListQuery(request),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*ipamv1.IPAddress, 0, len(page.Results))
	for _, address := range page.Results {
		results = append(results, typedIPAddressProto(address))
	}
	return &ipamv1.ListIPAddressesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *IPAddressRPCHandler) GetIPAddress(
	ctx context.Context,
	request *ipamv1.GetIPAddressRequest,
) (*ipamv1.GetIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.GetIPAddress(
		ctx,
		p,
		applicationipam.GetIPAddressQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.GetIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func (handler *IPAddressRPCHandler) CreateIPAddress(
	ctx context.Context,
	request *ipamv1.CreateIPAddressRequest,
) (*ipamv1.CreateIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.CreateIPAddress(
		ctx, p, typedIPAddressCreateCommand(request.GetIpAddress()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.CreateIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func (handler *IPAddressRPCHandler) ReplaceIPAddress(
	ctx context.Context,
	request *ipamv1.ReplaceIPAddressRequest,
) (*ipamv1.ReplaceIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.ReplaceIPAddress(
		ctx,
		p,
		typedIPAddressReplaceCommand(
			shared.ID(request.GetId()), request.GetIpAddress(),
		),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.ReplaceIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func (handler *IPAddressRPCHandler) UpdateIPAddress(
	ctx context.Context,
	request *ipamv1.UpdateIPAddressRequest,
) (*ipamv1.UpdateIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedIPAddressUpdateCommand(
		shared.ID(request.GetId()),
		request.GetIpAddress(),
		request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.UpdateIPAddress(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.UpdateIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func (handler *IPAddressRPCHandler) DeleteIPAddress(
	ctx context.Context,
	request *ipamv1.DeleteIPAddressRequest,
) (*ipamv1.DeleteIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteIPAddress(
		ctx,
		p,
		applicationipam.DeleteIPAddressCommand{
			ID: shared.ID(request.GetId()),
		},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.DeleteIPAddressResponse{}, nil
}

func (handler *IPAddressRPCHandler) AssignIPAddress(
	ctx context.Context,
	request *ipamv1.AssignIPAddressRequest,
) (*ipamv1.AssignIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.AssignIPAddress(
		ctx,
		p,
		applicationipam.AssignIPAddressCommand{
			ID:          shared.ID(request.GetId()),
			InterfaceID: shared.ID(request.GetInterfaceId()),
		},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.AssignIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func (handler *IPAddressRPCHandler) UnassignIPAddress(
	ctx context.Context,
	request *ipamv1.UnassignIPAddressRequest,
) (*ipamv1.UnassignIPAddressResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	address, err := handler.service.UnassignIPAddress(
		ctx,
		p,
		applicationipam.UnassignIPAddressCommand{
			ID: shared.ID(request.GetId()),
		},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.UnassignIPAddressResponse{
		IpAddress: typedIPAddressProto(address),
	}, nil
}

func typedIPAddressListQuery(
	request *ipamv1.ListIPAddressesRequest,
) applicationipam.ListIPAddressesQuery {
	query := applicationipam.ListIPAddressesQuery{}
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
	query.VRFIDs = grpcOptionalInt64(request.VrfId)
	query.VRFRDs = grpcOptionalString(request.VrfRd)
	query.Addresses = grpcOptionalString(request.Address)
	if request.Family != nil {
		value := int64(*request.Family)
		query.Family = &value
	}
	query.Parent = cloneStringPointer(request.Parent)
	query.Statuses = grpcOptionalString(request.Status)
	query.Assigned = cloneBoolPointer(request.Assigned)
	query.InterfaceIDs = grpcOptionalInt64(request.InterfaceId)
	query.DeviceIDs = grpcOptionalInt64(request.DeviceId)
	return query
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

type typedIPAddressFields struct {
	address            applicationipam.Field[string]
	vrf                applicationipam.Field[int64]
	status             applicationipam.Field[string]
	role               applicationipam.Field[string]
	dnsName            applicationipam.Field[string]
	description        applicationipam.Field[string]
	comments           applicationipam.Field[string]
	assignedObjectType applicationipam.Field[string]
	assignedObjectID   applicationipam.Field[int64]
}

func typedIPAddressInputFields(
	input *ipamv1.IPAddressInput,
) typedIPAddressFields {
	if input == nil {
		return typedIPAddressFields{}
	}
	return typedIPAddressFields{
		address:     typedVRFProtoString(input.Address),
		vrf:         typedIPAddressWrappedInt64(input.Vrf),
		status:      typedVRFProtoString(input.Status),
		role:        typedIPAddressWrappedString(input.Role),
		dnsName:     typedVRFProtoString(input.DnsName),
		description: typedVRFProtoString(input.Description),
		comments:    typedVRFProtoString(input.Comments),
		assignedObjectType: typedIPAddressWrappedString(
			input.AssignedObjectType,
		),
		assignedObjectID: typedIPAddressWrappedInt64(
			input.AssignedObjectId,
		),
	}
}

func typedIPAddressWrappedString(
	value *wrapperspb.StringValue,
) applicationipam.Field[string] {
	if value == nil {
		return applicationipam.OmittedField[string]()
	}
	return applicationipam.FieldValue(value.Value)
}

func typedIPAddressWrappedInt64(
	value *wrapperspb.Int64Value,
) applicationipam.Field[int64] {
	if value == nil {
		return applicationipam.OmittedField[int64]()
	}
	return applicationipam.FieldValue(value.Value)
}

func typedIPAddressCreateCommand(
	input *ipamv1.IPAddressInput,
) applicationipam.CreateIPAddressCommand {
	fields := typedIPAddressInputFields(input)
	return applicationipam.CreateIPAddressCommand{
		Address: fields.address, VRF: fields.vrf, Status: fields.status,
		Role: fields.role, DNSName: fields.dnsName,
		Description: fields.description, Comments: fields.comments,
		AssignedObjectType: fields.assignedObjectType,
		AssignedObjectID:   fields.assignedObjectID,
	}
}

func typedIPAddressReplaceCommand(
	id shared.ID,
	input *ipamv1.IPAddressInput,
) applicationipam.ReplaceIPAddressCommand {
	return applicationipam.ReplaceIPAddressCommand{
		ID: id, CreateIPAddressCommand: typedIPAddressCreateCommand(input),
	}
}

func typedIPAddressUpdateCommand(
	id shared.ID,
	input *ipamv1.IPAddressInput,
	mask *fieldmaskpb.FieldMask,
) (applicationipam.UpdateIPAddressCommand, error) {
	fields := typedIPAddressInputFields(input)
	command := applicationipam.UpdateIPAddressCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Address, command.VRF, command.Status =
			fields.address, fields.vrf, fields.status
		command.Role, command.DNSName = fields.role, fields.dnsName
		command.Description, command.Comments =
			fields.description, fields.comments
		command.AssignedObjectType = fields.assignedObjectType
		command.AssignedObjectID = fields.assignedObjectID
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "address":
			if input == nil || input.Address == nil {
				command.Address = applicationipam.NullField[string]()
			} else {
				command.Address = fields.address
			}
		case "vrf":
			if input == nil || input.Vrf == nil {
				command.VRF = applicationipam.NullField[int64]()
			} else {
				command.VRF = fields.vrf
			}
		case "status":
			if input == nil || input.Status == nil {
				command.Status = applicationipam.NullField[string]()
			} else {
				command.Status = fields.status
			}
		case "role":
			if input == nil || input.Role == nil {
				command.Role = applicationipam.NullField[string]()
			} else {
				command.Role = fields.role
			}
		case "dns_name":
			if input == nil || input.DnsName == nil {
				command.DNSName = applicationipam.NullField[string]()
			} else {
				command.DNSName = fields.dnsName
			}
		case "description":
			if input == nil || input.Description == nil {
				command.Description = applicationipam.NullField[string]()
			} else {
				command.Description = fields.description
			}
		case "comments":
			if input == nil || input.Comments == nil {
				command.Comments = applicationipam.NullField[string]()
			} else {
				command.Comments = fields.comments
			}
		case "assigned_object_type":
			if input == nil || input.AssignedObjectType == nil {
				command.AssignedObjectType =
					applicationipam.NullField[string]()
			} else {
				command.AssignedObjectType = fields.assignedObjectType
			}
		case "assigned_object_id":
			if input == nil || input.AssignedObjectId == nil {
				command.AssignedObjectID =
					applicationipam.NullField[int64]()
			} else {
				command.AssignedObjectID = fields.assignedObjectID
			}
		default:
			return applicationipam.UpdateIPAddressCommand{},
				invalidIPAddressUpdateMask()
		}
	}
	return command, nil
}

func invalidIPAddressUpdateMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value.",
		},
	)
}

func typedIPAddressProto(
	address *domainipam.IPAddress,
) *ipamv1.IPAddress {
	if address == nil {
		return nil
	}
	result := &ipamv1.IPAddress{
		Id:      address.ID().Int64(),
		Url:     "/api/ipam/ip-addresses/" + address.ID().String() + "/",
		Display: address.Display(), Address: address.Address().String(),
		VrfId:   typedPrefixVRFID(address.VRF()),
		Status:  address.Status().String(),
		Role:    typedIPAddressRole(address.Role()),
		DnsName: address.DNSName(), Description: address.Description(),
		Comments:    address.Comments(),
		Created:     timestamppb.New(address.Created().Time),
		LastUpdated: timestamppb.New(address.LastUpdated().Time),
		Family:      address.Family(),
	}
	if assignment, present := address.Assignment().Get(); present {
		result.AssignedObjectType = wrapperspb.String(
			domainipam.IPAddressAssignmentType,
		)
		result.AssignedObjectId = wrapperspb.Int64(assignment.ID().Int64())
		result.AssignedObject = &typesv1.ObjectReference{
			Id:      assignment.ID().Int64(),
			Url:     "/api/dcim/interfaces/" + assignment.ID().String() + "/",
			Display: assignment.Display(),
		}
	}
	return result
}

func typedIPAddressRole(
	nullable domainipam.NullableIPAddressRole,
) *wrapperspb.StringValue {
	role, present := nullable.Get()
	if !present || role.String() == "" {
		return nil
	}
	return wrapperspb.String(role.String())
}

var _ IPAMIPAddressService = (*applicationipam.IPAddressService)(nil)
