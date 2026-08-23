import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'
import { describe, expect, expectTypeOf, it } from 'vitest'
import DynamicField from '@/components/form/DynamicField.vue'
import DynamicForm from '@/components/form/DynamicForm.vue'
import { getCoreResourceAdapter } from '@/features/core/adapters'
import type { IPAddressDTO, IPAddressForm, SiteDTO, SiteForm } from '@/features/core/resources'
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
