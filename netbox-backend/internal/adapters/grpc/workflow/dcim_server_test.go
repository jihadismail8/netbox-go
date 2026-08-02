package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type typedDCIMTestServices struct {
	sites              DCIMSiteService
	manufacturers      DCIMManufacturerService
	rackRoles          DCIMRackRoleService
	rackTypes          DCIMRackTypeService
	racks              DCIMRackService
	deviceRoles        DCIMDeviceRoleService
	deviceTypes        DCIMDeviceTypeService
	interfaceTemplates DCIMInterfaceTemplateService
	devices            DCIMDeviceService
	interfaces         DCIMInterfaceService
}

func completeTypedDCIMTestServices() typedDCIMTestServices {
	organizations := &organizationGRPCServiceSpy{}
	return typedDCIMTestServices{
		sites:              &grpcTypedSiteCallSpy{},
		manufacturers:      organizations,
		rackRoles:          organizations,
		rackTypes:          &rackTypeGRPCServiceSpy{},
		racks:              &rackGRPCServiceSpy{},
		deviceRoles:        &deviceRoleGRPCServiceSpy{},
		deviceTypes:        &deviceTypeGRPCServiceSpy{},
		interfaceTemplates: &grpcInterfaceTemplateServiceSpy{},
		devices:            &deviceGRPCServiceSpy{},
		interfaces:         &grpcInterfaceServiceSpy{},
	}
}

func (services typedDCIMTestServices) server() *DCIMServer {
	return NewTypedDCIMServer(
		services.sites,
		services.manufacturers,
		services.rackRoles,
		services.rackTypes,
		services.racks,
		services.deviceRoles,
		services.deviceTypes,
		services.interfaceTemplates,
		services.devices,
		services.interfaces,
	)
}

func TestNewTypedDCIMServerRequiresEveryTypedService(t *testing.T) {
	tests := []struct {
		name string
		drop func(*typedDCIMTestServices)
	}{
		{name: "sites", drop: func(services *typedDCIMTestServices) { services.sites = nil }},
		{name: "manufacturers", drop: func(services *typedDCIMTestServices) { services.manufacturers = nil }},
		{name: "rack roles", drop: func(services *typedDCIMTestServices) { services.rackRoles = nil }},
		{name: "rack types", drop: func(services *typedDCIMTestServices) { services.rackTypes = nil }},
		{name: "racks", drop: func(services *typedDCIMTestServices) { services.racks = nil }},
		{name: "device roles", drop: func(services *typedDCIMTestServices) { services.deviceRoles = nil }},
		{name: "device types", drop: func(services *typedDCIMTestServices) { services.deviceTypes = nil }},
		{name: "interface templates", drop: func(services *typedDCIMTestServices) { services.interfaceTemplates = nil }},
		{name: "devices", drop: func(services *typedDCIMTestServices) { services.devices = nil }},
		{name: "interfaces", drop: func(services *typedDCIMTestServices) { services.interfaces = nil }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			services := completeTypedDCIMTestServices()
			test.drop(&services)
			require.Panics(t, func() {
				services.server()
			})
		})
	}

	server := completeTypedDCIMTestServices().server()
	require.NotNil(t, server)
	require.NotNil(t, server.sites)
	require.NotNil(t, server.organizations)
	require.NotNil(t, server.rackTypes)
	require.NotNil(t, server.racks)
	require.NotNil(t, server.deviceRoles)
	require.NotNil(t, server.deviceTypes)
	require.NotNil(t, server.interfaceTemplates)
	require.NotNil(t, server.devices)
	require.NotNil(t, server.interfaces)
}
