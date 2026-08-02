package workflow

import (
	"context"

	dcimv1 "netbox-go/gen/go/netbox/dcim/v1"
	typesv1 "netbox-go/gen/go/netbox/types/v1"
	"netbox-go/internal/adapters/grpc/statusmap"
	applicationdcim "netbox-go/internal/application/dcim"
	domaindcim "netbox-go/internal/domain/dcim"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/domain/shared"
)

type DCIMSiteService interface {
	ListSites(context.Context, identity.Principal, applicationdcim.ListSitesQuery) (applicationdcim.SitePage, error)
	GetSite(context.Context, identity.Principal, applicationdcim.GetSiteQuery) (*domaindcim.Site, error)
	CreateSite(context.Context, identity.Principal, applicationdcim.CreateSiteCommand) (*domaindcim.Site, error)
	ReplaceSite(context.Context, identity.Principal, applicationdcim.ReplaceSiteCommand) (*domaindcim.Site, error)
	UpdateSite(context.Context, identity.Principal, applicationdcim.UpdateSiteCommand) (*domaindcim.Site, error)
	DeleteSite(context.Context, identity.Principal, applicationdcim.DeleteSiteCommand) error
}

type DCIMServer struct {
	dcimv1.UnimplementedDCIMServiceServer
	sites              DCIMSiteService
	organizations      *DCIMOrganizationServer
	rackTypes          *RackTypeRPCHandler
	racks              *RackRPCHandler
	deviceRoles        *DCIMDeviceRoleServer
	deviceTypes        *DeviceTypeRPCHandler
	interfaceTemplates *InterfaceTemplateRPCHandler
	devices            *DeviceRPCHandler
	interfaces         *InterfaceRPCHandler
}

var _ dcimv1.DCIMServiceServer = (*DCIMServer)(nil)

// NewTypedDCIMServer is the only DCIM server constructor. Every one of the ten
// first-profile DCIM services is required, so no RPC can fall back to a
// map-shaped workflow service.
func NewTypedDCIMServer(
	sites DCIMSiteService,
	manufacturers DCIMManufacturerService,
	rackRoles DCIMRackRoleService,
	rackTypes DCIMRackTypeService,
	racks DCIMRackService,
	deviceRoles DCIMDeviceRoleService,
	deviceTypes DCIMDeviceTypeService,
	interfaceTemplates DCIMInterfaceTemplateService,
	devices DCIMDeviceService,
	interfaces DCIMInterfaceService,
) *DCIMServer {
	if sites == nil ||
		manufacturers == nil ||
		rackRoles == nil ||
		rackTypes == nil ||
		racks == nil ||
		deviceRoles == nil ||
		deviceTypes == nil ||
		interfaceTemplates == nil ||
		devices == nil ||
		interfaces == nil {
		panic("DCIM gRPC server requires all ten typed DCIM services")
	}
	return &DCIMServer{
		sites:              sites,
		organizations:      NewDCIMOrganizationServer(manufacturers, rackRoles),
		rackTypes:          NewRackTypeRPCHandler(rackTypes),
		racks:              NewRackRPCHandler(racks),
		deviceRoles:        NewDCIMDeviceRoleServer(deviceRoles),
		deviceTypes:        NewDeviceTypeRPCHandler(deviceTypes),
		interfaceTemplates: NewInterfaceTemplateRPCHandler(interfaceTemplates),
		devices:            NewDeviceRPCHandler(devices),
		interfaces:         NewInterfaceRPCHandler(interfaces),
	}
}

func (s *DCIMServer) ListSites(
	ctx context.Context,
	r *dcimv1.ListSitesRequest,
) (*dcimv1.ListSitesResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	page, err := s.sites.ListSites(ctx, p, typedSiteListQuery(r))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	out := make([]*dcimv1.Site, 0, len(page.Results))
	for _, site := range page.Results {
		out = append(out, typedSiteProto(site))
	}
	return &dcimv1.ListSitesResponse{
		Page:    &typesv1.PageInfo{Count: page.Count},
		Results: out,
	}, nil
}

