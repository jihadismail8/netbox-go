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

type IPAMPrefixService interface {
	ListPrefixes(context.Context, identity.Principal, applicationipam.ListPrefixesQuery) (applicationipam.PrefixPage, error)
	GetPrefix(context.Context, identity.Principal, applicationipam.GetPrefixQuery) (*domainipam.Prefix, error)
	CreatePrefix(context.Context, identity.Principal, applicationipam.CreatePrefixCommand) (*domainipam.Prefix, error)
	ReplacePrefix(context.Context, identity.Principal, applicationipam.ReplacePrefixCommand) (*domainipam.Prefix, error)
	UpdatePrefix(context.Context, identity.Principal, applicationipam.UpdatePrefixCommand) (*domainipam.Prefix, error)
	DeletePrefix(context.Context, identity.Principal, applicationipam.DeletePrefixCommand) error
}

type PrefixRPCHandler struct{ service IPAMPrefixService }

func NewPrefixRPCHandler(service IPAMPrefixService) *PrefixRPCHandler {
	if service == nil {
		panic("Prefix gRPC handler requires a typed Prefix service")
	}
	return &PrefixRPCHandler{service: service}
}

func (handler *PrefixRPCHandler) ListPrefixes(
	ctx context.Context,
	request *ipamv1.ListPrefixesRequest,
) (*ipamv1.ListPrefixesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := handler.service.ListPrefixes(ctx, p, typedPrefixListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*ipamv1.Prefix, 0, len(page.Results))
	for _, prefix := range page.Results {
		results = append(results, typedPrefixProto(prefix))
	}
	return &ipamv1.ListPrefixesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (handler *PrefixRPCHandler) GetPrefix(
	ctx context.Context,
	request *ipamv1.GetPrefixRequest,
) (*ipamv1.GetPrefixResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	prefix, err := handler.service.GetPrefix(
		ctx, p, applicationipam.GetPrefixQuery{ID: shared.ID(request.GetId())},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.GetPrefixResponse{Prefix: typedPrefixProto(prefix)}, nil
}

func (handler *PrefixRPCHandler) CreatePrefix(
	ctx context.Context,
	request *ipamv1.CreatePrefixRequest,
) (*ipamv1.CreatePrefixResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	prefix, err := handler.service.CreatePrefix(ctx, p, typedPrefixCreateCommand(request.GetPrefix()))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.CreatePrefixResponse{Prefix: typedPrefixProto(prefix)}, nil
}

func (handler *PrefixRPCHandler) ReplacePrefix(
	ctx context.Context,
	request *ipamv1.ReplacePrefixRequest,
) (*ipamv1.ReplacePrefixResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	prefix, err := handler.service.ReplacePrefix(
		ctx, p, typedPrefixReplaceCommand(shared.ID(request.GetId()), request.GetPrefix()),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.ReplacePrefixResponse{Prefix: typedPrefixProto(prefix)}, nil
}

func (handler *PrefixRPCHandler) UpdatePrefix(
	ctx context.Context,
	request *ipamv1.UpdatePrefixRequest,
) (*ipamv1.UpdatePrefixResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedPrefixUpdateCommand(
		shared.ID(request.GetId()), request.GetPrefix(), request.GetUpdateMask(),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	prefix, err := handler.service.UpdatePrefix(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.UpdatePrefixResponse{Prefix: typedPrefixProto(prefix)}, nil
}

func (handler *PrefixRPCHandler) DeletePrefix(
	ctx context.Context,
	request *ipamv1.DeletePrefixRequest,
) (*ipamv1.DeletePrefixResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := handler.service.DeletePrefix(
		ctx, p, applicationipam.DeletePrefixCommand{ID: shared.ID(request.GetId())},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &ipamv1.DeletePrefixResponse{}, nil
}

func typedPrefixListQuery(request *ipamv1.ListPrefixesRequest) applicationipam.ListPrefixesQuery {
	query := applicationipam.ListPrefixesQuery{}
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
	query.Prefixes = grpcOptionalString(request.Prefix)
	if request.Family != nil {
		value := int64(*request.Family)
		query.Family = &value
	}
	query.Statuses = grpcOptionalString(request.Status)
	query.Within = cloneStringPointer(request.Within)
	query.WithinInclude = cloneStringPointer(request.WithinInclude)
	query.Contains = cloneStringPointer(request.Contains)
	return query
}

func grpcOptionalInt64(value *int64) []int64 {
	if value == nil {
		return nil
	}
	return []int64{*value}
}

type typedPrefixFields struct {
	prefix       applicationipam.Field[string]
	vrf          applicationipam.Field[int64]
	status       applicationipam.Field[string]
	isPool       applicationipam.Field[bool]
	markUtilized applicationipam.Field[bool]
	description  applicationipam.Field[string]
	comments     applicationipam.Field[string]
}

func typedPrefixInputFields(input *ipamv1.PrefixInput) typedPrefixFields {
	if input == nil {
		return typedPrefixFields{}
	}
	return typedPrefixFields{
		prefix: typedVRFProtoString(input.Prefix), vrf: typedPrefixWrappedInt64(input.Vrf),
		status: typedVRFProtoString(input.Status), isPool: typedVRFProtoBool(input.IsPool),
		markUtilized: typedVRFProtoBool(input.MarkUtilized),
		description:  typedVRFProtoString(input.Description), comments: typedVRFProtoString(input.Comments),
	}
}

func typedPrefixWrappedInt64(value *wrapperspb.Int64Value) applicationipam.Field[int64] {
	if value == nil {
		return applicationipam.OmittedField[int64]()
	}
	return applicationipam.FieldValue(value.Value)
}

func typedPrefixCreateCommand(input *ipamv1.PrefixInput) applicationipam.CreatePrefixCommand {
	fields := typedPrefixInputFields(input)
	return applicationipam.CreatePrefixCommand{
		Prefix: fields.prefix, VRF: fields.vrf, Status: fields.status,
		IsPool: fields.isPool, MarkUtilized: fields.markUtilized,
		Description: fields.description, Comments: fields.comments,
	}
}

func typedPrefixReplaceCommand(id shared.ID, input *ipamv1.PrefixInput) applicationipam.ReplacePrefixCommand {
	return applicationipam.ReplacePrefixCommand{
		ID: id, CreatePrefixCommand: typedPrefixCreateCommand(input),
	}
}

func typedPrefixUpdateCommand(
	id shared.ID,
	input *ipamv1.PrefixInput,
	mask *fieldmaskpb.FieldMask,
) (applicationipam.UpdatePrefixCommand, error) {
	fields := typedPrefixInputFields(input)
	command := applicationipam.UpdatePrefixCommand{ID: id}
	if mask == nil || len(mask.Paths) == 0 {
		command.Prefix, command.VRF, command.Status = fields.prefix, fields.vrf, fields.status
		command.IsPool, command.MarkUtilized = fields.isPool, fields.markUtilized
		command.Description, command.Comments = fields.description, fields.comments
		return command, nil
	}
	for _, path := range mask.Paths {
		switch path {
		case "prefix":
			if fields.prefix.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.Prefix = fields.prefix
		case "vrf":
			if input == nil || input.Vrf == nil {
				command.VRF = applicationipam.NullField[int64]()
			} else {
				command.VRF = fields.vrf
			}
		case "status":
			if fields.status.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.Status = fields.status
		case "is_pool":
			if fields.isPool.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.IsPool = fields.isPool
		case "mark_utilized":
			if fields.markUtilized.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.MarkUtilized = fields.markUtilized
		case "description":
			if fields.description.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.Description = fields.description
		case "comments":
			if fields.comments.State() != applicationipam.FieldPresent {
				return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
			}
			command.Comments = fields.comments
		default:
			return applicationipam.UpdatePrefixCommand{}, invalidPrefixUpdateMask()
		}
	}
	return command, nil
}

func invalidPrefixUpdateMask() error {
	return shared.NewValidationError(
		shared.FieldViolation{
			Field:       "update_mask",
			Description: "Every update_mask path must name a supported field with an explicit value.",
		},
	)
}

func typedPrefixProto(prefix *domainipam.Prefix) *ipamv1.Prefix {
	if prefix == nil {
		return nil
	}
	return &ipamv1.Prefix{
		Id: prefix.ID().Int64(), Url: "/api/ipam/prefixes/" + prefix.ID().String() + "/",
		Display: prefix.Display(), Prefix: prefix.Network().String(), VrfId: typedPrefixVRFID(prefix.VRF()),
		Status: prefix.Status().String(), IsPool: prefix.IsPool(), MarkUtilized: prefix.MarkUtilized(),
		Description: prefix.Description(), Comments: prefix.Comments(),
		Created: timestamppb.New(prefix.Created().Time), LastUpdated: timestamppb.New(prefix.LastUpdated().Time),
		Family: prefix.Family(), Children: prefix.Children(), Depth: prefix.Depth(),
	}
}

func typedPrefixVRFID(nullable domainipam.NullableVRFReference) *wrapperspb.Int64Value {
	reference, present := nullable.Get()
	if !present {
		return nil
	}
	return wrapperspb.Int64(reference.ID().Int64())
}

var _ IPAMPrefixService = (*applicationipam.PrefixService)(nil)
