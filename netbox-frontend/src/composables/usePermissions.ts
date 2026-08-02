/**
 * usePermissions — RBAC permission checking for the UI.
 *
 * NetBox uses Django's auth model: users belong to groups, groups have
 * permissions formatted as `<app_label>.<action>_<model>`.
 * Actions: view, add, change, delete.
 *
 * Superusers bypass all checks. The auth store loads the user's effective
 * permissions on login and exposes them as a Set.
 */
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import type { AuthUserDTO } from '@/features/identity/api'

export type Action = 'view' | 'add' | 'change' | 'delete'

/**
 * Build the Django permission code: e.g. ('dcim', 'view', 'site') → 'dcim.view_site'
 */
export function permCode(app: string, action: Action, model: string): string {
  return `${app}.${action}_${model}`
}

export function hasPermission(
  user: AuthUserDTO | null,
  permissions: ReadonlySet<string>,
  action: Action,
  app: string,
  model: string,
): boolean {
  if (user?.is_superuser) return true
  return permissions.has(permCode(app, action, model))
}

/**
 * Check a single permission. Superusers always pass.
 */
export function usePermissions() {
  const auth = useAuthStore()
  const userPermissions = computed(() => auth.permissions)

  function can(action: Action, app: string, model: string): boolean {
    return hasPermission(auth.user, userPermissions.value, action, app, model)
  }

  function canView(app: string, model: string): boolean {
    return can('view', app, model)
  }
  function canAdd(app: string, model: string): boolean {
    return can('add', app, model)
  }
  function canChange(app: string, model: string): boolean {
    return can('change', app, model)
  }
  function canDelete(app: string, model: string): boolean {
    return can('delete', app, model)
  }

  return { can, canView, canAdd, canChange, canDelete }
}
