import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import CoreResourceDeleteView from './CoreResourceDeleteView.vue'
import type { ModelConfig } from '@/types'

const mocks = vi.hoisted(() => ({ fetchById: vi.fn(), deleteItem: vi.fn() }))

vi.mock('@/composables/useCoreResource', () => ({
  useCoreResource: () => ({
    fetchById: mocks.fetchById,
    deleteItem: mocks.deleteItem,
  }),
}))

const config: ModelConfig = {
  module: 'dcim',
  model: 'interface',
  display_name: 'Interface',
  display_name_plural: 'Interfaces',
  apiPath: '/dcim/interfaces/',
  routePath: '/dcim/interfaces/',
  writableFields: ['device', 'name'],
  columns: [
    { key: 'name', label: 'Name' },
    { key: 'count_ipaddresses', label: 'IP Addresses' },
  ],
  detailFields: [
    { key: 'name', label: 'Name' },
    { key: 'count_ipaddresses', label: 'IP Addresses' },
  ],
  filters: [],
  fields: [],
}

async function mountDeleteView() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/dcim/interfaces/:id/delete/', component: { template: '<div />' } },
      { path: '/dcim/interfaces/:id/', component: { template: '<div />' } },
      { path: '/dcim/interfaces/', component: { template: '<div />' } },
      { path: '/:pathMatch(.*)*', component: { template: '<div />' } },
    ],
  })
  await router.push('/dcim/interfaces/4/delete/')
  await router.isReady()
  const wrapper = mount(CoreResourceDeleteView, {
    props: { config },
    global: { plugins: [createPinia(), router], stubs: { RouterLink: true } },
  })
  await flushPromises()
  return { wrapper, router }
}

describe('CoreResourceDeleteView Interface cascade warning', () => {
  beforeEach(() => {
    mocks.fetchById.mockReset().mockResolvedValue({
      id: 4,
      display: 'edge-1: xe-0/0/0',
      name: 'xe-0/0/0',
      count_ipaddresses: 2,
    })
    mocks.deleteItem.mockReset().mockResolvedValue(undefined)
  })

  it('warns explicitly and allows the operator to cancel', async () => {
    const { wrapper, router } = await mountDeleteView()
    expect(wrapper.text()).toContain(
      'Deleting this Interface also deletes 2 assigned IP addresses.',
    )

    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Cancel')!
      .trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/dcim/interfaces/4/')
    expect(mocks.deleteItem).not.toHaveBeenCalled()
  })

  it('deletes only after explicit confirmation', async () => {
    const { wrapper, router } = await mountDeleteView()
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Confirm Delete')!
      .trigger('click')
    await flushPromises()

    expect(mocks.deleteItem).toHaveBeenCalledWith(4)
    expect(router.currentRoute.value.path).toBe('/dcim/interfaces/')
  })
})
