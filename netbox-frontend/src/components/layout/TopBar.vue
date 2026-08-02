<template>
  <header
    class="sticky top-0 z-20 flex h-14 items-center justify-between border-b border-gray-200 bg-white px-4 dark:border-gray-700 dark:bg-gray-800"
  >
    <!-- Mobile menu toggle -->
    <button
      class="rounded p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 lg:hidden"
      @click="uiStore.toggleSidebar()"
    >
      <Menu :size="20" />
    </button>

    <div class="flex-1" />

    <!-- Right actions -->
    <div class="flex items-center gap-2">
      <ThemeToggle />
      <UserMenu />
    </div>
  </header>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/ui'
import { Menu } from '@lucide/vue'
import ThemeToggle from './ThemeToggle.vue'
import UserMenu from './UserMenu.vue'

const uiStore = useUiStore()
</script>