func (s *DCIMServer) GetSite(
	ctx context.Context,
	r *dcimv1.GetSiteRequest,
) (*dcimv1.GetSiteResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	site, err := s.sites.GetSite(ctx, p, applicationdcim.GetSiteQuery{ID: shared.ID(r.Id)})
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.GetSiteResponse{Site: typedSiteProto(site)}, nil
}

func (s *DCIMServer) CreateSite(
	ctx context.Context,
	r *dcimv1.CreateSiteRequest,
) (*dcimv1.CreateSiteResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	site, err := s.sites.CreateSite(ctx, p, typedSiteCreateCommand(r.Site))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.CreateSiteResponse{Site: typedSiteProto(site)}, nil
}

func (s *DCIMServer) ReplaceSite(
	ctx context.Context,
	r *dcimv1.ReplaceSiteRequest,
) (*dcimv1.ReplaceSiteResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	site, err := s.sites.ReplaceSite(ctx, p, typedSiteReplaceCommand(shared.ID(r.Id), r.Site))
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.ReplaceSiteResponse{Site: typedSiteProto(site)}, nil
}

func (s *DCIMServer) UpdateSite(
	ctx context.Context,
	r *dcimv1.UpdateSiteRequest,
) (*dcimv1.UpdateSiteResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	command, err := typedSiteUpdateCommand(shared.ID(r.Id), r.Site, r.UpdateMask)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	site, err := s.sites.UpdateSite(ctx, p, command)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.UpdateSiteResponse{Site: typedSiteProto(site)}, nil
}

func (s *DCIMServer) DeleteSite(
	ctx context.Context,
	r *dcimv1.DeleteSiteRequest,
) (*dcimv1.DeleteSiteResponse, error) {
	p, err := principal(ctx)
	if err != nil {
		return nil, statusmap.Error(err)
	}
	if err := s.sites.DeleteSite(
		ctx,
		p,
		applicationdcim.DeleteSiteCommand{ID: shared.ID(r.Id)},
	); err != nil {
		return nil, statusmap.Error(err)
	}
	return &dcimv1.DeleteSiteResponse{}, nil
}

func (s *DCIMServer) ListManufacturers(ctx context.Context, r *dcimv1.ListManufacturersRequest) (*dcimv1.ListManufacturersResponse, error) {
	return s.organizations.ListManufacturers(ctx, r)
}

func (s *DCIMServer) GetManufacturer(ctx context.Context, r *dcimv1.GetManufacturerRequest) (*dcimv1.GetManufacturerResponse, error) {
	return s.organizations.GetManufacturer(ctx, r)
}

func (s *DCIMServer) CreateManufacturer(ctx context.Context, r *dcimv1.CreateManufacturerRequest) (*dcimv1.CreateManufacturerResponse, error) {
	return s.organizations.CreateManufacturer(ctx, r)
}

func (s *DCIMServer) ReplaceManufacturer(ctx context.Context, r *dcimv1.ReplaceManufacturerRequest) (*dcimv1.ReplaceManufacturerResponse, error) {
	return s.organizations.ReplaceManufacturer(ctx, r)
}

func (s *DCIMServer) UpdateManufacturer(ctx context.Context, r *dcimv1.UpdateManufacturerRequest) (*dcimv1.UpdateManufacturerResponse, error) {
	return s.organizations.UpdateManufacturer(ctx, r)
}

func (s *DCIMServer) DeleteManufacturer(ctx context.Context, r *dcimv1.DeleteManufacturerRequest) (*dcimv1.DeleteManufacturerResponse, error) {
	return s.organizations.DeleteManufacturer(ctx, r)
}

func (s *DCIMServer) ListRackRoles(ctx context.Context, r *dcimv1.ListRackRolesRequest) (*dcimv1.ListRackRolesResponse, error) {
	return s.organizations.ListRackRoles(ctx, r)
}

