import { ref, shallowRef, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTableStore } from '@/stores/table'
import type { ModelConfig } from '@/types'
import { getErrorDetail } from '@/api/errors'
import { getCoreResourceAdapter } from '@/features/core/adapters'
import type { CoreProfileResourceName } from '@/features/core/manifest'
import type {
  CoreFilterState,
  CoreResourceDTO,
  CoreResourceFilters,
} from '@/features/core/resources'

export function useTable<N extends CoreProfileResourceName>(config: ModelConfig, resource: N) {
  const route = useRoute()
  const router = useRouter()
  const tableStore = useTableStore()

  const data = shallowRef<CoreResourceDTO<N>[]>([])
  const total = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const page = computed(() => Math.max(1, parseInt(route.query.page as string) || 1))
  const pageSize = computed(() => tableStore.getPageSize(config.routePath))
  const sortBy = ref<string | null>(null)
  const sortDir = ref<'asc' | 'desc'>('asc')
  const filters = ref<CoreFilterState>({})
  const adapter = getCoreResourceAdapter(resource)
  let fetchSerial = 0

  const totalPages = computed(() => Math.ceil(total.value / pageSize.value))
  const selectedIds = computed(() => tableStore.getSelectedArray(config.routePath))

  async function fetchData(
    fetchFn: (
      params: CoreResourceFilters<N>,
    ) => Promise<{ count: number; results: CoreResourceDTO<N>[] }>,
  ) {
    const request = ++fetchSerial
    loading.value = true
    error.value = null
    try {
      const params = {
        limit: pageSize.value,
        offset: (page.value - 1) * pageSize.value,
        ...adapter.filtersFromState(filters.value),
      }
      if (sortBy.value) {
        Reflect.set(
          params,
          'ordering',
          sortDir.value === 'desc' ? `-${sortBy.value}` : sortBy.value,
        )
      }
      const response = await fetchFn(params as CoreResourceFilters<N>)
      if (request === fetchSerial) {
        const lastPage = Math.max(1, Math.ceil(response.count / pageSize.value))
        if (page.value > lastPage) {
          total.value = response.count
          data.value = []
          await router.replace({
            query: { ...route.query, page: lastPage.toString() },
          })
          return
        }
        data.value = response.results
        total.value = response.count
      }
    } catch (caught: unknown) {
      if (request === fetchSerial) {
        data.value = []
        total.value = 0
        error.value = getErrorDetail(caught, 'Failed to fetch data')
      }
    } finally {
      if (request === fetchSerial) loading.value = false
    }
  }

  function setSort(column: string) {
    if (sortBy.value === column) {
      sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortBy.value = column
      sortDir.value = 'asc'
    }
  }

  function setPage(newPage: number) {
    router.push({ query: { ...route.query, page: Math.max(1, newPage).toString() } })
  }

  function setPageSize(size: number) {
    tableStore.setPageSize(config.routePath, size)
  }

  function setFilter(key: string, value: unknown) {
    if (value === null || value === '' || (Array.isArray(value) && value.length === 0)) {
      Reflect.deleteProperty(filters.value, key)
    } else {
      Reflect.set(filters.value, key, value)
    }
  }

  function replaceFilters(values: CoreFilterState) {
    filters.value = {}
    for (const [key, value] of Object.entries(values)) setFilter(key, value)
  }

  function applyFilters() {
    if (page.value !== 1) {
      router.push({ query: { ...route.query, page: '1' } })
    }
  }

  function clearFilters() {
    filters.value = {}
    router.push({ query: {} })
  }

  function toggleRow(id: number) {
    tableStore.toggleRow(config.routePath, id)
  }

  function selectAllOnPage() {
    tableStore.selectAll(
      config.routePath,
      data.value.map((r) => r.id as number),
    )
  }

  function clearSelection() {
    tableStore.clearSelection(config.routePath)
  }

  function isRowSelected(id: number): boolean {
    return tableStore.getSelected(config.routePath).has(id)
  }

  return {
    data,
    total,
    loading,
    error,
    page,
    pageSize,
    totalPages,
    sortBy,
    sortDir,
    filters,
    selectedIds,
    fetchData,
    setSort,
    setPage,
    setPageSize,
    setFilter,
    replaceFilters,
    applyFilters,
    clearFilters,
    toggleRow,
    selectAllOnPage,
    clearSelection,
    isRowSelected,
  }
}
