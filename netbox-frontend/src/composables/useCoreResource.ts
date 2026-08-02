import { ref } from 'vue'
import type { CoreProfileResourceName } from '@/features/core/manifest'
import {
  createResource,
  deleteResource,
  getResource,
  listResources,
  updateResource,
} from '@/features/core/api'
import type {
  CoreCreateResponse,
  CorePage,
  CoreResourceDTO,
  CoreResourceFilters,
  CoreResourceMutation,
} from '@/features/core/resources'
import { getErrorDetail } from '@/api/errors'

export function useCoreResource<N extends CoreProfileResourceName>(resource: N) {
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function run<R>(operation: () => Promise<R>, fallback: string): Promise<R> {
    loading.value = true
    error.value = null
    try {
      return await operation()
    } catch (caught: unknown) {
      error.value = getErrorDetail(caught, fallback)
      throw caught
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    fetchList: (params?: CoreResourceFilters<N>): Promise<CorePage<N>> =>
      run(() => listResources(resource, params), 'Failed to fetch data'),
    fetchById: (id: number): Promise<CoreResourceDTO<N>> =>
      run(() => getResource(resource, id), 'Failed to fetch detail'),
    createItem: (data: CoreResourceMutation<N>): Promise<CoreCreateResponse<N>> =>
      run(() => createResource(resource, data), 'Failed to create'),
    updateItem: (id: number, data: CoreResourceMutation<N>): Promise<CoreResourceDTO<N>> =>
      run(() => updateResource(resource, id, data), 'Failed to update'),
    deleteItem: (id: number): Promise<void> =>
      run(() => deleteResource(resource, id), 'Failed to delete'),
  }
}
