import { describe, expect, it } from 'vitest'
import { CORE_PROFILE_RESOURCE_NAMES } from './manifest'
import { getCoreResourceAdapter } from './adapters'
import {
  resourceField,
  withFormField,
  type CoreReference,
  type IPAddressDTO,
  type IPAddressForm,
  type ManufacturerDTO,
  type ManufacturerForm,
  type RackRoleDTO,
  type RackRoleForm,
  type RackTypeDTO,
  type RackTypeForm,
  type SiteDTO,
  type SiteForm,
} from './resources'

const manufacturer: CoreReference = {
  id: 2,
  url: '/api/dcim/manufacturers/2/',
  display: 'Acme',
}
const site: CoreReference = { id: 3, url: '/api/dcim/sites/3/', display: 'Moscow' }
const rackType: CoreReference = {
  id: 4,
  url: '/api/dcim/rack-types/4/',
  display: 'Acme R42',
}
const rackRole: CoreReference = {
  id: 5,
  url: '/api/dcim/rack-roles/5/',
  display: 'Production',
}
const deviceType: CoreReference = {
  id: 6,
  url: '/api/dcim/device-types/6/',
  display: 'Acme Edge',
}
const deviceRole: CoreReference = {
  id: 7,
  url: '/api/dcim/device-roles/7/',
  display: 'Router',
}
const rack: CoreReference = { id: 8, url: '/api/dcim/racks/8/', display: 'R1' }
const device: CoreReference = { id: 9, url: '/api/dcim/devices/9/', display: 'edge-1' }
const vrf: CoreReference = { id: 10, url: '/api/ipam/vrfs/10/', display: 'Production' }
const assignedInterface: CoreReference = {
  id: 11,
  url: '/api/dcim/interfaces/11/',
  display: 'edge-1 (xe-0/0/0)',
}

