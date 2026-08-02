<template>
  <div
    class="p-6 bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800"
  >
    <div class="flex items-center justify-between">
      <div>
        <p class="text-sm font-medium text-gray-500">{{ label }}</p>
        <p class="mt-2 text-3xl font-bold text-gray-900 dark:text-white">
          <span v-if="loading" class="inline-block w-8 h-6 align-middle">
            <LoadingSpinner :small="true" />
          </span>
          <span v-else-if="error" class="text-sm text-gray-400">—</span>
          <span v-else>{{ formatCount(count) }}</span>
        </p>
      </div>
      <RouterLink
        :to="route"
        class="flex items-center justify-center w-12 h-12 rounded-full bg-primary/10 hover:bg-primary/20"
      >
        <component :is="icon" :size="24" class="text-primary" />
      </RouterLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { Component } from 'vue'
import { listResources } from '@/features/core/api'
import type { CoreProfileResourceName } from '@/features/core/manifest'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'

const props = defineProps<{
  label: string
  resource: CoreProfileResourceName
  route: string
  icon: Component
}>()

const count = ref(0)
const loading = ref(true)
const error = ref(false)

async function fetchCount() {
  loading.value = true
  error.value = false
  try {
    const resp = await listResources(props.resource, { limit: 1 })
    count.value = resp.count
  } catch {
    error.value = true
    count.value = 0
  } finally {
    loading.value = false
  }
}

function formatCount(n: number): string {
  return n.toLocaleString()
}

onMounted(fetchCount)
watch(() => props.resource, fetchCount)
</script>
