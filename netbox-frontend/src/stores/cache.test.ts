import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useCacheStore } from './cache'

describe('useCacheStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('stores and retrieves values', () => {
    const cache = useCacheStore()
    cache.set('key1', { data: 'hello' })
    const result = cache.get<{ data: string }>('key1')
    expect(result).toEqual({ data: 'hello' })
  })

  it('returns undefined for missing keys', () => {
    const cache = useCacheStore()
    expect(cache.get('nonexistent')).toBeUndefined()
  })

  it('returns stale data even after expiry via getStale', () => {
    const cache = useCacheStore()
    cache.set('key1', 'value')
    // getStale does not check TTL
    const stale = cache.getStale('key1')
    expect(stale).toBe('value')
  })

  it('invalidate removes a specific key', () => {
    const cache = useCacheStore()
    cache.set('a', 1)
    cache.set('b', 2)
    cache.invalidate('a')
    expect(cache.get('a')).toBeUndefined()
    expect(cache.get('b')).toBe(2)
  })

  it('invalidatePrefix removes all matching keys', () => {
    const cache = useCacheStore()
    cache.set('/dcim/sites/?page=1', 'page1')
    cache.set('/dcim/sites/?page=2', 'page2')
    cache.set('/dcim/devices/?page=1', 'devices')
    cache.invalidatePrefix('/dcim/sites/')
    expect(cache.get('/dcim/sites/?page=1')).toBeUndefined()
    expect(cache.get('/dcim/sites/?page=2')).toBeUndefined()
    expect(cache.get('/dcim/devices/?page=1')).toBe('devices')
  })

  it('clear empties the cache', () => {
    const cache = useCacheStore()
    cache.set('a', 1)
    cache.set('b', 2)
    cache.clear()
    expect(cache.get('a')).toBeUndefined()
    expect(cache.get('b')).toBeUndefined()
  })

  it('swrFetch caches and returns fresh data', async () => {
    const cache = useCacheStore()
    const fetcher = vi.fn().mockResolvedValue({ count: 5 })
    const result = await cache.swrFetch('/dcim/sites/', fetcher)
    expect(result.fresh).toEqual({ count: 5 })
    expect(fetcher).toHaveBeenCalledTimes(1)
  })

  it('swrFetch deduplicates concurrent calls', async () => {
    const cache = useCacheStore()
    const fetcher = vi.fn().mockResolvedValue('data')
    const [r1, r2] = await Promise.all([
      cache.swrFetch('key', fetcher),
      cache.swrFetch('key', fetcher),
    ])
    expect(r1.fresh).toBe('data')
    expect(r2.fresh).toBe('data')
    expect(fetcher).toHaveBeenCalledTimes(1)
  })
})
