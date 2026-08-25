import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { describe, expect, expectTypeOf, it } from 'vitest'
import DynamicField from '@/components/form/DynamicField.vue'
import DynamicForm from '@/components/form/DynamicForm.vue'
import { getCoreResourceAdapter } from '@/features/core/adapters'
import type {
  DeviceRoleDTO,
  DeviceRoleForm,
  DeviceTypeDTO,
  DeviceTypeForm,
  IPAddressDTO,
  IPAddressForm,
  ManufacturerDTO,
  ManufacturerForm,
  RackRoleDTO,
  RackRoleForm,
  RackTypeDTO,
  RackTypeForm,
  SiteDTO,
  SiteForm,
} from '@/features/core/resources'
import { CORE_PROFILE_MODELS } from './core-profile'

describe('Site core profile form contract', () => {
  it('requires Name, Slug, and Status while keeping Site text fields optional', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'site')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(fields.get('name')?.required).toBe(true)
    expect(fields.get('slug')?.required).toBe(true)
    expect(fields.get('status')).toEqual(
      expect.objectContaining({ required: true, default: 'active' }),
    )
    for (const optional of ['facility', 'description', 'comments']) {
      const field = fields.get(optional)
      expect(field, optional).toBeDefined()
      expect(field?.required, optional).not.toBe(true)
    }
  })

  it('routes field-scoped REST validation to the matching Site field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'site')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: { name: 'MOW', slug: '', status: 'active' },
        errors: {
          slug: ['This field may not be blank.'],
        },
      },
    })

    const slug = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'slug')
    const status = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'status')
    expect(slug).toBeDefined()
    expect(slug?.props('error')).toBe('This field may not be blank.')
    expect(status).toBeDefined()
    expect(status?.props('error')).toBeUndefined()
  })

  it('types nullable Site response timestamps exactly', () => {
    expectTypeOf<SiteDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<SiteDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the Site scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'site')!
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
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const facility = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'facility')
    expect(facility).toBeDefined()
    facility!.vm.$emit('update:modelValue', '')
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as SiteForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ facility: '' })
  })
})

describe('Manufacturer core profile form contract', () => {
  it('requires Name and Slug while keeping Description optional', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'manufacturer')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(fields.get('name')?.required).toBe(true)
    expect(fields.get('slug')?.required).toBe(true)
    expect(fields.get('description')).toBeDefined()
    expect(fields.get('description')?.required).not.toBe(true)
  })

  it('routes field-scoped REST validation to the matching Manufacturer field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'manufacturer')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: { name: 'Acme', slug: '' },
        errors: {
          slug: ['This field may not be blank.'],
        },
      },
    })

    const slug = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'slug')
    const description = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'description')
    expect(slug).toBeDefined()
    expect(slug?.props('error')).toBe('This field may not be blank.')
    expect(description).toBeDefined()
    expect(description?.props('error')).toBeUndefined()
  })

  it('types nullable Manufacturer response timestamps exactly', () => {
    expectTypeOf<ManufacturerDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<ManufacturerDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the Manufacturer scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'manufacturer')!
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
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const description = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'description')
    expect(description).toBeDefined()
    description!.vm.$emit('update:modelValue', '')
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as ManufacturerForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ description: '' })
  })
})

describe('RackRole core profile form contract', () => {
  it('requires Name and Slug while defaulting optional Color and keeping Description optional', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'rackrole')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(fields.get('name')?.required).toBe(true)
    expect(fields.get('slug')?.required).toBe(true)
    expect(fields.get('color')).toEqual(expect.objectContaining({ default: '9e9e9e' }))
    expect(fields.get('color')?.required).not.toBe(true)
    expect(fields.get('description')).toBeDefined()
    expect(fields.get('description')?.required).not.toBe(true)
  })

  it('routes field-scoped REST validation to the matching RackRole field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'rackrole')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: { name: 'Core', slug: 'core', color: '' },
        errors: {
          color: ['This field may not be blank.'],
        },
      },
    })

    const color = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'color')
    const slug = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'slug')
    expect(color).toBeDefined()
    expect(color?.props('error')).toBe('This field may not be blank.')
    expect(slug).toBeDefined()
    expect(slug?.props('error')).toBeUndefined()
  })

  it('types nullable RackRole response timestamps exactly', () => {
    expectTypeOf<RackRoleDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<RackRoleDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the RackRole scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'rackrole')!
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
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const color = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'color')
    expect(color).toBeDefined()
    color!.vm.$emit('update:modelValue', '00ff00')
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as RackRoleForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ color: '00ff00' })
  })
})

