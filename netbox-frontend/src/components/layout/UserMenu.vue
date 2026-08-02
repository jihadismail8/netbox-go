<template>
  <div class="relative">
    <button
      class="flex items-center rounded-lg p-1.5 text-sm text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-700"
      @click="open = !open"
    >
      <div
        class="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-xs font-bold text-white"
      >
        {{ initials }}
      </div>
      <span class="ml-2 hidden md:inline">{{ authStore.displayName || 'Guest' }}</span>
      <ChevronDown :size="14" class="ml-1" />
    </button>

    <div
      v-if="open"
      class="absolute right-0 mt-2 w-48 rounded-lg border border-gray-200 bg-white py-1 shadow-lg dark:border-gray-700 dark:bg-gray-800"
      @click="open = false"
    >
      <button
        class="block w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20"
        @click="handleLogout"
      >
        Log Out
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ChevronDown } from '@lucide/vue'

const authStore = useAuthStore()
const router = useRouter()
const open = ref(false)

const initials = computed(() => {
  const name = authStore.user?.first_name || authStore.user?.username || 'G'
  return name.charAt(0).toUpperCase()
})

async function handleLogout() {
  try {
    await authStore.logout()
  } catch {
    // The store clears local state in all cases; continue to the login screen.
  } finally {
    await router.push('/login')
  }
}
</script>
