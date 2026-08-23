import { describe, expect, it } from 'vitest'
import { CORE_PROFILE_RESOURCE_NAMES } from './manifest'
import { getCoreResourceAdapter } from './adapters'
import {
  resourceField,
  withFormField,
  type CoreReference,
  type IPAddressDTO,
  type IPAddressForm,
  type SiteDTO,
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

    expect(getCoreResourceAdapter('site').formFromDTO(dto)).toEqual({
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
