<template>
  <div>
    <label v-if="label" class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label }}
      <span v-if="required" class="text-red-500">*</span>
    </label>
    <div class="relative flex gap-2">
      <input
        ref="searchInput"
        v-model="search"
        type="text"
        :required="required"
        :disabled="disabled"
        :placeholder="placeholder || 'Search...'"
        class="min-w-0 flex-1 px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary disabled:cursor-not-allowed disabled:bg-gray-100 disabled:text-gray-500 dark:border-gray-600 dark:bg-gray-800 dark:disabled:bg-gray-900"
        @input="handleSearchInput"
        @focus="handleFocus"
        @blur="hideResults"
      />
      <button
        v-if="!disabled && !required && hasValue"
        type="button"
        class="rounded-lg border border-gray-300 px-3 py-2 text-sm text-gray-600 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
        aria-label="Clear selection"
        @mousedown.prevent="clearSelection"
      >
        Clear
      </button>
      <div
        v-if="showResults && results.length > 0"
        role="listbox"
        class="absolute left-0 right-0 top-full z-30 mt-1 overflow-y-auto bg-white border border-gray-200 rounded-lg shadow-lg max-h-60 dark:border-gray-700 dark:bg-gray-800"
      >
        <button
          v-for="item in results"
          :key="item.id"
          class="block w-full px-3 py-2 text-sm text-left text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
          @mousedown.prevent="selectItem(item)"
        >
          {{ item.display }}
        </button>
      </div>
    </div>
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="lookupError" class="mt-1 text-sm text-red-500">{{ lookupError }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { isCoreProfileResource, type CoreProfileResourceName } from '@/features/core/manifest'
import { searchResourceOptions } from '@/features/core/api'
import { getCoreResourceAdapter } from '@/features/core/adapters'
import type {
  CoreFilterState,
  CoreReference,
  CoreRelationSelection,
} from '@/features/core/resources'

const props = defineProps<{
  modelValue?: CoreRelationSelection | CoreRelationSelection[]
  label?: string
  relationResource?: string
  placeholder?: string
  required?: boolean
  multiple?: boolean
  helpText?: string
  error?: string
  disabled?: boolean
  relationFilters?: CoreFilterState
}>()

const emit = defineEmits<{ 'update:modelValue': [value: number | number[] | null] }>()

const search = ref('')
const searchInput = ref<HTMLInputElement | null>(null)
const results = ref<CoreReference[]>([])
const showResults = ref(false)
const lookupError = ref('')
const editingSearch = ref(false)

let debounce: ReturnType<typeof setTimeout>
let requestSerial = 0

const hasValue = computed(() =>
  Array.isArray(props.modelValue)
    ? props.modelValue.length > 0
    : props.modelValue !== null && props.modelValue !== undefined,
)

function syncSelectionValidity() {
  searchInput.value?.setCustomValidity(
    props.required && !props.disabled && !hasValue.value ? 'Select an item from the choices.' : '',
  )
}

function targetResource(): CoreProfileResourceName | null {
  return props.relationResource && isCoreProfileResource(props.relationResource)
    ? props.relationResource
    : null
}

async function loadResults(query: string) {
  const request = ++requestSerial
  const resource = targetResource()
  if (!resource) {
    results.value = []
    return
  }
  try {
    const nextResults =
      props.relationFilters && Object.keys(props.relationFilters).length > 0
        ? await searchResourceOptions(
            resource,
            query,
            [],
            getCoreResourceAdapter(resource).filtersFromState(props.relationFilters),
          )
        : await searchResourceOptions(resource, query)
    if (request === requestSerial) {
      results.value = nextResults
      lookupError.value = ''
    }
  } catch {
    if (request === requestSerial) {
      results.value = []
      lookupError.value = 'Unable to load choices.'
    }
  }
}

watch(search, (val) => {
  clearTimeout(debounce)
  if (props.disabled) return
  if (val.length === 1) return
  debounce = setTimeout(() => loadResults(val), 300)
})

watch(
  () => props.modelValue,
  (value) => {
    if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
      editingSearch.value = false
      search.value = value.display || String(value.id)
    } else if (Array.isArray(value)) {
      const labels = value
        .filter((item): item is CoreReference => typeof item === 'object' && item !== null)
        .map((item) => item.display || String(item.id))
      if (labels.length > 0) search.value = labels.join(', ')
    } else if (value === null || value === undefined) {
      if (!editingSearch.value) search.value = ''
    } else if (typeof value === 'number' && !search.value) {
      search.value = String(value)
    }
  },
  { immediate: true },
)

watch([hasValue, () => props.disabled, () => props.required], syncSelectionValidity)

watch(
  () => props.relationFilters,
  () => {
    results.value = []
    if (!props.disabled) void loadResults(search.value)
  },
  { deep: true },
)

function selectItem(item: CoreReference) {
  editingSearch.value = false
  if (props.multiple) {
    const current = Array.isArray(props.modelValue)
      ? props.modelValue.flatMap((value) => {
          if (value === null) return []
          return [typeof value === 'object' ? value.id : value]
        })
      : []
    current.push(item.id)
    emit('update:modelValue', current)
  } else {
    emit('update:modelValue', item.id)
  }
  searchInput.value?.setCustomValidity('')
  search.value = item.display
  showResults.value = false
}

function handleSearchInput() {
  editingSearch.value = true
  if (hasValue.value) emit('update:modelValue', null)
  if (props.required && !props.disabled) {
    searchInput.value?.setCustomValidity('Select an item from the choices.')
  }
  showResults.value = true
}

function handleFocus() {
  showResults.value = true
  if (results.value.length === 0) void loadResults(search.value)
}

function clearSelection() {
  clearTimeout(debounce)
  editingSearch.value = false
  search.value = ''
  results.value = []
  showResults.value = false
  emit('update:modelValue', null)
}

function hideResults() {
  setTimeout(() => {
    showResults.value = false
  }, 200)
}

onMounted(syncSelectionValidity)
onBeforeUnmount(() => clearTimeout(debounce))
</script>
