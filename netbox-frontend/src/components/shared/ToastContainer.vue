<template>
  <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
    <TransitionGroup name="toast">
      <div
        v-for="toast in notifications.toasts"
        :key="toast.id"
        class="flex items-center rounded-lg px-4 py-3 shadow-lg"
        :class="toastClasses(toast.type)"
      >
        <component :is="toastIcon(toast.type)" :size="18" class="mr-2" />
        <span class="text-sm font-medium">{{ toast.message }}</span>
        <button class="ml-3 opacity-70 hover:opacity-100" @click="notifications.dismiss(toast.id)">
          <X :size="14" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { useNotificationStore } from '@/stores/notifications'
import { CheckCircle, AlertCircle, AlertTriangle, Info, X } from '@lucide/vue'

const notifications = useNotificationStore()

function toastClasses(type: string) {
  return (
    {
      success: 'bg-green-600 text-white',
      error: 'bg-red-600 text-white',
      warning: 'bg-yellow-500 text-white',
      info: 'bg-blue-600 text-white',
    }[type] || 'bg-gray-700 text-white'
  )
}

function toastIcon(type: string) {
  return (
    {
      success: CheckCircle,
      error: AlertCircle,
      warning: AlertTriangle,
      info: Info,
    }[type] || Info
  )
}
</script>
