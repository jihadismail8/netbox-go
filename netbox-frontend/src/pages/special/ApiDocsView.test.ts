import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import ApiDocsView from './ApiDocsView.vue'

const mocks = vi.hoisted(() => ({ getOpenAPISchema: vi.fn() }))

mocks.getOpenAPISchema.mockResolvedValue({
  openapi: '3.1.0',
  paths: {
    '/api/dcim/sites/': {
      get: { operationId: 'ListSites', tags: ['dcim'], summary: 'List Sites' },
      post: { operationId: 'CreateSite', tags: ['dcim'], summary: 'Create Site' },
    },
    '/api/ipam/ip-addresses/': {
      get: { operationId: 'ListIPAddresses', tags: ['ipam'], summary: 'List IP Addresses' },
    },
  },
})

vi.mock('@/features/core/api', () => ({
  getOpenAPISchema: mocks.getOpenAPISchema,
  invokeReadOnlyOperation: vi.fn(),
}))

describe('ApiDocsView', () => {
  it('consumes the canonical schema endpoint and lists its operations', async () => {
    const wrapper = mount(ApiDocsView, {
      global: { stubs: { RouterLink: true } },
    })

    await flushPromises()

    expect(wrapper.get('a').attributes('href')).toBe('/api/schema/')
    expect(mocks.getOpenAPISchema).toHaveBeenCalledOnce()
    expect(wrapper.text()).toContain('DCIM')
    expect(wrapper.text()).toContain('/api/dcim/sites/')
    expect(wrapper.text()).toContain('IPAM')
    expect(wrapper.text()).not.toContain('GraphQL')
    expect(wrapper.text()).not.toContain('Reports')
  })
})
