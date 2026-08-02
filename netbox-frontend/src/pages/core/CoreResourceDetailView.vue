<template>
  <div v-if="!loading && item">
    <PageHeader :title="item.display" :breadcrumbs="breadcrumbs">
      <template #actions>
        <RouterLink
          v-if="canChange(config.module, config.model)"
          :to="config.routePath + id + '/edit/'"
          class="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium text-white rounded-lg bg-primary hover:bg-primary-dark"
        >
          <Edit :size="16" /> Edit
        </RouterLink>
        <RouterLink
          v-if="canDelete(config.module, config.model)"
          :to="config.routePath + id + '/delete/'"
          class="inline-flex items-center gap-1 px-3 py-2 text-sm font-medium text-red-600 border border-red-300 rounded-lg hover:bg-red-50 dark:border-red-700 dark:hover:bg-red-900/20"
        >
          <Trash2 :size="16" /> Delete
        </RouterLink>
      </template>
    </PageHeader>

    <div
      class="rounded-lg border border-gray-200 bg-white shadow-sm dark:border-gray-700 dark:bg-gray-800"
    >
      <div class="p-6">
        <div class="space-y-4">
          <PropertiesTable :item="item" :fields="config.detailFields" />
          <IPAddressAssignmentPanel
            v-if="ipAddress"
            :address="ipAddress"
            :can-change="canChange(config.module, config.model)"
            @updated="handleAssignmentUpdated"
          />
          <dl
            v-if="item.created || item.last_updated"
            class="grid grid-cols-1 gap-3 rounded-lg border border-gray-200 p-4 text-sm dark:border-gray-700 sm:grid-cols-2"
          >
            <div v-if="item.created">
              <dt class="text-xs font-semibold uppercase tracking-wide text-gray-500">Created</dt>
              <dd class="mt-1"><RelativeTime :datetime="item.created" /></dd>
            </div>
            <div v-if="item.last_updated">
              <dt class="text-xs font-semibold uppercase tracking-wide text-gray-500">
                Last updated
              </dt>
              <dd class="mt-1"><RelativeTime :datetime="item.last_updated" /></dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </div>

  <LoadingSpinner v-else-if="loading" />
  <ErrorState
    v-else-if="loadError"
    title="Unable to load object"
    :message="loadError"
    :retry="true"
    @retry="loadItem"
  />
</template>

<script setup lang="ts">
import { ref, shallowRef, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { Edit, Trash2 } from '@lucide/vue'
import type { ModelConfig } from '@/types'
import { useCoreResource } from '@/composables/useCoreResource'
import PageHeader from '@/components/layout/PageHeader.vue'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'
import PropertiesTable from '@/components/detail/PropertiesTable.vue'
import RelativeTime from '@/components/shared/RelativeTime.vue'
import ErrorState from '@/components/shared/ErrorState.vue'
import IPAddressAssignmentPanel from '@/features/ipam/components/IPAddressAssignmentPanel.vue'
import type { IPAddressDTO } from '@/features/ipam/assignment'
import type { CoreResourceDTO } from '@/features/core/resources'
import { usePermissions } from '@/composables/usePermissions'
import { getErrorDetail } from '@/api/errors'

const props = defineProps<{ config: ModelConfig }>()
const route = useRoute()

const resource = props.config.model
const { fetchById } = useCoreResource(resource)
const { canChange, canDelete } = usePermissions()

const item = shallowRef<CoreResourceDTO | null>(null)
const loading = ref(true)
const loadError = ref('')
const id = computed(() => Number(route.params.id))

const breadcrumbs = computed(() => [
  { label: props.config.display_name_plural, to: props.config.routePath },
  { label: item.value?.display || 'Detail' },
])

const ipAddress = computed<IPAddressDTO | null>(() => {
  if (resource !== 'ipaddress' || !item.value) return null
  return item.value as IPAddressDTO
})

async function loadItem() {
  loading.value = true
  loadError.value = ''
  item.value = null
  try {
    item.value = await fetchById(id.value)
  } catch (caught: unknown) {
    loadError.value = getErrorDetail(caught, 'Failed to load this object.')
  } finally {
    loading.value = false
  }
}

function handleAssignmentUpdated(updated: IPAddressDTO) {
  item.value = updated
}

onMounted(loadItem)
</script>
