import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { ToastNotification } from '@/types'

let nextId = 1

export const useNotificationStore = defineStore('notifications', () => {
  const toasts = ref<ToastNotification[]>([])

  function toast(message: string, type: ToastNotification['type'] = 'info', duration = 5000) {
    const id = nextId++
    toasts.value.push({ id, message, type, duration })
    if (duration > 0) {
      setTimeout(() => dismiss(id), duration)
    }
  }

  function success(message: string) {
    toast(message, 'success')
  }
  function error(message: string) {
    toast(message, 'error')
  }
  function warning(message: string) {
    toast(message, 'warning')
  }
  function info(message: string) {
    toast(message, 'info')
  }

  function dismiss(id: number) {
    const idx = toasts.value.findIndex((t) => t.id === id)
    if (idx > -1) toasts.value.splice(idx, 1)
  }

  return { toasts, toast, success, error, warning, info, dismiss }
})
