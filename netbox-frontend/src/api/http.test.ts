import { describe, expect, it } from 'vitest'
import http, { isApiFailure } from './http'

describe('HTTP boundary', () => {
  it('uses credentialed CSRF-aware browser requests', () => {
    expect(http.defaults.withCredentials).toBe(true)
    expect(http.defaults.withXSRFToken).toBe(true)
    expect(http.defaults.xsrfCookieName).toBe('csrftoken')
    expect(http.defaults.xsrfHeaderName).toBe('X-CSRFToken')
  })

  it('recognizes only normalized API failures', () => {
    expect(
      isApiFailure({
        kind: 'api-failure',
        status: 403,
        detail: 'Forbidden',
        fields: {},
        retryable: false,
      }),
    ).toBe(true)
    expect(isApiFailure({ detail: 'raw response' })).toBe(false)
  })
})
