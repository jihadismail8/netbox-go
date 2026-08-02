<template>
  <div>
    <label
      class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"
      :class="disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'"
    >
      <input
        type="checkbox"
        :checked="modelValue"
        :disabled="disabled"
        class="border-gray-300 rounded text-primary focus:ring-primary"
        @change="$emit('update:modelValue', ($event.target as HTMLInputElement).checked)"
      />
      {{ label }}
    </label>
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  modelValue?: boolean
  label?: string
  disabled?: boolean
  helpText?: string
  error?: string
}>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()
</script>
