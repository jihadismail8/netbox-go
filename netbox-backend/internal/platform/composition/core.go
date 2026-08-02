// Package composition explicitly wires the supported core capability. It is
// the only place that joins application ports to concrete persistence.
package composition

import (
	"gorm.io/gorm"

	changelogpostgres "netbox-go/internal/adapters/postgres/changelog"
	dcimpostgres "netbox-go/internal/adapters/postgres/dcim"
	identitypostgres "netbox-go/internal/adapters/postgres/identity"
	ipampostgres "netbox-go/internal/adapters/postgres/ipam"
	transactionpostgres "netbox-go/internal/adapters/postgres/transaction"
	"netbox-go/internal/application/authz"
	dcimapp "netbox-go/internal/application/dcim"
	identityapp "netbox-go/internal/application/identity"
	ipamapp "netbox-go/internal/application/ipam"
	"netbox-go/internal/domain/shared"
)

type Core struct {
	Identity           *identityapp.Service
	Sites              *dcimapp.SiteService
	Manufacturers      *dcimapp.ManufacturerService
	RackRoles          *dcimapp.RackRoleService
	RackTypes          *dcimapp.RackTypeService
	Racks              *dcimapp.RackService
	DeviceRoles        *dcimapp.DeviceRoleService
	DeviceTypes        *dcimapp.DeviceTypeService
	InterfaceTemplates *dcimapp.InterfaceTemplateService
	Devices            *dcimapp.DeviceService
	Interfaces         *dcimapp.InterfaceService
	VRFs               *ipamapp.VRFService
	Prefixes           *ipamapp.PrefixService
	IPAddresses        *ipamapp.IPAddressService
}

func NewCore(db *gorm.DB) Core {
	return NewCoreWithAuthorizer(db, authz.PermissionAuthorizer{})
}

// NewCoreWithAuthorizer wires the supported typed capability with an explicit
// authorization dependency. It exists for parity and application-boundary
// tests which must exercise the production composition with deterministic
// authorization decisions.
func NewCoreWithAuthorizer(db *gorm.DB, authorizer authz.ResourceAuthorizer) Core {
	identityService := identityapp.NewService(identitypostgres.NewStore(db), identityapp.RealClock{})
	unitOfWork := transactionpostgres.NewUnitOfWork(db)
	recorder := changelogpostgres.NewRecorder(db)
	clock := shared.SystemClock{}
	siteRepository := dcimpostgres.NewSiteRepository(db)
	siteService, err := dcimapp.NewSiteService(
		siteRepository,
		unitOfWork,
		recorder,
		authorizer,
		clock,
	)
	if err != nil {
		panic("compose typed Site service: " + err.Error())
	}
	manufacturerRepository := dcimpostgres.NewManufacturerRepository(db)
	manufacturerService, err := dcimapp.NewManufacturerService(
		manufacturerRepository, unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed Manufacturer service: " + err.Error())
	}
	rackRoleRepository := dcimpostgres.NewRackRoleRepository(db)
	rackRoleService, err := dcimapp.NewRackRoleService(
		rackRoleRepository, unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed RackRole service: " + err.Error())
	}
	rackTypeService, err := dcimapp.NewRackTypeService(
		dcimpostgres.NewRackTypeRepository(db), manufacturerRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed RackType service: " + err.Error())
	}
	rackRepository := dcimpostgres.NewRackRepository(db)
	rackService, err := dcimapp.NewRackService(
		rackRepository,
		siteRepository,
		dcimpostgres.NewRackTypeRepository(db),
		rackRoleRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed Rack service: " + err.Error())
	}
	deviceRoleRepository := dcimpostgres.NewDeviceRoleRepository(db)
	deviceRoleService, err := dcimapp.NewDeviceRoleService(
		deviceRoleRepository, unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed DeviceRole service: " + err.Error())
	}
	deviceTypeRepository := dcimpostgres.NewDeviceTypeRepository(db)
	deviceTypeService, err := dcimapp.NewDeviceTypeService(
		deviceTypeRepository, manufacturerRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed DeviceType service: " + err.Error())
	}
	interfaceTemplateRepository := dcimpostgres.NewInterfaceTemplateRepository(db)
	interfaceTemplateService, err := dcimapp.NewInterfaceTemplateService(
		interfaceTemplateRepository, deviceTypeRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed InterfaceTemplate service: " + err.Error())
	}
	vrfRepository := ipampostgres.NewVRFRepository(db)
	vrfService, err := ipamapp.NewVRFService(
		vrfRepository, unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed VRF service: " + err.Error())
	}
	prefixService, err := ipamapp.NewPrefixService(
		ipampostgres.NewPrefixRepository(db), vrfRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed Prefix service: " + err.Error())
	}
	interfaceRepository := dcimpostgres.NewInterfaceRepository(db)
	deviceRepository := dcimpostgres.NewDeviceRepository(db)
	ipAddressService, err := ipamapp.NewIPAddressService(
		ipampostgres.NewIPAddressRepository(db), vrfRepository, interfaceRepository,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed IPAddress service: " + err.Error())
	}
	interfaceService, err := dcimapp.NewInterfaceService(
		interfaceRepository, deviceRepository, ipAddressService,
		unitOfWork, recorder, authorizer, clock,
	)
	if err != nil {
		panic("compose typed Interface service: " + err.Error())
	}
	deviceService, err := dcimapp.NewDeviceService(
		deviceRepository,
		deviceTypeRepository,
		deviceRoleRepository,
		siteRepository,
		rackRepository,
		interfaceTemplateRepository,
		interfaceRepository,
		interfaceService,
		unitOfWork,
		recorder,
		authorizer,
		clock,
	)
	if err != nil {
		panic("compose typed Device service: " + err.Error())
	}
	return Core{
		Identity: identityService, Sites: siteService, Manufacturers: manufacturerService,
		RackRoles: rackRoleService, RackTypes: rackTypeService, Racks: rackService,
		DeviceRoles:        deviceRoleService,
		DeviceTypes:        deviceTypeService,
		InterfaceTemplates: interfaceTemplateService,
		Devices:            deviceService,
		Interfaces:         interfaceService,
		VRFs:               vrfService, Prefixes: prefixService,
		IPAddresses: ipAddressService,
	}
}
