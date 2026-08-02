import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import DynamicField from './DynamicField.vue'
import NumberField from './NumberField.vue'
import DateField from './DateField.vue'
import DateTimeField from './DateTimeField.vue'
import JsonField from './JsonField.vue'
import MarkdownField from './MarkdownField.vue'
import TextField from './TextField.vue'
import SlugField from './SlugField.vue'
import SelectField from './SelectField.vue'
import BooleanField from './BooleanField.vue'
import ApiSelectField from './ApiSelectField.vue'
import TagInputField from './TagInputField.vue'
import TextareaField from './TextareaField.vue'
import type { FormFieldDef } from '@/types'

function makeField(type: string): FormFieldDef {
  return { key: 'test', label: 'Test', type: type as FormFieldDef['type'] }
}

describe('DynamicField routing', () => {
  it('routes text to TextField', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('text'), modelValue: '' } })
    expect(wrapper.findComponent(TextField).exists()).toBe(true)
  })

  it('routes slug to SlugField', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('slug'), modelValue: '' } })
    expect(wrapper.findComponent(SlugField).exists()).toBe(true)
  })

  it('routes number to NumberField (not TextField fallback)', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('number'), modelValue: 0 } })
    expect(wrapper.findComponent(NumberField).exists()).toBe(true)
    expect(wrapper.findComponent(TextField).exists()).toBe(false)
  })

  it('forwards rack-position range and half-unit step metadata', () => {
    const field: FormFieldDef = {
      key: 'position',
      label: 'Position',
      type: 'number',
      min: 0,
      max: 100,
      step: 0.5,
    }
    const wrapper = mount(DynamicField, { props: { field, modelValue: 1.5 } })
    const input = wrapper.get('input')
    expect(input.attributes('min')).toBe('0')
    expect(input.attributes('max')).toBe('100')
    expect(input.attributes('step')).toBe('0.5')
  })

  it('routes boolean to BooleanField', () => {
    const wrapper = mount(DynamicField, {
      props: { field: makeField('boolean'), modelValue: false },
    })
    expect(wrapper.findComponent(BooleanField).exists()).toBe(true)
  })

  it('routes select to SelectField', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('select'), modelValue: '' } })
    expect(wrapper.findComponent(SelectField).exists()).toBe(true)
  })

  it('routes api-select to ApiSelectField', () => {
    const wrapper = mount(DynamicField, {
      props: { field: makeField('api-select'), modelValue: null },
    })
    expect(wrapper.findComponent(ApiSelectField).exists()).toBe(true)
  })

  it('routes tag to TagInputField', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('tag'), modelValue: [] } })
    expect(wrapper.findComponent(TagInputField).exists()).toBe(true)
  })

  it('routes markdown to MarkdownField (not TextareaField)', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('markdown'), modelValue: '' } })
    expect(wrapper.findComponent(MarkdownField).exists()).toBe(true)
    expect(wrapper.findComponent(TextareaField).exists()).toBe(false)
  })

  it('routes json to JsonField (not TextareaField)', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('json'), modelValue: null } })
    expect(wrapper.findComponent(JsonField).exists()).toBe(true)
    expect(wrapper.findComponent(TextareaField).exists()).toBe(false)
  })

  it('routes date to DateField (not TextField)', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('date'), modelValue: null } })
    expect(wrapper.findComponent(DateField).exists()).toBe(true)
    expect(wrapper.findComponent(TextField).exists()).toBe(false)
  })

  it('routes datetime to DateTimeField (not TextField)', () => {
    const wrapper = mount(DynamicField, {
      props: { field: makeField('datetime'), modelValue: null },
    })
    expect(wrapper.findComponent(DateTimeField).exists()).toBe(true)
    expect(wrapper.findComponent(TextField).exists()).toBe(false)
  })

  it('routes textarea to TextareaField', () => {
    const wrapper = mount(DynamicField, { props: { field: makeField('textarea'), modelValue: '' } })
    expect(wrapper.findComponent(TextareaField).exists()).toBe(true)
  })
})
