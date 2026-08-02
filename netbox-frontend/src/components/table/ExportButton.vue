<template>
  <button
    class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
    @click="exportCsv"
  >
    <Download :size="16" />
    Export
  </button>
</template>

<script setup lang="ts">
import { Download } from '@lucide/vue'
import type { TableColumn } from '@/types'
import { exportToCsv } from '@/utils/csv'

const props = defineProps<{
  data: object[]
  columns: TableColumn[]
  filename?: string
}>()

function exportCsv() {
  const filename = props.filename || 'export.csv'
  exportToCsv(props.data, props.columns, filename)
}
</script>
