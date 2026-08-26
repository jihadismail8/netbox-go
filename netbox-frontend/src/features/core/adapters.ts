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
  type DeviceRoleMutation,
  type DeviceTypeDTO,
  type DeviceTypeForm,
  type DeviceTypeMutation,
  type InterfaceDTO,
  type InterfaceForm,
  type InterfaceTemplateDTO,
  type InterfaceTemplateForm,
  type InterfaceTemplateMutation,
  type IPAddressDTO,
  type IPAddressForm,
  type IPAddressMutation,
  type ManufacturerDTO,
  type ManufacturerForm,
  type ManufacturerMutation,
  type PrefixDTO,
  type PrefixForm,
  type RackDTO,
  type RackForm,
  type RackMutation,
  type RackRoleDTO,
  type RackRoleForm,
  type RackRoleMutation,
  type RackTypeDTO,
  type RackTypeForm,
  type RackTypeMutation,
  type SiteDTO,
  type SiteForm,
  type SiteMutation,
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

const siteOriginalMutation = Symbol('site-original-mutation')

const siteScalarMutationFields = [
  'name',
  'slug',
  'status',
  'facility',
  'description',
  'comments',
] as const satisfies readonly (keyof SiteMutation)[]

type TrackedSiteForm = SiteForm & {
  [siteOriginalMutation]?: SiteMutation
}

function siteMutation(form: SiteForm): SiteMutation {
  const mutation: SiteMutation = {}
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'status')) mutation.status = form.status
  if (hasFormField(form, 'facility')) mutation.facility = form.facility
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  return mutation
}

