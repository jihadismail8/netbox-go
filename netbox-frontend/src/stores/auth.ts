import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  createSession,
  deleteSession,
  getSession,
  type AuthUserDTO,
  type LoginDTO,
  type SessionDTO,
} from '@/features/identity/api'

export type AuthUser = AuthUserDTO

export const useAuthStore = defineStore('auth', () => {
  const user = ref<AuthUser | null>(null)
  const permissions = ref<Set<string>>(new Set())
  const initialized = ref(false)
  let restorePromise: Promise<void> | null = null

  const isAuthenticated = computed(() => user.value !== null)
  const displayName = computed(() => {
    if (!user.value) return ''
    return user.value.first_name || user.value.username
  })

  function setSession(session: SessionDTO) {
    user.value = session.user
    permissions.value = new Set(session.permissions ?? [])
  }

  function clearSession() {
    user.value = null
    permissions.value = new Set()
  }

  function invalidateSession() {
    clearSession()
    initialized.value = true
  }

  function setPermissions(nextPermissions: string[]) {
    permissions.value = new Set(nextPermissions)
  }

  async function restoreSession(): Promise<void> {
    if (initialized.value) return
    if (restorePromise) return restorePromise

    restorePromise = (async () => {
      try {
        setSession(await getSession())
      } catch {
        clearSession()
      } finally {
        initialized.value = true
        restorePromise = null
      }
    })()

    return restorePromise
  }

  async function login(credentials: LoginDTO): Promise<void> {
    setSession(await createSession(credentials))
    initialized.value = true
  }

  async function logout(): Promise<void> {
    try {
      await deleteSession()
    } finally {
      clearSession()
      initialized.value = true
    }
  }

  return {
    user,
    permissions,
    initialized,
    isAuthenticated,
    displayName,
    setPermissions,
    invalidateSession,
    restoreSession,
    login,
    logout,
  }
})
