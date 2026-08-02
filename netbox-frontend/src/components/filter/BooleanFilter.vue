<template>
  <select
    :value="tristate"
    class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
    @change="handleChange($event)"
  >
    <option value="">— Any —</option>
    <option value="true">Yes</option>
    <option value="false">No</option>
  </select>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FilterField } from '@/types'

const props = defineProps<{ field: FilterField; modelValue?: unknown }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean | null] }>()

const tristate = computed(() => {
  if (props.modelValue === true) return 'true'
  if (props.modelValue === false) return 'false'
  return ''
})

function handleChange(event: Event) {
  const val = (event.target as HTMLSelectElement).value
  if (val === 'true') emit('update:modelValue', true)
  else if (val === 'false') emit('update:modelValue', false)
  else emit('update:modelValue', null)
}
</script>
