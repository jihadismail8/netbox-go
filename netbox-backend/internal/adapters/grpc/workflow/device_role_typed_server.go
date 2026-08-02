package workflow

import (
	"context"

	"google.golang.org/protobuf/types/known/fieldmaskpb"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DCIMDeviceRoleService interface {
	ListDeviceRoles(context.Context, identity.Principal, applicationdcim.ListDeviceRolesQuery) (applicationdcim.DeviceRolePage, error)
	GetDeviceRole(context.Context, identity.Principal, applicationdcim.GetDeviceRoleQuery) (*domaindcim.DeviceRole, error)
	CreateDeviceRole(context.Context, identity.Principal, applicationdcim.CreateDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	ReplaceDeviceRole(context.Context, identity.Principal, applicationdcim.ReplaceDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	UpdateDeviceRole(context.Context, identity.Principal, applicationdcim.UpdateDeviceRoleCommand) (*domaindcim.DeviceRole, error)
	DeleteDeviceRole(context.Context, identity.Principal, applicationdcim.DeleteDeviceRoleCommand) error
}

var _ DCIMDeviceRoleService = (*applicationdcim.DeviceRoleService)(nil)

// DCIMDeviceRoleServer translates DeviceRole RPCs without exposing workflow
// resource maps. DCIMServer delegates these methods after composition cutover.
type DCIMDeviceRoleServer struct {
	service DCIMDeviceRoleService
}

func NewDCIMDeviceRoleServer(service DCIMDeviceRoleService) *DCIMDeviceRoleServer {
	if service == nil {
		panic("DCIM DeviceRole gRPC server requires a typed service")
	}
	return &DCIMDeviceRoleServer{service: service}
}

func (server *DCIMDeviceRoleServer) ListDeviceRoles(
	ctx context.Context,
	request *dcimv1.ListDeviceRolesRequest,
) (*dcimv1.ListDeviceRolesResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := server.service.ListDeviceRoles(ctx, principal, typedDeviceRoleListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.DeviceRole, 0, len(page.Results))
	for _, role := range page.Results {
		results = append(results, typedDeviceRoleProto(role))
	}
	return &dcimv1.ListDeviceRolesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (server *DCIMDeviceRoleServer) GetDeviceRole(
	ctx context.Context,
	request *dcimv1.GetDeviceRoleRequest,
) (*dcimv1.GetDeviceRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	role, err := server.service.GetDeviceRole(
		ctx, principal, applicationdcim.GetDeviceRoleQuery{ID: shared.ID(id)},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetDeviceRoleResponse{DeviceRole: typedDeviceRoleProto(role)}, nil
}

func (server *DCIMDeviceRoleServer) CreateDeviceRole(
	ctx context.Context,
	request *dcimv1.CreateDeviceRoleRequest,
) (*dcimv1.CreateDeviceRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	var input *dcimv1.DeviceRoleInput
	if request != nil {
		input = request.DeviceRole
	}
	role, err := server.service.CreateDeviceRole(ctx, principal, typedDeviceRoleCreateCommand(input))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateDeviceRoleResponse{DeviceRole: typedDeviceRoleProto(role)}, nil
}

func (server *DCIMDeviceRoleServer) ReplaceDeviceRole(
	ctx context.Context,
	request *dcimv1.ReplaceDeviceRoleRequest,
) (*dcimv1.ReplaceDeviceRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.DeviceRoleInput
	if request != nil {
		id, input = request.Id, request.DeviceRole
	}
	role, err := server.service.ReplaceDeviceRole(
		ctx, principal, typedDeviceRoleReplaceCommand(shared.ID(id), input),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceDeviceRoleResponse{DeviceRole: typedDeviceRoleProto(role)}, nil
}

func (server *DCIMDeviceRoleServer) UpdateDeviceRole(
	ctx context.Context,
	request *dcimv1.UpdateDeviceRoleRequest,
) (*dcimv1.UpdateDeviceRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.DeviceRoleInput
	var mask *fieldmaskpb.FieldMask
	if request != nil {
		id, input, mask = request.Id, request.DeviceRole, request.UpdateMask
	}
	command, err := typedDeviceRoleUpdateCommand(shared.ID(id), input, mask)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	role, err := server.service.UpdateDeviceRole(ctx, principal, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateDeviceRoleResponse{DeviceRole: typedDeviceRoleProto(role)}, nil
}

func (server *DCIMDeviceRoleServer) DeleteDeviceRole(
	ctx context.Context,
	request *dcimv1.DeleteDeviceRoleRequest,
) (*dcimv1.DeleteDeviceRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	if err := server.service.DeleteDeviceRole(
		ctx, principal, applicationdcim.DeleteDeviceRoleCommand{ID: shared.ID(id)},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteDeviceRoleResponse{}, nil
}
