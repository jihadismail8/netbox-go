import type {
  ChoiceOption,
  DetailFieldDef,
  FilterField,
  FormFieldDef,
  ModelConfig,
  TableColumn,
} from '@/types'
import {
  CORE_PROFILE_RESOURCE_NAMES,
  getCoreProfileResource,
  type CoreProfileResourceName,
} from '@/features/core/manifest'
import { INTERFACE_TYPE_CHOICES } from '@/features/dcim/interface-types'
import type {
  CoreFilterFieldName,
  CoreFormFieldName,
  CoreResponseFieldName,
  CoreResponseFieldNameFor,
} from '@/features/core/resources'

export { CORE_PROFILE_RESOURCE_NAMES }
export type { CoreProfileResourceName }

/**
 * The only resources published by core-workflow-v1.
 *
 * Keep this registry explicit: adding a model to the legacy NetBox catalogue must
 * never make it reachable from the supported application by accident.
 */
const infrastructureStatuses: ChoiceOption[] = [
  { value: 'planned', label: 'Planned' },
  { value: 'staging', label: 'Staging' },
  { value: 'active', label: 'Active' },
  { value: 'decommissioning', label: 'Decommissioning' },
  { value: 'retired', label: 'Retired' },
]

const deviceStatuses: ChoiceOption[] = [
  { value: 'offline', label: 'Offline' },
  { value: 'active', label: 'Active' },
  { value: 'planned', label: 'Planned' },
  { value: 'staged', label: 'Staged' },
  { value: 'failed', label: 'Failed' },
  { value: 'inventory', label: 'Inventory' },
  { value: 'decommissioning', label: 'Decommissioning' },
]

const rackStatuses: ChoiceOption[] = [
  { value: 'reserved', label: 'Reserved' },
  { value: 'available', label: 'Available' },
  { value: 'planned', label: 'Planned' },
  { value: 'active', label: 'Active' },
  { value: 'deprecated', label: 'Deprecated' },
]

const prefixStatuses: ChoiceOption[] = [
  { value: 'container', label: 'Container' },
  { value: 'active', label: 'Active' },
  { value: 'reserved', label: 'Reserved' },
  { value: 'deprecated', label: 'Deprecated' },
]

const ipAddressStatuses: ChoiceOption[] = [
  { value: 'active', label: 'Active' },
  { value: 'reserved', label: 'Reserved' },
  { value: 'deprecated', label: 'Deprecated' },
  { value: 'dhcp', label: 'DHCP' },
  { value: 'slaac', label: 'SLAAC' },
]

const ipAddressRoles: ChoiceOption[] = [
  { value: 'loopback', label: 'Loopback' },
  { value: 'secondary', label: 'Secondary' },
  { value: 'anycast', label: 'Anycast' },
  { value: 'vip', label: 'VIP' },
  { value: 'vrrp', label: 'VRRP' },
  { value: 'hsrp', label: 'HSRP' },
  { value: 'glbp', label: 'GLBP' },
  { value: 'carp', label: 'CARP' },
]

const rackFormFactors: ChoiceOption[] = [
  { value: '2-post-frame', label: '2-post frame' },
  { value: '4-post-frame', label: '4-post frame' },
  { value: '4-post-cabinet', label: '4-post cabinet' },
  { value: 'wall-frame', label: 'Wall-mounted frame' },
  { value: 'wall-frame-vertical', label: 'Wall-mounted frame (vertical)' },
  { value: 'wall-cabinet', label: 'Wall-mounted cabinet' },
  { value: 'wall-cabinet-vertical', label: 'Wall-mounted cabinet (vertical)' },
]

const rackWidths: ChoiceOption[] = [
  { value: 10, label: '10 inches' },
  { value: 19, label: '19 inches' },
  { value: 21, label: '21 inches' },
  { value: 23, label: '23 inches' },
]

const airflowChoices: ChoiceOption[] = [
  { value: 'front-to-rear', label: 'Front to rear' },
  { value: 'rear-to-front', label: 'Rear to front' },
  { value: 'left-to-right', label: 'Left to right' },
  { value: 'right-to-left', label: 'Right to left' },
  { value: 'side-to-rear', label: 'Side to rear' },
  { value: 'rear-to-side', label: 'Rear to side' },
  { value: 'bottom-to-top', label: 'Bottom to top' },
  { value: 'top-to-bottom', label: 'Top to bottom' },
  { value: 'passive', label: 'Passive' },
  { value: 'mixed', label: 'Mixed' },
]

