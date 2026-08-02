import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ApiSelectField from './ApiSelectField.vue'

const mocks = vi.hoisted(() => ({ searchResourceOptions: vi.fn() }))

vi.mock('@/features/core/api', () => ({
  searchResourceOptions: mocks.searchResourceOptions,
}))

describe('ApiSelectField', () => {
  beforeEach(() => {
    mocks.searchResourceOptions
      .mockReset()
      .mockResolvedValue([{ id: 21, url: '/api/dcim/interfaces/21/', display: 'edge-1: xe-0/0/0' }])
  })

  it('hydrates a nested relationship and can explicitly clear it', async () => {
    const wrapper = mount(ApiSelectField, {
      props: {
        modelValue: {
          id: 21,
          url: '/api/dcim/interfaces/21/',
          display: 'edge-1: xe-0/0/0',
        },
        label: 'Interface',
        relationResource: 'interface',
      },
    })

    expect(wrapper.get('input').element.value).toBe('edge-1: xe-0/0/0')
    await wrapper.get('[aria-label="Clear selection"]').trigger('mousedown')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([null])
    expect(wrapper.get('input').element.value).toBe('')
  })

  it('loads manifest-backed choices and emits the selected ID', async () => {
    const wrapper = mount(ApiSelectField, {
      props: { modelValue: null, label: 'Interface', relationResource: 'interface' },
    })

    await wrapper.get('input').trigger('focus')
    await flushPromises()

    expect(mocks.searchResourceOptions).toHaveBeenCalledWith('interface', '')
    await wrapper.get('div[role="listbox"] button').trigger('mousedown')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([21])
  })

  it('keeps replacement search text after clearing the previous relationship', async () => {
    const wrapper = mount(ApiSelectField, {
      props: {
        modelValue: { id: 21, url: '/api/dcim/interfaces/21/', display: 'old Interface' },
        relationResource: 'interface',
      },
    })

    await wrapper.get('input').setValue('new Interface')
    await wrapper.setProps({ modelValue: null })

    expect(wrapper.get('input').element.value).toBe('new Interface')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([null])
  })

  it('requires choosing a result rather than accepting unmatched text', async () => {
    const wrapper = mount(ApiSelectField, {
      props: {
        required: true,
        relationResource: 'site',
      },
    })
    const input = wrapper.get('input')

    expect((input.element as HTMLInputElement).checkValidity()).toBe(false)

    await input.setValue('not a selected site')
    expect((input.element as HTMLInputElement).validationMessage).toBe(
      'Select an item from the choices.',
    )
  })

  it('passes parent relationship filters to the feature API', async () => {
    const wrapper = mount(ApiSelectField, {
      props: {
        modelValue: null,
        relationResource: 'rack',
        relationFilters: { site_id: 4 },
      },
    })

    await wrapper.get('input').trigger('focus')
    await flushPromises()

    expect(mocks.searchResourceOptions).toHaveBeenCalledWith('rack', '', [], { site_id: 4 })
  })
})
