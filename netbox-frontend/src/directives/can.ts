/**
 * v-can directive — Hide/disable elements based on user permissions.
 *
 * Usage:
 *   <button v-can="{ action: 'delete', app: 'dcim', model: 'site' }">Delete</button>
 *   <!-- Hide entirely (default) -->
 *   <button v-can.disable="{ action: 'add', app: 'dcim', model: 'site' }">Add</button>
 *   <!-- Disable instead of hide -->
 */
import type { Directive, DirectiveBinding } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { hasPermission, type Action } from '@/composables/usePermissions'

interface CanBinding {
  action: Action
  app: string
  model: string
}

function checkPermission(binding: CanBinding): boolean {
  const auth = useAuthStore()
  return hasPermission(auth.user, auth.permissions, binding.action, binding.app, binding.model)
}

export const vCan: Directive<HTMLElement, CanBinding> = {
  mounted(el: HTMLElement, binding: DirectiveBinding<CanBinding>) {
    const allowed = checkPermission(binding.value)
    const mode = binding.modifiers.disable ? 'disable' : 'hide'

    if (!allowed) {
      if (mode === 'hide') {
        el.style.display = 'none'
      } else {
        el.setAttribute('disabled', 'true')
        el.classList.add('opacity-50', 'cursor-not-allowed', 'pointer-events-none')
      }
    }
  },
}
