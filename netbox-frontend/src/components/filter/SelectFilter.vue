<template>
  <select
    :value="modelValue ?? ''"
    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
    @change="$emit('update:modelValue', ($event.target as HTMLSelectElement).value || null)"
  >
    <option value="">— Any —</option>
    <option v-for="opt in field.options" :key="String(opt.value)" :value="opt.value">
      {{ opt.label }}
    </option>
  </select>
</template>

<script setup lang="ts">
import type { FilterField } from '@/types'
defineProps<{ field: FilterField; modelValue?: unknown }>()
defineEmits<{ 'update:modelValue': [value: string | null] }>()
</script>
