<template>
  <Teleport to="body">
    <Transition name="fade">
      <div v-if="open" class="fixed inset-0 z-40 bg-black/30" @click="$emit('close')" />
    </Transition>
    <Transition name="slide-right">
      <div
        v-if="open"
        class="fixed top-0 right-0 z-50 flex h-full w-96 max-w-[90vw] flex-col bg-white shadow-2xl dark:bg-gray-800"
      >
        <!-- Header -->
        <div
          class="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-700"
        >
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Filters</h2>
          <button
            class="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700"
            @click="$emit('close')"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Filter fields -->
        <div class="flex-1 px-5 py-4 overflow-y-auto">
          <div class="space-y-4">
            <FilterField
              v-for="field in filters"
              :key="field.key"
              :field="field"
              :model-value="filterValue(field.key)"
              @update:model-value="updateValue(field.key, $event)"
            />
          </div>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-2 px-5 py-4 border-t border-gray-200 dark:border-gray-700">
          <button
            class="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
            @click="handleClear"
          >
            Clear All
          </button>
          <button
            class="px-4 py-2 text-sm font-medium text-white rounded-lg bg-primary hover:bg-primary-dark"
            @click="handleApply"
          >
            Apply Filters
          </button>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { X } from '@lucide/vue'
import type { FilterField as FilterFieldDef } from '@/types'
import type { CoreFilterState } from '@/features/core/resources'
import FilterField from './FilterField.vue'

const props = defineProps<{
  open: boolean
  filters: FilterFieldDef[]
  modelValue?: CoreFilterState
}>()

const emit = defineEmits<{
  close: []
  apply: [values: CoreFilterState]
  clear: []
}>()

const localValues = ref<CoreFilterState>({ ...(props.modelValue || {}) })

watch(
  () => props.open,
  (val) => {
    if (val) {
      localValues.value = { ...(props.modelValue || {}) }
    }
  },
)

function updateValue(key: string, value: unknown) {
  if (value === null || value === '' || (Array.isArray(value) && value.length === 0)) {
    Reflect.deleteProperty(localValues.value, key)
  } else {
    Reflect.set(localValues.value, key, value)
  }
}

function filterValue(key: string): unknown {
  return Reflect.get(localValues.value, key) as unknown
}

function handleApply() {
  emit('apply', { ...localValues.value })
  emit('close')
}

function handleClear() {
  localValues.value = {}
  emit('clear')
  emit('close')
}
</script>

<style scoped>
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.25s ease;
}
.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(100%);
}
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
