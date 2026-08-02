/**
 * Core API type definitions
 */

import type { CoreProfileModule, CoreProfileResourceName } from '@/features/core/manifest'

export interface PaginatedResponse<T> {
  count: number
  next: string | null
  previous: string | null
  results: T[]
}

export interface ApiError {
  detail?: string
  [key: string]: unknown
}

export interface ChoiceOption {
  value: string | number | boolean
  label: string
  color?: string
}

export interface NestedObject {
  id: number
  name?: string
  display?: string
  url?: string
  [key: string]: unknown
}

export interface InventoryItemNode {
  id: number
  name: string
  label?: string
  part_id?: string
  serial?: string
  parent?: { id: number } | null
  role?: NestedObject | null
  manufacturer?: NestedObject | null
  children: InventoryItemNode[]
}

export interface PrefixSummary {
  id: number
  prefix: string
  status: string
  vrf?: { name?: string; display?: string } | null
  utilization?: number
  address_count?: number
  _children?: PrefixSummary[]
}

export interface Tag {
  id: number
  name: string
  slug: string
  color: string
}

export interface BaseModel {
  id: number
  url?: string
  display?: string
  name?: string
  description?: string | null
  comments?: string | null
  slug?: string
  created?: string
  last_updated?: string
  tags?: Tag[]
  custom_field_data?: Record<string, unknown>
}

export interface TableColumn {
  key: string
  label: string
  sortable?: boolean
  width?: string
  formatter?: (value: unknown, row: object) => string
}

/**
 * One field in a resource detail projection.
 *
 * Detail projections are deliberately independent from list columns: list
 * views optimize for scanning while detail views expose the complete promoted
 * profile for one object.
 */
export interface DetailFieldDef<K extends string = string> {
  key: K
  label: string
  presentation?: 'default' | 'status' | 'markdown'
}

export interface FilterField {
  key: string
  label: string
  type: 'text' | 'select' | 'api-select' | 'boolean' | 'date-range' | 'integer-range'
  options?: ChoiceOption[]
  relationResource?: CoreProfileResourceName
  multiple?: boolean
  placeholder?: string
}

export interface ModelConfig {
  module: CoreProfileModule
  model: CoreProfileResourceName
  display_name: string
  display_name_plural: string
  apiPath: string
  routePath: string
  /** Exact baseline REST fields; UI-only semantic fields are mapped before transport. */
  writableFields?: string[]
  icon?: string
  columns: TableColumn[]
  detailFields: DetailFieldDef[]
  filters: FilterField[]
  fields: FormFieldDef[]
  tabs?: DetailTabDef[]
  features?: {
    contacts?: boolean
    journal?: boolean
    changelog?: boolean
    images?: boolean
    sync?: boolean
  }
  statusChoices?: ChoiceOption[]
}

export interface FormFieldDef {
  key: string
  /** Optional response field used to hydrate this writable field. */
  sourceKey?: string
  /** Form field used for automatic slug generation. */
  slugSource?: string
  label: string
  type:
    | 'text'
    | 'slug'
    | 'number'
    | 'boolean'
    | 'select'
    | 'api-select'
    | 'tag'
    | 'markdown'
    | 'json'
    | 'date'
    | 'datetime'
    | 'csv'
    | 'textarea'
  required?: boolean
  helpText?: string
  placeholder?: string
  options?: ChoiceOption[]
  /** A relationship target from the closed core capability manifest. */
  relationResource?: CoreProfileResourceName
  /** Baseline relation can be chosen at creation but cannot be moved later. */
  immutableOnEdit?: boolean
  /** Disable this control while the named form field has a value. */
  disabledWhenFieldTruthy?: string
  /** Disable this control until the named form field has a value. */
  disabledUnlessFieldTruthy?: string
  /** Make this control required while the named form field has a value. */
  requiredWhenFieldTruthy?: string
  /** Clear this field whenever one of these dependency fields changes. */
  clearWhenFieldChanges?: string[]
  /** Map REST relation filters to source fields in the current form. */
  relationFilterFields?: Record<string, string>
  multiple?: boolean
  group?: string
  default?: unknown
  min?: number
  max?: number
  step?: number
}

export interface DetailTabDef {
  key: string
  label: string
  component?: string
  badgeCount?: string
}

export interface ToastNotification {
  id: number
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}

export interface BreadcrumbItem {
  label: string
  to?: string
}
