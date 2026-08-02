<template>
  <div class="flex min-h-[60vh] flex-col items-center justify-center text-center">
    <div class="text-7xl font-bold text-red-500">403</div>
    <h1 class="mt-4 text-2xl font-bold text-gray-900 dark:text-white">Access Forbidden</h1>
    <p class="mt-2 text-sm text-gray-500">You don't have permission to access this page.</p>
    <RouterLink
      to="/"
      class="mt-6 inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-dark"
    >
      <Home :size="16" /> Go Home
    </RouterLink>
  </div>
</template>

<script setup lang="ts">
import { Home } from '@lucide/vue'
</script>
