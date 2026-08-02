import { describe, expect, it } from 'vitest'
import { NAVIGATION } from '@/config/navigation'
import { CORE_RESOURCE_REGISTRY } from './core-resource-registry'
import { CORE_PROFILE_RESOURCE_NAMES } from './models/core-profile'
import { routes } from './index'

const expectedPaths = [
  '/dcim/sites/',
  '/dcim/manufacturers/',
  '/dcim/rack-roles/',
  '/dcim/rack-types/',
  '/dcim/racks/',
  '/dcim/device-roles/',
  '/dcim/device-types/',
  '/dcim/interface-templates/',
  '/dcim/devices/',
  '/dcim/interfaces/',
  '/ipam/vrfs/',
  '/ipam/prefixes/',
  '/ipam/ip-addresses/',
]

const expectedWritableFields: Record<string, string[]> = {
  site: ['name', 'slug', 'status', 'facility', 'description', 'comments'],
  manufacturer: ['name', 'slug', 'description'],
  rackrole: ['name', 'slug', 'color', 'description'],
  racktype: [
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
  ],
  rack: [
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
  ],
  devicerole: ['parent', 'name', 'slug', 'color', 'vm_role', 'description', 'comments'],
  devicetype: [
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
  ],
  interfacetemplate: [
    'device_type',
    'name',
    'label',
    'type',
    'enabled',
    'mgmt_only',
    'description',
  ],
  device: [
    'device_type',
    'role',
    'name',
    'site',
    'rack',
    'position',
    'face',
    'status',
    'serial',
    'asset_tag',
    'airflow',
    'description',
    'comments',
  ],
  interface: [
    'device',
    'name',
    'label',
    'type',
    'enabled',
    'mgmt_only',
    'mtu',
    'speed',
    'duplex',
    'description',
  ],
  vrf: ['name', 'rd', 'enforce_unique', 'description', 'comments'],
  prefix: ['prefix', 'vrf', 'status', 'is_pool', 'mark_utilized', 'description', 'comments'],
  ipaddress: [
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
}

describe('core profile runtime registry', () => {
  it('publishes exactly the 13 core-workflow resources', () => {
    expect(CORE_RESOURCE_REGISTRY.map((model) => model.model)).toEqual(CORE_PROFILE_RESOURCE_NAMES)
    expect(CORE_RESOURCE_REGISTRY.map((model) => model.routePath)).toEqual(expectedPaths)
  })

  it('declares exactly the profile writable fields', () => {
    for (const model of CORE_RESOURCE_REGISTRY) {
      expect(model.writableFields, model.model).toEqual(expectedWritableFields[model.model])
      const visibleFields = model.fields.map((field) => field.key)
      if (model.model === 'ipaddress') {
        expect(visibleFields).toEqual([
          ...expectedWritableFields.ipaddress.slice(0, -2),
          'assigned_interface',
        ])
      } else {
        expect(visibleFields, model.model).toEqual(expectedWritableFields[model.model])
      }
    }
  })

  it('does not query relations outside the published registry', () => {
    const allowedResources = new Set(CORE_RESOURCE_REGISTRY.map((model) => model.model))
    for (const model of CORE_RESOURCE_REGISTRY) {
      for (const field of [...model.fields, ...model.filters]) {
        if (field.relationResource) {
          expect(allowedResources.has(field.relationResource), `${model.model}.${field.key}`).toBe(
            true,
          )
        }
      }
    }
  })

  it('uses profile-compatible rack choices and half-unit device heights', () => {
    const rack = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'rack')!
    const rackStatus = rack.fields.find((field) => field.key === 'status')!
    expect(rackStatus.options?.map((option) => option.value)).toEqual([
      'reserved',
      'available',
      'planned',
      'active',
      'deprecated',
    ])
    expect(rack.fields.find((field) => field.key === 'width')?.options?.[0].value).toBe(10)

    const deviceType = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'devicetype')!
    const deviceTypeHeight = deviceType.fields.find((field) => field.key === 'u_height')!
    expect(deviceTypeHeight.step).toBe(0.5)
    expect(deviceTypeHeight.max).toBe(999.9)
    const device = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'device')!
    const devicePosition = device.fields.find((field) => field.key === 'position')!
    expect(devicePosition).toEqual(expect.objectContaining({ min: 1, max: 100.5, step: 0.5 }))
    expect(devicePosition.disabledUnlessFieldTruthy).toBe('rack')
    const deviceRack = device.fields.find((field) => field.key === 'rack')!
    expect(deviceRack.relationFilterFields).toEqual({ site_id: 'site' })
    expect(deviceRack.clearWhenFieldChanges).toEqual(['site'])
    const deviceFace = device.fields.find((field) => field.key === 'face')!
    expect(deviceFace.requiredWhenFieldTruthy).toBe('position')

    const rackType = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'racktype')!
    expect(rackType.fields.find((field) => field.key === 'form_factor')?.required).toBe(true)
    const deviceRole = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'devicerole')!
    expect(deviceRole.fields.find((field) => field.key === 'vm_role')?.default).toBe(true)
    const vrf = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'vrf')!
    expect(vrf.fields.find((field) => field.key === 'enforce_unique')?.default).toBe(true)

    for (const modelName of ['interfacetemplate', 'interface']) {
      const model = CORE_RESOURCE_REGISTRY.find((candidate) => candidate.model === modelName)!
      const typeField = model.fields.find((field) => field.key === 'type')!
      expect(typeField.options).toHaveLength(206)
      expect(typeField.default).toBeUndefined()
    }

    const interfaceTemplate = CORE_RESOURCE_REGISTRY.find(
      (model) => model.model === 'interfacetemplate',
    )!
    expect(
      interfaceTemplate.fields.find((field) => field.key === 'device_type')?.immutableOnEdit,
    ).toBe(true)
    const interfaceModel = CORE_RESOURCE_REGISTRY.find((model) => model.model === 'interface')!
    expect(interfaceModel.fields.find((field) => field.key === 'device')?.immutableOnEdit).toBe(
      true,
    )

    for (const inheritedField of [
      'form_factor',
      'width',
      'u_height',
      'starting_unit',
      'desc_units',
    ]) {
      expect(
        rack.fields.find((field) => field.key === inheritedField)?.disabledWhenFieldTruthy,
      ).toBe('rack_type')
    }
  })

  it('exposes every displayed backend-supported ordering field and no others', () => {
    const allowedOrdering: Record<string, string[]> = {
      site: ['id', 'name', 'slug', 'status', 'created', 'last_updated'],
      manufacturer: ['id', 'name', 'slug', 'created', 'last_updated'],
      rackrole: ['id', 'name', 'slug', 'created', 'last_updated'],
      racktype: ['id', 'manufacturer', 'model', 'slug', 'u_height', 'created', 'last_updated'],
      rack: ['id', 'site', 'name', 'facility_id', 'status', 'u_height', 'created', 'last_updated'],
      devicerole: ['id', 'name', 'slug', 'created', 'last_updated'],
      devicetype: ['id', 'manufacturer', 'model', 'slug', 'u_height', 'created', 'last_updated'],
      interfacetemplate: ['id', 'device_type', 'name', 'type', 'created', 'last_updated'],
      device: ['id', 'site', 'rack', 'position', 'name', 'status', 'created', 'last_updated'],
      interface: ['id', 'device', 'name', 'type', 'created', 'last_updated'],
      vrf: ['id', 'name', 'rd', 'created', 'last_updated'],
      prefix: ['id', 'vrf', 'prefix', 'status', 'created', 'last_updated'],
      ipaddress: ['id', 'vrf', 'address', 'status', 'dns_name', 'created', 'last_updated'],
    }

    for (const model of CORE_RESOURCE_REGISTRY) {
      const expected = model.columns
        .filter((column) => allowedOrdering[model.model].includes(column.key))
        .map((column) => column.key)
      const actual = model.columns.filter((column) => column.sortable).map((column) => column.key)
      expect(actual, model.model).toEqual(expected)
    }
  })
})

describe('supported router and navigation', () => {
  it('registers only single-object CRUD routes for each profile resource', () => {
    const root = routes.find((route) => route.path === '/')
    const names = (root?.children ?? []).map((route) => String(route.name))

    for (const resource of CORE_PROFILE_RESOURCE_NAMES) {
      expect(names).toEqual(
        expect.arrayContaining([
          `${resource}-list`,
          `${resource}-add`,
          `${resource}-detail`,
          `${resource}-edit`,
          `${resource}-delete`,
        ]),
      )
    }

    expect(names).not.toEqual(
      expect.arrayContaining([
        'reports',
        'scripts',
        'graphql',
        'fallback-list',
        'site-import',
        'site-bulk-edit',
        'site-bulk-delete',
      ]),
    )
  })

  it('links only to the dashboard, profile resources, and API browser', () => {
    const items = NAVIGATION.flatMap((group) => group.items)
    expect(items.map((item) => item.route)).toEqual(['/', ...expectedPaths, '/api/'])

    for (const button of items.flatMap((item) => item.buttons ?? [])) {
      expect(button.label).toBe('Add')
      expect(button.route).toMatch(/\/add\/$/)
    }
  })
})
