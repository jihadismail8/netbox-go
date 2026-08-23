import type { PaginatedResponse } from '@/types'
import type { CoreProfileResourceName } from './manifest'

/** The compact relationship envelope emitted by the pinned NetBox REST profile. */
export interface CoreReference {
  id: number
  url: string
  display: string
}

/** NetBox's read-side representation for a model choice. Writes remain scalar. */
export interface ChoiceDTO<T extends string | number> {
  value: T
  label: string
}

export interface CoreResourceBase {
  id: number
  url: string
  display: string
  created: string
  last_updated: string
}

export type InfrastructureStatus = 'planned' | 'staging' | 'active' | 'decommissioning' | 'retired'
export type RackStatus = 'reserved' | 'available' | 'planned' | 'active' | 'deprecated'
export type DeviceStatus =
  'offline' | 'active' | 'planned' | 'staged' | 'failed' | 'inventory' | 'decommissioning'
export type PrefixStatus = 'container' | 'active' | 'reserved' | 'deprecated'
export type IPAddressStatus = 'active' | 'reserved' | 'deprecated' | 'dhcp' | 'slaac'
export type IPAddressRole =
  'loopback' | 'secondary' | 'anycast' | 'vip' | 'vrrp' | 'hsrp' | 'glbp' | 'carp'
export type RackFormFactor =
  | '2-post-frame'
  | '4-post-frame'
  | '4-post-cabinet'
  | 'wall-frame'
  | 'wall-frame-vertical'
  | 'wall-cabinet'
  | 'wall-cabinet-vertical'
export type Airflow =
  | 'front-to-rear'
  | 'rear-to-front'
  | 'left-to-right'
  | 'right-to-left'
  | 'side-to-rear'
  | 'rear-to-side'
  | 'bottom-to-top'
  | 'top-to-bottom'
  | 'passive'
  | 'mixed'
export type RackAirflow = 'front-to-rear' | 'rear-to-front'
export type InterfaceDuplex = 'half' | 'full' | 'auto'

export interface SiteDTO extends CoreResourceBase {
  name: string
  slug: string
  status: ChoiceDTO<InfrastructureStatus>
  facility: string
  description: string
  comments: string
  device_count: number
  prefix_count: number
  rack_count: number
}

export interface ManufacturerDTO extends CoreResourceBase {
  name: string
  slug: string
  description: string
  devicetype_count: number
}

export interface RackRoleDTO extends CoreResourceBase {
  name: string
  slug: string
  color: string
  description: string
  rack_count: number
}

export interface RackTypeDTO extends CoreResourceBase {
  manufacturer: CoreReference
  model: string
  slug: string
  form_factor: ChoiceDTO<RackFormFactor>
  width: ChoiceDTO<10 | 19 | 21 | 23>
  u_height: number
  starting_unit: number
  desc_units: boolean
  description: string
  comments: string
}

export interface RackDTO extends CoreResourceBase {
  site: CoreReference
  name: string
  facility_id: string | null
  rack_type: CoreReference | null
  status: ChoiceDTO<RackStatus>
  role: CoreReference | null
  serial: string
  asset_tag: string | null
  form_factor: ChoiceDTO<RackFormFactor> | null
  width: ChoiceDTO<10 | 19 | 21 | 23>
  u_height: number
  starting_unit: number
  desc_units: boolean
  airflow: ChoiceDTO<RackAirflow> | null
  description: string
  comments: string
  device_count: number
}

export interface DeviceRoleDTO extends CoreResourceBase {
  parent: CoreReference | null
  name: string
  slug: string
  color: string
  vm_role: boolean
  description: string
  comments: string
  device_count: number
  _depth: number
}

export interface DeviceTypeDTO extends CoreResourceBase {
  manufacturer: CoreReference
  model: string
  slug: string
  part_number: string
  u_height: number
  exclude_from_utilization: boolean
  is_full_depth: boolean
  airflow: ChoiceDTO<Airflow> | null
  description: string
  comments: string
  device_count: number
  interface_template_count: number
}

export interface InterfaceTemplateDTO extends CoreResourceBase {
  device_type: CoreReference
  name: string
  label: string
  type: ChoiceDTO<string>
  enabled: boolean
  mgmt_only: boolean
  description: string
}

