<template>
  <span
    class="cursor-pointer select-none text-sm"
    :class="expanded ? 'text-gray-600 dark:text-gray-300' : 'text-gray-400'"
    @click="expanded = !expanded"
  >
    <component :is="expanded ? ChevronDown : ChevronRight" :size="12" class="inline" />
    <span class="ml-1 text-primary">{{ name || 'root' }}:</span>
  </span>
  <span v-if="!expanded" class="ml-1 text-gray-400">{{ previewText }}</span>
  <div v-show="expanded" class="ml-4 border-l border-gray-200 pl-2 dark:border-gray-700">
    <template v-if="isArray">
      <div v-for="(item, idx) in arrayData" :key="idx">
        <JsonNode :data="item" :name="String(idx)" />
      </div>
      <div v-if="arrayData.length === 0" class="text-gray-400">[ ]</div>
    </template>
    <template v-else>
      <div v-for="(value, key) in objectData" :key="key">
        <JsonNode :data="value" :name="String(key)" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronDown, ChevronRight } from '@lucide/vue'

const props = withDefaults(
  defineProps<{
    data: unknown
    name?: string
    isRoot?: boolean
    defaultExpanded?: boolean
  }>(),
  {
    isRoot: false,
    defaultExpanded: false,
  },
)

const expanded = ref(props.isRoot || props.defaultExpanded)

const isArray = computed(() => Array.isArray(props.data))
const isObject = computed(
  () => props.data !== null && typeof props.data === 'object' && !isArray.value,
)

const arrayData = computed(() => (isArray.value ? (props.data as unknown[]) : []))
const objectData = computed(() => {
  if (!isObject.value) return {} as Record<string, unknown>
  return props.data as Record<string, unknown>
})

const primitiveText = computed(() => {
  if (props.data === null) return 'null'
  if (props.data === undefined) return 'undefined'
  if (typeof props.data === 'string') return `"${props.data}"`
  return String(props.data)
})

const previewText = computed(() => {
  if (isArray.value) {
    const len = arrayData.value.length
    return len === 0 ? '[ ]' : `[ ${len} item${len === 1 ? '' : 's'} ]`
  }
  if (isObject.value) {
    const keys = Object.keys(objectData.value)
    return keys.length === 0 ? '{ }' : `{ ${keys.length} key${keys.length === 1 ? '' : 's'} }`
  }
  return primitiveText.value
})
</script>
