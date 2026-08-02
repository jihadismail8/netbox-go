<template>
  <div class="space-y-2">
    <div v-if="loading" class="text-sm text-gray-400">Loading...</div>
    <div v-if="error" class="text-sm text-red-500">{{ error }}</div>
    <input
      v-model="search"
      type="search"
      placeholder="Search choices..."
      class="w-full rounded-lg border border-gray-300 px-3 py-1.5 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
    />
    <select
      v-if="!field.multiple"
      :value="modelValue ?? ''"
      class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700 dark:text-white"
      @change="handleSingleSelect($event)"
    >
      <option value="">— Any —</option>
      <option v-for="item in options" :key="item.id" :value="item.id">
        {{ item.display }}
      </option>
    </select>
    <div
      v-else
      class="space-y-1 max-h-48 overflow-y-auto rounded-lg border border-gray-300 p-2 dark:border-gray-600"
    >
      <label v-for="item in options" :key="item.id" class="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          :value="item.id"
          :checked="selectedArray.includes(item.id)"
          class="rounded border-gray-300"
          @change="toggleMulti(item.id)"
        />
        <span class="text-gray-700 dark:text-gray-300">{{ item.display }}</span>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onBeforeUnmount, onMounted } from 'vue'
import type { FilterField } from '@/types'
import type { CoreReference } from '@/features/core/resources'
import { isCoreProfileResource } from '@/features/core/manifest'
import { searchResourceOptions } from '@/features/core/api'
import { getErrorDetail } from '@/api/errors'

const props = defineProps<{ field: FilterField; modelValue?: unknown }>()
const emit = defineEmits<{ 'update:modelValue': [value: unknown] }>()

const options = ref<CoreReference[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const search = ref('')
let requestSerial = 0

const selectedArray = computed<number[]>(() => {
  if (Array.isArray(props.modelValue)) return props.modelValue as number[]
  return []
})

const selectedIDs = computed<number[]>(() => {
  if (Array.isArray(props.modelValue)) return props.modelValue as number[]
  return typeof props.modelValue === 'number' ? [props.modelValue] : []
})

async function loadOptions() {
  if (!props.field.relationResource || !isCoreProfileResource(props.field.relationResource)) return
  const request = ++requestSerial
  loading.value = true
  error.value = null
  try {
    const fetched = await searchResourceOptions(
      props.field.relationResource,
      search.value,
      selectedIDs.value,
    )
    if (request !== requestSerial) return
    // Merge existing selections that may not be in results
    const existingIds = options.value.filter((o) => selectedIDs.value.includes(o.id))
    const newIds = fetched.map((r) => r.id)
    const missingSelections = existingIds.filter((o) => !newIds.includes(o.id))
    options.value = [...missingSelections, ...fetched]
  } catch (caught: unknown) {
    if (request === requestSerial) {
      error.value = getErrorDetail(caught, 'Failed to load options')
    }
  } finally {
    if (request === requestSerial) loading.value = false
  }
}

function handleSingleSelect(event: Event) {
  const val = (event.target as HTMLSelectElement).value
  emit('update:modelValue', val ? Number(val) : null)
}

function toggleMulti(id: number) {
  const current = [...selectedArray.value]
  const idx = current.indexOf(id)
  if (idx >= 0) current.splice(idx, 1)
  else current.push(id)
  emit('update:modelValue', current)
}

let debounceTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(loadOptions, 300)
})

onMounted(loadOptions)
onBeforeUnmount(() => clearTimeout(debounceTimer))
</script>
