import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useTable } from './useTable'
import type { ModelConfig } from '@/types'
import type { SiteDTO } from '@/features/core/resources'

const routerMocks = vi.hoisted(() => ({
  route: { query: {} as Record<string, string> },
  push: vi.fn(),
  replace: vi.fn(),
}))

vi.mock('vue-router', () => ({
  useRoute: () => routerMocks.route,
  useRouter: () => ({ push: routerMocks.push, replace: routerMocks.replace }),
}))

const config: ModelConfig = {
  module: 'dcim',
  model: 'site',
  display_name: 'Site',
  display_name_plural: 'Sites',
  apiPath: '/dcim/sites/',
  routePath: '/dcim/sites/',
  writableFields: ['name'],
  columns: [{ key: 'name', label: 'Name', sortable: true }],
  detailFields: [{ key: 'name', label: 'Name' }],
  filters: [],
  fields: [],
}

function site(id: number, name: string): SiteDTO {
  return {
    id,
    url: `/api/dcim/sites/${id}/`,
    display: name,
    created: '2026-07-18T10:00:00Z',
    last_updated: '2026-07-18T10:00:00Z',
    name,
    slug: name,
    status: { value: 'active', label: 'Active' },
    facility: '',
    description: '',
    comments: '',
    device_count: 0,
    prefix_count: 0,
    rack_count: 0,
  }
}

describe('useTable request state', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    routerMocks.route.query = {}
    routerMocks.push.mockReset()
    routerMocks.replace.mockReset()
  })

  it('replaces removed filters instead of leaking stale query parameters', async () => {
    const table = useTable(config, 'site')
    table.setFilter('status', 'active')
    table.replaceFilters({ name: 'edge' })
    const fetch = vi.fn().mockResolvedValue({ count: 0, results: [] })

    await table.fetchData(fetch)

    expect(fetch).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'edge', limit: 50, offset: 0 }),
    )
    expect(fetch.mock.calls[0][0]).not.toHaveProperty('status')
  })

  it('does not let an older response overwrite the latest sort/filter request', async () => {
    const table = useTable(config, 'site')
    let resolveFirst!: (value: { count: number; results: SiteDTO[] }) => void
    let resolveSecond!: (value: { count: number; results: SiteDTO[] }) => void
    const first = new Promise<{ count: number; results: SiteDTO[] }>((resolve) => {
      resolveFirst = resolve
    })
    const second = new Promise<{ count: number; results: SiteDTO[] }>((resolve) => {
      resolveSecond = resolve
    })

    const firstRequest = table.fetchData(() => first)
    const secondRequest = table.fetchData(() => second)
    resolveSecond({ count: 1, results: [site(2, 'latest')] })
    await secondRequest
    resolveFirst({ count: 1, results: [site(1, 'stale')] })
    await firstRequest

    expect(table.data.value).toEqual([site(2, 'latest')])
  })

  it('clamps an out-of-range page to the final available page', async () => {
    routerMocks.route.query = { page: '99' }
    const table = useTable(config, 'site')

    await table.fetchData(async () => ({ count: 51, results: [] }))

    expect(routerMocks.replace).toHaveBeenCalledWith({ query: { page: '2' } })
    expect(table.total.value).toBe(51)
  })
})
