/**
 * useDebounce — Debounce a ref or a function.
 *
 * Usage:
 *   const debounced = useDebouncedRef('', 300)
 *   // or
 *   const { run } = useDebounce((val) => search(val), 300)
 */
import { ref, watch, type Ref } from 'vue'

/**
 * Returns a debounced copy of a source ref. The returned ref updates
 * `delay` ms after the source stops changing.
 */
export function useDebouncedRef<T>(source: Ref<T>, delay: number = 300): Ref<T> {
  const debounced = ref(source.value) as Ref<T>
  let timer: ReturnType<typeof setTimeout>

  watch(
    source,
    (val) => {
      clearTimeout(timer)
      timer = setTimeout(() => {
        debounced.value = val
      }, delay)
    },
    { flush: 'sync' },
  )

  return debounced
}

/**
 * Returns `{ run, cancel }` where `run(fn)` debounces the given function.
 * The last call within the delay window wins.
 */
export function useDebounce(delay: number = 300) {
  let timer: ReturnType<typeof setTimeout> | null = null

  function run(fn: () => void): void {
    if (timer) clearTimeout(timer)
    timer = setTimeout(fn, delay)
  }

  function cancel(): void {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  return { run, cancel }
}
