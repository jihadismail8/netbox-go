import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SlugField from './SlugField.vue'

describe('SlugField', () => {
  it('reactively generates a slug while the create value remains automatic', async () => {
    const wrapper = mount(SlugField, { props: { modelValue: '', sourceValue: '' } })

    await wrapper.setProps({ sourceValue: 'Edge Site One' })
    expect(wrapper.get('input').element.value).toBe('edge-site-one')
    await wrapper.setProps({ modelValue: 'edge-site-one', sourceValue: 'Edge Site Two' })
    expect(wrapper.get('input').element.value).toBe('edge-site-two')
  })

  it('preserves a custom slug loaded asynchronously for edit', async () => {
    const wrapper = mount(SlugField, { props: { modelValue: '', sourceValue: '' } })

    await wrapper.setProps({ modelValue: 'custom-edge', sourceValue: 'Edge Site' })
    expect(wrapper.get('input').element.value).toBe('custom-edge')
  })

  it('renders backend slug validation errors beside the field', () => {
    const wrapper = mount(SlugField, {
      props: { modelValue: 'Invalid Slug', error: 'Enter a valid lowercase slug.' },
    })
    expect(wrapper.text()).toContain('Enter a valid lowercase slug.')
  })
})