func (s *DCIMServer) GetRackRole(ctx context.Context, r *dcimv1.GetRackRoleRequest) (*dcimv1.GetRackRoleResponse, error) {
	return s.organizations.GetRackRole(ctx, r)
}

func (s *DCIMServer) CreateRackRole(ctx context.Context, r *dcimv1.CreateRackRoleRequest) (*dcimv1.CreateRackRoleResponse, error) {
	return s.organizations.CreateRackRole(ctx, r)
}

func (s *DCIMServer) ReplaceRackRole(ctx context.Context, r *dcimv1.ReplaceRackRoleRequest) (*dcimv1.ReplaceRackRoleResponse, error) {
	return s.organizations.ReplaceRackRole(ctx, r)
}

func (s *DCIMServer) UpdateRackRole(ctx context.Context, r *dcimv1.UpdateRackRoleRequest) (*dcimv1.UpdateRackRoleResponse, error) {
	return s.organizations.UpdateRackRole(ctx, r)
}

func (s *DCIMServer) DeleteRackRole(ctx context.Context, r *dcimv1.DeleteRackRoleRequest) (*dcimv1.DeleteRackRoleResponse, error) {
	return s.organizations.DeleteRackRole(ctx, r)
}

func (s *DCIMServer) ListRackTypes(ctx context.Context, r *dcimv1.ListRackTypesRequest) (*dcimv1.ListRackTypesResponse, error) {
	return s.rackTypes.ListRackTypes(ctx, r)
}

func (s *DCIMServer) GetRackType(ctx context.Context, r *dcimv1.GetRackTypeRequest) (*dcimv1.GetRackTypeResponse, error) {
	return s.rackTypes.GetRackType(ctx, r)
}

func (s *DCIMServer) CreateRackType(ctx context.Context, r *dcimv1.CreateRackTypeRequest) (*dcimv1.CreateRackTypeResponse, error) {
	return s.rackTypes.CreateRackType(ctx, r)
}

func (s *DCIMServer) ReplaceRackType(ctx context.Context, r *dcimv1.ReplaceRackTypeRequest) (*dcimv1.ReplaceRackTypeResponse, error) {
	return s.rackTypes.ReplaceRackType(ctx, r)
}

func (s *DCIMServer) UpdateRackType(ctx context.Context, r *dcimv1.UpdateRackTypeRequest) (*dcimv1.UpdateRackTypeResponse, error) {
	return s.rackTypes.UpdateRackType(ctx, r)
}

func (s *DCIMServer) DeleteRackType(ctx context.Context, r *dcimv1.DeleteRackTypeRequest) (*dcimv1.DeleteRackTypeResponse, error) {
	return s.rackTypes.DeleteRackType(ctx, r)
}

func (s *DCIMServer) ListRacks(ctx context.Context, r *dcimv1.ListRacksRequest) (*dcimv1.ListRacksResponse, error) {
	return s.racks.ListRacks(ctx, r)
}

func (s *DCIMServer) GetRack(ctx context.Context, r *dcimv1.GetRackRequest) (*dcimv1.GetRackResponse, error) {
	return s.racks.GetRack(ctx, r)
}

func (s *DCIMServer) CreateRack(ctx context.Context, r *dcimv1.CreateRackRequest) (*dcimv1.CreateRackResponse, error) {
	return s.racks.CreateRack(ctx, r)
}

func (s *DCIMServer) ReplaceRack(ctx context.Context, r *dcimv1.ReplaceRackRequest) (*dcimv1.ReplaceRackResponse, error) {
	return s.racks.ReplaceRack(ctx, r)
}

func (s *DCIMServer) UpdateRack(ctx context.Context, r *dcimv1.UpdateRackRequest) (*dcimv1.UpdateRackResponse, error) {
	return s.racks.UpdateRack(ctx, r)
}

func (s *DCIMServer) DeleteRack(ctx context.Context, r *dcimv1.DeleteRackRequest) (*dcimv1.DeleteRackResponse, error) {
	return s.racks.DeleteRack(ctx, r)
}

