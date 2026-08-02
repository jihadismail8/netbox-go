import { request } from '@/api/http'
import { getCoreProfileResource, type CoreProfileResourceName } from '@/features/core/manifest'
import type {
  CoreCreateResponse,
  CorePage,
  CoreReference,
  CoreResourceDTO,
  CoreResourceFilters,
  CoreResourceMutation,
} from './resources'

function pathFor(resourceName: CoreProfileResourceName): string {
  return getCoreProfileResource(resourceName).apiPath
}

export async function listResources<N extends CoreProfileResourceName>(
  resourceName: N,
  params?: CoreResourceFilters<N>,
): Promise<CorePage<N>> {
  const response = await request<CorePage<N>>({
    method: 'GET',
    url: pathFor(resourceName),
    params,
  })
  return response.data
}

export async function getResource<N extends CoreProfileResourceName>(
  resourceName: N,
  id: number,
): Promise<CoreResourceDTO<N>> {
  const response = await request<CoreResourceDTO<N>>({
    method: 'GET',
    url: `${pathFor(resourceName)}${id}/`,
  })
  return response.data
}

export async function createResource<N extends CoreProfileResourceName>(
  resourceName: N,
  payload: CoreResourceMutation<N>,
): Promise<CoreCreateResponse<N>> {
  const response = await request<CoreCreateResponse<N>>({
    method: 'POST',
    url: pathFor(resourceName),
    data: payload,
  })
  return response.data
}

export async function updateResource<N extends CoreProfileResourceName>(
  resourceName: N,
  id: number,
  payload: CoreResourceMutation<N>,
): Promise<CoreResourceDTO<N>> {
  const response = await request<CoreResourceDTO<N>>({
    method: 'PATCH',
    url: `${pathFor(resourceName)}${id}/`,
    data: payload,
  })
  return response.data
}

export async function deleteResource(
  resourceName: CoreProfileResourceName,
  id: number,
): Promise<void> {
  await request({ method: 'DELETE', url: `${pathFor(resourceName)}${id}/` })
}

export async function searchResourceOptions<N extends CoreProfileResourceName>(
  resourceName: N,
  query: string,
  selectedIDs: number[] = [],
  filters: CoreResourceFilters<N> = {},
): Promise<CoreReference[]> {
  const params: CoreResourceFilters<N> = { ...filters, limit: 100 }
  if (query.trim()) params.q = query.trim()
  const response = await listResources(resourceName, params)
  if (selectedIDs.length === 0) return response.results.map(toReference)

  const selected = await listResources(resourceName, {
    id: selectedIDs.join(','),
    limit: selectedIDs.length,
  })
  const results = selected.results.map(toReference)
  for (const candidate of response.results) {
    if (!results.some((existing) => existing.id === candidate.id))
      results.push(toReference(candidate))
  }
  return results
}

function toReference(resource: CoreResourceDTO): CoreReference {
  return { id: resource.id, url: resource.url, display: resource.display }
}

export interface OpenAPIOperation {
  operationId?: string
  summary?: string
  description?: string
  tags?: string[]
}

export interface OpenAPISchema {
  openapi: string
  paths: {
    [path: string]: Partial<{
      get: OpenAPIOperation
      post: OpenAPIOperation
      put: OpenAPIOperation
      patch: OpenAPIOperation
      delete: OpenAPIOperation
    }>
  }
}

export async function getOpenAPISchema(): Promise<OpenAPISchema> {
  const response = await request<OpenAPISchema>({ method: 'GET', url: '/schema/' })
  return response.data
}

export async function invokeReadOnlyOperation(
  path: string,
): Promise<{ status: number; data: unknown }> {
  const relativePath = path.slice('/api'.length)
  if (
    !path.startsWith('/api/') ||
    path.includes('{') ||
    path.includes('?') ||
    path.includes('#') ||
    relativePath.startsWith('//')
  ) {
    throw new Error('Only concrete API paths may be requested.')
  }
  const response = await request<unknown>({ method: 'GET', url: relativePath })
  return { status: response.status, data: response.data }
}
