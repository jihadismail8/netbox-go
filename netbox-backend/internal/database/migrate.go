// Package database provides database auto-migration.
package database

import (
	"context"
	"reflect"

	"github.com/go-dev-frame/sponge/pkg/logger"
	"gorm.io/gorm"

	"netbox-go/internal/adapters/postgres/bootstrap"
	postgreschangelog "netbox-go/internal/adapters/postgres/changelog"
	dcimrow "netbox-go/internal/adapters/postgres/dcim/row"
	identityrows "netbox-go/internal/adapters/postgres/identity"
	ipamrow "netbox-go/internal/adapters/postgres/ipam/row"
	"netbox-go/internal/model"
)

// AutoMigrate runs GORM auto-migration for all registered models.
// It creates each missing table independently and never passes an existing
// table to AutoMigrate, so existing schemas remain untouched.
func AutoMigrate() error {
	return AutoMigrateWithDB(GetDB())
}

// getAllModels returns every legacy model in deterministic bootstrap order.
func getAllModels() []any {
	allModels := []any{
		// auth (Django)
		&model.AuthGroup{},
		&model.AuthGroupPermissions{},
		&model.AuthPermission{},

		// circuits
		&model.CircuitsCircuit{},
		&model.CircuitsCircuitgroup{},
		&model.CircuitsCircuitgroupassignment{},
		&model.CircuitsCircuittermination{},
		&model.CircuitsCircuittype{},
		&model.CircuitsProvider{},
		&model.CircuitsProvideraccount{},
		&model.CircuitsProviderAsns{},
		&model.CircuitsProvidernetwork{},
		&model.CircuitsVirtualcircuit{},
		&model.CircuitsVirtualcircuittermination{},
		&model.CircuitsVirtualcircuittype{},

		// core
		&model.CoreAutosyncrecord{},
		&model.CoreConfigrevision{},
		&model.CoreDatafile{},
		&model.CoreDatasource{},
		&model.CoreJob{},
		&model.CoreManagedfile{},
		&model.CoreObjectchange{},
		&model.CoreObjecttype{},

		// dcim
		&model.DcimCable{},
		&model.DcimCablepath{},
		&model.DcimCabletermination{},
		&model.DcimConsoleport{},
		&model.DcimConsoleporttemplate{},
		&model.DcimConsoleserverport{},
		&model.DcimConsoleserverporttemplate{},
		&model.DcimDevicebay{},
		&model.DcimDevicebaytemplate{},
		&model.DcimFrontport{},
		&model.DcimFrontporttemplate{},
		&model.DcimInterfaceTaggedVlans{},
		&model.DcimInterfaceVdcs{},
		&model.DcimInterfaceWirelessLans{},
		&model.DcimInventoryitem{},
		&model.DcimInventoryitemrole{},
		&model.DcimInventoryitemtemplate{},
		&model.DcimLocation{},
		&model.DcimMacaddress{},
		&model.DcimModule{},
		&model.DcimModulebay{},
		&model.DcimModulebaytemplate{},
		&model.DcimModuletype{},
		&model.DcimModuletypeprofile{},
		&model.DcimPlatform{},
		&model.DcimPowerfeed{},
		&model.DcimPoweroutlet{},
		&model.DcimPoweroutlettemplate{},
		&model.DcimPowerpanel{},
		&model.DcimPowerport{},
		&model.DcimPowerporttemplate{},
		&model.DcimRackreservation{},
		&model.DcimRearport{},
		&model.DcimRearporttemplate{},
		&model.DcimRegion{},
		&model.DcimSiteAsns{},
		&model.DcimSitegroup{},
		&model.DcimVirtualchassis{},
		&model.DcimVirtualdevicecontext{},

		// django
		&model.DjangoContentType{},
		&model.DjangoMigrations{},
		&model.DjangoSession{},

		// extras
		&model.ExtrasBookmark{},
		&model.ExtrasCachedvalue{},
		&model.ExtrasConfigcontext{},
		&model.ExtrasConfigcontextClusterGroups{},
		&model.ExtrasConfigcontextClusters{},
		&model.ExtrasConfigcontextClusterTypes{},
		&model.ExtrasConfigcontextDeviceTypes{},
		&model.ExtrasConfigcontextLocations{},
		&model.ExtrasConfigcontextPlatforms{},
		&model.ExtrasConfigcontextprofile{},
		&model.ExtrasConfigcontextRegions{},
		&model.ExtrasConfigcontextRoles{},
		&model.ExtrasConfigcontextSiteGroups{},
		&model.ExtrasConfigcontextSites{},
		&model.ExtrasConfigcontextTags{},
		&model.ExtrasConfigcontextTenantGroups{},
		&model.ExtrasConfigcontextTenants{},
		&model.ExtrasConfigtemplate{},
		&model.ExtrasCustomfield{},
		&model.ExtrasCustomfieldchoiceset{},
		&model.ExtrasCustomfieldObjectTypes{},
		&model.ExtrasCustomlink{},
		&model.ExtrasCustomlinkObjectTypes{},
		&model.ExtrasDashboard{},
		&model.ExtrasEventrule{},
		&model.ExtrasEventruleObjectTypes{},
		&model.ExtrasExporttemplate{},
		&model.ExtrasExporttemplateObjectTypes{},
		&model.ExtrasImageattachment{},
		&model.ExtrasJournalentry{},
		&model.ExtrasNotification{},
		&model.ExtrasNotificationgroup{},
		&model.ExtrasNotificationgroupGroups{},
		&model.ExtrasNotificationgroupUsers{},
		&model.ExtrasSavedfilter{},
		&model.ExtrasSavedfilterObjectTypes{},
		&model.ExtrasScript{},
		&model.ExtrasSubscription{},
		&model.ExtrasTableconfig{},
		&model.ExtrasTag{},
		&model.ExtrasTaggeditem{},
		&model.ExtrasTagObjectTypes{},
		&model.ExtrasWebhook{},

		// ipam
		&model.IpamAggregate{},
		&model.IpamAsn{},
		&model.IpamAsnrange{},
		&model.IpamFhrpgroup{},
		&model.IpamFhrpgroupassignment{},
		&model.IpamIprange{},
		&model.IpamRir{},
		&model.IpamRole{},
		&model.IpamRoutetarget{},
		&model.IpamService{},
		&model.IpamServiceIpaddresses{},
		&model.IpamServicetemplate{},
		&model.IpamVlan{},
		&model.IpamVlangroup{},
		&model.IpamVlantranslationpolicy{},
		&model.IpamVlantranslationrule{},
		&model.IpamVrfExportTargets{},
		&model.IpamVrfImportTargets{},

		// social_auth
		&model.SocialAuthAssociation{},
		&model.SocialAuthCode{},
		&model.SocialAuthNonce{},
		&model.SocialAuthPartial{},
		&model.SocialAuthUsersocialauth{},

		// taggit
		&model.TaggitTag{},
		&model.TaggitTaggeditem{},

		// tenancy
		&model.TenancyContact{},
		&model.TenancyContactassignment{},
		&model.TenancyContactgroup{},
		&model.TenancyContactGroups{},
		&model.TenancyContactrole{},
		&model.TenancyTenant{},
		&model.TenancyTenantgroup{},

		// thumbnail
		&model.ThumbnailKvstore{},

		// users
		&model.UsersGroup{},
		&model.UsersGroupObjectPermissions{},
		&model.UsersGroupPermissions{},
		&model.UsersObjectpermission{},
		&model.UsersObjectpermissionObjectTypes{},
		&model.UsersToken{},
		&model.UsersUser{},
		&model.UsersUserconfig{},
		&model.UsersUserGroups{},
		&model.UsersUserObjectPermissions{},
		&model.UsersUserUserPermissions{},

		// virtualization
		&model.VirtualizationCluster{},
		&model.VirtualizationClustergroup{},
		&model.VirtualizationClustertype{},
		&model.VirtualizationVirtualdisk{},
		&model.VirtualizationVirtualmachine{},
		&model.VirtualizationVminterface{},
		&model.VirtualizationVminterfaceTaggedVlans{},

		// vpn
		&model.VpnIkepolicy{},
		&model.VpnIkepolicyProposals{},
		&model.VpnIkeproposal{},
		&model.VpnIpsecpolicy{},
		&model.VpnIpsecpolicyProposals{},
		&model.VpnIpsecprofile{},
		&model.VpnIpsecproposal{},
		&model.VpnL2Vpn{},
		&model.VpnL2VpnExportTargets{},
		&model.VpnL2VpnImportTargets{},
		&model.VpnL2Vpntermination{},
		&model.VpnTunnel{},
		&model.VpnTunnelgroup{},
		&model.VpnTunneltermination{},

		// wireless
		&model.WirelessWirelesslan{},
		&model.WirelessWirelesslangroup{},
		&model.WirelessWirelesslink{},
	}

	return allModels
}