export interface DeviceDTO extends CoreResourceBase {
  device_type: CoreReference
  role: CoreReference
  name: string | null
  site: CoreReference
  rack: CoreReference | null
  position: number | null
  face: ChoiceDTO<'front' | 'rear'> | null
  status: ChoiceDTO<DeviceStatus>
  serial: string
  asset_tag: string | null
  airflow: ChoiceDTO<Airflow> | null
  description: string
  comments: string
  interface_count: number
}

export interface InterfaceDTO extends CoreResourceBase {
  device: CoreReference
  name: string
  label: string
  type: ChoiceDTO<string>
  enabled: boolean
  mgmt_only: boolean
  mtu: number | null
  speed: number | null
  duplex: ChoiceDTO<InterfaceDuplex> | null
  description: string
  count_ipaddresses: number
}

export interface VRFDTO extends CoreResourceBase {
  name: string
  rd: string | null
  enforce_unique: boolean
  description: string
  comments: string
  ipaddress_count: number
  prefix_count: number
}

export interface PrefixDTO extends CoreResourceBase {
  prefix: string
  vrf: CoreReference | null
  status: ChoiceDTO<PrefixStatus>
  is_pool: boolean
  mark_utilized: boolean
  description: string
  comments: string
  family: ChoiceDTO<4 | 6>
  children: number
  _depth: number
}

export interface IPAddressDTO extends Omit<CoreResourceBase, 'created' | 'last_updated'> {
  created: string | null
  last_updated: string | null
  address: string
  vrf: CoreReference | null
  status: ChoiceDTO<IPAddressStatus>
  role: ChoiceDTO<IPAddressRole> | null
  dns_name: string
  description: string
  comments: string
  assigned_object_type: 'dcim.interface' | null
  assigned_object_id: number | null
  family: ChoiceDTO<4 | 6>
  assigned_object: CoreReference | null
}

export interface CoreResourceDTOMap {
  site: SiteDTO
  manufacturer: ManufacturerDTO
  rackrole: RackRoleDTO
  racktype: RackTypeDTO
  rack: RackDTO
  devicerole: DeviceRoleDTO
  devicetype: DeviceTypeDTO
  interfacetemplate: InterfaceTemplateDTO
  device: DeviceDTO
  interface: InterfaceDTO
  vrf: VRFDTO
  prefix: PrefixDTO
  ipaddress: IPAddressDTO
}

export type CoreResourceDTO<N extends CoreProfileResourceName = CoreProfileResourceName> =
  CoreResourceDTOMap[N]
export type CorePage<N extends CoreProfileResourceName> = PaginatedResponse<CoreResourceDTO<N>>
export type CoreResponseFieldNameFor<N extends CoreProfileResourceName> = Extract<
  keyof CoreResourceDTOMap[N],
  string
>
export type CoreResponseFieldName = {
  [N in CoreProfileResourceName]: CoreResponseFieldNameFor<N>
}[CoreProfileResourceName]

type SiteCreateDTO = Omit<SiteDTO, 'device_count' | 'prefix_count' | 'rack_count'>
type ManufacturerCreateDTO = Omit<ManufacturerDTO, 'devicetype_count'>
type RackRoleCreateDTO = Omit<RackRoleDTO, 'rack_count'>
type RackCreateDTO = Omit<RackDTO, 'device_count'>
type DeviceTypeCreateDTO = Omit<DeviceTypeDTO, 'device_count'>
type VRFCreateDTO = Omit<VRFDTO, 'ipaddress_count' | 'prefix_count'>

/** POST responses intentionally omit NetBox queryset-only relationship counters. */
export interface CoreCreateResponseMap {
  site: SiteCreateDTO
  manufacturer: ManufacturerCreateDTO
  rackrole: RackRoleCreateDTO
  racktype: RackTypeDTO
  rack: RackCreateDTO
  devicerole: DeviceRoleDTO
  devicetype: DeviceTypeCreateDTO
  interfacetemplate: InterfaceTemplateDTO
  device: DeviceDTO
  interface: InterfaceDTO
  vrf: VRFCreateDTO
  prefix: PrefixDTO
  ipaddress: IPAddressDTO
}
export type CoreCreateResponse<N extends CoreProfileResourceName> = CoreCreateResponseMap[N]

