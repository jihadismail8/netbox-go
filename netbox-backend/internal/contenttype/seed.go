// Package contenttype seeds Django's django_content_type table and provides
// a resolver for GenericFK lookups.
//
// django_content_type is Django's registry of "which models exist" — a tiny
// lookup table of (id, app_label, model). Each concrete Python model class
// gets exactly one row, allocated a stable numeric ID at migration time. The
// table powers Django's GenericForeignKey: a column like object_type_id points
// at a row in this table, which says "the object is a dcim.device", and a
// companion object_id column says "...row #42".
//
// NetBox uses GenericFKs in 27+ models: cables/cable-paths (a cable can
// terminate on an interface OR a console port OR ...), custom fields
// (applied to specific content types), tags, change log, bookmarks, journal,
// webhooks, scoped objects (prefix/vlan/cluster scope), and more.
//
// Python populates this table via `python manage.py migrate`. Since this is a
// Go-only deployment, we seed it ourselves from a static list derived from the
// Python model classes — exactly the rows Django would create, minus abstract
// base classes, junction/M2M tables, and Django/third-party infrastructure.
package contenttype

// Type describes one content-type entry: a (app_label, model) pair plus the
// PostgreSQL table that backs it.
type Type struct {
	AppLabel string // Django app label, e.g. "dcim"
	Model    string // lowercase model name, e.g. "interface" (Django's convention)
	Table    string // PostgreSQL table name, e.g. "dcim_interface"
}

