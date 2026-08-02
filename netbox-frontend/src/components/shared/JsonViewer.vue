<template>
  <div
    class="overflow-x-auto rounded-lg border border-gray-200 bg-gray-50 font-mono text-sm dark:border-gray-700 dark:bg-gray-900"
  >
    <div
      v-if="!data || (typeof data === 'object' && Object.keys(data).length === 0)"
      class="p-4 text-gray-400"
    >
      {{ typeof data === 'string' ? data : 'Empty' }}
    </div>
    <div v-else class="p-3">
      <JsonNode :data="data" :name="rootName" :is-root="true" :default-expanded="expanded" />
    </div>
  </div>
</template>

<script setup lang="ts">
import JsonNode from './JsonNode.vue'

withDefaults(
  defineProps<{
    data: unknown
    rootName?: string
    expanded?: boolean
  }>(),
  {
    rootName: '',
    expanded: false,
  },
)
</script>