export type CoreRelationSelection = number | CoreReference | null

export interface SiteForm {
  name?: string
  slug?: string
  status?: InfrastructureStatus
  facility?: string
  description?: string
  comments?: string
}
export interface ManufacturerForm {
  name?: string
  slug?: string
  description?: string
}
export interface RackRoleForm {
  name?: string
  slug?: string
  color?: string
  description?: string
}
export interface RackTypeForm {
  manufacturer?: CoreRelationSelection
  model?: string
  slug?: string
  form_factor?: RackFormFactor
  width?: 10 | 19 | 21 | 23
  u_height?: number
  starting_unit?: number
  desc_units?: boolean
  description?: string
  comments?: string
}
export interface RackForm {
  site?: CoreRelationSelection
  name?: string
  facility_id?: string | null
  rack_type?: CoreRelationSelection
  status?: RackStatus
  role?: CoreRelationSelection
  serial?: string
  asset_tag?: string | null
  form_factor?: RackFormFactor | null
  width?: 10 | 19 | 21 | 23
  u_height?: number
  starting_unit?: number
  desc_units?: boolean
  airflow?: RackAirflow | null
  description?: string
  comments?: string
}
export interface DeviceRoleForm {
  parent?: CoreRelationSelection
  name?: string
  slug?: string
  color?: string
  vm_role?: boolean
  description?: string
  comments?: string
}
export interface DeviceTypeForm {
  manufacturer?: CoreRelationSelection
  model?: string
  slug?: string
  part_number?: string
  u_height?: number
  exclude_from_utilization?: boolean
  is_full_depth?: boolean
  airflow?: Airflow | null
  description?: string
  comments?: string
}
export interface InterfaceTemplateForm {
  device_type?: CoreRelationSelection
  name?: string
  label?: string
  type?: string
  enabled?: boolean
  mgmt_only?: boolean
  description?: string
}
export interface DeviceForm {
  device_type?: CoreRelationSelection
  role?: CoreRelationSelection
  name?: string | null
  site?: CoreRelationSelection
  rack?: CoreRelationSelection
  position?: number | null
  face?: 'front' | 'rear' | '' | null
  status?: DeviceStatus
  serial?: string
  asset_tag?: string | null
  airflow?: Airflow | null
  description?: string
  comments?: string
}
export interface InterfaceForm {
  device?: CoreRelationSelection
  name?: string
  label?: string
  type?: string
  enabled?: boolean
  mgmt_only?: boolean
  mtu?: number | null
  speed?: number | null
  duplex?: InterfaceDuplex | null
  description?: string
}
export interface VRFForm {
  name?: string
  rd?: string | null
  enforce_unique?: boolean
  description?: string
  comments?: string
}
export interface PrefixForm {
  prefix?: string
  vrf?: CoreRelationSelection
  status?: PrefixStatus
  is_pool?: boolean
  mark_utilized?: boolean
  description?: string
  comments?: string
}
export interface IPAddressForm {
  address?: string
  vrf?: CoreRelationSelection
  status?: IPAddressStatus
  role?: IPAddressRole | null
  dns_name?: string
  description?: string
  comments?: string
  assigned_interface?: CoreRelationSelection
}

export interface CoreResourceFormMap {
  site: SiteForm
  manufacturer: ManufacturerForm
  rackrole: RackRoleForm
  racktype: RackTypeForm
  rack: RackForm
  devicerole: DeviceRoleForm
  devicetype: DeviceTypeForm
  interfacetemplate: InterfaceTemplateForm
  device: DeviceForm
  interface: InterfaceForm
  vrf: VRFForm
  prefix: PrefixForm
  ipaddress: IPAddressForm
}
export type CoreResourceForm<N extends CoreProfileResourceName = CoreProfileResourceName> =
  CoreResourceFormMap[N]
export type CoreFormFieldName = {
  [N in CoreProfileResourceName]: Extract<keyof CoreResourceFormMap[N], string>
}[CoreProfileResourceName]

