<template>
  <div class="group relative flex items-center">
    <RouterLink
      :to="item.route"
      class="flex flex-1 items-center rounded px-4 py-1.5 text-sm text-gray-300 hover:bg-primary/20 hover:text-white"
      :class="{
        'bg-primary/30 text-white font-medium': isActive,
        'justify-center px-2': collapsed,
      }"
    >
      <span class="truncate">{{ item.label }}</span>
    </RouterLink>

    <!-- Action buttons (Add, Import) -->
    <div
      v-if="item.buttons && !collapsed"
      class="ml-auto flex items-center pr-2 opacity-0 group-hover:opacity-100 transition-opacity"
    >
      <button
        v-for="btn in item.buttons"
        :key="btn.route"
        :title="btn.label"
        class="ml-0.5 rounded p-1 text-gray-500 hover:bg-white/10 hover:text-white"
        @click.prevent="$router.push(btn.route)"
      >
        <component :is="getIcon(btn.icon)" :size="13" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import type { MenuItem } from '@/config/navigation'
import { getNavigationIcon } from './navigationIcons'

const props = defineProps<{
  item: MenuItem
  collapsed: boolean
}>()

const route = useRoute()
const isActive = computed(() =>
  props.item.route === '/' ? route.path === '/' : route.path.startsWith(props.item.route),
)

function getIcon(name: string) {
  return getNavigationIcon(name)
}
</script>
