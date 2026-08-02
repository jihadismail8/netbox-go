<template>
  <div
    class="flex items-center justify-between border-t border-gray-200 px-4 py-3 dark:border-gray-700"
  >
    <div class="flex items-center gap-2 text-sm text-gray-500">
      <span>Showing {{ startItem }}-{{ endItem }} of {{ total }}</span>
      <select
        :value="String(pageSize)"
        class="rounded border-gray-200 text-sm dark:border-gray-600 dark:bg-gray-800"
        @change="onPageSizeChange"
      >
        <option v-for="size in sizes" :key="size" :value="String(size)">{{ size }} / page</option>
      </select>
    </div>
    <div class="flex items-center gap-1">
      <button
        :disabled="page <= 1"
        class="rounded px-2 py-1 text-sm text-gray-500 hover:bg-gray-100 disabled:opacity-50 dark:hover:bg-gray-700"
        @click="emit('update:page', page - 1)"
      >
        <ChevronLeft :size="16" />
      </button>
      <template v-for="p in visiblePages" :key="p">
        <span v-if="p === '...'" class="px-2 text-gray-400">...</span>
        <button
          v-else
          class="min-w-[2rem] rounded px-2 py-1 text-sm"
          :class="
            p === page
              ? 'bg-primary text-white'
              : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700'
          "
          @click="emit('update:page', p as number)"
        >
          {{ p }}
        </button>
      </template>
      <button
        :disabled="page >= totalPages"
        class="rounded px-2 py-1 text-sm text-gray-500 hover:bg-gray-100 disabled:opacity-50 dark:hover:bg-gray-700"
        @click="emit('update:page', page + 1)"
      >
        <ChevronRight :size="16" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import { APP_CONFIG } from '@/config/app'

const props = defineProps<{ total: number; page: number; pageSize: number; totalPages: number }>()
const emit = defineEmits<{ 'update:page': [page: number]; 'update:pageSize': [size: number] }>()

const sizes = APP_CONFIG.pageSizes
const startItem = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const endItem = computed(() => Math.min(props.page * props.pageSize, props.total))

function onPageSizeChange(e: Event) {
  emit('update:pageSize', Number((e.target as HTMLSelectElement).value))
}

const visiblePages = computed<(number | string)[]>(() => {
  const pages: (number | string)[] = []
  const tp = props.totalPages
  const cp = props.page
  if (tp <= 7) {
    for (let i = 1; i <= tp; i++) pages.push(i)
  } else {
    pages.push(1)
    if (cp > 3) pages.push('...')
    for (let i = Math.max(2, cp - 1); i <= Math.min(tp - 1, cp + 1); i++) pages.push(i)
    if (cp < tp - 2) pages.push('...')
    pages.push(tp)
  }
  return pages
})
</script>
