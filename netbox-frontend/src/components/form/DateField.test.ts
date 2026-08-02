import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import DateField from './DateField.vue'

describe('DateField', () => {
  it('renders a date input', () => {
    const wrapper = mount(DateField, { props: { modelValue: '2026-06-27' } })
    expect(wrapper.find('input').attributes('type')).toBe('date')
    expect(wrapper.find('input').element.value).toBe('2026-06-27')
  })

  it('emits ISO date string on input', async () => {
    const wrapper = mount(DateField, { props: { modelValue: null } })
    await wrapper.find('input').setValue('2026-01-15')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['2026-01-15'])
  })

  it('emits null when cleared', async () => {
    const wrapper = mount(DateField, { props: { modelValue: '2026-06-27' } })
    await wrapper.find('input').setValue('')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
  })

  it('shows label and required indicator', () => {
    const wrapper = mount(DateField, {
      props: { label: 'Install Date', required: true },
    })
    expect(wrapper.text()).toContain('Install Date')
    expect(wrapper.text()).toContain('*')
  })
})
