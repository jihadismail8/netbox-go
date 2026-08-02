import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LoginPage from './LoginPage.vue'

const mocks = vi.hoisted(() => ({ login: vi.fn() }))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({ login: mocks.login }),
}))

async function mountLogin(next: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/login', component: { template: '<div />' } },
      { path: '/dcim/sites/', component: { template: '<div />' } },
      { path: '/', component: { template: '<div />' } },
    ],
  })
  await router.push({ path: '/login', query: { next } })
  await router.isReady()
  const wrapper = mount(LoginPage, { global: { plugins: [router] } })
  return { wrapper, router }
}

describe('LoginPage', () => {
  beforeEach(() => {
    mocks.login.mockReset().mockResolvedValue(undefined)
    localStorage.clear()
  })

  it('uses the server session API, clears the password, and returns to a local next route', async () => {
    const { wrapper, router } = await mountLogin('/dcim/sites/')
    const inputs = wrapper.findAll('input')
    await inputs[0].setValue('operator')
    await inputs[1].setValue('secret-not-persisted')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(mocks.login).toHaveBeenCalledWith({
      username: 'operator',
      password: 'secret-not-persisted',
    })
    expect(inputs[1].element.value).toBe('')
    expect(router.currentRoute.value.path).toBe('/dcim/sites/')
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('rejects an external-looking next route', async () => {
    const { wrapper, router } = await mountLogin('//attacker.example/')
    await wrapper.findAll('input')[0].setValue('operator')
    await wrapper.findAll('input')[1].setValue('secret')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/')
  })
})