describe('RackType core profile form contract', () => {
  it('pins required fields, create defaults, choices, and numeric ranges', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'racktype')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    for (const required of ['manufacturer', 'model', 'slug', 'form_factor']) {
      expect(fields.get(required)?.required, required).toBe(true)
    }
    expect(fields.get('form_factor')?.options?.map((option) => option.value)).toEqual([
      '2-post-frame',
      '4-post-frame',
      '4-post-cabinet',
      'wall-frame',
      'wall-frame-vertical',
      'wall-cabinet',
      'wall-cabinet-vertical',
    ])
    expect(fields.get('width')).toEqual(
      expect.objectContaining({ default: 19, options: expect.any(Array) }),
    )
    expect(fields.get('width')?.options?.map((option) => option.value)).toEqual([10, 19, 21, 23])
    expect(fields.get('u_height')).toEqual(
      expect.objectContaining({ default: 42, min: 1, max: 100 }),
    )
    expect(fields.get('starting_unit')).toEqual(
      expect.objectContaining({ default: 1, min: 1, max: 32767 }),
    )
    expect(fields.get('desc_units')).toEqual(expect.objectContaining({ default: false }))
    for (const optional of [
      'width',
      'u_height',
      'starting_unit',
      'desc_units',
      'description',
      'comments',
    ]) {
      expect(fields.get(optional), optional).toBeDefined()
      expect(fields.get(optional)?.required, optional).not.toBe(true)
    }
  })

  it('routes field-scoped REST validation to the matching RackType field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'racktype')!
    const visibleFields = config.fields.filter((field) =>
      ['form_factor', 'starting_unit'].includes(field.key),
    )
    const wrapper = mount(DynamicForm, {
      props: {
        fields: visibleFields,
        modelValue: { form_factor: '4-post-cabinet', starting_unit: 0 },
        errors: {
          starting_unit: ['Ensure this value is greater than or equal to 1.'],
        },
      },
    })

    const startingUnit = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'starting_unit')
    const formFactor = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'form_factor')
    expect(startingUnit).toBeDefined()
    expect(startingUnit?.props('error')).toBe('Ensure this value is greater than or equal to 1.')
    expect(formFactor).toBeDefined()
    expect(formFactor?.props('error')).toBeUndefined()
  })

  it('types nullable RackType response timestamps exactly', () => {
    expectTypeOf<RackTypeDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<RackTypeDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the RackType scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'racktype')!
    const adapter = getCoreResourceAdapter('racktype')
    const dto: RackTypeDTO = {
      id: 4,
      url: '/api/dcim/rack-types/4/',
      display: 'Acme R42',
      created: null,
      last_updated: null,
      manufacturer: {
        id: 2,
        url: '/api/dcim/manufacturers/2/',
        display: 'Acme',
      },
      model: 'R42',
      slug: 'r42',
      form_factor: { value: '4-post-cabinet', label: '4-post cabinet' },
      width: { value: 19, label: '19 inches' },
      u_height: 42,
      starting_unit: 1,
      desc_units: false,
      description: 'Core rack',
      comments: 'Operator note',
    }
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields.filter((field) => field.key === 'description'),
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const description = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'description')
    expect(description).toBeDefined()
    description!.vm.$emit('update:modelValue', '')
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as RackTypeForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ description: '' })
  })
})

