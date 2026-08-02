package parity

import (
	"github.com/gin-gonic/gin"

	grpcadapter "netbox-go/internal/adapters/grpc/workflow"
	restadapter "netbox-go/internal/adapters/rest/netbox/workflow"
	"netbox-go/internal/domain/identity"
	"netbox-go/internal/platform/composition"
)

// newParityDCIMServer mirrors production composition and deliberately exposes
// no transitional map-shaped workflow service to the parity suite.
func newParityDCIMServer(core composition.Core) *grpcadapter.DCIMServer {
	return grpcadapter.NewTypedDCIMServer(
		core.Sites,
		core.Manufacturers,
		core.RackRoles,
		core.RackTypes,
		core.Racks,
		core.DeviceRoles,
		core.DeviceTypes,
		core.InterfaceTemplates,
		core.Devices,
		core.Interfaces,
	)
}

func newParityIPAMServer(core composition.Core) *grpcadapter.IPAMServer {
	return grpcadapter.NewIPAMServer(core.VRFs, core.Prefixes, core.IPAddresses)
}

func newParityRESTRouter(
	core composition.Core,
	principal identity.Principal,
) *gin.Engine {
	router := gin.New()
	restadapter.NewHandler(
		core.Sites,
		restadapter.WithOrganizationServices(core.Manufacturers, core.RackRoles),
		restadapter.WithRackTypeService(core.RackTypes),
		restadapter.WithRackService(core.Racks),
		restadapter.WithDeviceRoleService(core.DeviceRoles),
		restadapter.WithDeviceTypeService(core.DeviceTypes),
		restadapter.WithInterfaceTemplateService(core.InterfaceTemplates),
		restadapter.WithDeviceService(core.Devices),
		restadapter.WithInterfaceService(core.Interfaces),
		restadapter.WithVRFService(core.VRFs),
		restadapter.WithPrefixService(core.Prefixes),
		restadapter.WithIPAddressService(core.IPAddresses),
	).Register(router, func(c *gin.Context) {
		restadapter.SetPrincipal(c, principal)
		c.Next()
	})
	return router
}
