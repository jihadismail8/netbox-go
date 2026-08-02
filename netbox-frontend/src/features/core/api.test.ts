import { beforeEach, describe, expect, it, vi } from 'vitest'
import { invokeReadOnlyOperation, listResources, searchResourceOptions } from './api'

const mocks = vi.hoisted(() => ({ request: vi.fn() }))

vi.mock('@/api/http', () => ({ request: mocks.request }))

describe('core resource API', () => {
  beforeEach(() => mocks.request.mockReset())

  it('resolves endpoints only through the closed capability manifest', async () => {
    mocks.request.mockResolvedValue({
      data: { count: 0, next: null, previous: null, results: [] },
    })

    await listResources('site', { limit: 1 })

    expect(mocks.request).toHaveBeenCalledWith({
      method: 'GET',
      url: '/dcim/sites/',
      params: { limit: 1 },
    })
  })

  it('keeps selected relationship options visible alongside search results', async () => {
    mocks.request
      .mockResolvedValueOnce({
        data: {
          count: 1,
          next: null,
          previous: null,
          results: [{ id: 2, url: '/api/dcim/sites/2/', display: 'search result' }],
        },
      })
      .mockResolvedValueOnce({
        data: {
          count: 1,
          next: null,
          previous: null,
          results: [{ id: 7, url: '/api/dcim/sites/7/', display: 'selected value' }],
        },
      })

    await expect(searchResourceOptions('site', 'search', [7])).resolves.toEqual([
      { id: 7, url: '/api/dcim/sites/7/', display: 'selected value' },
      { id: 2, url: '/api/dcim/sites/2/', display: 'search result' },
    ])
    expect(mocks.request).toHaveBeenNthCalledWith(2, {
      method: 'GET',
      url: '/dcim/sites/',
      params: { id: '7', limit: 1 },
    })
  })

  it('scopes relationship choices with declared parent filters', async () => {
    mocks.request.mockResolvedValue({
      data: { count: 0, next: null, previous: null, results: [] },
    })

    await searchResourceOptions('rack', 'edge', [], { site_id: 4 })

    expect(mocks.request).toHaveBeenCalledWith({
      method: 'GET',
      url: '/dcim/racks/',
      params: { site_id: 4, limit: 100, q: 'edge' },
    })
  })

  it('rejects parameterized or cross-origin-looking Try it out paths', async () => {
    await expect(invokeReadOnlyOperation('/api/dcim/sites/{id}/')).rejects.toThrow(
      'Only concrete API paths',
    )
    await expect(invokeReadOnlyOperation('/api//example.com/')).rejects.toThrow(
      'Only concrete API paths',
    )
    expect(mocks.request).not.toHaveBeenCalled()
  })
})