describe('DeviceRole core profile form contract', () => {
  it('pins required fields, create defaults, and optional parent and text fields', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicerole')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(fields.get('parent')).toEqual(
      expect.objectContaining({ type: 'api-select', relationResource: 'devicerole' }),
    )
    expect(fields.get('name')?.required).toBe(true)
    expect(fields.get('slug')?.required).toBe(true)
    expect(fields.get('color')).toEqual(expect.objectContaining({ default: '9e9e9e' }))
    expect(fields.get('vm_role')).toEqual(expect.objectContaining({ default: true }))
    for (const optional of ['parent', 'color', 'vm_role', 'description', 'comments']) {
      expect(fields.get(optional), optional).toBeDefined()
      expect(fields.get(optional)?.required, optional).not.toBe(true)
    }
  })

  it('routes field-scoped REST validation to the matching DeviceRole field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicerole')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: { name: 'Access', slug: 'access', color: '' },
        errors: {
          color: ['This field may not be blank.'],
        },
      },
    })

    const color = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'color')
    const slug = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'slug')
    expect(color).toBeDefined()
    expect(color?.props('error')).toBe('This field may not be blank.')
    expect(slug).toBeDefined()
    expect(slug?.props('error')).toBeUndefined()
  })

  it('types nullable DeviceRole response timestamps exactly', () => {
    expectTypeOf<DeviceRoleDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<DeviceRoleDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the DeviceRole scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicerole')!
    const adapter = getCoreResourceAdapter('devicerole')
    const dto: DeviceRoleDTO = {
      id: 13,
      url: '/api/dcim/device-roles/13/',
      display: 'Access',
      created: null,
      last_updated: null,
      parent: {
        id: 7,
        url: '/api/dcim/device-roles/7/',
        display: 'Infrastructure',
      },
      name: 'Access',
      slug: 'access',
      color: '9e9e9e',
      vm_role: true,
      description: 'Access devices',
      comments: 'Operator note',
      device_count: 4,
      _depth: 1,
    }
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields.filter((field) => field.key === 'parent'),
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const parent = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'parent')
    expect(parent).toBeDefined()
    parent!.vm.$emit('update:modelValue', null)
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as DeviceRoleForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ parent: null })
  })
})

describe('DeviceType core profile form contract', () => {
  it('pins owned fields, required inputs, defaults, airflow choices, and height metadata', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicetype')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(config.writableFields).toEqual([
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
    ])
    expect(fields.get('manufacturer')).toEqual(
      expect.objectContaining({
        required: true,
        type: 'api-select',
        relationResource: 'manufacturer',
      }),
    )
    expect(fields.get('model')?.required).toBe(true)
    expect(fields.get('slug')).toEqual(
      expect.objectContaining({ required: true, type: 'slug', slugSource: 'model' }),
    )
    expect(fields.get('u_height')).toEqual(
      expect.objectContaining({ default: 1, min: 0, max: 999.9, step: 0.5 }),
    )
    expect(fields.get('exclude_from_utilization')).toEqual(
      expect.objectContaining({ default: false }),
    )
    expect(fields.get('is_full_depth')).toEqual(expect.objectContaining({ default: true }))
    expect(fields.get('airflow')?.options?.map((option) => option.value)).toEqual([
      'front-to-rear',
      'rear-to-front',
      'left-to-right',
      'right-to-left',
      'side-to-rear',
      'rear-to-side',
      'bottom-to-top',
      'top-to-bottom',
      'passive',
      'mixed',
    ])
    for (const optional of [
      'part_number',
      'u_height',
      'exclude_from_utilization',
      'is_full_depth',
      'airflow',
      'description',
      'comments',
    ]) {
      expect(fields.get(optional), optional).toBeDefined()
      expect(fields.get(optional)?.required, optional).not.toBe(true)
    }
  })

  it('routes field-scoped REST validation to the matching DeviceType field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicetype')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields.filter((field) => ['u_height', 'airflow'].includes(field.key)),
        modelValue: { u_height: 0.25, airflow: null },
        errors: {
          u_height: ['Height must be a whole or half unit.'],
        },
      },
    })

    const height = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'u_height')
    const airflow = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'airflow')
    expect(height).toBeDefined()
    expect(height?.props('error')).toBe('Height must be a whole or half unit.')
    expect(airflow).toBeDefined()
    expect(airflow?.props('error')).toBeUndefined()
  })

  it('types nullable DeviceType response timestamps exactly', () => {
    expectTypeOf<DeviceTypeDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<DeviceTypeDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the DeviceType scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'devicetype')!
    const adapter = getCoreResourceAdapter('devicetype')
    const dto: DeviceTypeDTO = {
      id: 6,
      url: '/api/dcim/device-types/6/',
      display: 'Acme Edge',
      created: null,
      last_updated: null,
      manufacturer: {
        id: 2,
        url: '/api/dcim/manufacturers/2/',
        display: 'Acme',
      },
      model: 'Edge',
      slug: 'edge',
      part_number: 'PN-1',
      u_height: 1,
      exclude_from_utilization: false,
      is_full_depth: true,
      airflow: { value: 'front-to-rear', label: 'Front to rear' },
      description: 'Edge appliance',
      comments: 'Operator note',
      device_count: 3,
      interface_template_count: 8,
    }
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields.filter((field) => field.key === 'airflow'),
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const airflow = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'airflow')
    expect(airflow).toBeDefined()
    airflow!.vm.$emit('update:modelValue', null)
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as DeviceTypeForm
    expect(adapter.mutationFromForm(form, true)).toEqual({ airflow: null })
  })
})

