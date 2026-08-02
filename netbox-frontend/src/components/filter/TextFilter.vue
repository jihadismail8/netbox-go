<template>
  <input
    type="text"
    :value="modelValue ?? ''"
    :placeholder="field.placeholder || 'Search...'"
    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
    @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>

<script setup lang="ts">
import type { FilterField } from '@/types'
defineProps<{ field: FilterField; modelValue?: unknown }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>