// AutoMigrateWithDB allows running migration with a specific *gorm.DB instance.
// Useful for testing.
func AutoMigrateWithDB(db *gorm.DB) error {
	registry, err := modelRegistry()
	if err != nil {
		return err
	}
	result, err := bootstrap.Run(context.Background(), db, registry)
	if err != nil {
		return err
	}
	logger.Info(
		"[auto-migrate] database bootstrap completed",
		logger.Int("createdTableCount", len(result.Created)),
		logger.Int("existingTableCount", len(result.Existing)),
	)
	return nil
}

func modelRegistry() (bootstrap.Registry, error) {
	models := getAllModels()
	entries := make([]bootstrap.Entry, 0, len(models)+22)
	for _, registeredModel := range models {
		entries = append(entries, bootstrap.Entry{
			Name:  reflect.TypeOf(registeredModel).String(),
			Model: registeredModel,
		})
	}
	// Go-owned tables for the supported core workflow. These are intentionally
	// private adapter rows rather than public/domain models.
	entries = append(entries,
		bootstrap.Entry{Name: "go_identity_users", Model: &identityrows.UserRow{}},
		bootstrap.Entry{Name: "go_identity_groups", Model: &identityrows.GroupRow{}},
		bootstrap.Entry{Name: "go_identity_permission_grants", Model: &identityrows.PermissionGrantRow{}},
		bootstrap.Entry{Name: "go_identity_group_memberships", Model: &identityrows.GroupMembershipRow{}, Dependencies: []string{"go_identity_users", "go_identity_groups"}},
		bootstrap.Entry{Name: "go_identity_user_permission_grants", Model: &identityrows.UserPermissionGrantRow{}, Dependencies: []string{"go_identity_users", "go_identity_permission_grants"}},
		bootstrap.Entry{Name: "go_identity_group_permission_grants", Model: &identityrows.GroupPermissionGrantRow{}, Dependencies: []string{"go_identity_groups", "go_identity_permission_grants"}},
		bootstrap.Entry{Name: "go_identity_tokens", Model: &identityrows.TokenRow{}, Dependencies: []string{"go_identity_users"}},
		bootstrap.Entry{Name: "go_identity_sessions", Model: &identityrows.SessionRow{}, Dependencies: []string{"go_identity_users"}},
	)
	dcimTables := dcimrow.Descriptors()
	ipamTables := ipamrow.Descriptors()
	changeDependencies := make([]string, 1, len(dcimTables)+len(ipamTables)+1)
	changeDependencies[0] = "go_identity_users"
	for _, table := range dcimTables {
		entries = append(entries, bootstrap.Entry{
			Name:         table.Name,
			Model:        table.Model,
			Dependencies: table.Dependencies,
		})
		changeDependencies = append(changeDependencies, table.Name)
	}
	for _, table := range ipamTables {
		entries = append(entries, bootstrap.Entry{
			Name:         table.Name,
			Model:        table.Model,
			Dependencies: table.Dependencies,
		})
		changeDependencies = append(changeDependencies, table.Name)
	}
	entries = append(entries, bootstrap.Entry{
		Name:         "go_object_changes",
		Model:        &postgreschangelog.ChangeRow{},
		Dependencies: changeDependencies,
	})
	return bootstrap.NewRegistry(entries...)
}
