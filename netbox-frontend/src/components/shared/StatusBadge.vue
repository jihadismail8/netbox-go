<template>
  <span
    class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium"
    :class="colorClass"
  >
    {{ displayLabel }}
  </span>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { STATUS_COLORS } from '@/config/app'

const props = defineProps<{ status?: string; label?: string }>()

const colorMap: Record<string, string> = {
  success: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  danger: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  secondary: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
}

const colorKey = computed(() =>
  props.status ? STATUS_COLORS[props.status] || 'secondary' : 'secondary',
)
const colorClass = computed(() => colorMap[colorKey.value] || colorMap.secondary)
const displayLabel = computed(
  () =>
    props.label ||
    (props.status ? props.status.charAt(0).toUpperCase() + props.status.slice(1) : ''),
)
</script>
