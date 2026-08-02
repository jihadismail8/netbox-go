/**
 * Content type utilities for NetBox feature mixins.
 * NetBox identifies objects by "app_label.model" content types (e.g. "dcim.site").
 */

const SINGULAR_MAP: Record<string, string> = {
  sites: 'site',
  regions: 'region',
  site_groups: 'sitegroup',
  locations: 'location',
  manufacturers: 'manufacturer',
  device_types: 'devicetype',
  devices: 'device',
  device_roles: 'devicerole',
  racks: 'rack',
  rack_roles: 'rackrole',
  rack_types: 'racktype',
  rack_reservations: 'rackreservation',
  clusters: 'cluster',
  cluster_types: 'clustertype',
  cluster_groups: 'clustergroup',
  virtual_machines: 'virtualmachine',
  interfaces: 'interface',
  vrfs: 'vrf',
  route_targets: 'routetarget',
  prefixes: 'prefix',
  ip_addresses: 'ipaddress',
  ip_ranges: 'iprange',
  aggregates: 'aggregate',
  vlans: 'vlan',
  vlan_groups: 'vlangroup',
  asns: 'asn',
  asn_ranges: 'asnrange',
  tenants: 'tenant',
  tenant_groups: 'tenantgroup',
  contacts: 'contact',
  contact_groups: 'contactgroup',
  contact_roles: 'contactrole',
  circuits: 'circuit',
  circuit_types: 'circuittype',
  circuit_terminations: 'circuittermination',
  providers: 'provider',
  provider_networks: 'providernetwork',
  wireless_lans: 'wirelesslan',
  wireless_lan_groups: 'wirelesslangroup',
  wireless_links: 'wirelesslink',
  tunnels: 'tunnel',
  tunnel_terminations: 'tunneltermination',
  ipsec_profiles: 'ipsecprofile',
  l2vpns: 'l2vpn',
  l2vpn_terminations: 'l2vpntermination',
  platforms: 'platform',
  rirs: 'rir',
  roles: 'role',
  power_panels: 'powerpanel',
  power_feeds: 'powerfeed',
  cables: 'cable',
  consoleserverports: 'consoleserverport',
  consoleports: 'consoleport',
  poweroutlets: 'poweroutlet',
  powerports: 'powerport',
  frontports: 'frontport',
  rearports: 'rearport',
  device_bays: 'devicebay',
  inventory_items: 'inventoryitem',
  inventory_item_roles: 'inventoryitemrole',
  modules: 'module',
  module_bays: 'modulebay',
  module_types: 'moduletype',
  mac_addresses: 'macaddress',
  virtual_device_contexts: 'virtualdevicecontext',
  services: 'service',
  service_templates: 'servicetemplate',
  fhrp_groups: 'fhrpgroup',
  accounts: 'account',
}

/**
 * Build a NetBox content type string ("app_label.model") from a route path.
 * e.g. "/dcim/sites/" -> "dcim.site"
 */
export function buildContentType(routePath: string): string {
  const parts = routePath.split('/').filter(Boolean)
  if (parts.length < 2) return ''
  const app = parts[0]
  // Route paths use hyphens (ip-addresses) but internal names use underscores
  const modelPlural = parts[1].replace(/-/g, '_')
  const model = SINGULAR_MAP[modelPlural] || modelPlural.replace(/s$/, '')
  // Content types never contain underscores (ipam.ipaddress, not ipam.ip_address)
  return `${app}.${model.replace(/_/g, '')}`
}

/**
 * Build the full API content type app_label.model from module + model.
 */
export function contentTypeOf(module: string, model: string): string {
  return `${module}.${model}`
}
