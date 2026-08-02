<template>
  <div>
    <label v-if="label" class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <input
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :required="required"
      class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-800"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    modelValue?: string | number
    label?: string
    type?: string
    placeholder?: string
    required?: boolean
    helpText?: string
    error?: string
  }>(),
  { type: 'text', required: false },
)

defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
