import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import type { FormFieldDef } from '@/types'
import DynamicForm from './DynamicForm.vue'
import ApiSelectField from './ApiSelectField.vue'

describe('DynamicForm edit hydration', () => {
  it('updates its controls when edit data arrives asynchronously', async () => {
    const fields: FormFieldDef[] = [{ key: 'name', label: 'Name', type: 'text' }]
    const wrapper = mount(DynamicForm, { props: { fields, modelValue: {} } })

    expect((wrapper.get('input').element as HTMLInputElement).value).toBe('')
    await wrapper.setProps({ modelValue: { name: 'loaded-site' } })
    expect((wrapper.get('input').element as HTMLInputElement).value).toBe('loaded-site')
  })

  it('shows the display value of a hydrated nested relation', () => {
    const wrapper = mount(ApiSelectField, {
      props: {
        modelValue: { id: 12, url: '/api/dcim/racks/12/', display: 'Rack A' },
        label: 'Rack',
      },
    })

    expect((wrapper.get('input').element as HTMLInputElement).value).toBe('Rack A')
  })

  it('keeps non-field validation and conflict errors visible', () => {
    const wrapper = mount(DynamicForm, {
      props: {
        fields: [],
        errors: { non_field_errors: ['The requested placement conflicts with another Device.'] },
      },
    })

    expect(wrapper.get('[role="alert"]').text()).toContain('placement conflicts')
  })

  it('locks immutable relationships only while editing', async () => {
    const fields: FormFieldDef[] = [
      {
        key: 'device',
        label: 'Device',
        type: 'api-select',
        relationResource: 'device',
        immutableOnEdit: true,
      },
    ]
    const wrapper = mount(DynamicForm, {
      props: {
        fields,
        editing: true,
        modelValue: {
          device: { id: 1, url: '/api/dcim/devices/1/', display: 'edge-1' },
        },
      },
    })

    expect((wrapper.get('input').element as HTMLInputElement).disabled).toBe(true)
    await wrapper.setProps({ editing: false })
    expect((wrapper.get('input').element as HTMLInputElement).disabled).toBe(false)
  })

  it('locks fields inherited from a selected relationship and unlocks them when cleared', async () => {
    const fields: FormFieldDef[] = [
      { key: 'rack_type', label: 'Rack Type', type: 'api-select', relationResource: 'racktype' },
      {
        key: 'width',
        label: 'Width',
        type: 'select',
        options: [{ value: 19, label: '19 inches' }],
        disabledWhenFieldTruthy: 'rack_type',
      },
    ]
    const wrapper = mount(DynamicForm, {
      props: {
        fields,
        modelValue: {
          rack_type: {
            id: 2,
            url: '/api/dcim/rack-types/2/',
            display: 'Standard rack',
          },
          width: 19,
        },
      },
    })

    expect((wrapper.get('select').element as HTMLSelectElement).disabled).toBe(true)

    await wrapper.getComponent(ApiSelectField).vm.$emit('update:modelValue', null)
    expect((wrapper.get('select').element as HTMLSelectElement).disabled).toBe(false)
  })

  it('scopes Racks by Site and clears placement when the Site changes', async () => {
    const fields: FormFieldDef[] = [
      { key: 'site', label: 'Site', type: 'api-select', relationResource: 'site' },
      {
        key: 'rack',
        label: 'Rack',
        type: 'api-select',
        relationResource: 'rack',
        relationFilterFields: { site_id: 'site' },
        clearWhenFieldChanges: ['site'],
      },
      {
        key: 'position',
        label: 'Position',
        type: 'number',
        disabledUnlessFieldTruthy: 'rack',
        clearWhenFieldChanges: ['rack'],
      },
      {
        key: 'face',
        label: 'Face',
        type: 'select',
        options: [{ value: 'front', label: 'Front' }],
        disabledUnlessFieldTruthy: 'rack',
        requiredWhenFieldTruthy: 'position',
        clearWhenFieldChanges: ['rack'],
      },
    ]
    const wrapper = mount(DynamicForm, {
      props: {
        fields,
        modelValue: {
          site: { id: 1, url: '/api/dcim/sites/1/', display: 'Site 1' },
          rack: { id: 2, url: '/api/dcim/racks/2/', display: 'Rack 2' },
          position: 4,
          face: 'front',
        },
      },
    })
    const relations = wrapper.findAllComponents(ApiSelectField)

    expect(relations[1].props('relationFilters')).toEqual({ site_id: 1 })
    expect((wrapper.get('select').element as HTMLSelectElement).required).toBe(true)

    relations[0].vm.$emit('update:modelValue', 3)
    await wrapper.vm.$nextTick()

    expect(wrapper.emitted('update:modelValue')?.at(-1)?.[0]).toEqual({
      site: 3,
      rack: null,
      position: null,
      face: null,
    })
    expect(relations[1].props('relationFilters')).toEqual({ site_id: 3 })
    expect((wrapper.get('input[type="number"]').element as HTMLInputElement).disabled).toBe(true)
    expect((wrapper.get('select').element as HTMLSelectElement).disabled).toBe(true)
  })
})
