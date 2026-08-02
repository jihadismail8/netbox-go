<template>
  <div>
    <label v-if="label" class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <input
      type="number"
      :value="modelValue ?? ''"
      :placeholder="placeholder"
      :required="required"
      :disabled="disabled"
      :min="min"
      :max="max"
      :step="step"
      class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-500 dark:border-gray-600 dark:bg-gray-800 dark:disabled:bg-gray-900"
      @input="onInput"
    />
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue?: number | null
    label?: string
    placeholder?: string
    required?: boolean
    helpText?: string
    error?: string
    min?: number
    max?: number
    step?: number
    disabled?: boolean
  }>(),
  { required: false, step: 1 },
)

const emit = defineEmits<{ 'update:modelValue': [value: number | null] }>()

function onInput(event: Event) {
  const raw = (event.target as HTMLInputElement).value
  if (raw === '') {
    emit('update:modelValue', null)
    return
  }
  const num = Number(raw)
  if (!isNaN(num)) {
    emit('update:modelValue', num)
  }
}
</script>
