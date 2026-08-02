import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SelectField from './SelectField.vue'

describe('SelectField wire values', () => {
  it('preserves numeric option types for the REST payload', async () => {
    const wrapper = mount(SelectField, {
      props: {
        modelValue: null,
        options: [
          { value: 10, label: '10 inches' },
          { value: 19, label: '19 inches' },
        ],
      },
    })

    await wrapper.get('select').setValue('19')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([19])
  })

  it('emits null when an optional selection is cleared', async () => {
    const wrapper = mount(SelectField, {
      props: { modelValue: 'active', options: [{ value: 'active', label: 'Active' }] },
    })

    await wrapper.get('select').setValue('')
    expect(wrapper.emitted('update:modelValue')?.at(-1)).toEqual([null])
  })

  it('does not visually invent a value for an unset required choice', () => {
    const wrapper = mount(SelectField, {
      props: {
        required: true,
        options: [
          { value: 'virtual', label: 'Virtual' },
          { value: 'bridge', label: 'Bridge' },
        ],
      },
    })

    const select = wrapper.get('select').element as HTMLSelectElement
    expect(select.value).toBe('')
    expect(select.checkValidity()).toBe(false)
    expect(wrapper.get('option').attributes('disabled')).toBeDefined()
  })
})
