<template>
  <div class="flex min-h-[60vh] flex-col items-center justify-center text-center">
    <LoadingSpinner />
    <p class="mt-4 text-sm text-gray-500">Logging out...</p>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  try {
    await authStore.logout()
  } catch {
    // The store clears local state in all cases; continue to the login screen.
  } finally {
    await router.push('/login')
  }
})
</script>