func (s *DCIMServer) ListDeviceRoles(ctx context.Context, r *dcimv1.ListDeviceRolesRequest) (*dcimv1.ListDeviceRolesResponse, error) {
	return s.deviceRoles.ListDeviceRoles(ctx, r)
}

func (s *DCIMServer) GetDeviceRole(ctx context.Context, r *dcimv1.GetDeviceRoleRequest) (*dcimv1.GetDeviceRoleResponse, error) {
	return s.deviceRoles.GetDeviceRole(ctx, r)
}

func (s *DCIMServer) CreateDeviceRole(ctx context.Context, r *dcimv1.CreateDeviceRoleRequest) (*dcimv1.CreateDeviceRoleResponse, error) {
	return s.deviceRoles.CreateDeviceRole(ctx, r)
}

func (s *DCIMServer) ReplaceDeviceRole(ctx context.Context, r *dcimv1.ReplaceDeviceRoleRequest) (*dcimv1.ReplaceDeviceRoleResponse, error) {
	return s.deviceRoles.ReplaceDeviceRole(ctx, r)
}

func (s *DCIMServer) UpdateDeviceRole(ctx context.Context, r *dcimv1.UpdateDeviceRoleRequest) (*dcimv1.UpdateDeviceRoleResponse, error) {
	return s.deviceRoles.UpdateDeviceRole(ctx, r)
}

func (s *DCIMServer) DeleteDeviceRole(ctx context.Context, r *dcimv1.DeleteDeviceRoleRequest) (*dcimv1.DeleteDeviceRoleResponse, error) {
	return s.deviceRoles.DeleteDeviceRole(ctx, r)
}

func (s *DCIMServer) ListDeviceTypes(ctx context.Context, r *dcimv1.ListDeviceTypesRequest) (*dcimv1.ListDeviceTypesResponse, error) {
	return s.deviceTypes.ListDeviceTypes(ctx, r)
}

func (s *DCIMServer) GetDeviceType(ctx context.Context, r *dcimv1.GetDeviceTypeRequest) (*dcimv1.GetDeviceTypeResponse, error) {
	return s.deviceTypes.GetDeviceType(ctx, r)
}

func (s *DCIMServer) CreateDeviceType(ctx context.Context, r *dcimv1.CreateDeviceTypeRequest) (*dcimv1.CreateDeviceTypeResponse, error) {
	return s.deviceTypes.CreateDeviceType(ctx, r)
}

func (s *DCIMServer) ReplaceDeviceType(ctx context.Context, r *dcimv1.ReplaceDeviceTypeRequest) (*dcimv1.ReplaceDeviceTypeResponse, error) {
	return s.deviceTypes.ReplaceDeviceType(ctx, r)
}

func (s *DCIMServer) UpdateDeviceType(ctx context.Context, r *dcimv1.UpdateDeviceTypeRequest) (*dcimv1.UpdateDeviceTypeResponse, error) {
	return s.deviceTypes.UpdateDeviceType(ctx, r)
}

func (s *DCIMServer) DeleteDeviceType(ctx context.Context, r *dcimv1.DeleteDeviceTypeRequest) (*dcimv1.DeleteDeviceTypeResponse, error) {
	return s.deviceTypes.DeleteDeviceType(ctx, r)
}

func (s *DCIMServer) ListInterfaceTemplates(ctx context.Context, r *dcimv1.ListInterfaceTemplatesRequest) (*dcimv1.ListInterfaceTemplatesResponse, error) {
	return s.interfaceTemplates.ListInterfaceTemplates(ctx, r)
}

func (s *DCIMServer) GetInterfaceTemplate(ctx context.Context, r *dcimv1.GetInterfaceTemplateRequest) (*dcimv1.GetInterfaceTemplateResponse, error) {
	return s.interfaceTemplates.GetInterfaceTemplate(ctx, r)
}

