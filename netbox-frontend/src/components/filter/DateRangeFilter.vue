<template>
  <div class="grid grid-cols-2 gap-2">
    <div>
      <input
        type="date"
        :value="rangeAfter"
        class="w-full rounded-lg border border-gray-300 px-2.5 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
        placeholder="From"
        @input="updateRange('after', ($event.target as HTMLInputElement).value)"
      />
      <span class="text-xs text-gray-400">After</span>
    </div>
    <div>
      <input
        type="date"
        :value="rangeBefore"
        class="w-full rounded-lg border border-gray-300 px-2.5 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
        placeholder="To"
        @input="updateRange('before', ($event.target as HTMLInputElement).value)"
      />
      <span class="text-xs text-gray-400">Before</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FilterField } from '@/types'

const props = defineProps<{ field: FilterField; modelValue?: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string | null] }>()

const rangeAfter = computed(() => {
  const v = props.modelValue || ''
  const parts = v.split(',')
  return parts[0] || ''
})
const rangeBefore = computed(() => {
  const v = props.modelValue || ''
  const parts = v.split(',')
  return parts[1] || ''
})

function updateRange(side: 'after' | 'before', val: string) {
  const after = side === 'after' ? val : rangeAfter.value
  const before = side === 'before' ? val : rangeBefore.value
  if (!after && !before) {
    emit('update:modelValue', null)
  } else {
    emit('update:modelValue', `${after},${before}`)
  }
}
</script>
