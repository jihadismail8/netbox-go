import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import JsonField from './JsonField.vue'

describe('JsonField', () => {
  it('initializes with empty modelValue', () => {
    const wrapper = mount(JsonField, { props: { modelValue: null } })
    expect(wrapper.find('textarea').element.value).toBe('')
  })

  it('serializes object modelValue to JSON string', () => {
    const wrapper = mount(JsonField, {
      props: { modelValue: { name: 'test', value: 42 } },
    })
    const val = wrapper.find('textarea').element.value
    expect(JSON.parse(val)).toEqual({ name: 'test', value: 42 })
  })

  it('emits parsed object on valid JSON input', async () => {
    const wrapper = mount(JsonField, { props: { modelValue: null } })
    await wrapper.find('textarea').setValue('{"key":"value"}')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([{ key: 'value' }])
  })

  it('emits null when textarea is cleared', async () => {
    const wrapper = mount(JsonField, {
      props: { modelValue: { a: 1 } },
    })
    await wrapper.find('textarea').setValue('')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
  })

  it('does not emit for invalid JSON', async () => {
    const wrapper = mount(JsonField, { props: { modelValue: null } })
    await wrapper.find('textarea').setValue('{invalid json}')
    expect(wrapper.emitted('update:modelValue')).toBeUndefined()
  })
})
