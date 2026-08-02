<template>
  <div class="flex min-h-screen bg-gray-50 dark:bg-gray-900">
    <SidebarNav />
    <div
      class="flex flex-col flex-1 transition-all duration-200"
      :class="uiStore.sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-60'"
    >
      <TopBar />
      <main class="flex-1 p-6 overflow-y-auto">
        <RouterView v-slot="{ Component, route }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </Transition>
        </RouterView>
      </main>
    </div>
    <ToastContainer />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUiStore } from '@/stores/ui'
import SidebarNav from '@/components/layout/SidebarNav.vue'
import TopBar from '@/components/layout/TopBar.vue'
import ToastContainer from '@/components/shared/ToastContainer.vue'

const uiStore = useUiStore()

onMounted(() => {
  uiStore.initTheme()
})
</script>

<style>
.page-enter-active,
.page-leave-active {
  transition:
    opacity 0.15s ease,
    transform 0.15s ease;
}
.page-enter-from {
  opacity: 0;
  transform: translateY(8px);
}
.page-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
