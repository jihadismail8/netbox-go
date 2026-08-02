<template>
  <div>
    <PageHeader :title="'Delete ' + config.display_name" :breadcrumbs="breadcrumbs" />

    <div v-if="loading" class="py-8"><LoadingSpinner /></div>

    <div v-else-if="item" class="space-y-6">
      <div
        class="rounded-lg border border-red-300 bg-red-50 p-4 dark:border-red-700 dark:bg-red-900/30"
      >
        <div class="flex items-center gap-3">
          <AlertTriangle :size="24" class="text-red-600 dark:text-red-400" />
          <div>
            <p class="font-semibold text-red-800 dark:text-red-200">Delete "{{ item.display }}"?</p>
            <p class="text-sm text-red-600 dark:text-red-400">
              This action cannot be undone. Related objects may also be affected.
            </p>
            <p
              v-if="config.model === 'interface'"
              class="mt-2 text-sm font-semibold text-red-700 dark:text-red-300"
            >
              Deleting this Interface also deletes
              {{ interfaceAddressCount }} assigned IP
              {{ interfaceAddressCount === 1 ? 'address' : 'addresses' }}.
            </p>
          </div>
        </div>
      </div>

      <div
        class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-gray-700 dark:bg-gray-800"
      >
        <h3 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-300">Summary</h3>
        <dl class="grid grid-cols-1 gap-x-6 gap-y-2 sm:grid-cols-2">
          <div
            v-for="col in config.columns.slice(0, 6)"
            :key="col.key"
            class="flex justify-between border-b border-gray-100 py-1 dark:border-gray-800"
          >
            <dt class="text-sm text-gray-500">{{ col.label }}</dt>
            <dd class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatValue(resourceField(item, col.key)) }}
            </dd>
          </div>
        </dl>
      </div>

      <div
        v-if="errors"
        class="rounded-lg border border-red-300 bg-red-50 p-4 dark:border-red-700 dark:bg-red-900/30"
      >
        <pre class="text-sm text-red-700 dark:text-red-300">{{
          JSON.stringify(errors, null, 2)
        }}</pre>
      </div>

      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
          @click="handleCancel"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="deleting"
          class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:opacity-50"
          @click="handleConfirm"
        >
          {{ deleting ? 'Deleting...' : 'Confirm Delete' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AlertTriangle } from '@lucide/vue'
import { useCoreResource } from '@/composables/useCoreResource'
import { useNotificationStore } from '@/stores/notifications'
import type { ModelConfig } from '@/types'
import PageHeader from '@/components/layout/PageHeader.vue'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'
import { getErrorDetail, getErrorRecord, type ValidationErrors } from '@/api/errors'
import {
  isCoreReference,
  resourceField,
  type CoreFieldValue,
  type CoreResourceDTO,
} from '@/features/core/resources'

const props = defineProps<{ config: ModelConfig }>()
const route = useRoute()
const router = useRouter()
const notifications = useNotificationStore()

const { fetchById, deleteItem } = useCoreResource(props.config.model)

const item = shallowRef<CoreResourceDTO | null>(null)
const loading = ref(true)
const deleting = ref(false)
const errors = ref<ValidationErrors | null>(null)
const id = computed(() => Number(route.params.id))

const breadcrumbs = computed(() => [
  { label: props.config.display_name_plural, to: props.config.routePath },
  {
    label: item.value?.display || 'Detail',
    to: `${props.config.routePath}${id.value}/`,
  },
  { label: 'Delete' },
])

const interfaceAddressCount = computed(() => {
  if (!item.value || props.config.model !== 'interface') return 0
  const value = resourceField(item.value, 'count_ipaddresses')
  return typeof value === 'number' ? value : 0
})

function formatValue(v: CoreFieldValue | undefined): string {
  if (v === null || v === undefined) return '—'
  if (isCoreReference(v)) return v.display
  if (typeof v === 'boolean') return v ? 'Yes' : 'No'
  return String(v)
}

async function handleConfirm() {
  deleting.value = true
  errors.value = null
  try {
    await deleteItem(id.value)
    notifications.success(`${props.config.display_name} deleted`)
    router.push(props.config.routePath)
  } catch (caught: unknown) {
    errors.value = getErrorRecord(caught)
    notifications.error(getErrorDetail(caught, 'Failed to delete'))
  } finally {
    deleting.value = false
  }
}

function handleCancel() {
  router.push(`${props.config.routePath}${id.value}/`)
}

onMounted(async () => {
  try {
    item.value = await fetchById(id.value)
  } catch (caught: unknown) {
    notifications.error(getErrorDetail(caught, 'Failed to load'))
    router.push(props.config.routePath)
  } finally {
    loading.value = false
  }
})
</script>