export type SiteMutation = SiteForm
export type ManufacturerMutation = ManufacturerForm
export type RackRoleMutation = RackRoleForm
export interface RackTypeMutation extends Omit<RackTypeForm, 'manufacturer'> {
  manufacturer?: number | null
}
export interface RackMutation extends Omit<RackForm, 'site' | 'rack_type' | 'role'> {
  site?: number | null
  rack_type?: number | null
  role?: number | null
}
export interface DeviceRoleMutation extends Omit<DeviceRoleForm, 'parent'> {
  parent?: number | null
}
export interface DeviceTypeMutation extends Omit<DeviceTypeForm, 'manufacturer'> {
  manufacturer?: number | null
}
export interface InterfaceTemplateMutation extends Omit<InterfaceTemplateForm, 'device_type'> {
  device_type?: number | null
}
export interface DeviceMutation extends Omit<DeviceForm, 'device_type' | 'role' | 'site' | 'rack'> {
  device_type?: number | null
  role?: number | null
  site?: number | null
  rack?: number | null
}
export interface InterfaceMutation extends Omit<InterfaceForm, 'device'> {
  device?: number | null
}
export type VRFMutation = VRFForm
export interface PrefixMutation extends Omit<PrefixForm, 'vrf'> {
  vrf?: number | null
}
export interface IPAddressMutation extends Omit<IPAddressForm, 'vrf' | 'assigned_interface'> {
  vrf?: number | null
  assigned_object_type?: 'dcim.interface' | null
  assigned_object_id?: number | null
}

export interface CoreResourceMutationMap {
  site: SiteMutation
  manufacturer: ManufacturerMutation
  rackrole: RackRoleMutation
  racktype: RackTypeMutation
  rack: RackMutation
  devicerole: DeviceRoleMutation
  devicetype: DeviceTypeMutation
  interfacetemplate: InterfaceTemplateMutation
  device: DeviceMutation
  interface: InterfaceMutation
  vrf: VRFMutation
  prefix: PrefixMutation
  ipaddress: IPAddressMutation
}
export type CoreResourceMutation<N extends CoreProfileResourceName = CoreProfileResourceName> =
  CoreResourceMutationMap[N]

export interface CoreListControls {
  id?: number | string
  q?: string
  limit?: number
  offset?: number
  ordering?: string
}
export interface SiteFilters extends CoreListControls {
  name?: string
  slug?: string
  status?: InfrastructureStatus
}
export interface ManufacturerFilters extends CoreListControls {
  name?: string
  slug?: string
}
export type RackRoleFilters = ManufacturerFilters
export interface RackTypeFilters extends CoreListControls {
  manufacturer_id?: number
  manufacturer_slug?: string
  model?: string
  slug?: string
}
export interface RackFilters extends CoreListControls {
  site_id?: number
  site_slug?: string
  name?: string
  status?: RackStatus
  role_id?: number
  role_slug?: string
  rack_type_id?: number
  rack_type_slug?: string
}
export type DeviceRoleFilters = ManufacturerFilters
export type DeviceTypeFilters = RackTypeFilters
export interface InterfaceTemplateFilters extends CoreListControls {
  device_type_id?: number
  name?: string
  type?: string
  enabled?: boolean
  mgmt_only?: boolean
}
export interface DeviceFilters extends CoreListControls {
  site_id?: number
  site_slug?: string
  rack_id?: number
  device_type_id?: number
  device_type_slug?: string
  role_id?: number
  role_slug?: string
  name?: string
  status?: DeviceStatus
}
export interface InterfaceFilters extends CoreListControls {
  device_id?: number
  device_name?: string
  name?: string
  type?: string
  enabled?: boolean
  mgmt_only?: boolean
}
export interface VRFFilters extends CoreListControls {
  name?: string
  rd?: string
  enforce_unique?: boolean
}
export interface PrefixFilters extends CoreListControls {
  vrf_id?: number
  vrf_rd?: string
  prefix?: string
  family?: 4 | 6
  status?: PrefixStatus
  within?: string
  within_include?: string
  contains?: string
}
export interface IPAddressFilters extends CoreListControls {
  vrf_id?: number
  vrf_rd?: string
  address?: string
  family?: 4 | 6
  parent?: string
  status?: IPAddressStatus
  assigned?: boolean
  interface_id?: number
  device_id?: number
}

