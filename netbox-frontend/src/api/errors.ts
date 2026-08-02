import { isApiFailure } from '@/api/http'

export interface ValidationErrors {
  [field: string]: unknown
}

export function getErrorDetail(error: unknown, fallback: string): string {
  if (isApiFailure(error)) return error.detail || fallback
  if (error !== null && typeof error === 'object' && 'detail' in error) {
    const detail = (error as { detail?: unknown }).detail
    if (typeof detail === 'string' && detail.trim()) return detail
  }
  return fallback
}

export function getErrorRecord(error: unknown): ValidationErrors {
  if (isApiFailure(error)) {
    return Object.keys(error.fields).length > 0 ? error.fields : { detail: error.detail }
  }
  if (error !== null && typeof error === 'object' && !Array.isArray(error)) {
    return error as ValidationErrors
  }
  return { detail: getErrorDetail(error, String(error)) }
}
