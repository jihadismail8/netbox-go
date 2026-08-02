/**
 * Runtime capability manifest for the published core-workflow-v1 profile.
 *
 * This is intentionally a closed list. Legacy model files may remain as
 * migration reference, but no route, navigation item, or API call may become
 * reachable unless its resource is declared here.
 */
export const CORE_PROFILE_RESOURCES = [
  { module: 'dcim', model: 'site', apiPath: '/dcim/sites/', routePath: '/dcim/sites/' },
  {
    module: 'dcim',
    model: 'manufacturer',
    apiPath: '/dcim/manufacturers/',
    routePath: '/dcim/manufacturers/',
  },
  {
    module: 'dcim',
    model: 'rackrole',
    apiPath: '/dcim/rack-roles/',
    routePath: '/dcim/rack-roles/',
  },
  {
    module: 'dcim',
    model: 'racktype',
    apiPath: '/dcim/rack-types/',
    routePath: '/dcim/rack-types/',
  },
  { module: 'dcim', model: 'rack', apiPath: '/dcim/racks/', routePath: '/dcim/racks/' },
  {
    module: 'dcim',
    model: 'devicerole',
    apiPath: '/dcim/device-roles/',
    routePath: '/dcim/device-roles/',
  },
  {
    module: 'dcim',
    model: 'devicetype',
    apiPath: '/dcim/device-types/',
    routePath: '/dcim/device-types/',
  },
  {
    module: 'dcim',
    model: 'interfacetemplate',
    apiPath: '/dcim/interface-templates/',
    routePath: '/dcim/interface-templates/',
  },
  { module: 'dcim', model: 'device', apiPath: '/dcim/devices/', routePath: '/dcim/devices/' },
  {
    module: 'dcim',
    model: 'interface',
    apiPath: '/dcim/interfaces/',
    routePath: '/dcim/interfaces/',
  },
  { module: 'ipam', model: 'vrf', apiPath: '/ipam/vrfs/', routePath: '/ipam/vrfs/' },
  {
    module: 'ipam',
    model: 'prefix',
    apiPath: '/ipam/prefixes/',
    routePath: '/ipam/prefixes/',
  },
  {
    module: 'ipam',
    model: 'ipaddress',
    apiPath: '/ipam/ip-addresses/',
    routePath: '/ipam/ip-addresses/',
  },
] as const

export type CoreProfileResource = (typeof CORE_PROFILE_RESOURCES)[number]
export type CoreProfileResourceName = CoreProfileResource['model']
export type CoreProfileModule = CoreProfileResource['module']

export const CORE_PROFILE_RESOURCE_NAMES = CORE_PROFILE_RESOURCES.map(
  (resource) => resource.model,
) as CoreProfileResourceName[]

export function getCoreProfileResource(model: CoreProfileResourceName): CoreProfileResource {
  const resource = CORE_PROFILE_RESOURCES.find((candidate) => candidate.model === model)
  if (!resource) throw new Error(`Undeclared core profile resource: ${model}`)
  return resource
}

export function isCoreProfileResource(value: string): value is CoreProfileResourceName {
  return CORE_PROFILE_RESOURCES.some((resource) => resource.model === value)
}
