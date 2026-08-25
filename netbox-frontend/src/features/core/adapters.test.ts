import { describe, expect, it } from 'vitest'
import { CORE_PROFILE_RESOURCE_NAMES } from './manifest'
import { getCoreResourceAdapter } from './adapters'
import {
  resourceField,
  withFormField,
  type CoreReference,
  type DeviceRoleDTO,
  type DeviceRoleForm,
  type DeviceTypeDTO,
  type DeviceTypeForm,
  type InterfaceTemplateDTO,
  type InterfaceTemplateForm,
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

  it('preserves DeviceRole create omissions, defaults, IDs, and concrete scalar states', () => {
    const adapter = getCoreResourceAdapter('devicerole')

    expect(adapter.mutationFromForm({}, false)).toEqual({})
    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({
      color: '9e9e9e',
      vm_role: true,
    })
    expect(
      adapter.mutationFromForm(
        {
          parent: deviceRole,
          name: 'Access',
          slug: 'access',
          color: '00ff00',
          vm_role: true,
          description: 'Access devices',
          comments: 'Operator note',
        },
        false,
      ),
    ).toEqual({
      parent: 7,
      name: 'Access',
      slug: 'access',
      color: '00ff00',
      vm_role: true,
      description: 'Access devices',
      comments: 'Operator note',
    })
    expect(
      adapter.mutationFromForm(
        {
          parent: null,
          color: '',
          vm_role: false,
          description: '',
          comments: '',
        },
        false,
      ),
    ).toEqual({
      parent: null,
      color: '',
      vm_role: false,
      description: '',
      comments: '',
    })
  })

  it('omits unchanged DeviceRole fields and emits one normalized dirty PATCH field', () => {
    const adapter = getCoreResourceAdapter('devicerole')
    const dto: DeviceRoleDTO = {
      id: 13,
      url: '/api/dcim/device-roles/13/',
      display: 'Access',
      created: null,
      last_updated: null,
      parent: deviceRole,
      name: 'Access',
      slug: 'access',
      color: '9e9e9e',
      vm_role: true,
      description: 'Access devices',
      comments: 'Operator note',
      device_count: 4,
      _depth: 1,
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})

    const otherParent: CoreReference = {
      id: 14,
      url: '/api/dcim/device-roles/14/',
      display: 'Infrastructure',
    }
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'parent', otherParent) as DeviceRoleForm,
        true,
      ),
    ).toEqual({ parent: 14 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'parent', null) as DeviceRoleForm, true),
    ).toEqual({ parent: null })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'name', 'Edge') as DeviceRoleForm, true),
    ).toEqual({ name: 'Edge' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'slug', 'edge') as DeviceRoleForm, true),
    ).toEqual({ slug: 'edge' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'color', '00ff00') as DeviceRoleForm, true),
    ).toEqual({ color: '00ff00' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'vm_role', false) as DeviceRoleForm, true),
    ).toEqual({ vm_role: false })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'description', '') as DeviceRoleForm, true),
    ).toEqual({ description: '' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'comments', '') as DeviceRoleForm, true),
    ).toEqual({ comments: '' })
  })

  it('preserves DeviceType create omissions, defaults, IDs, and concrete scalar states', () => {
    const adapter = getCoreResourceAdapter('devicetype')

    expect(adapter.mutationFromForm({}, false)).toEqual({})
    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({
      u_height: 1,
      exclude_from_utilization: false,
      is_full_depth: true,
    })
    expect(
      adapter.mutationFromForm(
        {
          manufacturer,
          model: 'Edge',
          slug: 'edge',
          part_number: '',
          u_height: 0,
          exclude_from_utilization: false,
          is_full_depth: false,
          airflow: null,
          description: '',
          comments: '',
        },
        false,
      ),
    ).toEqual({
      manufacturer: 2,
      model: 'Edge',
      slug: 'edge',
      part_number: '',
      u_height: 0,
      exclude_from_utilization: false,
      is_full_depth: false,
      airflow: null,
      description: '',
      comments: '',
    })
    expect(
      adapter.mutationFromForm(
        {
          manufacturer: 2,
          model: 'Core',
          slug: 'core',
          airflow: 'front-to-rear',
        },
        false,
      ),
    ).toEqual({ manufacturer: 2, model: 'Core', slug: 'core', airflow: 'front-to-rear' })

    const formWithResponseFields = {
      manufacturer,
      model: 'Owned fields only',
      slug: 'owned-fields-only',
      id: 99,
      url: '/api/dcim/device-types/99/',
      display: 'Must not leak',
      device_count: 12,
      interface_template_count: 4,
    } as DeviceTypeForm
    expect(adapter.mutationFromForm(formWithResponseFields, false)).toEqual({
      manufacturer: 2,
      model: 'Owned fields only',
      slug: 'owned-fields-only',
    })
  })

  it('keeps a private normalized DeviceType baseline and emits only dirty PATCH fields', () => {
    const adapter = getCoreResourceAdapter('devicetype')
    const dto: DeviceTypeDTO = {
      id: 6,
      url: '/api/dcim/device-types/6/',
      display: 'Acme Edge',
      created: null,
      last_updated: null,
      manufacturer,
      model: 'Edge',
      slug: 'edge',
      part_number: 'PN-1',
      u_height: 1,
      exclude_from_utilization: true,
      is_full_depth: true,
      airflow: { value: 'front-to-rear', label: 'Front to rear' },
      description: 'Edge appliance',
      comments: 'Operator note',
      device_count: 3,
      interface_template_count: 8,
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(Object.fromEntries(Object.entries(hydrated))).toEqual({
      manufacturer,
      model: 'Edge',
      slug: 'edge',
      part_number: 'PN-1',
      u_height: 1,
      exclude_from_utilization: true,
      is_full_depth: true,
      airflow: 'front-to-rear',
      description: 'Edge appliance',
      comments: 'Operator note',
    })
    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'manufacturer', { ...manufacturer }) as DeviceTypeForm,
        true,
      ),
    ).toEqual({})

    const otherManufacturer: CoreReference = {
      id: 12,
      url: '/api/dcim/manufacturers/12/',
      display: 'Globex',
    }
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'manufacturer', otherManufacturer) as DeviceTypeForm,
        true,
      ),
    ).toEqual({ manufacturer: 12 })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'u_height', 0) as DeviceTypeForm, true),
    ).toEqual({ u_height: 0 })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'exclude_from_utilization', false) as DeviceTypeForm,
        true,
      ),
    ).toEqual({ exclude_from_utilization: false })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'is_full_depth', false) as DeviceTypeForm,
        true,
      ),
    ).toEqual({ is_full_depth: false })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'airflow', null) as DeviceTypeForm, true),
    ).toEqual({ airflow: null })
    for (const field of ['part_number', 'description', 'comments'] as const) {
      expect(
        adapter.mutationFromForm(withFormField(hydrated, field, '') as DeviceTypeForm, true),
      ).toEqual({ [field]: '' })
    }
  })

  it('preserves InterfaceTemplate create omissions, defaults, IDs, and concrete scalar states', () => {
    const adapter = getCoreResourceAdapter('interfacetemplate')

    expect(adapter.mutationFromForm({}, false)).toEqual({})
    expect(adapter.mutationFromForm(adapter.emptyForm(), false)).toEqual({
      enabled: true,
      mgmt_only: false,
    })
    expect(
      adapter.mutationFromForm(
        {
          device_type: deviceType,
          name: 'xe-0/0/0',
          label: '',
          type: '10gbase-x-sfpp',
          enabled: false,
          mgmt_only: false,
          description: '',
        },
        false,
      ),
    ).toEqual({
      device_type: 6,
      name: 'xe-0/0/0',
      label: '',
      type: '10gbase-x-sfpp',
      enabled: false,
      mgmt_only: false,
      description: '',
    })
    expect(
      adapter.mutationFromForm({ device_type: 6, name: 'mgmt0', type: '1000base-t' }, false),
    ).toEqual({ device_type: 6, name: 'mgmt0', type: '1000base-t' })

    const formWithResponseFields = {
      device_type: deviceType,
      name: 'Owned fields only',
      id: 99,
      url: '/api/dcim/interface-templates/99/',
      display: 'Must not leak',
      created: null,
      last_updated: null,
    } as InterfaceTemplateForm
    expect(adapter.mutationFromForm(formWithResponseFields, false)).toEqual({
      device_type: 6,
      name: 'Owned fields only',
    })
  })

  it('keeps a private normalized InterfaceTemplate baseline and emits only dirty PATCH fields', () => {
    const adapter = getCoreResourceAdapter('interfacetemplate')
    const dto: InterfaceTemplateDTO = {
      id: 15,
      url: '/api/dcim/interface-templates/15/',
      display: 'xe-0/0/0',
      created: null,
      last_updated: null,
      device_type: deviceType,
      name: 'xe-0/0/0',
      label: 'Uplink',
      type: { value: '10gbase-x-sfpp', label: 'SFP+ (10GE)' },
      enabled: true,
      mgmt_only: true,
      description: 'Provider uplink',
    }
    const hydrated = adapter.formFromDTO(dto)

    expect(Object.fromEntries(Object.entries(hydrated))).toEqual({
      device_type: deviceType,
      name: 'xe-0/0/0',
      label: 'Uplink',
      type: '10gbase-x-sfpp',
      enabled: true,
      mgmt_only: true,
      description: 'Provider uplink',
    })
    expect(adapter.mutationFromForm({ ...hydrated }, true)).toEqual({})
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'device_type', { ...deviceType }) as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({})

    const otherDeviceType: CoreReference = {
      id: 12,
      url: '/api/dcim/device-types/12/',
      display: 'Acme Core',
    }
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'device_type', otherDeviceType) as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ device_type: 12 })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'device_type', null) as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ device_type: null })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'name', 'xe-0/0/1') as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ name: 'xe-0/0/1' })
    expect(
      adapter.mutationFromForm(withFormField(hydrated, 'label', '') as InterfaceTemplateForm, true),
    ).toEqual({ label: '' })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'type', '1000base-t') as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ type: '1000base-t' })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'enabled', false) as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ enabled: false })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'mgmt_only', false) as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ mgmt_only: false })
    expect(
      adapter.mutationFromForm(
        withFormField(hydrated, 'description', '') as InterfaceTemplateForm,
        true,
      ),
    ).toEqual({ description: '' })
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