describe('typed core resource adapters', () => {
  it('owns exactly one adapter for every promoted resource', () => {
    expect(
      CORE_PROFILE_RESOURCE_NAMES.map((name) => getCoreResourceAdapter(name).resource),
    ).toEqual(CORE_PROFILE_RESOURCE_NAMES)
  })

  it('serializes all 13 form models through their resource-specific mutation contracts', () => {
    expect(
      getCoreResourceAdapter('site').mutationFromForm({ name: 'MOW', slug: 'mow' }, false),
    ).toEqual({ name: 'MOW', slug: 'mow' })
    expect(
      getCoreResourceAdapter('manufacturer').mutationFromForm(
        { name: 'Acme', slug: 'acme' },
        false,
      ),
    ).toEqual({ name: 'Acme', slug: 'acme' })
    expect(
      getCoreResourceAdapter('rackrole').mutationFromForm(
        { name: 'Production', slug: 'production', color: '00ff00' },
        false,
      ),
    ).toEqual({ name: 'Production', slug: 'production', color: '00ff00' })
    expect(
      getCoreResourceAdapter('racktype').mutationFromForm(
        { manufacturer, model: 'R42', slug: 'r42', form_factor: '4-post-cabinet' },
        false,
      ),
    ).toEqual({ manufacturer: 2, model: 'R42', slug: 'r42', form_factor: '4-post-cabinet' })
    expect(
      getCoreResourceAdapter('rack').mutationFromForm(
        { site, name: 'R1', rack_type: rackType, role: rackRole, width: 23 },
        false,
      ),
    ).toEqual({ site: 3, name: 'R1', rack_type: 4, role: 5 })
    expect(
      getCoreResourceAdapter('devicerole').mutationFromForm(
        { parent: deviceRole, name: 'Child', slug: 'child' },
        false,
      ),
    ).toEqual({ parent: 7, name: 'Child', slug: 'child' })
    expect(
      getCoreResourceAdapter('devicetype').mutationFromForm(
        { manufacturer, model: 'Edge', slug: 'edge' },
        false,
      ),
    ).toEqual({ manufacturer: 2, model: 'Edge', slug: 'edge' })
    expect(
      getCoreResourceAdapter('interfacetemplate').mutationFromForm(
        { device_type: deviceType, name: 'xe-0/0/0', type: '1000base-t' },
        false,
      ),
    ).toEqual({ device_type: 6, name: 'xe-0/0/0', type: '1000base-t' })
    expect(
      getCoreResourceAdapter('device').mutationFromForm(
        {
          device_type: deviceType,
          role: deviceRole,
          site,
          rack,
          position: 10,
          face: 'front',
        },
        false,
      ),
    ).toEqual({ device_type: 6, role: 7, site: 3, rack: 8, position: 10, face: 'front' })
    expect(
      getCoreResourceAdapter('interface').mutationFromForm(
        { device, name: 'xe-0/0/0', type: '1000base-t' },
        false,
      ),
    ).toEqual({ device: 9, name: 'xe-0/0/0', type: '1000base-t' })
    expect(
      getCoreResourceAdapter('vrf').mutationFromForm({ name: 'Production', rd: '65000:10' }, false),
    ).toEqual({ name: 'Production', rd: '65000:10' })
    expect(
      getCoreResourceAdapter('prefix').mutationFromForm({ prefix: '192.0.2.0/24', vrf }, false),
    ).toEqual({ prefix: '192.0.2.0/24', vrf: 10 })
    expect(
      getCoreResourceAdapter('ipaddress').mutationFromForm(
        { address: '192.0.2.1/24', vrf, assigned_interface: assignedInterface },
        false,
      ),
    ).toEqual({
      address: '192.0.2.1/24',
      vrf: 10,
      assigned_object_type: 'dcim.interface',
      assigned_object_id: 11,
    })
  })

  it('enforces edit immutability and dependent-field omission in the adapter', () => {
    expect(
      getCoreResourceAdapter('interface').mutationFromForm(
        { device, name: 'xe-0/0/1', type: '1000base-t' },
        true,
      ),
    ).toEqual({ name: 'xe-0/0/1', type: '1000base-t' })
    expect(
      getCoreResourceAdapter('device').mutationFromForm(
        {
          device_type: deviceType,
          role: deviceRole,
          site,
          rack: null,
          position: 10,
          face: 'front',
        },
        false,
      ),
    ).toEqual({ device_type: 6, role: 7, site: 3, rack: null })
  })

  it('preserves IPAddress scalar presence in create and PATCH mutations', () => {
    const adapter = getCoreResourceAdapter('ipaddress')

    const create = adapter.mutationFromForm({ address: '192.0.2.17', status: 'active' }, false)
    expect(create).toEqual({ address: '192.0.2.17', status: 'active' })
    for (const omitted of ['role', 'dns_name', 'description', 'comments']) {
      expect(create).not.toHaveProperty(omitted)
    }
    expect(
      adapter.mutationFromForm(
        {
          address: '192.0.2.18',
          status: 'reserved',
          role: null,
          dns_name: 'edge.example',
          description: '',
          comments: '',
        },
        false,
      ),
    ).toEqual({
      address: '192.0.2.18',
      status: 'reserved',
      role: null,
      dns_name: 'edge.example',
      description: '',
      comments: '',
    })

    expect(adapter.mutationFromForm({}, true)).toEqual({})
    expect(adapter.mutationFromForm({ role: null }, true)).toEqual({ role: null })
    expect(adapter.mutationFromForm({ role: 'loopback' }, true)).toEqual({ role: 'loopback' })
    expect(adapter.mutationFromForm({ dns_name: '', description: '', comments: '' }, true)).toEqual(
      { dns_name: '', description: '', comments: '' },
    )
    expect(
      adapter.mutationFromForm(
        {
          address: '2001:db8::17',
          status: 'reserved',
          role: 'anycast',
          dns_name: 'edge.example',
          description: 'Edge address',
          comments: 'Operator note',
        },
        true,
      ),
    ).toEqual({
      address: '2001:db8::17',
      status: 'reserved',
      role: 'anycast',
      dns_name: 'edge.example',
      description: 'Edge address',
      comments: 'Operator note',
    })
  })

  it('preserves Site scalar presence in create mutations', () => {
    const adapter = getCoreResourceAdapter('site')

    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({ status: 'active' })
    expect(
      adapter.mutationFromForm(
        {
          name: 'MOW',
          slug: 'mow',
          status: 'planned',
          facility: '',
          description: '',
          comments: '',
        },
        false,
      ),
    ).toEqual({
      name: 'MOW',
      slug: 'mow',
      status: 'planned',
      facility: '',
      description: '',
      comments: '',
    })
    expect(
      adapter.mutationFromForm(
        {
          name: 'LED',
          slug: 'led',
          status: 'staging',
          facility: 'LED1',
          description: 'Staging site',
          comments: 'Operator note',
        },
        false,
      ),
    ).toEqual({
      name: 'LED',
      slug: 'led',
      status: 'staging',
      facility: 'LED1',
      description: 'Staging site',
      comments: 'Operator note',
    })
  })

  it('omits unchanged Site scalars and preserves explicit clears in PATCH mutations', () => {
    const adapter = getCoreResourceAdapter('site')
    const dto: SiteDTO = {
      id: 3,
      url: '/api/dcim/sites/3/',
      display: 'MOW',
      created: null,
      last_updated: null,
      name: 'MOW',
      slug: 'mow',
      status: { value: 'active', label: 'Active' },
      facility: 'MOW1',
      description: 'Primary',
      comments: 'Operator note',
      device_count: 4,
      prefix_count: 2,
      rack_count: 1,
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})
    expect(
      adapter.mutationFromForm({ ...hydrated, facility: '', description: '', comments: '' }, true),
    ).toEqual({ facility: '', description: '', comments: '' })
    expect(adapter.mutationFromForm({ ...hydrated, status: 'planned' }, true)).toEqual({
      status: 'planned',
    })

    const renamed = withFormField(hydrated, 'name', 'Moscow') as SiteForm
    expect(adapter.mutationFromForm(renamed, true)).toEqual({ name: 'Moscow' })
  })

  it('preserves Manufacturer scalar presence in create mutations', () => {
    const adapter = getCoreResourceAdapter('manufacturer')

    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({})
    expect(adapter.mutationFromForm({ name: 'Acme', slug: 'acme' }, false)).toEqual({
      name: 'Acme',
      slug: 'acme',
    })
    expect(
      adapter.mutationFromForm(
        { name: 'Empty Description', slug: 'empty-description', description: '' },
        false,
      ),
    ).toEqual({ name: 'Empty Description', slug: 'empty-description', description: '' })
    expect(
      adapter.mutationFromForm(
        { name: 'Juniper', slug: 'juniper', description: 'Network manufacturer' },
        false,
      ),
    ).toEqual({ name: 'Juniper', slug: 'juniper', description: 'Network manufacturer' })
  })

  it('omits unchanged Manufacturer scalars and preserves explicit clears in PATCH mutations', () => {
    const adapter = getCoreResourceAdapter('manufacturer')
    const dto: ManufacturerDTO = {
      id: 2,
      url: '/api/dcim/manufacturers/2/',
      display: 'Acme',
      created: null,
      last_updated: null,
      name: 'Acme',
      slug: 'acme',
      description: 'Network manufacturer',
      devicetype_count: 4,
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})

    const cleared = withFormField(hydrated, 'description', '') as ManufacturerForm
    expect(adapter.mutationFromForm(cleared, true)).toEqual({ description: '' })

    const renamed = withFormField(hydrated, 'name', 'Acme Networks') as ManufacturerForm
    expect(adapter.mutationFromForm(renamed, true)).toEqual({ name: 'Acme Networks' })

    const reslugged = withFormField(hydrated, 'slug', 'acme-networks') as ManufacturerForm
    expect(adapter.mutationFromForm(reslugged, true)).toEqual({ slug: 'acme-networks' })
  })

  it('preserves RackRole scalar presence and the pinned create default', () => {
    const adapter = getCoreResourceAdapter('rackrole')

    expect(adapter.mutationFromForm({}, false)).toEqual({})
    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({ color: '9e9e9e' })
    expect(adapter.mutationFromForm({ name: 'Core', slug: 'core' }, false)).toEqual({
      name: 'Core',
      slug: 'core',
    })
    expect(
      adapter.mutationFromForm(
        { name: 'Blank Fields', slug: 'blank-fields', color: '', description: '' },
        false,
      ),
    ).toEqual({ name: 'Blank Fields', slug: 'blank-fields', color: '', description: '' })
    expect(
      adapter.mutationFromForm(
        {
          name: 'Distribution',
          slug: 'distribution',
          color: '00ff00',
          description: 'Distribution racks',
        },
        false,
      ),
    ).toEqual({
      name: 'Distribution',
      slug: 'distribution',
      color: '00ff00',
      description: 'Distribution racks',
    })
  })

  it('omits unchanged RackRole scalars and emits only dirty PATCH fields', () => {
    const adapter = getCoreResourceAdapter('rackrole')
    const dto: RackRoleDTO = {
      id: 5,
      url: '/api/dcim/rack-roles/5/',
      display: 'Core',
      created: null,
      last_updated: null,
      name: 'Core',
      slug: 'core',
      color: '9e9e9e',
      description: 'Core racks',
      rack_count: 3,
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})

    const cleared = withFormField(hydrated, 'description', '') as RackRoleForm
    expect(adapter.mutationFromForm(cleared, true)).toEqual({ description: '' })

    const renamed = withFormField(hydrated, 'name', 'Core Infrastructure') as RackRoleForm
    expect(adapter.mutationFromForm(renamed, true)).toEqual({ name: 'Core Infrastructure' })

    const reslugged = withFormField(hydrated, 'slug', 'core-infrastructure') as RackRoleForm
    expect(adapter.mutationFromForm(reslugged, true)).toEqual({ slug: 'core-infrastructure' })

    const recolored = withFormField(hydrated, 'color', '00ff00') as RackRoleForm
    expect(adapter.mutationFromForm(recolored, true)).toEqual({ color: '00ff00' })
  })

  it('preserves RackType create omissions, defaults, IDs, and concrete scalar states', () => {
    const adapter = getCoreResourceAdapter('racktype')

    expect(adapter.mutationFromForm({}, false)).toEqual({})
    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({
      width: 19,
      u_height: 42,
      starting_unit: 1,
      desc_units: false,
    })
    expect(
      adapter.mutationFromForm(
        {
          manufacturer,
          model: 'R42',
          slug: 'r42',
          form_factor: '4-post-cabinet',
          width: 23,
          u_height: 48,
          starting_unit: 2,
          desc_units: true,
          description: 'Core rack',
          comments: 'Operator note',
        },
        false,
      ),
    ).toEqual({
      manufacturer: 2,
      model: 'R42',
      slug: 'r42',
      form_factor: '4-post-cabinet',
      width: 23,
      u_height: 48,
      starting_unit: 2,
      desc_units: true,
      description: 'Core rack',
      comments: 'Operator note',
    })
    expect(
      adapter.mutationFromForm(
        {
          manufacturer: null,
          u_height: 0,
          starting_unit: 0,
          desc_units: false,
          description: '',
          comments: '',
        },
        false,
      ),
    ).toEqual({
      manufacturer: null,
      u_height: 0,
      starting_unit: 0,
      desc_units: false,
      description: '',
      comments: '',
    })
  })

  it('omits unchanged RackType scalars and emits one normalized dirty PATCH field', () => {
    const adapter = getCoreResourceAdapter('racktype')
    const dto: RackTypeDTO = {
      id: 4,
      url: '/api/dcim/rack-types/4/',
      display: 'Acme R42',
      created: null,
      last_updated: null,
      manufacturer,
      model: 'R42',
      slug: 'r42',
      form_factor: { value: '4-post-cabinet', label: '4-post cabinet' },
      width: { value: 19, label: '19 inches' },
      u_height: 42,
      starting_unit: 1,
      desc_units: true,
      description: 'Core rack',
      comments: 'Operator note',
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})

    const otherManufacturer: CoreReference = {
      id: 12,
      url: '/api/dcim/manufacturers/12/',
      display: 'Globex',
    }
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'manufacturer', otherManufacturer) as RackTypeForm,
        true,
      ),
    ).toEqual({ manufacturer: 12 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'model', 'R48') as RackTypeForm, true),
    ).toEqual({ model: 'R48' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'slug', 'r48') as RackTypeForm, true),
    ).toEqual({ slug: 'r48' })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'form_factor', '2-post-frame') as RackTypeForm,
        true,
      ),
    ).toEqual({ form_factor: '2-post-frame' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'width', 23) as RackTypeForm, true),
    ).toEqual({ width: 23 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'u_height', 0) as RackTypeForm, true),
    ).toEqual({ u_height: 0 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'starting_unit', 2) as RackTypeForm, true),
    ).toEqual({ starting_unit: 2 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'desc_units', false) as RackTypeForm, true),
    ).toEqual({ desc_units: false })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'description', '') as RackTypeForm, true),
    ).toEqual({ description: '' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'comments', '') as RackTypeForm, true),
    ).toEqual({ comments: '' })
  })

  it('omits unchanged IPAddress scalars without changing relationship serialization', () => {
    const adapter = getCoreResourceAdapter('ipaddress')
    const dto: IPAddressDTO = {
      id: 17,
      url: '/api/ipam/ip-addresses/17/',
      display: '192.0.2.17/24',
      created: '2026-08-19T10:00:00Z',
      last_updated: '2026-08-19T10:00:00Z',
      address: '192.0.2.17/24',
      vrf,
      status: { value: 'active', label: 'Active' },
      role: { value: 'loopback', label: 'Loopback' },
      dns_name: 'edge.example',
      description: 'Existing description',
      comments: 'Existing comment',
      assigned_object_type: 'dcim.interface',
      assigned_object_id: assignedInterface.id,
      family: { value: 4, label: 'IPv4' },
      assigned_object: assignedInterface,
    }
    const hydrated = adapter.formFromDTO(dto)

    const unchangedRelationships = {
      vrf: vrf.id,
      assigned_object_type: 'dcim.interface',
      assigned_object_id: assignedInterface.id,
    } as const
    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual(unchangedRelationships)

    const nullableBlankHydrated = adapter.formFromDTO({
      ...dto,
      role: null,
      dns_name: '',
      description: '',
      comments: '',
    })
    expect(adapter.mutationFromForm({ ...nullableBlankHydrated }, true)).toEqual(
      unchangedRelationships,
    )

    const clearedRole = withFormField(hydrated, 'role', null) as IPAddressForm
    expect(adapter.mutationFromForm(clearedRole, true)).toEqual({
      ...unchangedRelationships,
      role: null,
    })
    expect(adapter.mutationFromForm({ ...hydrated, description: '' }, true)).toEqual({
      ...unchangedRelationships,
      description: '',
    })
    expect(adapter.mutationFromForm({ ...hydrated, role: 'anycast' }, true)).toEqual({
      ...unchangedRelationships,
      role: 'anycast',
    })
    expect(adapter.mutationFromForm({ ...hydrated, dns_name: 'new.example' }, true)).toEqual({
      ...unchangedRelationships,
      dns_name: 'new.example',
    })
  })

  it('narrows panel state to the declared resource filter contract', () => {
    expect(
      getCoreResourceAdapter('rack').filtersFromState({
        site_id: 3,
        status: 'active',
        address: 'must-not-leak',
      }),
    ).toEqual({ site_id: 3, status: 'active' })
  })

  it('hydrates a typed form without response-only fields', () => {
    const dto: SiteDTO = {
      id: 3,
      url: '/api/dcim/sites/3/',
      display: 'MOW',
      created: '2026-07-18T10:00:00Z',
      last_updated: '2026-07-18T10:00:00Z',
      name: 'MOW',
      slug: 'mow',
      status: { value: 'active', label: 'Active' },
      facility: 'MOW1',
      description: 'Primary',
      comments: '',
      device_count: 4,
      prefix_count: 2,
      rack_count: 1,
    }

    const hydrated = getCoreResourceAdapter('site').formFromDTO(dto)
    expect(Object.fromEntries(Object.entries(hydrated))).toEqual({
      name: 'MOW',
      slug: 'mow',
      status: 'active',
      facility: 'MOW1',
      description: 'Primary',
      comments: '',
    })
    expect(resourceField(dto, 'status')).toBe('active')
  })
})
