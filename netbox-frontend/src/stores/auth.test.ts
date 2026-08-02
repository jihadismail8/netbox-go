import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useAuthStore } from './auth'

const mocks = vi.hoisted(() => ({
  getSession: vi.fn(),
  createSession: vi.fn(),
  deleteSession: vi.fn(),
}))

vi.mock('@/features/identity/api', () => ({
  getSession: mocks.getSession,
  createSession: mocks.createSession,
  deleteSession: mocks.deleteSession,
}))

const session = {
  user: {
    id: 7,
    username: 'operator',
    email: 'operator@example.test',
    first_name: 'Op',
    last_name: 'Erator',
    is_staff: true,
    is_superuser: false,
  },
  permissions: ['dcim.view_site'],
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mocks.getSession.mockReset().mockResolvedValue(session)
    mocks.createSession.mockReset().mockResolvedValue(session)
    mocks.deleteSession.mockReset().mockResolvedValue(undefined)
    localStorage.clear()
  })

  it('restores one server-owned session and its effective permissions', async () => {
    const store = useAuthStore()
    await Promise.all([store.restoreSession(), store.restoreSession()])

    expect(mocks.getSession).toHaveBeenCalledOnce()
    expect(store.user?.username).toBe('operator')
    expect(store.permissions.has('dcim.view_site')).toBe(true)
    expect(localStorage.length).toBe(0)
  })

  it('logs in without persisting credentials and clears state on logout', async () => {
    const store = useAuthStore()
    await store.login({ username: 'operator', password: 'not-stored' })
    expect(mocks.createSession).toHaveBeenCalledWith({
      username: 'operator',
      password: 'not-stored',
    })
    expect(localStorage.length).toBe(0)

    await store.logout()
    expect(mocks.deleteSession).toHaveBeenCalledOnce()
    expect(store.user).toBeNull()
    expect(store.permissions.size).toBe(0)
  })

  it('clears local session state even when the logout request fails', async () => {
    const store = useAuthStore()
    await store.login({ username: 'operator', password: 'not-stored' })
    mocks.deleteSession.mockRejectedValueOnce(new Error('network unavailable'))

    await expect(store.logout()).rejects.toThrow('network unavailable')
    expect(store.user).toBeNull()
    expect(store.permissions.size).toBe(0)
  })
})
