<template>
  <div>
    <PageHeader :title="title" :breadcrumbs="breadcrumbs">
      <template #actions>
        <div class="flex items-center gap-2">
          <button
            class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
            @click="filterPanelOpen = true"
          >
            <Filter :size="16" /> Filter
            <span
              v-if="activeFilterCount > 0"
              class="ml-1 rounded-full bg-primary px-1.5 py-0.5 text-xs text-white"
              >{{ activeFilterCount }}</span
            >
          </button>
          <ExportButton
            :data="tableData"
            :columns="config.columns"
            :filename="config.model + '-export.csv'"
          />
          <RouterLink
            v-if="canAdd(config.module, config.model)"
            :to="config.routePath + 'add/'"
            class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-sm font-medium text-white hover:bg-primary-dark"
          >
            <Plus :size="16" /> Add
          </RouterLink>
        </div>
      </template>
    </PageHeader>

    <!-- Active filter chips -->
    <div v-if="activeFilterEntries.length > 0" class="flex flex-wrap items-center gap-2 mb-3">
      <span class="text-xs font-medium text-gray-500">Active filters:</span>
      <span
        v-for="[key, value] in activeFilterEntries"
        :key="key"
        class="inline-flex items-center gap-1 rounded-full bg-gray-100 px-2.5 py-1 text-xs text-gray-700 dark:bg-gray-700 dark:text-gray-300"
      >
        {{ getFilterLabel(key) }}: {{ value }}
        <button class="ml-0.5 text-gray-400 hover:text-red-500" @click="removeFilter(key)">
          <X :size="12" />
        </button>
      </span>
      <button class="text-xs text-primary hover:underline" @click="clearAllFilters">
        Clear all
      </button>
    </div>

    <div
      class="bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800"
    >
      <DataTable
        :data="tableData"
        :columns="config.columns"
        :selectable="false"
        @sort="handleSort"
        @row-click="goToDetail"
      />
      <div v-if="loading" class="py-8"><LoadingSpinner /></div>
      <ErrorState
        v-else-if="error"
        title="Unable to load results"
        :message="error"
        :retry="true"
        @retry="loadData"
      />
      <EmptyState v-if="!loading && !error && tableData.length === 0" message="No results found" />
      <PaginationControls
        v-if="total > 0"
        :total="total"
        :page="page"
        :page-size="pageSize"
        :total-pages="totalPages"
        @update:page="setPage"
        @update:page-size="handlePageSize"
      />
    </div>

    <!-- Filter Panel -->
    <FilterPanel
      v-model="activeFilters"
      :open="filterPanelOpen"
      :filters="config.filters"
      @close="filterPanelOpen = false"
      @apply="handleApplyFilters"
      @clear="handleClearFilters"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Filter, X } from '@lucide/vue'
import { useTable } from '@/composables/useTable'
import { useCoreResource } from '@/composables/useCoreResource'
import type { ModelConfig } from '@/types'
import PageHeader from '@/components/layout/PageHeader.vue'
import DataTable from '@/components/table/DataTable.vue'
import ExportButton from '@/components/table/ExportButton.vue'
import PaginationControls from '@/components/shared/PaginationControls.vue'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'
import EmptyState from '@/components/shared/EmptyState.vue'
import FilterPanel from '@/components/filter/FilterPanel.vue'
import ErrorState from '@/components/shared/ErrorState.vue'
import { usePermissions } from '@/composables/usePermissions'
import type { CoreFilterState, CoreResourceDTO } from '@/features/core/resources'

const props = defineProps<{ config: ModelConfig }>()
const router = useRouter()
const resource = props.config.model
const { fetchList } = useCoreResource(resource)
const { canAdd } = usePermissions()

const table = useTable(props.config, resource)
const tableData = table.data
const total = table.total
const loading = table.loading
const error = table.error
const page = table.page
const pageSize = table.pageSize
const totalPages = table.totalPages
const setSort = table.setSort
const setPage = table.setPage
const setPageSize = table.setPageSize
const setFilter = table.setFilter
const replaceFilters = table.replaceFilters
const applyFilters = table.applyFilters
const clearFilters = table.clearFilters
const fetchData = table.fetchData

// Filter state
const filterPanelOpen = ref(false)
const activeFilters = ref<CoreFilterState>({})

const title = computed(() => props.config.display_name_plural)
const breadcrumbs = computed(() => [{ label: props.config.display_name_plural }])

const activeFilterEntries = computed(() => {
  return Object.entries(activeFilters.value).filter(
    ([, v]) => v !== null && v !== '' && !(Array.isArray(v) && v.length === 0),
  )
})

const activeFilterCount = computed(() => activeFilterEntries.value.length)

function getFilterLabel(key: string): string {
  const f = props.config.filters.find((candidate) => candidate.key === key)
  return f?.label || key
}

function removeFilter(key: string) {
  Reflect.deleteProperty(activeFilters.value, key)
  setFilter(key, null)
  activeFilters.value = { ...activeFilters.value }
  applyFilters()
  if (page.value === 1) void loadData()
}

function clearAllFilters() {
  activeFilters.value = {}
  clearFilters()
  if (page.value === 1) void loadData()
}

function handleApplyFilters(values: CoreFilterState) {
  activeFilters.value = { ...values }
  replaceFilters(values)
  applyFilters()
  if (page.value === 1) void loadData()
}

function handleClearFilters() {
  activeFilters.value = {}
  clearFilters()
  if (page.value === 1) void loadData()
}

async function loadData() {
  await fetchData(fetchList)
}

function handleSort(column: string) {
  setSort(column)
  void loadData()
}

function handlePageSize(size: number) {
  setPageSize(size)
  if (page.value === 1) void loadData()
  else setPage(1)
}

function goToDetail(row: CoreResourceDTO) {
  router.push(props.config.routePath + row.id + '/')
}

watch(() => page.value, loadData)
onMounted(loadData)
</script>
