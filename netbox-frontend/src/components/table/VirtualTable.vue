<template>
  <div
    ref="containerRef"
    class="overflow-auto"
    :style="{ maxHeight: typeof maxHeight === 'number' ? maxHeight + 'px' : maxHeight }"
    @scroll.passive="handleScroll"
  >
    <div :style="{ height: totalHeight + 'px', position: 'relative' }">
      <table class="w-full text-sm">
        <thead v-if="stickyHeader" class="sticky top-0 z-10 bg-white dark:bg-gray-800">
          <tr class="border-b border-gray-200 dark:border-gray-700">
            <th
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3 text-left font-semibold text-gray-700 dark:text-gray-300"
              :style="{ width: col.width }"
            >
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in visibleItems"
            :key="item.index"
            :style="{
              position: 'absolute',
              top: item.offset + 'px',
              width: '100%',
              height: rowHeight + 'px',
            }"
            class="flex border-b border-gray-100 dark:border-gray-800"
            @click="$emit('rowClick', item.data)"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              class="flex items-center px-4 py-2 text-gray-700 dark:text-gray-300"
              :style="{ width: col.width || 'auto' }"
            >
              {{ formatCell(item.data[col.key]) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import type { TableColumn } from '@/types'

const props = withDefaults(
  defineProps<{
    data: Record<string, unknown>[]
    columns: TableColumn[]
    rowHeight?: number
    overscan?: number
    maxHeight?: number | string
    stickyHeader?: boolean
  }>(),
  {
    rowHeight: 40,
    overscan: 10,
    maxHeight: 600,
    stickyHeader: true,
  },
)

defineEmits<{ rowClick: [row: Record<string, unknown>] }>()

const containerRef = ref<HTMLElement | null>(null)
const scrollTop = ref(0)
const viewportHeight = ref(600)

const totalHeight = computed(() => props.data.length * props.rowHeight)

const startIndex = computed(() => {
  return Math.max(0, Math.floor(scrollTop.value / props.rowHeight) - props.overscan)
})

const endIndex = computed(() => {
  const visibleCount = Math.ceil(viewportHeight.value / props.rowHeight)
  return Math.min(props.data.length, startIndex.value + visibleCount + props.overscan * 2)
})

interface VirtualItem {
  index: number
  offset: number
  data: Record<string, unknown>
}

const visibleItems = computed<VirtualItem[]>(() => {
  const items: VirtualItem[] = []
  for (let i = startIndex.value; i < endIndex.value; i++) {
    items.push({
      index: i,
      offset: i * props.rowHeight,
      data: props.data[i],
    })
  }
  return items
})

function handleScroll(e: Event) {
  scrollTop.value = (e.target as HTMLElement).scrollTop
}

function updateViewport() {
  if (containerRef.value) {
    viewportHeight.value = containerRef.value.clientHeight
  }
}

let resizeObserver: ResizeObserver | null = null
onMounted(() => {
  updateViewport()
  if (containerRef.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(updateViewport)
    resizeObserver.observe(containerRef.value)
  }
})
onUnmounted(() => {
  resizeObserver?.disconnect()
})

function formatCell(value: unknown): string {
  if (value === null || value === undefined) return '—'
  if (typeof value === 'object') {
    if (Array.isArray(value)) return String(value.length)
    const obj = value as Record<string, unknown>
    return String(obj.display || obj.name || '—')
  }
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  return String(value)
}
</script>
