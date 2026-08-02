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

type DCIMManufacturerService interface {
	ListManufacturers(context.Context, identity.Principal, applicationdcim.ListManufacturersQuery) (applicationdcim.ManufacturerPage, error)
	GetManufacturer(context.Context, identity.Principal, applicationdcim.GetManufacturerQuery) (*domaindcim.Manufacturer, error)
	CreateManufacturer(context.Context, identity.Principal, applicationdcim.CreateManufacturerCommand) (*domaindcim.Manufacturer, error)
	ReplaceManufacturer(context.Context, identity.Principal, applicationdcim.ReplaceManufacturerCommand) (*domaindcim.Manufacturer, error)
	UpdateManufacturer(context.Context, identity.Principal, applicationdcim.UpdateManufacturerCommand) (*domaindcim.Manufacturer, error)
	DeleteManufacturer(context.Context, identity.Principal, applicationdcim.DeleteManufacturerCommand) error
}

type DCIMRackRoleService interface {
	ListRackRoles(context.Context, identity.Principal, applicationdcim.ListRackRolesQuery) (applicationdcim.RackRolePage, error)
	GetRackRole(context.Context, identity.Principal, applicationdcim.GetRackRoleQuery) (*domaindcim.RackRole, error)
	CreateRackRole(context.Context, identity.Principal, applicationdcim.CreateRackRoleCommand) (*domaindcim.RackRole, error)
	ReplaceRackRole(context.Context, identity.Principal, applicationdcim.ReplaceRackRoleCommand) (*domaindcim.RackRole, error)
	UpdateRackRole(context.Context, identity.Principal, applicationdcim.UpdateRackRoleCommand) (*domaindcim.RackRole, error)
	DeleteRackRole(context.Context, identity.Principal, applicationdcim.DeleteRackRoleCommand) error
}

var (
	_ DCIMManufacturerService = (*applicationdcim.ManufacturerService)(nil)
	_ DCIMRackRoleService     = (*applicationdcim.RackRoleService)(nil)
)

// DCIMOrganizationServer translates the twelve typed Manufacturer and
// RackRole RPCs delegated by DCIMServer.
type DCIMOrganizationServer struct {
	manufacturers DCIMManufacturerService
	rackRoles     DCIMRackRoleService
}

func NewDCIMOrganizationServer(
	manufacturers DCIMManufacturerService,
	rackRoles DCIMRackRoleService,
) *DCIMOrganizationServer {
	if manufacturers == nil || rackRoles == nil {
		panic("DCIM organization gRPC server requires typed Manufacturer and RackRole services")
	}
	return &DCIMOrganizationServer{manufacturers: manufacturers, rackRoles: rackRoles}
}

