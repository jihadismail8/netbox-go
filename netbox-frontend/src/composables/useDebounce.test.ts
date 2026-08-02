import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ref, nextTick } from 'vue'
import { useDebounce, useDebouncedRef } from './useDebounce'

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('calls the function after the delay', () => {
    const fn = vi.fn()
    const { run } = useDebounce(300)
    run(fn)
    expect(fn).not.toHaveBeenCalled()
    vi.advanceTimersByTime(300)
    expect(fn).toHaveBeenCalledTimes(1)
  })

  it('only calls the last function within the window', () => {
    const fn1 = vi.fn()
    const fn2 = vi.fn()
    const fn3 = vi.fn()
    const { run } = useDebounce(300)
    run(fn1)
    vi.advanceTimersByTime(100)
    run(fn2)
    vi.advanceTimersByTime(100)
    run(fn3)
    vi.advanceTimersByTime(300)
    expect(fn1).not.toHaveBeenCalled()
    expect(fn2).not.toHaveBeenCalled()
    expect(fn3).toHaveBeenCalledTimes(1)
  })

  it('cancel prevents the pending call', () => {
    const fn = vi.fn()
    const { run, cancel } = useDebounce(300)
    run(fn)
    cancel()
    vi.advanceTimersByTime(500)
    expect(fn).not.toHaveBeenCalled()
  })
})

describe('useDebouncedRef', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })
  afterEach(() => {
    vi.useRealTimers()
  })

  it('initially mirrors the source value', () => {
    const source = ref('hello')
    const debounced = useDebouncedRef(source, 300)
    expect(debounced.value).toBe('hello')
  })

  it('updates after the delay', async () => {
    const source = ref('a')
    const debounced = useDebouncedRef(source, 300)
    source.value = 'b'
    expect(debounced.value).toBe('a')
    vi.advanceTimersByTime(300)
    await nextTick()
    expect(debounced.value).toBe('b')
  })

  it('only reflects the latest change within the window', async () => {
    const source = ref('a')
    const debounced = useDebouncedRef(source, 300)
    source.value = 'b'
    vi.advanceTimersByTime(100)
    source.value = 'c'
    vi.advanceTimersByTime(100)
    source.value = 'd'
    vi.advanceTimersByTime(300)
    await nextTick()
    expect(debounced.value).toBe('d')
  })
})
