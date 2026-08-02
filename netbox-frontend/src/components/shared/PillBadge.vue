<template>
  <span
    class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium"
    :class="colorClass"
  >
    <component :is="icon" v-if="icon" :size="11" />
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Component } from 'vue'

const props = withDefaults(
  defineProps<{
    label?: string
    color?: 'primary' | 'success' | 'info' | 'warning' | 'danger' | 'secondary'
    icon?: Component
  }>(),
  {
    color: 'secondary',
  },
)

const colorMap: Record<string, string> = {
  primary: 'bg-primary/10 text-primary dark:bg-primary/20 dark:text-primary',
  success: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
  info: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
  warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
  danger: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  secondary: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
}

const colorClass = computed(() => colorMap[props.color] || colorMap.secondary)
</script>
