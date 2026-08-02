<template>
  <div>
    <label v-if="label" class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <select
      :value="modelValue"
      :required="required"
      :disabled="disabled"
      :multiple="multiple"
      class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-500 dark:border-gray-600 dark:bg-gray-800 dark:disabled:bg-gray-900"
      @change="
        $emit('update:modelValue', multiple ? getMultiSelect($event) : getSingleSelect($event))
      "
    >
      <option v-if="!multiple" value="" :disabled="required">
        {{ required ? '— Select —' : '—' }}
      </option>
      <option v-for="opt in options" :key="String(opt.value)" :value="opt.value">
        {{ opt.label }}
      </option>
    </select>
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import type { ChoiceOption } from '@/types'

const props = withDefaults(
  defineProps<{
    modelValue?: string | number | boolean | Array<string | number | boolean> | null
    label?: string
    options: ChoiceOption[]
    required?: boolean
    multiple?: boolean
    helpText?: string
    error?: string
    disabled?: boolean
  }>(),
  { required: false, multiple: false },
)

defineEmits<{
  'update:modelValue': [value: string | number | boolean | Array<string | number | boolean> | null]
}>()

function getSingleSelect(e: Event): string | number | boolean | null {
  const value = (e.target as HTMLSelectElement).value
  if (value === '') return null
  return props.options.find((option) => String(option.value) === value)?.value ?? value
}

function getMultiSelect(e: Event): Array<string | number | boolean> {
  return Array.from((e.target as HTMLSelectElement).selectedOptions).map(
    (option) =>
      props.options.find((candidate) => String(candidate.value) === option.value)?.value ??
      option.value,
  )
}
</script>
