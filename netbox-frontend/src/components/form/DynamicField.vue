<template>
  <component
    :is="componentName"
    :model-value="modelValue"
    :label="field.label"
    :required="required ?? field.required"
    :help-text="field.helpText"
    :placeholder="field.placeholder"
    :options="field.options || []"
    :multiple="field.multiple"
    :min="field.min"
    :max="field.max"
    :step="field.step"
    :disabled="disabled"
    :source-value="field.type === 'slug' ? formFieldValue(field.slugSource || 'name') : undefined"
    :relation-resource="field.relationResource"
    :relation-filters="relationFilters"
    :error="error"
    @update:model-value="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import type { FormFieldDef } from '@/types'
import {
  formField,
  isCoreReference,
  type CoreFilterState,
  type CoreResourceForm,
} from '@/features/core/resources'
import TextField from './TextField.vue'
import SlugField from './SlugField.vue'
import SelectField from './SelectField.vue'
import BooleanField from './BooleanField.vue'
import TextareaField from './TextareaField.vue'
import ApiSelectField from './ApiSelectField.vue'
import TagInputField from './TagInputField.vue'
import NumberField from './NumberField.vue'
import MarkdownField from './MarkdownField.vue'
import JsonField from './JsonField.vue'
import DateField from './DateField.vue'
import DateTimeField from './DateTimeField.vue'

const props = defineProps<{
  field: FormFieldDef
  modelValue: unknown
  formData?: CoreResourceForm
  disabled?: boolean
  required?: boolean
  error?: string
}>()

defineEmits<{ 'update:modelValue': [value: unknown] }>()

const componentName = computed(() => {
  const map: { [type: string]: Component } = {
    text: TextField,
    slug: SlugField,
    number: NumberField,
    boolean: BooleanField,
    select: SelectField,
    'api-select': ApiSelectField,
    tag: TagInputField,
    markdown: MarkdownField,
    json: JsonField,
    textarea: TextareaField,
    date: DateField,
    datetime: DateTimeField,
    csv: TextareaField,
  }
  return map[props.field.type] || TextField
})

const relationFilters = computed<CoreFilterState | undefined>(() => {
  if (!props.field.relationFilterFields) return undefined
  const result: CoreFilterState = {}
  for (const [filter, sourceField] of Object.entries(props.field.relationFilterFields)) {
    if (!sourceField) continue
    const source = props.formData ? formField(props.formData, sourceField) : undefined
    const value = isCoreReference(source) ? source.id : source
    if (value !== null && value !== undefined && value !== '') Reflect.set(result, filter, value)
  }
  return result
})

function formFieldValue(key: string): unknown {
  return props.formData ? formField(props.formData, key) : undefined
}
</script>
