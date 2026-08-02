import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import NumberField from './NumberField.vue'

describe('NumberField', () => {
  it('renders with label and required indicator', () => {
    const wrapper = mount(NumberField, {
      props: { label: 'Count', required: true, modelValue: 5 },
    })
    expect(wrapper.text()).toContain('Count')
    expect(wrapper.text()).toContain('*')
  })

  it('displays the current value', () => {
    const wrapper = mount(NumberField, { props: { modelValue: 42 } })
    expect(wrapper.find('input').element.value).toBe('42')
  })

  it('emits number (not string) on input', async () => {
    const wrapper = mount(NumberField, { props: { modelValue: null } })
    await wrapper.find('input').setValue('123')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([123])
    expect(typeof wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe('number')
  })

  it('emits null when input is cleared', async () => {
    const wrapper = mount(NumberField, { props: { modelValue: 5 } })
    await wrapper.find('input').setValue('')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
  })

  it('emits null for invalid numeric input (browser clears number input)', async () => {
    const wrapper = mount(NumberField, { props: { modelValue: null } })
    await wrapper.find('input').setValue('abc')
    // Browsers/jSDOM clear invalid number input to empty, so null is emitted
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
  })

  it('passes min/max/step attributes to input', () => {
    const wrapper = mount(NumberField, {
      props: { modelValue: 0, min: 1, max: 100, step: 5 },
    })
    const input = wrapper.find('input')
    expect(input.attributes('min')).toBe('1')
    expect(input.attributes('max')).toBe('100')
    expect(input.attributes('step')).toBe('5')
  })
})
