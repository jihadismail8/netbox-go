<template>
  <div
    class="p-6 bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800"
  >
    <h3 class="mb-4 text-base font-semibold text-gray-900 dark:text-white">{{ title }}</h3>
    <div v-if="data.length === 0" class="py-8 text-center text-sm text-gray-400">No data</div>
    <div v-else class="space-y-2.5">
      <div v-for="(item, idx) in sortedData" :key="idx" class="flex items-center gap-3">
        <span class="w-24 text-sm text-gray-600 dark:text-gray-400 truncate">{{ item.label }}</span>
        <div class="flex-1 h-5 bg-gray-100 rounded dark:bg-gray-700 overflow-hidden">
          <div
            class="h-full bg-primary transition-all duration-500 rounded"
            :style="{ width: barWidth(item.value) + '%' }"
          />
        </div>
        <span class="w-12 text-right text-sm font-medium text-gray-700 dark:text-gray-300">{{
          item.value
        }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface ChartDatum {
  label: string
  value: number
}

const props = defineProps<{
  title: string
  data: ChartDatum[]
  maxItems?: number
}>()

const sortedData = computed(() => {
  const limit = props.maxItems || 8
  return [...props.data].sort((a, b) => b.value - a.value).slice(0, limit)
})

function barWidth(value: number): number {
  const max = Math.max(...sortedData.value.map((d) => d.value), 1)
  return Math.round((value / max) * 100)
}
</script>
