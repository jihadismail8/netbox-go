import { CORE_PROFILE_RESOURCES } from '@/features/core/manifest'

/** Convert a profile REST object URL to its declared Vue detail route. */
export function routeForObjectURL(value: string): string {
  let pathname: string
  try {
    pathname = new URL(value, 'http://netbox.local').pathname
  } catch {
    return '#'
  }
  if (!pathname.startsWith('/api/')) return '#'
  const route = pathname.slice('/api'.length)
  const declared = CORE_PROFILE_RESOURCES.some((resource) => {
    if (!route.startsWith(resource.routePath)) return false
    const suffix = route.slice(resource.routePath.length)
    return /^\d+\/$/.test(suffix)
  })
  return declared ? route : '#'
}
