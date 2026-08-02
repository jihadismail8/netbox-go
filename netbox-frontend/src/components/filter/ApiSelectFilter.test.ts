import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ApiSelectFilter from './ApiSelectFilter.vue'

const mocks = vi.hoisted(() => ({ searchResourceOptions: vi.fn() }))

vi.mock('@/features/core/api', () => ({
  searchResourceOptions: mocks.searchResourceOptions,
}))

describe('ApiSelectFilter', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mocks.searchResourceOptions.mockReset().mockResolvedValue([{ id: 7, display: 'Site 7' }])
  })

  afterEach(() => vi.useRealTimers())

  it('lets single-value relationship filters search beyond the first result page', async () => {
    const wrapper = mount(ApiSelectFilter, {
      props: {
        field: { key: 'site_id', label: 'Site', type: 'api-select', relationResource: 'site' },
      },
    })
    await flushPromises()

    expect(mocks.searchResourceOptions).toHaveBeenCalledWith('site', '', [])

    await wrapper.get('input[type="search"]').setValue('edge')
    await vi.advanceTimersByTimeAsync(300)
    await flushPromises()

    expect(mocks.searchResourceOptions).toHaveBeenLastCalledWith('site', 'edge', [])
    expect(wrapper.get('option[value="7"]').text()).toBe('Site 7')
  })
})