function siteMutationDelta(current: SiteMutation, original: SiteMutation): SiteMutation {
  const delta = { ...current }
  for (const field of siteScalarMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const siteAdapter: CoreResourceAdapter<'site'> = {
  resource: 'site',
  emptyForm: () => ({ status: 'active' }),
  formFromDTO: (dto: SiteDTO): SiteForm => {
    const form: TrackedSiteForm = {
      name: dto.name,
      slug: dto.slug,
      status: choiceValue(dto.status),
      facility: dto.facility,
      description: dto.description,
      comments: dto.comments,
    }
    form[siteOriginalMutation] = siteMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = siteMutation(form)
    const original = (form as TrackedSiteForm)[siteOriginalMutation]
    return editing && original ? siteMutationDelta(mutation, original) : mutation
  },
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

const manufacturerOriginalMutation = Symbol('manufacturer-original-mutation')

const manufacturerScalarMutationFields = [
  'name',
  'slug',
  'description',
] as const satisfies readonly (keyof ManufacturerMutation)[]

type TrackedManufacturerForm = ManufacturerForm & {
  [manufacturerOriginalMutation]?: ManufacturerMutation
}

function manufacturerMutation(form: ManufacturerForm): ManufacturerMutation {
  const mutation: ManufacturerMutation = {}
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'description')) mutation.description = form.description
  return mutation
}

function manufacturerMutationDelta(
  current: ManufacturerMutation,
  original: ManufacturerMutation,
): ManufacturerMutation {
  const delta = { ...current }
  for (const field of manufacturerScalarMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const manufacturerAdapter: CoreResourceAdapter<'manufacturer'> = {
  resource: 'manufacturer',
  emptyForm: () => ({}),
  formFromDTO: (dto: ManufacturerDTO): ManufacturerForm => {
    const form: TrackedManufacturerForm = {
      name: dto.name,
      slug: dto.slug,
      description: dto.description,
    }
    form[manufacturerOriginalMutation] = manufacturerMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = manufacturerMutation(form)
    const original = (form as TrackedManufacturerForm)[manufacturerOriginalMutation]
    return editing && original ? manufacturerMutationDelta(mutation, original) : mutation
  },
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const rackRoleOriginalMutation = Symbol('rack-role-original-mutation')

const rackRoleScalarMutationFields = [
  'name',
  'slug',
  'color',
  'description',
] as const satisfies readonly (keyof RackRoleMutation)[]

type TrackedRackRoleForm = RackRoleForm & {
  [rackRoleOriginalMutation]?: RackRoleMutation
}

function rackRoleMutation(form: RackRoleForm): RackRoleMutation {
  const mutation: RackRoleMutation = {}
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'color')) mutation.color = form.color
  if (hasFormField(form, 'description')) mutation.description = form.description
  return mutation
}

function rackRoleMutationDelta(
  current: RackRoleMutation,
  original: RackRoleMutation,
): RackRoleMutation {
  const delta = { ...current }
  for (const field of rackRoleScalarMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const rackRoleAdapter: CoreResourceAdapter<'rackrole'> = {
  resource: 'rackrole',
  emptyForm: () => ({ color: '9e9e9e' }),
  formFromDTO: (dto: RackRoleDTO): RackRoleForm => {
    const form: TrackedRackRoleForm = {
      name: dto.name,
      slug: dto.slug,
      color: dto.color,
      description: dto.description,
    }
    form[rackRoleOriginalMutation] = rackRoleMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = rackRoleMutation(form)
    const original = (form as TrackedRackRoleForm)[rackRoleOriginalMutation]
    return editing && original ? rackRoleMutationDelta(mutation, original) : mutation
  },
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const rackTypeOriginalMutation = Symbol('rack-type-original-mutation')

const rackTypeMutationFields = [
  'manufacturer',
  'model',
  'slug',
  'form_factor',
  'width',
  'u_height',
  'starting_unit',
  'desc_units',
  'description',
  'comments',
] as const satisfies readonly (keyof RackTypeMutation)[]

type TrackedRackTypeForm = RackTypeForm & {
  [rackTypeOriginalMutation]?: RackTypeMutation
}

function rackTypeMutation(form: RackTypeForm): RackTypeMutation {
  const mutation: RackTypeMutation = {}
  if (hasFormField(form, 'manufacturer')) mutation.manufacturer = relationID(form.manufacturer)
  if (hasFormField(form, 'model')) mutation.model = form.model
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'form_factor')) mutation.form_factor = form.form_factor
  if (hasFormField(form, 'width')) mutation.width = form.width
  if (hasFormField(form, 'u_height')) mutation.u_height = form.u_height
  if (hasFormField(form, 'starting_unit')) mutation.starting_unit = form.starting_unit
  if (hasFormField(form, 'desc_units')) mutation.desc_units = form.desc_units
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  return mutation
}

function rackTypeMutationDelta(
  current: RackTypeMutation,
  original: RackTypeMutation,
): RackTypeMutation {
  const delta = { ...current }
  for (const field of rackTypeMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const rackTypeAdapter: CoreResourceAdapter<'racktype'> = {
  resource: 'racktype',
  emptyForm: () => ({ width: 19, u_height: 42, starting_unit: 1, desc_units: false }),
  formFromDTO: (dto: RackTypeDTO): RackTypeForm => {
    const form: TrackedRackTypeForm = {
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
    }
    form[rackTypeOriginalMutation] = rackTypeMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = rackTypeMutation(form)
    const original = (form as TrackedRackTypeForm)[rackTypeOriginalMutation]
    return editing && original ? rackTypeMutationDelta(mutation, original) : mutation
  },
  filtersFromState: (state) => ({
    manufacturer_id: state.manufacturer_id,
    manufacturer_slug: state.manufacturer_slug,
    model: state.model,
    slug: state.slug,
  }),
}

const rackOriginalMutation = Symbol('rack-original-mutation')

const rackMutationFields = [
  'site',
  'name',
  'facility_id',
  'rack_type',
  'status',
  'role',
  'serial',
  'asset_tag',
  'form_factor',
  'width',
  'u_height',
  'starting_unit',
  'desc_units',
  'airflow',
  'description',
  'comments',
] as const satisfies readonly (keyof RackMutation)[]

const rackTypePhysicalFields = [
  'form_factor',
  'width',
  'u_height',
  'starting_unit',
  'desc_units',
] as const satisfies readonly (keyof RackMutation)[]

type TrackedRackForm = RackForm & {
  [rackOriginalMutation]?: RackMutation
}

function rackMutation(form: RackForm): RackMutation {
  const mutation: RackMutation = {}
  if (hasFormField(form, 'site')) mutation.site = relationID(form.site)
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'facility_id')) mutation.facility_id = form.facility_id
  if (hasFormField(form, 'rack_type')) mutation.rack_type = relationID(form.rack_type)
  if (hasFormField(form, 'status')) mutation.status = form.status
  if (hasFormField(form, 'role')) mutation.role = relationID(form.role)
  if (hasFormField(form, 'serial')) mutation.serial = form.serial
  if (hasFormField(form, 'asset_tag')) mutation.asset_tag = form.asset_tag
  if (hasFormField(form, 'form_factor')) mutation.form_factor = form.form_factor
  if (hasFormField(form, 'width')) mutation.width = form.width
  if (hasFormField(form, 'u_height')) mutation.u_height = form.u_height
  if (hasFormField(form, 'starting_unit')) mutation.starting_unit = form.starting_unit
  if (hasFormField(form, 'desc_units')) mutation.desc_units = form.desc_units
  if (hasFormField(form, 'airflow')) mutation.airflow = form.airflow
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  return mutation
}

function rackMutationDelta(current: RackMutation, original: RackMutation): RackMutation {
  const delta = { ...current }
  for (const field of rackMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

function suppressRackTypePhysicalFields(
  mutation: RackMutation,
  rackType: number | null | undefined,
): RackMutation {
  if (rackType === null || rackType === undefined) return mutation
  const result = { ...mutation }
  for (const field of rackTypePhysicalFields) Reflect.deleteProperty(result, field)
  return result
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
  formFromDTO: (dto: RackDTO): RackForm => {
    const form: TrackedRackForm = {
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
    }
    form[rackOriginalMutation] = rackMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = rackMutation(form)
    const original = (form as TrackedRackForm)[rackOriginalMutation]
    const outgoing = editing && original ? rackMutationDelta(mutation, original) : mutation
    return suppressRackTypePhysicalFields(outgoing, relationID(form.rack_type))
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

const deviceRoleOriginalMutation = Symbol('device-role-original-mutation')

const deviceRoleMutationFields = [
  'parent',
  'name',
  'slug',
  'color',
  'vm_role',
  'description',
  'comments',
] as const satisfies readonly (keyof DeviceRoleMutation)[]

type TrackedDeviceRoleForm = DeviceRoleForm & {
  [deviceRoleOriginalMutation]?: DeviceRoleMutation
}

function deviceRoleMutation(form: DeviceRoleForm): DeviceRoleMutation {
  const mutation: DeviceRoleMutation = {}
  if (hasFormField(form, 'parent')) mutation.parent = relationID(form.parent)
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'color')) mutation.color = form.color
  if (hasFormField(form, 'vm_role')) mutation.vm_role = form.vm_role
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  return mutation
}

function deviceRoleMutationDelta(
  current: DeviceRoleMutation,
  original: DeviceRoleMutation,
): DeviceRoleMutation {
  const delta = { ...current }
  for (const field of deviceRoleMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const deviceRoleAdapter: CoreResourceAdapter<'devicerole'> = {
  resource: 'devicerole',
  emptyForm: () => ({ color: '9e9e9e', vm_role: true }),
  formFromDTO: (dto: DeviceRoleDTO): DeviceRoleForm => {
    const form: TrackedDeviceRoleForm = {
      parent: dto.parent,
      name: dto.name,
      slug: dto.slug,
      color: dto.color,
      vm_role: dto.vm_role,
      description: dto.description,
      comments: dto.comments,
    }
    form[deviceRoleOriginalMutation] = deviceRoleMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = deviceRoleMutation(form)
    const original = (form as TrackedDeviceRoleForm)[deviceRoleOriginalMutation]
    return editing && original ? deviceRoleMutationDelta(mutation, original) : mutation
  },
  filtersFromState: (state) => ({ name: state.name, slug: state.slug }),
}

const deviceTypeOriginalMutation = Symbol('device-type-original-mutation')

const deviceTypeMutationFields = [
  'manufacturer',
  'model',
  'slug',
  'part_number',
  'u_height',
  'exclude_from_utilization',
  'is_full_depth',
  'airflow',
  'description',
  'comments',
] as const satisfies readonly (keyof DeviceTypeMutation)[]

type TrackedDeviceTypeForm = DeviceTypeForm & {
  [deviceTypeOriginalMutation]?: DeviceTypeMutation
}

function deviceTypeMutation(form: DeviceTypeForm): DeviceTypeMutation {
  const mutation: DeviceTypeMutation = {}
  if (hasFormField(form, 'manufacturer')) mutation.manufacturer = relationID(form.manufacturer)
  if (hasFormField(form, 'model')) mutation.model = form.model
  if (hasFormField(form, 'slug')) mutation.slug = form.slug
  if (hasFormField(form, 'part_number')) mutation.part_number = form.part_number
  if (hasFormField(form, 'u_height')) mutation.u_height = form.u_height
  if (hasFormField(form, 'exclude_from_utilization')) {
    mutation.exclude_from_utilization = form.exclude_from_utilization
  }
  if (hasFormField(form, 'is_full_depth')) mutation.is_full_depth = form.is_full_depth
  if (hasFormField(form, 'airflow')) mutation.airflow = form.airflow
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  return mutation
}

function deviceTypeMutationDelta(
  current: DeviceTypeMutation,
  original: DeviceTypeMutation,
): DeviceTypeMutation {
  const delta = { ...current }
  for (const field of deviceTypeMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const deviceTypeAdapter: CoreResourceAdapter<'devicetype'> = {
  resource: 'devicetype',
  emptyForm: () => ({ u_height: 1, exclude_from_utilization: false, is_full_depth: true }),
  formFromDTO: (dto: DeviceTypeDTO): DeviceTypeForm => {
    const form: TrackedDeviceTypeForm = {
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
    }
    form[deviceTypeOriginalMutation] = deviceTypeMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = deviceTypeMutation(form)
    const original = (form as TrackedDeviceTypeForm)[deviceTypeOriginalMutation]
    return editing && original ? deviceTypeMutationDelta(mutation, original) : mutation
  },
  filtersFromState: (state) => ({
    manufacturer_id: state.manufacturer_id,
    manufacturer_slug: state.manufacturer_slug,
    model: state.model,
    slug: state.slug,
  }),
}

const interfaceTemplateOriginalMutation = Symbol('interface-template-original-mutation')

const interfaceTemplateMutationFields = [
  'device_type',
  'name',
  'label',
  'type',
  'enabled',
  'mgmt_only',
  'description',
] as const satisfies readonly (keyof InterfaceTemplateMutation)[]

type TrackedInterfaceTemplateForm = InterfaceTemplateForm & {
  [interfaceTemplateOriginalMutation]?: InterfaceTemplateMutation
}

function interfaceTemplateMutation(form: InterfaceTemplateForm): InterfaceTemplateMutation {
  const mutation: InterfaceTemplateMutation = {}
  if (hasFormField(form, 'device_type')) {
    mutation.device_type = relationID(form.device_type)
  }
  if (hasFormField(form, 'name')) mutation.name = form.name
  if (hasFormField(form, 'label')) mutation.label = form.label
  if (hasFormField(form, 'type')) mutation.type = form.type
  if (hasFormField(form, 'enabled')) mutation.enabled = form.enabled
  if (hasFormField(form, 'mgmt_only')) mutation.mgmt_only = form.mgmt_only
  if (hasFormField(form, 'description')) mutation.description = form.description
  return mutation
}

function interfaceTemplateMutationDelta(
  current: InterfaceTemplateMutation,
  original: InterfaceTemplateMutation,
): InterfaceTemplateMutation {
  const delta = { ...current }
  for (const field of interfaceTemplateMutationFields) {
    if (Object.is(current[field], original[field])) Reflect.deleteProperty(delta, field)
  }
  return delta
}

const interfaceTemplateAdapter: CoreResourceAdapter<'interfacetemplate'> = {
  resource: 'interfacetemplate',
  emptyForm: () => ({ enabled: true, mgmt_only: false }),
  formFromDTO: (dto: InterfaceTemplateDTO): InterfaceTemplateForm => {
    const form: TrackedInterfaceTemplateForm = {
      device_type: dto.device_type,
      name: dto.name,
      label: dto.label,
      type: choiceValue(dto.type),
      enabled: dto.enabled,
      mgmt_only: dto.mgmt_only,
      description: dto.description,
    }
    form[interfaceTemplateOriginalMutation] = interfaceTemplateMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = interfaceTemplateMutation(form)
    const original = (form as TrackedInterfaceTemplateForm)[interfaceTemplateOriginalMutation]
    return editing && original ? interfaceTemplateMutationDelta(mutation, original) : mutation
  },
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

const ipAddressOriginalMutation = Symbol('ip-address-original-mutation')

const ipAddressScalarMutationFields = [
  'address',
  'status',
  'role',
  'dns_name',
  'description',
  'comments',
] as const satisfies readonly (keyof IPAddressMutation)[]

type TrackedIPAddressForm = IPAddressForm & {
  [ipAddressOriginalMutation]?: IPAddressMutation
}

function ipAddressMutation(form: IPAddressForm): IPAddressMutation {
  const mutation: IPAddressMutation = {}
  if (hasFormField(form, 'address')) mutation.address = form.address
  if (hasFormField(form, 'vrf')) mutation.vrf = relationID(form.vrf)
  if (hasFormField(form, 'status')) mutation.status = form.status
  if (hasFormField(form, 'role')) mutation.role = form.role
  if (hasFormField(form, 'dns_name')) mutation.dns_name = form.dns_name
  if (hasFormField(form, 'description')) mutation.description = form.description
  if (hasFormField(form, 'comments')) mutation.comments = form.comments
  if (!hasFormField(form, 'assigned_interface')) return mutation

  const assignedID = relationID(form.assigned_interface)
  if (
    assignedID !== null &&
    assignedID !== undefined &&
    (!Number.isInteger(assignedID) || assignedID <= 0)
  ) {
    throw new Error('Select a valid Interface.')
  }
  mutation.assigned_object_type = assignedID ? 'dcim.interface' : null
  mutation.assigned_object_id = assignedID ?? null
  return mutation
}

function ipAddressMutationDelta(
  current: IPAddressMutation,
  original: IPAddressMutation,
): IPAddressMutation {
  const delta = { ...current }
  for (const key of ipAddressScalarMutationFields) {
    if (Object.is(current[key], original[key])) Reflect.deleteProperty(delta, key)
  }
  return delta
}

const ipAddressAdapter: CoreResourceAdapter<'ipaddress'> = {
  resource: 'ipaddress',
  emptyForm: () => ({ status: 'active' }),
  formFromDTO: (dto: IPAddressDTO): IPAddressForm => {
    const form: TrackedIPAddressForm = {
      address: dto.address,
      vrf: dto.vrf,
      status: choiceValue(dto.status),
      role: choiceValue(dto.role),
      dns_name: dto.dns_name,
      description: dto.description,
      comments: dto.comments,
      assigned_interface: dto.assigned_object,
    }
    form[ipAddressOriginalMutation] = ipAddressMutation(form)
    return form
  },
  mutationFromForm: (form, editing) => {
    const mutation = ipAddressMutation(form)
    const original = (form as TrackedIPAddressForm)[ipAddressOriginalMutation]
    return editing && original ? ipAddressMutationDelta(mutation, original) : mutation
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