func (server *DCIMOrganizationServer) ListManufacturers(
	ctx context.Context,
	request *dcimv1.ListManufacturersRequest,
) (*dcimv1.ListManufacturersResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := server.manufacturers.ListManufacturers(ctx, principal, typedManufacturerListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.Manufacturer, 0, len(page.Results))
	for _, manufacturer := range page.Results {
		results = append(results, typedManufacturerProto(manufacturer))
	}
	return &dcimv1.ListManufacturersResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (server *DCIMOrganizationServer) GetManufacturer(
	ctx context.Context,
	request *dcimv1.GetManufacturerRequest,
) (*dcimv1.GetManufacturerResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	manufacturer, err := server.manufacturers.GetManufacturer(
		ctx, principal, applicationdcim.GetManufacturerQuery{ID: shared.ID(id)},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetManufacturerResponse{Manufacturer: typedManufacturerProto(manufacturer)}, nil
}

func (server *DCIMOrganizationServer) CreateManufacturer(
	ctx context.Context,
	request *dcimv1.CreateManufacturerRequest,
) (*dcimv1.CreateManufacturerResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	var input *dcimv1.ManufacturerInput
	if request != nil {
		input = request.Manufacturer
	}
	manufacturer, err := server.manufacturers.CreateManufacturer(
		ctx, principal, typedManufacturerCreateCommand(input),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateManufacturerResponse{Manufacturer: typedManufacturerProto(manufacturer)}, nil
}

func (server *DCIMOrganizationServer) ReplaceManufacturer(
	ctx context.Context,
	request *dcimv1.ReplaceManufacturerRequest,
) (*dcimv1.ReplaceManufacturerResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.ManufacturerInput
	if request != nil {
		id, input = request.Id, request.Manufacturer
	}
	manufacturer, err := server.manufacturers.ReplaceManufacturer(
		ctx, principal, typedManufacturerReplaceCommand(shared.ID(id), input),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceManufacturerResponse{Manufacturer: typedManufacturerProto(manufacturer)}, nil
}

func (server *DCIMOrganizationServer) UpdateManufacturer(
	ctx context.Context,
	request *dcimv1.UpdateManufacturerRequest,
) (*dcimv1.UpdateManufacturerResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.ManufacturerInput
	var mask = manufacturerRequestUpdateMask(request)
	if request != nil {
		id, input = request.Id, request.Manufacturer
	}
	command, err := typedManufacturerUpdateCommand(shared.ID(id), input, mask)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	manufacturer, err := server.manufacturers.UpdateManufacturer(ctx, principal, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateManufacturerResponse{Manufacturer: typedManufacturerProto(manufacturer)}, nil
}

func (server *DCIMOrganizationServer) DeleteManufacturer(
	ctx context.Context,
	request *dcimv1.DeleteManufacturerRequest,
) (*dcimv1.DeleteManufacturerResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	if err := server.manufacturers.DeleteManufacturer(
		ctx, principal, applicationdcim.DeleteManufacturerCommand{ID: shared.ID(id)},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteManufacturerResponse{}, nil
}

func (server *DCIMOrganizationServer) ListRackRoles(
	ctx context.Context,
	request *dcimv1.ListRackRolesRequest,
) (*dcimv1.ListRackRolesResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := server.rackRoles.ListRackRoles(ctx, principal, typedRackRoleListQuery(request))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	results := make([]*dcimv1.RackRole, 0, len(page.Results))
	for _, role := range page.Results {
		results = append(results, typedRackRoleProto(role))
	}
	return &dcimv1.ListRackRolesResponse{
		Page: &typesv1.PageInfo{Count: page.Count}, Results: results,
	}, nil
}

func (server *DCIMOrganizationServer) GetRackRole(
	ctx context.Context,
	request *dcimv1.GetRackRoleRequest,
) (*dcimv1.GetRackRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	role, err := server.rackRoles.GetRackRole(
		ctx, principal, applicationdcim.GetRackRoleQuery{ID: shared.ID(id)},
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetRackRoleResponse{RackRole: typedRackRoleProto(role)}, nil
}

func (server *DCIMOrganizationServer) CreateRackRole(
	ctx context.Context,
	request *dcimv1.CreateRackRoleRequest,
) (*dcimv1.CreateRackRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	var input *dcimv1.RackRoleInput
	if request != nil {
		input = request.RackRole
	}
	role, err := server.rackRoles.CreateRackRole(ctx, principal, typedRackRoleCreateCommand(input))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateRackRoleResponse{RackRole: typedRackRoleProto(role)}, nil
}

func (server *DCIMOrganizationServer) ReplaceRackRole(
	ctx context.Context,
	request *dcimv1.ReplaceRackRoleRequest,
) (*dcimv1.ReplaceRackRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.RackRoleInput
	if request != nil {
		id, input = request.Id, request.RackRole
	}
	role, err := server.rackRoles.ReplaceRackRole(
		ctx, principal, typedRackRoleReplaceCommand(shared.ID(id), input),
	)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceRackRoleResponse{RackRole: typedRackRoleProto(role)}, nil
}

func (server *DCIMOrganizationServer) UpdateRackRole(
	ctx context.Context,
	request *dcimv1.UpdateRackRoleRequest,
) (*dcimv1.UpdateRackRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	var input *dcimv1.RackRoleInput
	var mask = rackRoleRequestUpdateMask(request)
	if request != nil {
		id, input = request.Id, request.RackRole
	}
	command, err := typedRackRoleUpdateCommand(shared.ID(id), input, mask)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	role, err := server.rackRoles.UpdateRackRole(ctx, principal, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateRackRoleResponse{RackRole: typedRackRoleProto(role)}, nil
}

func (server *DCIMOrganizationServer) DeleteRackRole(
	ctx context.Context,
	request *dcimv1.DeleteRackRoleRequest,
) (*dcimv1.DeleteRackRoleResponse, error) {
	principal, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	id := int64(0)
	if request != nil {
		id = request.Id
	}
	if err := server.rackRoles.DeleteRackRole(
		ctx, principal, applicationdcim.DeleteRackRoleCommand{ID: shared.ID(id)},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteRackRoleResponse{}, nil
}

func manufacturerRequestUpdateMask(request *dcimv1.UpdateManufacturerRequest) *fieldmaskpb.FieldMask {
	if request == nil {
		return nil
	}
	return request.UpdateMask
}

func rackRoleRequestUpdateMask(request *dcimv1.UpdateRackRoleRequest) *fieldmaskpb.FieldMask {
	if request == nil {
		return nil
	}
	return request.UpdateMask
}
