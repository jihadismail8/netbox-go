/**
 * API Response Cache Store (Stale-While-Revalidate pattern)
 *
 * Caches GET responses so that navigations back to a previously-visited
 * list/detail page show cached data instantly while a fresh fetch happens
 * in the background. Entries auto-expire after `TTL` ms.
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'

interface CacheEntry<T = unknown> {
  data: T
  timestamp: number
  promise?: Promise<T>
}

const DEFAULT_TTL = 30_000 // 30 seconds

export const useCacheStore = defineStore('cache', () => {
  const entries = ref<Map<string, CacheEntry>>(new Map())

  /**
   * Read a cached value if it exists and is still fresh (within TTL).
   * Returns `undefined` if not present or stale.
   */
  function get<T>(key: string, ttl: number = DEFAULT_TTL): T | undefined {
    const entry = entries.value.get(key)
    if (!entry) return undefined
    if (Date.now() - entry.timestamp > ttl) {
      entries.value.delete(key)
      return undefined
    }
    return entry.data as T
  }

  /**
   * Read a cached value even if stale (for instant display while revalidating).
   */
  function getStale<T>(key: string): T | undefined {
    const entry = entries.value.get(key)
    return entry ? (entry.data as T) : undefined
  }

  /**
   * Store a value in the cache with the current timestamp.
   */
  function set<T>(key: string, data: T): void {
    entries.value.set(key, { data, timestamp: Date.now() })
  }

  /**
   * SWR fetch: return stale data immediately (if any), then fetch fresh data.
   * Deduplicates concurrent requests for the same key.
   */
  async function swrFetch<T>(
    key: string,
    fetcher: () => Promise<T>,
    ttl: number = DEFAULT_TTL,
  ): Promise<{ stale?: T; fresh: T }> {
    const stale = getStale<T>(key)

    // Deduplicate FIRST: if a fetch is already in-flight for this key, await it.
    // This must be checked before the freshness check below, otherwise a
    // placeholder entry (null data) would short-circuit as "fresh".
    const existing = entries.value.get(key)
    if (existing?.promise) {
      const data = (await existing.promise) as T
      return { stale, fresh: data }
    }

    // If we have fresh cached data, return it immediately
    const cached = get<T>(key, ttl)
    if (cached !== undefined && cached !== null) {
      return { stale, fresh: cached }
    }

    // Start a new fetch
    const promise = fetcher()
    const placeholder = (stale !== undefined ? stale : null) as unknown as T
    entries.value.set(key, { data: placeholder, timestamp: Date.now(), promise })
    try {
      const data = await promise
      set(key, data)
      return { stale, fresh: data }
    } finally {
      const e = entries.value.get(key)
      if (e) e.promise = undefined
    }
  }

  /**
   * Invalidate a specific cache key (e.g., after a mutation).
   */
  function invalidate(key: string): void {
    entries.value.delete(key)
  }

  /**
   * Invalidate all keys matching a prefix (e.g., '/dcim/sites').
   */
  function invalidatePrefix(prefix: string): void {
    for (const key of entries.value.keys()) {
      if (key.startsWith(prefix)) {
        entries.value.delete(key)
      }
    }
  }

  /**
   * Clear the entire cache.
   */
  function clear(): void {
    entries.value.clear()
  }

  return { entries, get, getStale, set, swrFetch, invalidate, invalidatePrefix, clear }
})
