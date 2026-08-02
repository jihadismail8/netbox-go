import type { CoreProfileResourceName } from './manifest'
import {
  choiceValue,
  hasFormField,
  relationID,
  type CoreFilterState,
  type CoreResourceDTO,
  type CoreResourceFilters,
  type CoreResourceForm,
  type CoreResourceMutation,
  type DeviceDTO,
  type DeviceForm,
  type DeviceRoleDTO,
  type DeviceRoleForm,
  type DeviceTypeDTO,
  type DeviceTypeForm,
  type InterfaceDTO,
  type InterfaceForm,
  type InterfaceTemplateDTO,
  type InterfaceTemplateForm,
  type IPAddressDTO,
  type IPAddressForm,
  type ManufacturerDTO,
  type ManufacturerForm,
  type PrefixDTO,
  type PrefixForm,
  type RackDTO,
  type RackForm,
  type RackRoleDTO,
  type RackRoleForm,
  type RackTypeDTO,
  type RackTypeForm,
  type SiteDTO,
  type SiteForm,
  type VRFDTO,
  type VRFForm,
} from './resources'

export interface CoreResourceAdapter<N extends CoreProfileResourceName> {
  readonly resource: N
  emptyForm(): CoreResourceForm<N>
  formFromDTO(dto: CoreResourceDTO<N>): CoreResourceForm<N>
  mutationFromForm(form: CoreResourceForm<N>, editing: boolean): CoreResourceMutation<N>
  filtersFromState(state: CoreFilterState): CoreResourceFilters<N>
}

