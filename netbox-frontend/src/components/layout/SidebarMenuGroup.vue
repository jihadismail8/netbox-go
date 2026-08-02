<template>
  <div class="mb-1">
    <!-- Group header (hidden when collapsed) -->
    <button
      v-if="!collapsed"
      class="flex w-full items-center px-4 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-500 hover:text-gray-300"
      @click="isOpen = !isOpen"
    >
      <component :is="getIcon(group.icon)" :size="14" class="mr-2" />
      <span class="flex-1 text-left">{{ group.label }}</span>
      <ChevronDown :size="12" :class="{ 'rotate-180': isOpen }" class="transition-transform" />
    </button>

    <!-- Collapsed: show icon only -->
    <button
      v-else
      class="mx-auto flex w-10 items-center justify-center rounded py-2 text-gray-400 hover:bg-white/5 hover:text-white"
      :title="group.label"
      @click="isOpen = !isOpen"
    >
      <component :is="getIcon(group.icon)" :size="18" />
    </button>

    <!-- Menu items -->
    <div
      v-show="isOpen || collapsed"
      class="mt-0.5"
      :class="{ 'absolute left-16 z-50 w-56 rounded-r-lg bg-rich-black shadow-xl p-1': collapsed }"
    >
      <SidebarMenuItem
        v-for="item in group.items"
        :key="item.route"
        :item="item"
        :collapsed="collapsed"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { MenuGroup } from '@/config/navigation'
import SidebarMenuItem from './SidebarMenuItem.vue'
import { ChevronDown } from '@lucide/vue'
import { getNavigationIcon } from './navigationIcons'

defineProps<{
  group: MenuGroup
  collapsed: boolean
}>()

const isOpen = ref(true)

function getIcon(name: string) {
  return getNavigationIcon(name)
}
</script>