// allTypes is the authoritative list of concrete NetBox models that get a
// django_content_type row. Derived from the Python model classes (see
// docs/contenttype-seed-derivation.txt for the extraction method), excluding:
//   - abstract base classes (PathEndpoint, ComponentModel, RackBase, ...)
//   - junction/M2M tables (dcim_interface_tagged_vlans, extras_configcontext_sites, ...)
//   - Django infrastructure (auth_group, django_session, django_content_type itself)
//   - third-party (social_auth_*, taggit_tag, thumbnail_kvstore)
//
// Django's model-name rule: class ConsolePort → "consoleport" (lowercase, no
// separator). The table name follows the app_model convention.
var allTypes = []Type{
	// ── dcim (44) ──────────────────────────────────────────────
	{"dcim", "cable", "dcim_cable"},
	{"dcim", "cablepath", "dcim_cablepath"},
	{"dcim", "cabletermination", "dcim_cabletermination"},
	{"dcim", "consoleport", "dcim_consoleport"},
	{"dcim", "consoleporttemplate", "dcim_consoleporttemplate"},
	{"dcim", "consoleserverport", "dcim_consoleserverport"},
	{"dcim", "consoleserverporttemplate", "dcim_consoleserverporttemplate"},
	{"dcim", "device", "dcim_device"},
	{"dcim", "devicebay", "dcim_devicebay"},
	{"dcim", "devicebaytemplate", "dcim_devicebaytemplate"},
	{"dcim", "devicerole", "dcim_devicerole"},
	{"dcim", "devicetype", "dcim_devicetype"},
	{"dcim", "frontport", "dcim_frontport"},
	{"dcim", "frontporttemplate", "dcim_frontporttemplate"},
	{"dcim", "interface", "dcim_interface"},
	{"dcim", "interfacetemplate", "dcim_interfacetemplate"},
	{"dcim", "inventoryitem", "dcim_inventoryitem"},
	{"dcim", "inventoryitemrole", "dcim_inventoryitemrole"},
	{"dcim", "inventoryitemtemplate", "dcim_inventoryitemtemplate"},
	{"dcim", "location", "dcim_location"},
	{"dcim", "macaddress", "dcim_macaddress"},
	{"dcim", "manufacturer", "dcim_manufacturer"},
	{"dcim", "module", "dcim_module"},
	{"dcim", "modulebay", "dcim_modulebay"},
	{"dcim", "modulebaytemplate", "dcim_modulebaytemplate"},
	{"dcim", "moduletype", "dcim_moduletype"},
	{"dcim", "moduletypeprofile", "dcim_moduletypeprofile"},
	{"dcim", "platform", "dcim_platform"},
	{"dcim", "powerfeed", "dcim_powerfeed"},
	{"dcim", "poweroutlet", "dcim_poweroutlet"},
	{"dcim", "poweroutlettemplate", "dcim_poweroutlettemplate"},
	{"dcim", "powerpanel", "dcim_powerpanel"},
	{"dcim", "powerport", "dcim_powerport"},
	{"dcim", "powerporttemplate", "dcim_powerporttemplate"},
	{"dcim", "rack", "dcim_rack"},
	{"dcim", "rackreservation", "dcim_rackreservation"},
	{"dcim", "rackrole", "dcim_rackrole"},
	{"dcim", "racktype", "dcim_racktype"},
	{"dcim", "rearport", "dcim_rearport"},
	{"dcim", "rearporttemplate", "dcim_rearporttemplate"},
	{"dcim", "region", "dcim_region"},
	{"dcim", "site", "dcim_site"},
	{"dcim", "sitegroup", "dcim_sitegroup"},
	{"dcim", "virtualchassis", "dcim_virtualchassis"},
	{"dcim", "virtualdevicecontext", "dcim_virtualdevicecontext"},

	// ── ipam (18) ──────────────────────────────────────────────
	{"ipam", "aggregate", "ipam_aggregate"},
	{"ipam", "asn", "ipam_asn"},
	{"ipam", "asnrange", "ipam_asnrange"},
	{"ipam", "fhrpgroup", "ipam_fhrpgroup"},
	{"ipam", "fhrpgroupassignment", "ipam_fhrpgroupassignment"},
	{"ipam", "ipaddress", "ipam_ipaddress"},
	{"ipam", "iprange", "ipam_iprange"},
	{"ipam", "prefix", "ipam_prefix"},
	{"ipam", "rir", "ipam_rir"},
	{"ipam", "role", "ipam_role"},
	{"ipam", "routetarget", "ipam_routetarget"},
	{"ipam", "service", "ipam_service"},
	{"ipam", "servicetemplate", "ipam_servicetemplate"},
	{"ipam", "vlan", "ipam_vlan"},
	{"ipam", "vlangroup", "ipam_vlangroup"},
	{"ipam", "vlantranslationpolicy", "ipam_vlantranslationpolicy"},
	{"ipam", "vlantranslationrule", "ipam_vlantranslationrule"},
	{"ipam", "vrf", "ipam_vrf"},

	// ── circuits (11) ──────────────────────────────────────────
	{"circuits", "circuit", "circuits_circuit"},
	{"circuits", "circuitgroup", "circuits_circuitgroup"},
	{"circuits", "circuitgroupassignment", "circuits_circuitgroupassignment"},
	{"circuits", "circuittermination", "circuits_circuittermination"},
	{"circuits", "circuittype", "circuits_circuittype"},
	{"circuits", "provider", "circuits_provider"},
	{"circuits", "provideraccount", "circuits_provideraccount"},
	{"circuits", "providernetwork", "circuits_providernetwork"},
	{"circuits", "virtualcircuit", "circuits_virtualcircuit"},
	{"circuits", "virtualcircuittermination", "circuits_virtualcircuittermination"},
	{"circuits", "virtualcircuittype", "circuits_virtualcircuittype"},

	// ── tenancy (6) ────────────────────────────────────────────
	{"tenancy", "contact", "tenancy_contact"},
	{"tenancy", "contactassignment", "tenancy_contactassignment"},
	{"tenancy", "contactgroup", "tenancy_contactgroup"},
	{"tenancy", "contactrole", "tenancy_contactrole"},
	{"tenancy", "tenant", "tenancy_tenant"},
	{"tenancy", "tenantgroup", "tenancy_tenantgroup"},

	// ── virtualization (6) ─────────────────────────────────────
	{"virtualization", "cluster", "virtualization_cluster"},
	{"virtualization", "clustergroup", "virtualization_clustergroup"},
	{"virtualization", "clustertype", "virtualization_clustertype"},
	{"virtualization", "virtualdisk", "virtualization_virtualdisk"},
	{"virtualization", "virtualmachine", "virtualization_virtualmachine"},
	{"virtualization", "vminterface", "virtualization_vminterface"},

	// ── vpn (10) ───────────────────────────────────────────────
	{"vpn", "ikepolicy", "vpn_ikepolicy"},
	{"vpn", "ikeproposal", "vpn_ikeproposal"},
	{"vpn", "ipsecpolicy", "vpn_ipsecpolicy"},
	{"vpn", "ipsecprofile", "vpn_ipsecprofile"},
	{"vpn", "ipsecproposal", "vpn_ipsecproposal"},
	{"vpn", "l2vpn", "vpn_l2vpn"},
	{"vpn", "l2vpntermination", "vpn_l2vpntermination"},
	{"vpn", "tunnel", "vpn_tunnel"},
	{"vpn", "tunnelgroup", "vpn_tunnelgroup"},
	{"vpn", "tunneltermination", "vpn_tunneltermination"},

	// ── wireless (3) ───────────────────────────────────────────
	{"wireless", "wirelesslan", "wireless_wirelesslan"},
	{"wireless", "wirelesslangroup", "wireless_wirelesslangroup"},
	{"wireless", "wirelesslink", "wireless_wirelesslink"},

	// ── extras (18) ────────────────────────────────────────────
	{"extras", "bookmark", "extras_bookmark"},
	{"extras", "cachedvalue", "extras_cachedvalue"},
	{"extras", "configcontext", "extras_configcontext"},
	{"extras", "configcontextprofile", "extras_configcontextprofile"},
	{"extras", "configtemplate", "extras_configtemplate"},
	{"extras", "customfield", "extras_customfield"},
	{"extras", "customfieldchoiceset", "extras_customfieldchoiceset"},
	{"extras", "customlink", "extras_customlink"},
	{"extras", "dashboard", "extras_dashboard"},
	{"extras", "eventrule", "extras_eventrule"},
	{"extras", "exporttemplate", "extras_exporttemplate"},
	{"extras", "imageattachment", "extras_imageattachment"},
	{"extras", "journalentry", "extras_journalentry"},
	{"extras", "notification", "extras_notification"},
	{"extras", "notificationgroup", "extras_notificationgroup"},
	{"extras", "script", "extras_script"},
	{"extras", "subscription", "extras_subscription"},
	{"extras", "tag", "extras_tag"},
	{"extras", "taggeditem", "extras_taggeditem"},
	{"extras", "webhook", "extras_webhook"},

	// ── users (6) ──────────────────────────────────────────────
	{"users", "user", "users_user"},
	{"users", "group", "users_group"},
	{"users", "objectpermission", "users_objectpermission"},
	{"users", "token", "users_token"},
	{"users", "userconfig", "users_userconfig"},

	// ── core (8) ───────────────────────────────────────────────
	{"core", "autosyncrecord", "core_autosyncrecord"},
	{"core", "configrevision", "core_configrevision"},
	{"core", "datafile", "core_datafile"},
	{"core", "datasource", "core_datasource"},
	{"core", "job", "core_job"},
	{"core", "managedfile", "core_managedfile"},
	{"core", "objectchange", "core_objectchange"},
	{"core", "objecttype", "core_objecttype"},
}

// All returns the full static seed list. Exposed for inspection/tests.
func All() []Type {
	return allTypes
}
