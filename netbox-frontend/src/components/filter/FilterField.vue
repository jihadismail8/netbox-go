<template>
  <div>
    <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{{
      field.label
    }}</label>
    <component
      :is="componentName"
      :field="field"
      :model-value="modelValue"
      @update:model-value="$emit('update:modelValue', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FilterField } from '@/types'
import TextFilter from './TextFilter.vue'
import SelectFilter from './SelectFilter.vue'
import ApiSelectFilter from './ApiSelectFilter.vue'
import BooleanFilter from './BooleanFilter.vue'
import DateRangeFilter from './DateRangeFilter.vue'
import IntegerRangeFilter from './IntegerRangeFilter.vue'

const props = defineProps<{
  field: FilterField
  modelValue?: unknown
}>()

defineEmits<{ 'update:modelValue': [value: unknown] }>()

const componentName = computed(() => {
  switch (props.field.type) {
    case 'text':
      return TextFilter
    case 'select':
      return SelectFilter
    case 'api-select':
      return ApiSelectFilter
    case 'boolean':
      return BooleanFilter
    case 'date-range':
      return DateRangeFilter
    case 'integer-range':
      return IntegerRangeFilter
    default:
      return TextFilter
  }
})
</script>