export interface CoreResourceFilterMap {
  site: SiteFilters
  manufacturer: ManufacturerFilters
  rackrole: RackRoleFilters
  racktype: RackTypeFilters
  rack: RackFilters
  devicerole: DeviceRoleFilters
  devicetype: DeviceTypeFilters
  interfacetemplate: InterfaceTemplateFilters
  device: DeviceFilters
  interface: InterfaceFilters
  vrf: VRFFilters
  prefix: PrefixFilters
  ipaddress: IPAddressFilters
}
export type CoreResourceFilters<N extends CoreProfileResourceName = CoreProfileResourceName> =
  CoreResourceFilterMap[N]
export type CoreFilterFieldName = {
  [N in CoreProfileResourceName]: Extract<keyof CoreResourceFilterMap[N], string>
}[CoreProfileResourceName]

/** Typed filter-panel state; adapters narrow this superset before transport. */
export interface CoreFilterState {
  name?: string
  slug?: string
  status?: InfrastructureStatus | RackStatus | DeviceStatus | PrefixStatus | IPAddressStatus
  manufacturer_id?: number
  manufacturer_slug?: string
  model?: string
  site_id?: number
  site_slug?: string
  role_id?: number
  role_slug?: string
  rack_type_id?: number
  rack_type_slug?: string
  rack_id?: number
  device_type_id?: number
  device_type_slug?: string
  device_id?: number
  device_name?: string
  type?: string
  enabled?: boolean
  mgmt_only?: boolean
  enforce_unique?: boolean
  rd?: string
  vrf_id?: number
  vrf_rd?: string
  prefix?: string
  address?: string
  family?: 4 | 6
  within?: string
  within_include?: string
  contains?: string
  parent?: string
  assigned?: boolean
  interface_id?: number
}

export type CoreFieldValue = string | number | boolean | null | CoreReference
export type CoreDetailFieldValue = CoreFieldValue | ChoiceDTO<string | number>

export function resourceField(resource: CoreResourceDTO, key: string): CoreFieldValue | undefined {
  const value: unknown = Reflect.get(resource, key)
  if (
    value === undefined ||
    value === null ||
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  ) {
    return value
  }
  if (isChoiceDTO(value)) return value.value
  return isCoreReference(value) ? value : undefined
}

/**
 * Read a field without flattening a choice envelope.
 *
 * Lists and mutation adapters consume a choice's scalar value, whereas detail
 * pages must preserve the API-provided human label.
 */
export function resourceDetailField(
  resource: CoreResourceDTO,
  key: string,
): CoreDetailFieldValue | undefined {
  const value: unknown = Reflect.get(resource, key)
  if (
    value === undefined ||
    value === null ||
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  ) {
    return value
  }
  if (isChoiceDTO(value) || isCoreReference(value)) return value
  return undefined
}

export function isChoiceDTO(value: unknown): value is ChoiceDTO<string | number> {
  return (
    value !== null &&
    typeof value === 'object' &&
    (typeof Reflect.get(value, 'value') === 'string' ||
      typeof Reflect.get(value, 'value') === 'number') &&
    typeof Reflect.get(value, 'label') === 'string'
  )
}

export function choiceValue<T extends string | number>(choice: ChoiceDTO<T>): T
export function choiceValue<T extends string | number>(choice: ChoiceDTO<T> | null): T | null
export function choiceValue<T extends string | number>(choice: ChoiceDTO<T> | null): T | null {
  return choice?.value ?? null
}

export function isCoreReference(value: unknown): value is CoreReference {
  return (
    value !== null &&
    typeof value === 'object' &&
    typeof Reflect.get(value, 'id') === 'number' &&
    typeof Reflect.get(value, 'display') === 'string' &&
    typeof Reflect.get(value, 'url') === 'string'
  )
}

export function relationID(value: CoreRelationSelection | undefined): number | null | undefined {
  return typeof value === 'object' && value !== null ? value.id : value
}

export function formField(form: CoreResourceForm, key: string): unknown {
  return Reflect.get(form, key) as unknown
}

export function withFormField(
  form: CoreResourceForm,
  key: string,
  value: unknown,
): CoreResourceForm {
  const next = { ...form }
  Reflect.set(next, key, value)
  return next
}

export function hasFormField(form: CoreResourceForm, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(form, key)
}