describe('IPAddress core profile form contract', () => {
  it('requires Address and Status while keeping Role and text fields optional', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'ipaddress')!
    const fields = new Map(config.fields.map((field) => [field.key, field]))

    expect(fields.get('address')?.required).toBe(true)
    expect(fields.get('status')).toEqual(
      expect.objectContaining({ required: true, default: 'active' }),
    )
    for (const optional of ['role', 'dns_name', 'description', 'comments']) {
      const field = fields.get(optional)
      expect(field, optional).toBeDefined()
      expect(field?.required, optional).not.toBe(true)
    }
  })

  it('routes field-scoped REST validation to the matching IPAddress field', () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'ipaddress')!
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields,
        modelValue: { address: '', status: 'active' },
        errors: {
          address: ['This field may not be blank.'],
        },
      },
    })

    const address = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'address')
    const status = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'status')
    expect(address).toBeDefined()
    expect(address?.props('error')).toBe('This field may not be blank.')
    expect(status).toBeDefined()
    expect(status?.props('error')).toBeUndefined()
  })

  it('types nullable IPAddress response timestamps exactly', () => {
    expectTypeOf<IPAddressDTO['created']>().toEqualTypeOf<string | null>()
    expectTypeOf<IPAddressDTO['last_updated']>().toEqualTypeOf<string | null>()
  })

  it('preserves the scalar dirty baseline through the real DynamicForm edit flow', async () => {
    const config = CORE_PROFILE_MODELS.find((model) => model.model === 'ipaddress')!
    const scalarFields = new Set([
      'address',
      'status',
      'role',
      'dns_name',
      'description',
      'comments',
    ])
    const adapter = getCoreResourceAdapter('ipaddress')
    const dto: IPAddressDTO = {
      id: 17,
      url: '/api/ipam/ip-addresses/17/',
      display: '192.0.2.17/24',
      created: '2026-08-19T10:00:00Z',
      last_updated: '2026-08-19T10:00:00Z',
      address: '192.0.2.17/24',
      vrf: { id: 10, url: '/api/ipam/vrfs/10/', display: 'Production' },
      status: { value: 'active', label: 'Active' },
      role: { value: 'loopback', label: 'Loopback' },
      dns_name: 'edge.example',
      description: 'Existing description',
      comments: 'Existing comment',
      assigned_object_type: 'dcim.interface',
      assigned_object_id: 11,
      family: { value: 4, label: 'IPv4' },
      assigned_object: {
        id: 11,
        url: '/api/dcim/interfaces/11/',
        display: 'edge-01 Ethernet1',
      },
    }
    const wrapper = mount(DynamicForm, {
      props: {
        fields: config.fields.filter((field) => scalarFields.has(field.key)),
        modelValue: adapter.formFromDTO(dto),
        editing: true,
      },
    })

    const role = wrapper
      .findAllComponents(DynamicField)
      .find((field) => field.props('field').key === 'role')
    expect(role).toBeDefined()
    role!.vm.$emit('update:modelValue', null)
    await nextTick()

    const form = wrapper.emitted('update:modelValue')?.at(-1)?.[0] as IPAddressForm
    expect(adapter.mutationFromForm(form, true)).toEqual({
      vrf: 10,
      assigned_object_type: 'dcim.interface',
      assigned_object_id: 11,
      role: null,
    })
  })
})
