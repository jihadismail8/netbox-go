<template>
  <button
    class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700"
    :title="uiStore.darkMode ? 'Switch to light mode' : 'Switch to dark mode'"
    @click="uiStore.toggleDarkMode()"
  >
    <Sun v-if="uiStore.darkMode" :size="18" />
    <Moon v-else :size="18" />
  </button>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/ui'
import { Sun, Moon } from '@lucide/vue'
const uiStore = useUiStore()
</script>
