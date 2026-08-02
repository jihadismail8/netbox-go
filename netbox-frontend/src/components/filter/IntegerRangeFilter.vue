<template>
  <div class="grid grid-cols-2 gap-2">
    <div>
      <input
        type="number"
        :value="rangeMin"
        class="w-full rounded-lg border border-gray-300 px-2.5 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
        placeholder="Min"
        @input="updateRange('min', ($event.target as HTMLInputElement).value)"
      />
    </div>
    <div>
      <input
        type="number"
        :value="rangeMax"
        class="w-full rounded-lg border border-gray-300 px-2.5 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
        placeholder="Max"
        @input="updateRange('max', ($event.target as HTMLInputElement).value)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FilterField } from '@/types'

const props = defineProps<{ field: FilterField; modelValue?: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string | null] }>()

const rangeMin = computed(() => {
  const v = props.modelValue || ''
  const parts = v.split(',')
  return parts[0] || ''
})
const rangeMax = computed(() => {
  const v = props.modelValue || ''
  const parts = v.split(',')
  return parts[1] || ''
})

function updateRange(side: 'min' | 'max', val: string) {
  const min = side === 'min' ? val : rangeMin.value
  const max = side === 'max' ? val : rangeMax.value
  if (!min && !max) {
    emit('update:modelValue', null)
  } else {
    emit('update:modelValue', `${min},${max}`)
  }
}
</script>
