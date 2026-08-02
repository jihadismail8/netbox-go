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

// IPAMVRFService is the typed VRF application contract.
type IPAMVRFService interface {
	ListVRFs(context.Context, identity.Principal, applicationipam.ListVRFsQuery) (applicationipam.VRFPage, error)
	GetVRF(context.Context, identity.Principal, applicationipam.GetVRFQuery) (*domainipam.VRF, error)
	CreateVRF(context.Context, identity.Principal, applicationipam.CreateVRFCommand) (*domainipam.VRF, error)
	ReplaceVRF(context.Context, identity.Principal, applicationipam.ReplaceVRFCommand) (*domainipam.VRF, error)
	UpdateVRF(context.Context, identity.Principal, applicationipam.UpdateVRFCommand) (*domainipam.VRF, error)
	DeleteVRF(context.Context, identity.Principal, applicationipam.DeleteVRFCommand) error
}

// VRFRPCHandler implements the six VRF RPCs delegated by IPAMServer.
type VRFRPCHandler struct {
	service IPAMVRFService
}

func NewVRFRPCHandler(service IPAMVRFService) *VRFRPCHandler {
	if service == nil {
		panic("VRF gRPC handler requires a typed VRF service")
	}
	return &VRFRPCHandler{service: service}
}

func (handler *VRFRPCHandler) ListVRFs(
	ctx context.Context,
	request *ipamv1.ListVRFsRequest,
) (*ipamv1.ListVRFsResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListVRFs(ctx, p, typedVRFListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*ipamv1.VRF, 0, len(page.Results))
	for _, vrf := range page.Results {
		results = append(results, typedVRFProto(vrf))
	}
	return &ipamv1.ListVRFsResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *VRFRPCHandler) GetVRF(
	ctx context.Context,
	request *ipamv1.GetVRFRequest,
) (*ipamv1.GetVRFResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	vrf, err := handler.service.GetVRF(
		ctx, p, applicationipam.GetVRFQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.GetVRFResponse{Vrf: typedVRFProto(vrf)}, nil
}

func (handler *VRFRPCHandler) CreateVRF(
	ctx context.Context,
	request *ipamv1.CreateVRFRequest,
) (*ipamv1.CreateVRFResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	vrf, err := handler.service.CreateVRF(ctx, p, typedVRFCreateCommand(request.GetVrf()))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.CreateVRFResponse{Vrf: typedVRFProto(vrf)}, nil
}

func (handler *VRFRPCHandler) ReplaceVRF(
	ctx context.Context,
	request *ipamv1.ReplaceVRFRequest,
) (*ipamv1.ReplaceVRFResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	vrf, err := handler.service.ReplaceVRF(
		ctx, p, typedVRFReplaceCommand(shared.ID(request.GetId()), request.GetVrf()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.ReplaceVRFResponse{Vrf: typedVRFProto(vrf)}, nil
}

func (handler *VRFRPCHandler) UpdateVRF(
	ctx context.Context,
	request *ipamv1.UpdateVRFRequest,
) (*ipamv1.UpdateVRFResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedVRFUpdateCommand(
		shared.ID(request.GetId()), request.GetVrf(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	vrf, err := handler.service.UpdateVRF(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.UpdateVRFResponse{Vrf: typedVRFProto(vrf)}, nil
}

func (handler *VRFRPCHandler) DeleteVRF(
	ctx context.Context,
	request *ipamv1.DeleteVRFRequest,
) (*ipamv1.DeleteVRFResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeleteVRF(
		ctx, p, applicationipam.DeleteVRFCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.DeleteVRFResponse{}, nil
}

func typedVRFListQuery(request *ipamv1.ListVRFsRequest) applicationipam.ListVRFsQuery {
	query := applicationipam.ListVRFsQuery{}
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
	query.Names = grpcOptionalString(request.Name)
	query.RDs = grpcOptionalString(request.Rd)
	if request.EnforceUnique != nil {
		value := *request.EnforceUnique
		query.EnforceUnique = &value
	}
	return query
}

func grpcOptionalString(value *string) []string {
	if value == nil {
		return nil
	}
	return []string{*value}
}

func typedVRFCreateCommand(input *ipamv1.VRFInput) applicationipam.CreateVRFCommand {
	fields := typedVRFInputFields(input)
	return applicationipam.CreateVRFCommand{
		Name: fields.name, RD: fields.rd, EnforceUnique: fields.enforceUnique,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedVRFReplaceCommand(id shared.ID, input *ipamv1.VRFInput) applicationipam.ReplaceVRFCommand {
	fields := typedVRFInputFields(input)
	return applicationipam.ReplaceVRFCommand{
		ID: id, Name: fields.name, RD: fields.rd, EnforceUnique: fields.enforceUnique,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedVRFUpdateCommand(
	id shared.ID,
	input *ipamv1.VRFInput,
	mask *fieldmaskpb.FieldMask,
) (applicationipam.UpdateVRFCommand, error) {
	fields := typedVRFInputFields(input)
	command := applicationipam.UpdateVRFCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Name = fields.name
		command.RD = fields.rd
		command.EnforceUnique = fields.enforceUnique
		command.Description = fields.description
		command.Comments = fields.comments
		return command, nil
	}

	for _, path := range mask.Paths {
		switch path {
		case "name":
			if fields.name.State() != applicationipam.FieldPresent {
				return applicationipam.UpdateVRFCommand{}, invalidVRFUpdateMask()
			}
			command.Name = fields.name
		case "rd":
			// A nil wrapper plus an explicit rd mask is protobuf's representation
			// of clearing this nullable field.
			if input == nil || input.Rd == nil {
				command.RD = applicationipam.NullField[string]()
			} else {
				command.RD = fields.rd
			}
		case "enforce_unique":
			if fields.enforceUnique.State() != applicationipam.FieldPresent {
				return applicationipam.UpdateVRFCommand{}, invalidVRFUpdateMask()
			}
			command.EnforceUnique = fields.enforceUnique
		case "description":
			if fields.description.State() != applicationipam.FieldPresent {
				return applicationipam.UpdateVRFCommand{}, invalidVRFUpdateMask()
			}
			command.Description = fields.description
		case "comments":
			if fields.comments.State() != applicationipam.FieldPresent {
				return applicationipam.UpdateVRFCommand{}, invalidVRFUpdateMask()
			}
			command.Comments = fields.comments
		default:
			return applicationipam.UpdateVRFCommand{}, invalidVRFUpdateMask()
		}
	}
	return command, nil
}

func invalidVRFUpdateMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value.",
		},
	)
}

type typedVRFFields struct {
	name          applicationipam.Field[string]
	rd            applicationipam.Field[string]
	enforceUnique applicationipam.Field[bool]
	description   applicationipam.Field[string]
	comments      applicationipam.Field[string]
}

func typedVRFInputFields(input *ipamv1.VRFInput) typedVRFFields {
	if input == nil {
		return typedVRFFields{}
	}
	return typedVRFFields{
		name:          typedVRFProtoString(input.Name),
		rd:            typedVRFProtoWrappedString(input.Rd),
		enforceUnique: typedVRFProtoBool(input.EnforceUnique),
		description:   typedVRFProtoString(input.Description),
		comments:      typedVRFProtoString(input.Comments),
	}
}

func typedVRFProtoString(value *string) applicationipam.Field[string] {
	if value == nil {
		return applicationipam.OmittedField[string]()
	}
	return applicationipam.FieldValue(*value)
}

func typedVRFProtoWrappedString(value *wrapperspb.StringValue) applicationipam.Field[string] {
	if value == nil {
		return applicationipam.OmittedField[string]()
	}
	return applicationipam.FieldValue(value.Value)
}

func typedVRFProtoBool(value *bool) applicationipam.Field[bool] {
	if value == nil {
		return applicationipam.OmittedField[bool]()
	}
	return applicationipam.FieldValue(*value)
}

func typedVRFProto(vrf *domainipam.VRF) *ipamv1.VRF {
	if vrf == nil {
		return nil
	}
	return &ipamv1.VRF{
		Id: vrf.ID().Int64(), Url: "/api/ipam/vrfs/" + vrf.ID().String() + "/",
		Display: vrf.Display(), Name: vrf.Name(), Rd: typedVRFRD(vrf.RD()),
		EnforceUnique: vrf.EnforceUnique(), Description: vrf.Description(), Comments: vrf.Comments(),
		Created: timestamppb.New(vrf.Created().Time), LastUpdated: timestamppb.New(vrf.LastUpdated().Time),
		IpaddressCount: vrf.IPAddressCount(), PrefixCount: vrf.PrefixCount(),
	}
}

func typedVRFRD(nullable domainipam.NullableRouteDistinguisher) *wrapperspb.StringValue {
	rd, present := nullable.Get()
	if !present {
		return nil
	}
	return wrapperspb.String(rd.String())
}

var _ IPAMVRFService = (*applicationipam.VRFService)(nil)