const siteAdapter: CoreResourceAdapter<'site'> = {
  resource: 'site',
  emptyForm: () => ({ status: 'active' }),
  formFromDTO: (dto: SiteDTO): SiteForm => ({
    name: dto.name,
    slug: dto.slug,
    status: choiceValue(dto.status),
    facility: dto.facility,
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({ ...form }),
  filtersFromState: (state) => ({
    name: state.name,
    slug: state.slug,
    status:
      state.status === 'planned' ||
      state.status === 'staging' ||
      state.status === 'active' ||
      state.status === 'decommissioning' ||
      state.status === 'retired'
        ? state.status
        : undefined,
  }),
}

const manufacturerAdapter: CoreResourceAdapter<'manufacturer'> = {
  resource: 'manufacturer',
  emptyForm: () => ({}),
  formFromDTO: (dto: ManufacturerDTO): ManufacturerForm => ({
    name: dto.name,
    slug: dto.slug,
    description: dto.description,
  }),
  mutationFromForm: (form) => ({ ...form }),
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const rackRoleAdapter: CoreResourceAdapter<'rackrole'> = {
  resource: 'rackrole',
  emptyForm: () => ({ color: '9e9e9e' }),
  formFromDTO: (dto: RackRoleDTO): RackRoleForm => ({
    name: dto.name,
    slug: dto.slug,
    color: dto.color,
    description: dto.description,
  }),
  mutationFromForm: (form) => ({ ...form }),
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const rackTypeAdapter: CoreResourceAdapter<'racktype'> = {
  resource: 'racktype',
  emptyForm: () => ({ width: 19, u_height: 42, starting_unit: 1, desc_units: false }),
  formFromDTO: (dto: RackTypeDTO): RackTypeForm => ({
    manufacturer: dto.manufacturer,
    model: dto.model,
    slug: dto.slug,
    form_factor: choiceValue(dto.form_factor),
    width: choiceValue(dto.width),
    u_height: dto.u_height,
    starting_unit: dto.starting_unit,
    desc_units: dto.desc_units,
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({
    manufacturer: relationID(form.manufacturer),
    model: form.model,
    slug: form.slug,
    form_factor: form.form_factor,
    width: form.width,
    u_height: form.u_height,
    starting_unit: form.starting_unit,
    desc_units: form.desc_units,
    description: form.description,
    comments: form.comments,
  }),
  filtersFromState: (state) => ({
    manufacturer_id: state.manufacturer_id,
    manufacturer_slug: state.manufacturer_slug,
    model: state.model,
    slug: state.slug,
  }),
}

const rackAdapter: CoreResourceAdapter<'rack'> = {
  resource: 'rack',
  emptyForm: () => ({
    status: 'active',
    width: 19,
    u_height: 42,
    starting_unit: 1,
    desc_units: false,
  }),
  formFromDTO: (dto: RackDTO): RackForm => ({
    site: dto.site,
    name: dto.name,
    facility_id: dto.facility_id,
    rack_type: dto.rack_type,
    status: choiceValue(dto.status),
    role: dto.role,
    serial: dto.serial,
    asset_tag: dto.asset_tag,
    form_factor: choiceValue(dto.form_factor),
    width: choiceValue(dto.width),
    u_height: dto.u_height,
    starting_unit: dto.starting_unit,
    desc_units: dto.desc_units,
    airflow: choiceValue(dto.airflow),
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => {
    const rackType = relationID(form.rack_type)
    return {
      site: relationID(form.site),
      name: form.name,
      facility_id: form.facility_id,
      rack_type: rackType,
      status: form.status,
      role: relationID(form.role),
      serial: form.serial,
      asset_tag: form.asset_tag,
      ...(rackType
        ? {}
        : {
            form_factor: form.form_factor,
            width: form.width,
            u_height: form.u_height,
            starting_unit: form.starting_unit,
            desc_units: form.desc_units,
          }),
      airflow: form.airflow,
      description: form.description,
      comments: form.comments,
    }
  },
  filtersFromState: (state) => ({
    site_id: state.site_id,
    site_slug: state.site_slug,
    name: state.name,
    status:
      state.status === 'reserved' ||
      state.status === 'available' ||
      state.status === 'planned' ||
      state.status === 'active' ||
      state.status === 'deprecated'
        ? state.status
        : undefined,
    role_id: state.role_id,
    role_slug: state.role_slug,
    rack_type_id: state.rack_type_id,
    rack_type_slug: state.rack_type_slug,
  }),
}

const deviceRoleAdapter: CoreResourceAdapter<'devicerole'> = {
  resource: 'devicerole',
  emptyForm: () => ({ color: '9e9e9e', vm_role: true }),
  formFromDTO: (dto: DeviceRoleDTO): DeviceRoleForm => ({
    parent: dto.parent,
    name: dto.name,
    slug: dto.slug,
    color: dto.color,
    vm_role: dto.vm_role,
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({ ...form, parent: relationID(form.parent) }),
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const deviceTypeAdapter: CoreResourceAdapter<'devicetype'> = {
  resource: 'devicetype',
  emptyForm: () => ({ u_height: 1, exclude_from_utilization: false, is_full_depth: true }),
  formFromDTO: (dto: DeviceTypeDTO): DeviceTypeForm => ({
    manufacturer: dto.manufacturer,
    model: dto.model,
    slug: dto.slug,
    part_number: dto.part_number,
    u_height: dto.u_height,
    exclude_from_utilization: dto.exclude_from_utilization,
    is_full_depth: dto.is_full_depth,
    airflow: choiceValue(dto.airflow),
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({ ...form, manufacturer: relationID(form.manufacturer) }),
  filtersFromState: (state) => ({
    manufacturer_id: state.manufacturer_id,
    manufacturer_slug: state.manufacturer_slug,
    model: state.model,
    slug: state.slug,
  }),
}

const interfaceTemplateAdapter: CoreResourceAdapter<'interfacetemplate'> = {
  resource: 'interfacetemplate',
  emptyForm: () => ({ enabled: true, mgmt_only: false }),
  formFromDTO: (dto: InterfaceTemplateDTO): InterfaceTemplateForm => ({
    device_type: dto.device_type,
    name: dto.name,
    label: dto.label,
    type: choiceValue(dto.type),
    enabled: dto.enabled,
    mgmt_only: dto.mgmt_only,
    description: dto.description,
  }),
  mutationFromForm: (form, editing) => ({
    ...(!editing ? { device_type: relationID(form.device_type) } : {}),
    name: form.name,
    label: form.label,
    type: form.type,
    enabled: form.enabled,
    mgmt_only: form.mgmt_only,
    description: form.description,
  }),
  filtersFromState: (state) => ({
    device_type_id: state.device_type_id,
    name: state.name,
    type: state.type,
    enabled: state.enabled,
    mgmt_only: state.mgmt_only,
  }),
}

const deviceAdapter: CoreResourceAdapter<'device'> = {
  resource: 'device',
  emptyForm: () => ({ status: 'active' }),
  formFromDTO: (dto: DeviceDTO): DeviceForm => ({
    device_type: dto.device_type,
    role: dto.role,
    name: dto.name,
    site: dto.site,
    rack: dto.rack,
    position: dto.position,
    face: choiceValue(dto.face),
    status: choiceValue(dto.status),
    serial: dto.serial,
    asset_tag: dto.asset_tag,
    airflow: choiceValue(dto.airflow),
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => {
    const rack = relationID(form.rack)
    return {
      device_type: relationID(form.device_type),
      role: relationID(form.role),
      name: form.name,
      site: relationID(form.site),
      rack,
      ...(rack ? { position: form.position, face: form.face } : {}),
      status: form.status,
      serial: form.serial,
      asset_tag: form.asset_tag,
      airflow: form.airflow,
      description: form.description,
      comments: form.comments,
    }
  },
  filtersFromState: (state) => ({
    site_id: state.site_id,
    site_slug: state.site_slug,
    rack_id: state.rack_id,
    device_type_id: state.device_type_id,
    device_type_slug: state.device_type_slug,
    role_id: state.role_id,
    role_slug: state.role_slug,
    name: state.name,
    status:
      state.status === 'offline' ||
      state.status === 'active' ||
      state.status === 'planned' ||
      state.status === 'staged' ||
      state.status === 'failed' ||
      state.status === 'inventory' ||
      state.status === 'decommissioning'
        ? state.status
        : undefined,
  }),
}

const interfaceAdapter: CoreResourceAdapter<'interface'> = {
  resource: 'interface',
  emptyForm: () => ({ enabled: true, mgmt_only: false }),
  formFromDTO: (dto: InterfaceDTO): InterfaceForm => ({
    device: dto.device,
    name: dto.name,
    label: dto.label,
    type: choiceValue(dto.type),
    enabled: dto.enabled,
    mgmt_only: dto.mgmt_only,
    mtu: dto.mtu,
    speed: dto.speed,
    duplex: choiceValue(dto.duplex),
    description: dto.description,
  }),
  mutationFromForm: (form, editing) => ({
    ...(!editing ? { device: relationID(form.device) } : {}),
    name: form.name,
    label: form.label,
    type: form.type,
    enabled: form.enabled,
    mgmt_only: form.mgmt_only,
    mtu: form.mtu,
    speed: form.speed,
    duplex: form.duplex,
    description: form.description,
  }),
  filtersFromState: (state) => ({
    device_id: state.device_id,
    device_name: state.device_name,
    name: state.name,
    type: state.type,
    enabled: state.enabled,
    mgmt_only: state.mgmt_only,
  }),
}

const vrfAdapter: CoreResourceAdapter<'vrf'> = {
  resource: 'vrf',
  emptyForm: () => ({ enforce_unique: true }),
  formFromDTO: (dto: VRFDTO): VRFForm => ({
    name: dto.name,
    rd: dto.rd,
    enforce_unique: dto.enforce_unique,
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({ ...form }),
  filtersFromState: (state) => ({
    name: state.name,
    rd: state.rd,
    enforce_unique: state.enforce_unique,
  }),
}

const prefixAdapter: CoreResourceAdapter<'prefix'> = {
  resource: 'prefix',
  emptyForm: () => ({ status: 'active', is_pool: false, mark_utilized: false }),
  formFromDTO: (dto: PrefixDTO): PrefixForm => ({
    prefix: dto.prefix,
    vrf: dto.vrf,
    status: choiceValue(dto.status),
    is_pool: dto.is_pool,
    mark_utilized: dto.mark_utilized,
    description: dto.description,
    comments: dto.comments,
  }),
  mutationFromForm: (form) => ({ ...form, vrf: relationID(form.vrf) }),
  filtersFromState: (state) => ({
    vrf_id: state.vrf_id,
    vrf_rd: state.vrf_rd,
    prefix: state.prefix,
    family: state.family,
    status:
      state.status === 'container' ||
      state.status === 'active' ||
      state.status === 'reserved' ||
      state.status === 'deprecated'
        ? state.status
        : undefined,
    within: state.within,
    within_include: state.within_include,
    contains: state.contains,
  }),
}

const ipAddressAdapter: CoreResourceAdapter<'ipaddress'> = {
  resource: 'ipaddress',
  emptyForm: () => ({ status: 'active' }),
  formFromDTO: (dto: IPAddressDTO): IPAddressForm => ({
    address: dto.address,
    vrf: dto.vrf,
    status: choiceValue(dto.status),
    role: choiceValue(dto.role),
    dns_name: dto.dns_name,
    description: dto.description,
    comments: dto.comments,
    assigned_interface: dto.assigned_object,
  }),
  mutationFromForm: (form) => {
    const mutation = {
      address: form.address,
      vrf: relationID(form.vrf),
      status: form.status,
      role: form.role,
      dns_name: form.dns_name,
      description: form.description,
      comments: form.comments,
    }
    if (!hasFormField(form, 'assigned_interface')) return mutation
    const assignedID = relationID(form.assigned_interface)
    if (
      assignedID !== null &&
      assignedID !== undefined &&
      (!Number.isInteger(assignedID) || assignedID <= 0)
    ) {
      throw new Error('Select a valid Interface.')
    }
    return {
      ...mutation,
      assigned_object_type: assignedID ? 'dcim.interface' : null,
      assigned_object_id: assignedID ?? null,
    }
  },
  filtersFromState: (state) => ({
    vrf_id: state.vrf_id,
    vrf_rd: state.vrf_rd,
    address: state.address,
    family: state.family,
    parent: state.parent,
    status:
      state.status === 'active' ||
      state.status === 'reserved' ||
      state.status === 'deprecated' ||
      state.status === 'dhcp' ||
      state.status === 'slaac'
        ? state.status
        : undefined,
    assigned: state.assigned,
    interface_id: state.interface_id,
    device_id: state.device_id,
  }),
}

const adapters = {
  site: siteAdapter,
  manufacturer: manufacturerAdapter,
  rackrole: rackRoleAdapter,
  racktype: rackTypeAdapter,
  rack: rackAdapter,
  devicerole: deviceRoleAdapter,
  devicetype: deviceTypeAdapter,
  interfacetemplate: interfaceTemplateAdapter,
  device: deviceAdapter,
  interface: interfaceAdapter,
  vrf: vrfAdapter,
  prefix: prefixAdapter,
  ipaddress: ipAddressAdapter,
} satisfies { [N in CoreProfileResourceName]: CoreResourceAdapter<N> }

/** Resolve the one typed adapter owned by each promoted resource. */
export function getCoreResourceAdapter<N extends CoreProfileResourceName>(
  resource: N,
): CoreResourceAdapter<N> {
  // TypeScript cannot preserve the key/value correlation of an indexed mapped object.
  const adapter = adapters[resource] as CoreResourceAdapter<N>
  return {
    resource,
    emptyForm: adapter.emptyForm,
    formFromDTO: adapter.formFromDTO,
    mutationFromForm: (form, editing) => withoutUndefined(adapter.mutationFromForm(form, editing)),
    filtersFromState: (state) => withoutUndefined(adapter.filtersFromState(state)),
  }
}

function withoutUndefined<T extends object>(value: T): T {
  const result = { ...value }
  for (const [key, item] of Object.entries(result)) {
    if (item === undefined) Reflect.deleteProperty(result, key)
  }
  return result
}
