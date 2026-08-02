import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createSession, deleteSession, getSession } from './api'

const mocks = vi.hoisted(() => ({ request: vi.fn() }))

vi.mock('@/api/http', () => ({ request: mocks.request }))

const session = {
  user: {
    id: 1,
    username: 'operator',
    email: '',
    first_name: '',
    last_name: '',
    is_staff: true,
    is_superuser: false,
  },
  permissions: ['dcim.view_site'],
}

describe('identity API', () => {
  beforeEach(() => mocks.request.mockReset())

  it('establishes CSRF before submitting credentials', async () => {
    mocks.request.mockResolvedValueOnce({ data: {} }).mockResolvedValueOnce({ data: session })

    await expect(createSession({ username: 'operator', password: 'secret' })).resolves.toEqual(
      session,
    )
    expect(mocks.request.mock.calls).toEqual([
      [{ method: 'GET', url: '/auth/csrf/' }],
      [
        {
          method: 'POST',
          url: '/auth/login/',
          data: { username: 'operator', password: 'secret' },
        },
      ],
    ])
  })

  it('uses cookie-authenticated session and logout endpoints', async () => {
    mocks.request
      .mockResolvedValueOnce({ data: session })
      .mockResolvedValueOnce({ data: undefined })

    await expect(getSession()).resolves.toEqual(session)
    await expect(deleteSession()).resolves.toBeUndefined()
    expect(mocks.request).toHaveBeenNthCalledWith(1, { method: 'GET', url: '/auth/session/' })
    expect(mocks.request).toHaveBeenNthCalledWith(2, { method: 'POST', url: '/auth/logout/' })
  })
})
