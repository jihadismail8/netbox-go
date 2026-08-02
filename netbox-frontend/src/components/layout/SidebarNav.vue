<template>
  <aside
    class="fixed inset-y-0 left-0 z-30 flex flex-col bg-rich-black text-gray-300 transition-all duration-200"
    :class="collapsed ? 'w-16' : 'w-60'"
  >
    <!-- Logo -->
    <div class="flex h-14 items-center border-b border-white/10 px-4">
      <span v-if="!collapsed" class="text-lg font-bold text-white">NetBox</span>
      <span v-else class="text-lg font-bold text-white mx-auto">N</span>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-2">
      <SidebarMenuGroup
        v-for="group in navigation"
        :key="group.label"
        :group="group"
        :collapsed="collapsed"
      />
    </nav>

    <!-- Footer -->
    <div class="border-t border-white/10 p-2">
      <button
        class="flex w-full items-center justify-center rounded px-2 py-1.5 text-sm text-gray-400 hover:bg-white/5 hover:text-white"
        @click="uiStore.toggleSidebar()"
      >
        <ChevronsLeft v-if="!collapsed" :size="16" />
        <ChevronsRight v-else :size="16" />
        <span v-if="!collapsed" class="ml-2">Collapse</span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUiStore } from '@/stores/ui'
import { NAVIGATION, type MenuGroup } from '@/config/navigation'
import SidebarMenuGroup from './SidebarMenuGroup.vue'
import { ChevronsLeft, ChevronsRight } from '@lucide/vue'
import { getCoreResourceConfig } from '@/router/core-resource-registry'
import { usePermissions } from '@/composables/usePermissions'

const uiStore = useUiStore()
const { canView, canAdd } = usePermissions()
const navigation = computed<MenuGroup[]>(() =>
  NAVIGATION.map((group) => ({
    ...group,
    items: group.items
      .filter((item) => {
        const config = getCoreResourceConfig(item.route)
        return !config || canView(config.module, config.model)
      })
      .map((item) => {
        const config = getCoreResourceConfig(item.route)
        return {
          ...item,
          buttons: config && !canAdd(config.module, config.model) ? undefined : item.buttons,
        }
      }),
  })).filter((group) => group.items.length > 0),
)
const collapsed = computed(() => uiStore.sidebarCollapsed)
</script>
