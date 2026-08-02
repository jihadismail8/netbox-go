<template>
  <div class="overflow-x-auto">
    <table class="w-full text-sm">
      <thead>
        <tr class="border-b border-gray-200 dark:border-gray-700">
          <th v-if="selectable" class="w-10 px-4 py-3">
            <input
              type="checkbox"
              :checked="allSelected"
              class="border-gray-300 rounded"
              @change="toggleSelectAll"
            />
          </th>
          <th
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-3 font-semibold text-left text-gray-700 dark:text-gray-300"
            :class="{ 'cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800': col.sortable }"
            @click="col.sortable && $emit('sort', col.key)"
          >
            <div class="flex items-center gap-1">
              {{ col.label }}
              <ArrowUpDown v-if="col.sortable" :size="12" class="text-gray-400" />
            </div>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, idx) in data"
          :key="row.id || idx"
          class="transition-colors border-b border-gray-100 hover:bg-gray-50 dark:border-gray-800 dark:hover:bg-gray-800/50"
          :class="{ 'bg-primary/5': isSelected(row.id) }"
          @click="$emit('rowClick', row)"
        >
          <td v-if="selectable" class="px-4 py-3" @click.stop>
            <input
              type="checkbox"
              :checked="isSelected(row.id)"
              class="border-gray-300 rounded"
              @change="$emit('toggleRow', row.id as number)"
            />
          </td>
          <td
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-3 text-gray-700 dark:text-gray-300"
          >
            <slot :name="'cell-' + col.key" :row="row" :value="resourceField(row, col.key)">
              {{ formatCell(resourceField(row, col.key), col, row) }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ArrowUpDown } from '@lucide/vue'
import type { TableColumn } from '@/types'
import { isCoreReference, resourceField, type CoreResourceDTO } from '@/features/core/resources'

const props = defineProps<{
  data: CoreResourceDTO[]
  columns: TableColumn[]
  selectable?: boolean
  selectedIds?: Set<number>
}>()

const emit = defineEmits<{
  sort: [column: string]
  toggleRow: [id: number]
  toggleSelectAll: []
  rowClick: [row: CoreResourceDTO]
}>()

const allSelected = computed(() => {
  if (!props.selectedIds || props.data.length === 0) return false
  return props.data.every((r) => props.selectedIds!.has(r.id as number))
})

function isSelected(id: unknown): boolean {
  return props.selectedIds?.has(id as number) || false
}

function toggleSelectAll() {
  emit('toggleSelectAll')
}

function formatCell(value: unknown, col: TableColumn, row: CoreResourceDTO): string {
  if (value === null || value === undefined) return '—'
  if (col.formatter) return col.formatter(value, row)
  if (isCoreReference(value)) return value.display
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  return String(value)
}
</script>