const rackAirflowChoices: ChoiceOption[] = [
  { value: 'front-to-rear', label: 'Front to rear' },
  { value: 'rear-to-front', label: 'Rear to front' },
]

type CoreDetailProjectionMap = {
  [N in CoreProfileResourceName]: readonly DetailFieldDef<CoreResponseFieldNameFor<N>>[]
}

/**
 * Detail projections cover every promoted business field. Common transport
 * metadata is owned by the page header and timestamp footer instead.
 *
 * Keep these separate from table columns: collection views are intentionally
 * compact and must not determine what is visible on a single-object page.
 */
export const CORE_PROFILE_DETAIL_PROJECTIONS = {
  site: [
    { key: 'name', label: 'Name' },
    { key: 'slug', label: 'Slug' },
    { key: 'status', label: 'Status', presentation: 'status' },
    { key: 'facility', label: 'Facility' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'device_count', label: 'Devices' },
    { key: 'prefix_count', label: 'Prefixes' },
    { key: 'rack_count', label: 'Racks' },
  ],
  manufacturer: [
    { key: 'name', label: 'Name' },
    { key: 'slug', label: 'Slug' },
    { key: 'description', label: 'Description' },
    { key: 'devicetype_count', label: 'Device Types' },
  ],
  rackrole: [
    { key: 'name', label: 'Name' },
    { key: 'slug', label: 'Slug' },
    { key: 'color', label: 'Color' },
    { key: 'description', label: 'Description' },
    { key: 'rack_count', label: 'Racks' },
  ],
  racktype: [
    { key: 'manufacturer', label: 'Manufacturer' },
    { key: 'model', label: 'Model' },
    { key: 'slug', label: 'Slug' },
    { key: 'form_factor', label: 'Form Factor' },
    { key: 'width', label: 'Width' },
    { key: 'u_height', label: 'Height (U)' },
    { key: 'starting_unit', label: 'Starting Unit' },
    { key: 'desc_units', label: 'Descending Units' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
  ],
  rack: [
    { key: 'site', label: 'Site' },
    { key: 'name', label: 'Name' },
    { key: 'facility_id', label: 'Facility ID' },
    { key: 'rack_type', label: 'Rack Type' },
    { key: 'status', label: 'Status', presentation: 'status' },
    { key: 'role', label: 'Role' },
    { key: 'serial', label: 'Serial Number' },
    { key: 'asset_tag', label: 'Asset Tag' },
    { key: 'form_factor', label: 'Form Factor' },
    { key: 'width', label: 'Width' },
    { key: 'u_height', label: 'Height (U)' },
    { key: 'starting_unit', label: 'Starting Unit' },
    { key: 'desc_units', label: 'Descending Units' },
    { key: 'airflow', label: 'Airflow' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'device_count', label: 'Devices' },
  ],
  devicerole: [
    { key: 'parent', label: 'Parent Role' },
    { key: 'name', label: 'Name' },
    { key: 'slug', label: 'Slug' },
    { key: 'color', label: 'Color' },
    { key: 'vm_role', label: 'VM Role' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'device_count', label: 'Devices' },
    { key: '_depth', label: 'Hierarchy Depth' },
  ],
  devicetype: [
    { key: 'manufacturer', label: 'Manufacturer' },
    { key: 'model', label: 'Model' },
    { key: 'slug', label: 'Slug' },
    { key: 'part_number', label: 'Part Number' },
    { key: 'u_height', label: 'Height (U)' },
    { key: 'exclude_from_utilization', label: 'Exclude from Utilization' },
    { key: 'is_full_depth', label: 'Full Depth' },
    { key: 'airflow', label: 'Airflow' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'device_count', label: 'Devices' },
    { key: 'interface_template_count', label: 'Interface Templates' },
  ],
  interfacetemplate: [
    { key: 'device_type', label: 'Device Type' },
    { key: 'name', label: 'Name' },
    { key: 'label', label: 'Label' },
    { key: 'type', label: 'Type' },
    { key: 'enabled', label: 'Enabled' },
    { key: 'mgmt_only', label: 'Management Only' },
    { key: 'description', label: 'Description' },
  ],
  device: [
    { key: 'device_type', label: 'Device Type' },
    { key: 'role', label: 'Device Role' },
    { key: 'name', label: 'Name' },
    { key: 'site', label: 'Site' },
    { key: 'rack', label: 'Rack' },
    { key: 'position', label: 'Position (U)' },
    { key: 'face', label: 'Face' },
    { key: 'status', label: 'Status', presentation: 'status' },
    { key: 'serial', label: 'Serial Number' },
    { key: 'asset_tag', label: 'Asset Tag' },
    { key: 'airflow', label: 'Airflow' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'interface_count', label: 'Interfaces' },
  ],
  interface: [
    { key: 'device', label: 'Device' },
    { key: 'name', label: 'Name' },
    { key: 'label', label: 'Label' },
    { key: 'type', label: 'Type' },
    { key: 'enabled', label: 'Enabled' },
    { key: 'mgmt_only', label: 'Management Only' },
    { key: 'mtu', label: 'MTU' },
    { key: 'speed', label: 'Speed (Kbps)' },
    { key: 'duplex', label: 'Duplex' },
    { key: 'description', label: 'Description' },
    { key: 'count_ipaddresses', label: 'IP Addresses' },
  ],
  vrf: [
    { key: 'name', label: 'Name' },
    { key: 'rd', label: 'Route Distinguisher' },
    { key: 'enforce_unique', label: 'Enforce Unique' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'prefix_count', label: 'Prefixes' },
    { key: 'ipaddress_count', label: 'IP Addresses' },
  ],
  prefix: [
    { key: 'prefix', label: 'Prefix' },
    { key: 'vrf', label: 'VRF' },
    { key: 'status', label: 'Status', presentation: 'status' },
    { key: 'is_pool', label: 'Pool' },
    { key: 'mark_utilized', label: 'Mark Utilized' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'family', label: 'Family' },
    { key: 'children', label: 'Children' },
    { key: '_depth', label: 'Hierarchy Depth' },
  ],
  ipaddress: [
    { key: 'address', label: 'Address' },
    { key: 'vrf', label: 'VRF' },
    { key: 'status', label: 'Status', presentation: 'status' },
    { key: 'role', label: 'Role' },
    { key: 'dns_name', label: 'DNS Name' },
    { key: 'description', label: 'Description' },
    { key: 'comments', label: 'Comments', presentation: 'markdown' },
    { key: 'assigned_object_type', label: 'Assigned Object Type' },
    { key: 'assigned_object_id', label: 'Assigned Object ID' },
    { key: 'family', label: 'Family' },
    { key: 'assigned_object', label: 'Assigned Interface' },
  ],
} satisfies CoreDetailProjectionMap

const column = (key: CoreResponseFieldName, label: string, sortable = false): TableColumn => ({
  key,
  label,
  sortable,
})

const textFilter = (key: CoreFilterFieldName, label: string): FilterField => ({
  key,
  label,
  type: 'text',
})
const booleanFilter = (key: CoreFilterFieldName, label: string): FilterField => ({
  key,
  label,
  type: 'boolean',
})
const selectFilter = (
  key: CoreFilterFieldName,
  label: string,
  options: ChoiceOption[],
): FilterField => ({
  key,
  label,
  type: 'select',
  options,
})
const relationFilter = (
  key: CoreFilterFieldName,
  label: string,
  relationResource: CoreProfileResourceName,
): FilterField => ({
  key,
  label,
  type: 'api-select',
  relationResource,
})

const textField = (
  key: CoreFormFieldName,
  label: string,
  group: string,
  options: Partial<FormFieldDef> = {},
): FormFieldDef => ({ key, label, type: 'text', group, ...options })
const numberField = (
  key: CoreFormFieldName,
  label: string,
  group: string,
  options: Partial<FormFieldDef> = {},
): FormFieldDef => ({ key, label, type: 'number', group, ...options })
const booleanField = (
  key: CoreFormFieldName,
  label: string,
  group: string,
  options: Partial<FormFieldDef> = {},
): FormFieldDef => ({ key, label, type: 'boolean', group, ...options })
const selectField = (
  key: CoreFormFieldName,
  label: string,
  group: string,
  options: ChoiceOption[],
  extra: Partial<FormFieldDef> = {},
): FormFieldDef => ({ key, label, type: 'select', group, options, ...extra })
const relationField = (
  key: CoreFormFieldName,
  label: string,
  group: string,
  relationResource: CoreProfileResourceName,
  options: Partial<FormFieldDef> = {},
): FormFieldDef => ({ key, label, type: 'api-select', group, relationResource, ...options })
const descriptionField = (group: string): FormFieldDef =>
  textField('description', 'Description', group)
const commentsField = (group: string): FormFieldDef => ({
  key: 'comments',
  label: 'Comments',
  type: 'markdown',
  group,
})

const coreModel = (
  module: 'dcim' | 'ipam',
  model: CoreProfileResourceName,
  displayName: string,
  displayNamePlural: string,
  path: string,
  columns: TableColumn[],
  filters: FilterField[],
  fields: FormFieldDef[],
  statusChoices?: ChoiceOption[],
  writableFields: string[] = fields.map((field) => field.key),
): ModelConfig => {
  const manifestResource = getCoreProfileResource(model)
  if (manifestResource.module !== module || manifestResource.routePath !== `/${module}/${path}/`) {
    throw new Error(`Core profile config drift for ${model}`)
  }
  return {
    module,
    model,
    display_name: displayName,
    display_name_plural: displayNamePlural,
    apiPath: manifestResource.apiPath,
    routePath: manifestResource.routePath,
    writableFields,
    columns,
    detailFields: [...CORE_PROFILE_DETAIL_PROJECTIONS[model]],
    filters,
    fields,
    ...(statusChoices ? { statusChoices } : {}),
  }
}

export const CORE_PROFILE_MODELS: ModelConfig[] = [
  coreModel(
    'dcim',
    'site',
    'Site',
    'Sites',
    'sites',
    [
      column('name', 'Name', true),
      column('slug', 'Slug', true),
      column('status', 'Status', true),
      column('facility', 'Facility'),
      column('device_count', 'Devices'),
      column('rack_count', 'Racks'),
      column('prefix_count', 'Prefixes'),
    ],
    [
      textFilter('name', 'Name'),
      textFilter('slug', 'Slug'),
      selectFilter('status', 'Status', infrastructureStatuses),
    ],
    [
      textField('name', 'Name', 'Site', { required: true }),
      { key: 'slug', label: 'Slug', type: 'slug', group: 'Site', required: true },
      selectField('status', 'Status', 'Site', infrastructureStatuses, {
        required: true,
        default: 'active',
      }),
      textField('facility', 'Facility', 'Site'),
      descriptionField('Site'),
      commentsField('Site'),
    ],
    infrastructureStatuses,
  ),
  coreModel(
    'dcim',
    'manufacturer',
    'Manufacturer',
    'Manufacturers',
    'manufacturers',
    [
      column('name', 'Name', true),
      column('slug', 'Slug', true),
      column('description', 'Description'),
      column('devicetype_count', 'Device Types'),
    ],
    [textFilter('name', 'Name'), textFilter('slug', 'Slug')],
    [
      textField('name', 'Name', 'Manufacturer', { required: true }),
      { key: 'slug', label: 'Slug', type: 'slug', group: 'Manufacturer', required: true },
      descriptionField('Manufacturer'),
    ],
  ),
  coreModel(
    'dcim',
    'rackrole',
    'Rack Role',
    'Rack Roles',
    'rack-roles',
    [
      column('name', 'Name', true),
      column('slug', 'Slug', true),
      column('color', 'Color'),
      column('rack_count', 'Racks'),
      column('description', 'Description'),
    ],
    [textFilter('name', 'Name'), textFilter('slug', 'Slug')],
    [
      textField('name', 'Name', 'Rack Role', { required: true }),
      { key: 'slug', label: 'Slug', type: 'slug', group: 'Rack Role', required: true },
      textField('color', 'Color', 'Rack Role', { default: '9e9e9e' }),
      descriptionField('Rack Role'),
    ],
  ),
  coreModel(
    'dcim',
    'racktype',
    'Rack Type',
    'Rack Types',
    'rack-types',
    [
      column('manufacturer', 'Manufacturer', true),
      column('model', 'Model', true),
      column('slug', 'Slug', true),
      column('form_factor', 'Form Factor'),
      column('width', 'Width'),
      column('u_height', 'Height (U)', true),
    ],
    [
      relationFilter('manufacturer_id', 'Manufacturer', 'manufacturer'),
      textFilter('manufacturer_slug', 'Manufacturer Slug'),
      textFilter('model', 'Model'),
      textFilter('slug', 'Slug'),
    ],
    [
      relationField('manufacturer', 'Manufacturer', 'Rack Type', 'manufacturer', {
        required: true,
      }),
      textField('model', 'Model', 'Rack Type', { required: true }),
      {
        key: 'slug',
        label: 'Slug',
        type: 'slug',
        group: 'Rack Type',
        required: true,
        slugSource: 'model',
      },
      selectField('form_factor', 'Form Factor', 'Rack Type', rackFormFactors, { required: true }),
      selectField('width', 'Width', 'Rack Type', rackWidths, { default: 19 }),
      numberField('u_height', 'Height (U)', 'Rack Type', { default: 42, min: 1, max: 100 }),
      numberField('starting_unit', 'Starting Unit', 'Rack Type', {
        default: 1,
        min: 1,
        max: 32767,
      }),
      booleanField('desc_units', 'Descending Units', 'Rack Type', { default: false }),
      descriptionField('Rack Type'),
      commentsField('Rack Type'),
    ],
  ),
  coreModel(
    'dcim',
    'rack',
    'Rack',
    'Racks',
    'racks',
    [
      column('name', 'Name', true),
      column('site', 'Site', true),
      column('status', 'Status', true),
      column('rack_type', 'Rack Type'),
      column('role', 'Role'),
      column('u_height', 'Height (U)', true),
      column('device_count', 'Devices'),
    ],
    [
      relationFilter('site_id', 'Site', 'site'),
      textFilter('site_slug', 'Site Slug'),
      textFilter('name', 'Name'),
      selectFilter('status', 'Status', rackStatuses),
      relationFilter('role_id', 'Role', 'rackrole'),
      textFilter('role_slug', 'Role Slug'),
      relationFilter('rack_type_id', 'Rack Type', 'racktype'),
      textFilter('rack_type_slug', 'Rack Type Slug'),
    ],
    [
      relationField('site', 'Site', 'Rack', 'site', { required: true }),
      textField('name', 'Name', 'Rack', { required: true }),
      textField('facility_id', 'Facility ID', 'Rack'),
      relationField('rack_type', 'Rack Type', 'Rack', 'racktype'),
      selectField('status', 'Status', 'Rack', rackStatuses, {
        required: true,
        default: 'active',
      }),
      relationField('role', 'Role', 'Rack', 'rackrole'),
      textField('serial', 'Serial Number', 'Rack'),
      textField('asset_tag', 'Asset Tag', 'Rack'),
      selectField('form_factor', 'Form Factor', 'Rack', rackFormFactors, {
        disabledWhenFieldTruthy: 'rack_type',
        helpText: 'Inherited from the Rack Type when one is selected.',
      }),
      selectField('width', 'Width', 'Rack', rackWidths, {
        default: 19,
        disabledWhenFieldTruthy: 'rack_type',
        helpText: 'Inherited from the Rack Type when one is selected.',
      }),
      numberField('u_height', 'Height (U)', 'Rack', {
        default: 42,
        min: 1,
        max: 100,
        disabledWhenFieldTruthy: 'rack_type',
        helpText: 'Inherited from the Rack Type when one is selected.',
      }),
      numberField('starting_unit', 'Starting Unit', 'Rack', {
        default: 1,
        min: 1,
        disabledWhenFieldTruthy: 'rack_type',
        helpText: 'Inherited from the Rack Type when one is selected.',
      }),
      booleanField('desc_units', 'Descending Units', 'Rack', {
        default: false,
        disabledWhenFieldTruthy: 'rack_type',
        helpText: 'Inherited from the Rack Type when one is selected.',
      }),
      selectField('airflow', 'Airflow', 'Rack', rackAirflowChoices),
      descriptionField('Rack'),
      commentsField('Rack'),
    ],
    rackStatuses,
  ),
  coreModel(
    'dcim',
    'devicerole',
    'Device Role',
    'Device Roles',
    'device-roles',
    [
      column('name', 'Name', true),
      column('slug', 'Slug', true),
      column('parent', 'Parent'),
      column('color', 'Color'),
      column('vm_role', 'VM Role'),
      column('device_count', 'Devices'),
    ],
    [textFilter('name', 'Name'), textFilter('slug', 'Slug')],
    [
      relationField('parent', 'Parent Role', 'Device Role', 'devicerole'),
      textField('name', 'Name', 'Device Role', { required: true }),
      { key: 'slug', label: 'Slug', type: 'slug', group: 'Device Role', required: true },
      textField('color', 'Color', 'Device Role', { default: '9e9e9e' }),
      booleanField('vm_role', 'VM Role', 'Device Role', { default: true }),
      descriptionField('Device Role'),
      commentsField('Device Role'),
    ],
  ),
  coreModel(
    'dcim',
    'devicetype',
    'Device Type',
    'Device Types',
    'device-types',
    [
      column('model', 'Model', true),
      column('manufacturer', 'Manufacturer', true),
      column('slug', 'Slug', true),
      column('part_number', 'Part Number'),
      column('u_height', 'Height (U)', true),
      column('device_count', 'Devices'),
      column('interface_template_count', 'Interface Templates'),
    ],
    [
      relationFilter('manufacturer_id', 'Manufacturer', 'manufacturer'),
      textFilter('manufacturer_slug', 'Manufacturer Slug'),
      textFilter('model', 'Model'),
      textFilter('slug', 'Slug'),
    ],
    [
      relationField('manufacturer', 'Manufacturer', 'Device Type', 'manufacturer', {
        required: true,
      }),
      textField('model', 'Model', 'Device Type', { required: true }),
      {
        key: 'slug',
        label: 'Slug',
        type: 'slug',
        group: 'Device Type',
        required: true,
        slugSource: 'model',
      },
      textField('part_number', 'Part Number', 'Device Type'),
      numberField('u_height', 'Height (U)', 'Device Type', {
        default: 1,
        min: 0,
        max: 999.9,
        step: 0.5,
      }),
      booleanField('exclude_from_utilization', 'Exclude from Utilization', 'Device Type', {
        default: false,
      }),
      booleanField('is_full_depth', 'Full Depth', 'Device Type', { default: true }),
      selectField('airflow', 'Airflow', 'Device Type', airflowChoices),
      descriptionField('Device Type'),
      commentsField('Device Type'),
    ],
  ),
  coreModel(
    'dcim',
    'interfacetemplate',
    'Interface Template',
    'Interface Templates',
    'interface-templates',
    [
      column('device_type', 'Device Type', true),
      column('name', 'Name', true),
      column('label', 'Label'),
      column('type', 'Type', true),
      column('enabled', 'Enabled'),
      column('mgmt_only', 'Management Only'),
    ],
    [
      relationFilter('device_type_id', 'Device Type', 'devicetype'),
      textFilter('name', 'Name'),
      selectFilter('type', 'Type', INTERFACE_TYPE_CHOICES),
      booleanFilter('enabled', 'Enabled'),
      booleanFilter('mgmt_only', 'Management Only'),
    ],
    [
      relationField('device_type', 'Device Type', 'Interface Template', 'devicetype', {
        required: true,
        immutableOnEdit: true,
        helpText: 'An Interface Template cannot be moved to another Device Type.',
      }),
      textField('name', 'Name', 'Interface Template', { required: true }),
      textField('label', 'Label', 'Interface Template'),
      selectField('type', 'Type', 'Interface Template', INTERFACE_TYPE_CHOICES, {
        required: true,
      }),
      booleanField('enabled', 'Enabled', 'Interface Template', { default: true }),
      booleanField('mgmt_only', 'Management Only', 'Interface Template', { default: false }),
      descriptionField('Interface Template'),
    ],
  ),
  coreModel(
    'dcim',
    'device',
    'Device',
    'Devices',
    'devices',
    [
      column('name', 'Name', true),
      column('device_type', 'Device Type'),
      column('role', 'Role'),
      column('site', 'Site', true),
      column('rack', 'Rack', true),
      column('position', 'Position', true),
      column('status', 'Status', true),
      column('interface_count', 'Interfaces'),
    ],
    [
      relationFilter('site_id', 'Site', 'site'),
      textFilter('site_slug', 'Site Slug'),
      relationFilter('rack_id', 'Rack', 'rack'),
      relationFilter('device_type_id', 'Device Type', 'devicetype'),
      textFilter('device_type_slug', 'Device Type Slug'),
      relationFilter('role_id', 'Role', 'devicerole'),
      textFilter('role_slug', 'Role Slug'),
      textFilter('name', 'Name'),
      selectFilter('status', 'Status', deviceStatuses),
    ],
    [
      relationField('device_type', 'Device Type', 'Device', 'devicetype', {
        required: true,
      }),
      relationField('role', 'Device Role', 'Device', 'devicerole', { required: true }),
      textField('name', 'Name', 'Device'),
      relationField('site', 'Site', 'Location', 'site', { required: true }),
      relationField('rack', 'Rack', 'Location', 'rack', {
        relationFilterFields: { site_id: 'site' },
        clearWhenFieldChanges: ['site'],
        helpText: 'Choices are limited to Racks in the selected Site.',
      }),
      numberField('position', 'Position (U)', 'Location', {
        min: 1,
        max: 100.5,
        step: 0.5,
        disabledUnlessFieldTruthy: 'rack',
        clearWhenFieldChanges: ['rack'],
        helpText: 'A Rack and Face are required when a position is set.',
      }),
      selectField(
        'face',
        'Face',
        'Location',
        [
          { value: 'front', label: 'Front' },
          { value: 'rear', label: 'Rear' },
        ],
        {
          disabledUnlessFieldTruthy: 'rack',
          requiredWhenFieldTruthy: 'position',
          clearWhenFieldChanges: ['rack'],
        },
      ),
      selectField('status', 'Status', 'Device', deviceStatuses, {
        required: true,
        default: 'active',
      }),
      textField('serial', 'Serial Number', 'Device'),
      textField('asset_tag', 'Asset Tag', 'Device'),
      selectField('airflow', 'Airflow', 'Device', airflowChoices),
      descriptionField('Device'),
      commentsField('Device'),
    ],
    deviceStatuses,
  ),
  coreModel(
    'dcim',
    'interface',
    'Interface',
    'Interfaces',
    'interfaces',
    [
      column('device', 'Device', true),
      column('name', 'Name', true),
      column('label', 'Label'),
      column('type', 'Type', true),
      column('enabled', 'Enabled'),
      column('mgmt_only', 'Management Only'),
      column('mtu', 'MTU'),
      column('count_ipaddresses', 'IP Addresses'),
    ],
    [
      relationFilter('device_id', 'Device', 'device'),
      textFilter('device_name', 'Device Name'),
      textFilter('name', 'Name'),
      selectFilter('type', 'Type', INTERFACE_TYPE_CHOICES),
      booleanFilter('enabled', 'Enabled'),
      booleanFilter('mgmt_only', 'Management Only'),
    ],
    [
      relationField('device', 'Device', 'Interface', 'device', {
        required: true,
        immutableOnEdit: true,
        helpText: 'An Interface cannot be moved to another Device.',
      }),
      textField('name', 'Name', 'Interface', { required: true }),
      textField('label', 'Label', 'Interface'),
      selectField('type', 'Type', 'Interface', INTERFACE_TYPE_CHOICES, {
        required: true,
      }),
      booleanField('enabled', 'Enabled', 'Interface', { default: true }),
      booleanField('mgmt_only', 'Management Only', 'Interface', { default: false }),
      numberField('mtu', 'MTU', 'Interface', { min: 1, max: 65536 }),
      numberField('speed', 'Speed (Kbps)', 'Interface', { min: 0 }),
      selectField('duplex', 'Duplex', 'Interface', [
        { value: 'half', label: 'Half' },
        { value: 'full', label: 'Full' },
        { value: 'auto', label: 'Auto' },
      ]),
      descriptionField('Interface'),
    ],
  ),
  coreModel(
    'ipam',
    'vrf',
    'VRF',
    'VRFs',
    'vrfs',
    [
      column('name', 'Name', true),
      column('rd', 'Route Distinguisher', true),
      column('enforce_unique', 'Enforce Unique'),
      column('prefix_count', 'Prefixes'),
      column('ipaddress_count', 'IP Addresses'),
      column('description', 'Description'),
    ],
    [
      textFilter('name', 'Name'),
      textFilter('rd', 'Route Distinguisher'),
      booleanFilter('enforce_unique', 'Enforce Unique'),
    ],
    [
      textField('name', 'Name', 'VRF', { required: true }),
      textField('rd', 'Route Distinguisher', 'VRF', { placeholder: '65000:100' }),
      booleanField('enforce_unique', 'Enforce Unique Space', 'VRF', { default: true }),
      descriptionField('VRF'),
      commentsField('VRF'),
    ],
  ),
  coreModel(
    'ipam',
    'prefix',
    'Prefix',
    'Prefixes',
    'prefixes',
    [
      column('prefix', 'Prefix', true),
      column('vrf', 'VRF', true),
      column('family', 'Family'),
      column('status', 'Status', true),
      column('is_pool', 'Pool'),
      column('mark_utilized', 'Mark Utilized'),
      column('description', 'Description'),
    ],
    [
      relationFilter('vrf_id', 'VRF', 'vrf'),
      textFilter('vrf_rd', 'VRF Route Distinguisher'),
      textFilter('prefix', 'Prefix'),
      selectFilter('family', 'Family', [
        { value: 4, label: 'IPv4' },
        { value: 6, label: 'IPv6' },
      ]),
      selectFilter('status', 'Status', prefixStatuses),
      textFilter('within', 'Within'),
      textFilter('within_include', 'Within (including self)'),
      textFilter('contains', 'Contains'),
    ],
    [
      textField('prefix', 'Prefix', 'Prefix', {
        required: true,
        placeholder: '192.0.2.0/24',
      }),
      relationField('vrf', 'VRF', 'Prefix', 'vrf'),
      selectField('status', 'Status', 'Prefix', prefixStatuses, {
        required: true,
        default: 'active',
      }),
      booleanField('is_pool', 'Is a Pool', 'Prefix', { default: false }),
      booleanField('mark_utilized', 'Mark Fully Utilized', 'Prefix', { default: false }),
      descriptionField('Prefix'),
      commentsField('Prefix'),
    ],
    prefixStatuses,
  ),
  coreModel(
    'ipam',
    'ipaddress',
    'IP Address',
    'IP Addresses',
    'ip-addresses',
    [
      column('address', 'Address', true),
      column('vrf', 'VRF', true),
      column('family', 'Family'),
      column('status', 'Status', true),
      column('role', 'Role'),
      column('dns_name', 'DNS Name', true),
      column('assigned_object', 'Assigned Interface'),
    ],
    [
      relationFilter('vrf_id', 'VRF', 'vrf'),
      textFilter('vrf_rd', 'VRF Route Distinguisher'),
      textFilter('address', 'Address'),
      selectFilter('family', 'Family', [
        { value: 4, label: 'IPv4' },
        { value: 6, label: 'IPv6' },
      ]),
      textFilter('parent', 'Parent Prefix'),
      selectFilter('status', 'Status', ipAddressStatuses),
      booleanFilter('assigned', 'Assigned'),
      relationFilter('interface_id', 'Interface', 'interface'),
      relationFilter('device_id', 'Device', 'device'),
    ],
    [
      textField('address', 'Address', 'IP Address', {
        required: true,
        placeholder: '192.0.2.1/24',
      }),
      relationField('vrf', 'VRF', 'IP Address', 'vrf'),
      selectField('status', 'Status', 'IP Address', ipAddressStatuses, {
        required: true,
        default: 'active',
      }),
      selectField('role', 'Role', 'IP Address', ipAddressRoles),
      textField('dns_name', 'DNS Name', 'IP Address'),
      descriptionField('IP Address'),
      commentsField('IP Address'),
      relationField('assigned_interface', 'Assigned Interface', 'Assignment', 'interface', {
        sourceKey: 'assigned_object',
        helpText: 'Clear this field to unassign the address from its Interface.',
      }),
    ],
    ipAddressStatuses,
    [
      'address',
      'vrf',
      'status',
      'role',
      'dns_name',
      'description',
      'comments',
      'assigned_object_type',
      'assigned_object_id',
    ],
  ),
]
