import axios, { AxiosError, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { APP_CONFIG } from '@/config/app'

export interface ApiFailure {
  kind: 'api-failure'
  status: number
  detail: string
  fields: Record<string, unknown>
  retryable: boolean
}

function failureFrom(error: unknown): ApiFailure {
  if (!axios.isAxiosError(error)) {
    return {
      kind: 'api-failure',
      status: 0,
      detail: error instanceof Error ? error.message : 'The request failed.',
      fields: {},
      retryable: true,
    }
  }

  const axiosError = error as AxiosError<unknown>
  const status = axiosError.response?.status ?? 0
  const body = axiosError.response?.data
  const fields =
    body !== null && typeof body === 'object' && !Array.isArray(body)
      ? { ...(body as Record<string, unknown>) }
      : {}
  const detailValue = fields.detail
  const detail =
    typeof detailValue === 'string' && detailValue.trim()
      ? detailValue
      : status === 0
        ? 'The server could not be reached.'
        : `The request failed with status ${status}.`
  delete fields.detail

  return {
    kind: 'api-failure',
    status,
    detail,
    fields,
    retryable: status === 0 || status === 429 || status >= 500,
  }
}

export function isApiFailure(error: unknown): error is ApiFailure {
  return (
    error !== null &&
    typeof error === 'object' &&
    'kind' in error &&
    (error as { kind?: unknown }).kind === 'api-failure'
  )
}

export const http = axios.create({
  baseURL: APP_CONFIG.apiBaseUrl,
  timeout: 30000,
  withCredentials: true,
  withXSRFToken: true,
  xsrfCookieName: 'csrftoken',
  xsrfHeaderName: 'X-CSRFToken',
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    const failure = failureFrom(error)
    const requestURL = axios.isAxiosError(error) ? error.config?.url || '' : ''
    if (
      failure.status === 401 &&
      !requestURL.includes('/auth/session') &&
      !requestURL.includes('/auth/login') &&
      typeof window !== 'undefined'
    ) {
      window.dispatchEvent(new CustomEvent('netbox:unauthorized'))
    }
    return Promise.reject(failure)
  },
)

export async function request<T>(config: AxiosRequestConfig): Promise<AxiosResponse<T>> {
  return http.request<T>(config)
}

export default http
