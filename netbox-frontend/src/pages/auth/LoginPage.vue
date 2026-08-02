<template>
  <div
    class="p-8 bg-white border border-gray-200 rounded-lg shadow-lg dark:border-gray-700 dark:bg-gray-800"
  >
    <form class="space-y-4" @submit.prevent="handleLogin">
      <div>
        <label class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300"
          >Username</label
        >
        <input
          v-model="username"
          type="text"
          required
          autofocus
          class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700"
        />
      </div>
      <div>
        <label class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300"
          >Password</label
        >
        <input
          v-model="password"
          type="password"
          required
          class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-700"
        />
      </div>
      <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
      <button
        type="submit"
        :disabled="loading"
        class="w-full px-4 py-2 text-sm font-medium text-white rounded-lg bg-primary hover:bg-primary-dark disabled:opacity-50"
      >
        {{ loading ? 'Signing in...' : 'Sign In' }}
      </button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getErrorDetail } from '@/api/errors'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    await authStore.login({ username: username.value, password: password.value })
    password.value = ''
    const requested = typeof route.query.next === 'string' ? route.query.next : ''
    const destination = requested.startsWith('/') && !requested.startsWith('//') ? requested : '/'
    await router.push(destination)
  } catch (caught: unknown) {
    error.value = getErrorDetail(caught, 'Login failed')
  } finally {
    password.value = ''
    loading.value = false
  }
}
</script>
