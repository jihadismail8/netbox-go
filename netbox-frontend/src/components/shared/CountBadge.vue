<template>
  <span
    v-if="count > 0"
    class="inline-flex min-w-[1.25rem] items-center justify-center rounded-full px-1.5 py-0.5 text-xs font-semibold leading-none"
    :class="colorClass"
  >
    {{ count > 999 ? '999+' : count }}
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    count: number
    color?: 'primary' | 'danger' | 'gray'
  }>(),
  {
    color: 'gray',
  },
)

const colorMap: Record<string, string> = {
  primary: 'bg-primary text-white',
  danger: 'bg-red-500 text-white',
  gray: 'bg-gray-200 text-gray-700 dark:bg-gray-600 dark:text-gray-200',
}

const colorClass = computed(() => colorMap[props.color] || colorMap.gray)
</script>