func (s *DCIMServer) CreateInterfaceTemplate(ctx context.Context, r *dcimv1.CreateInterfaceTemplateRequest) (*dcimv1.CreateInterfaceTemplateResponse, error) {
	return s.interfaceTemplates.CreateInterfaceTemplate(ctx, r)
}

func (s *DCIMServer) ReplaceInterfaceTemplate(ctx context.Context, r *dcimv1.ReplaceInterfaceTemplateRequest) (*dcimv1.ReplaceInterfaceTemplateResponse, error) {
	return s.interfaceTemplates.ReplaceInterfaceTemplate(ctx, r)
}

func (s *DCIMServer) UpdateInterfaceTemplate(ctx context.Context, r *dcimv1.UpdateInterfaceTemplateRequest) (*dcimv1.UpdateInterfaceTemplateResponse, error) {
	return s.interfaceTemplates.UpdateInterfaceTemplate(ctx, r)
}

func (s *DCIMServer) DeleteInterfaceTemplate(ctx context.Context, r *dcimv1.DeleteInterfaceTemplateRequest) (*dcimv1.DeleteInterfaceTemplateResponse, error) {
	return s.interfaceTemplates.DeleteInterfaceTemplate(ctx, r)
}

func (s *DCIMServer) ListDevices(ctx context.Context, r *dcimv1.ListDevicesRequest) (*dcimv1.ListDevicesResponse, error) {
	return s.devices.ListDevices(ctx, r)
}

func (s *DCIMServer) GetDevice(ctx context.Context, r *dcimv1.GetDeviceRequest) (*dcimv1.GetDeviceResponse, error) {
	return s.devices.GetDevice(ctx, r)
}

func (s *DCIMServer) CreateDevice(ctx context.Context, r *dcimv1.CreateDeviceRequest) (*dcimv1.CreateDeviceResponse, error) {
	return s.devices.CreateDevice(ctx, r)
}

func (s *DCIMServer) ReplaceDevice(ctx context.Context, r *dcimv1.ReplaceDeviceRequest) (*dcimv1.ReplaceDeviceResponse, error) {
	return s.devices.ReplaceDevice(ctx, r)
}

func (s *DCIMServer) UpdateDevice(ctx context.Context, r *dcimv1.UpdateDeviceRequest) (*dcimv1.UpdateDeviceResponse, error) {
	return s.devices.UpdateDevice(ctx, r)
}

func (s *DCIMServer) DeleteDevice(ctx context.Context, r *dcimv1.DeleteDeviceRequest) (*dcimv1.DeleteDeviceResponse, error) {
	return s.devices.DeleteDevice(ctx, r)
}

func (s *DCIMServer) ListInterfaces(ctx context.Context, r *dcimv1.ListInterfacesRequest) (*dcimv1.ListInterfacesResponse, error) {
	return s.interfaces.ListInterfaces(ctx, r)
}

func (s *DCIMServer) GetInterface(ctx context.Context, r *dcimv1.GetInterfaceRequest) (*dcimv1.GetInterfaceResponse, error) {
	return s.interfaces.GetInterface(ctx, r)
}

func (s *DCIMServer) CreateInterface(ctx context.Context, r *dcimv1.CreateInterfaceRequest) (*dcimv1.CreateInterfaceResponse, error) {
	return s.interfaces.CreateInterface(ctx, r)
}

func (s *DCIMServer) ReplaceInterface(ctx context.Context, r *dcimv1.ReplaceInterfaceRequest) (*dcimv1.ReplaceInterfaceResponse, error) {
	return s.interfaces.ReplaceInterface(ctx, r)
}

func (s *DCIMServer) UpdateInterface(ctx context.Context, r *dcimv1.UpdateInterfaceRequest) (*dcimv1.UpdateInterfaceResponse, error) {
	return s.interfaces.UpdateInterface(ctx, r)
}

func (s *DCIMServer) DeleteInterface(ctx context.Context, r *dcimv1.DeleteInterfaceRequest) (*dcimv1.DeleteInterfaceResponse, error) {
	return s.interfaces.DeleteInterface(ctx, r)
}
