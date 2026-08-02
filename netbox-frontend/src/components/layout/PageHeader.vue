<template>
  <div class="mb-4">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ title }}</h1>
        <nav
          v-if="breadcrumbs && breadcrumbs.length"
          class="mt-1 flex items-center text-sm text-gray-500"
        >
          <template v-for="(crumb, i) in breadcrumbs" :key="i">
            <RouterLink v-if="crumb.to" :to="crumb.to" class="hover:text-primary">{{
              crumb.label
            }}</RouterLink>
            <span v-else>{{ crumb.label }}</span>
            <ChevronRight v-if="i < breadcrumbs.length - 1" :size="12" class="mx-1.5" />
          </template>
        </nav>
      </div>
      <div v-if="$slots.actions" class="flex items-center gap-2">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ChevronRight } from '@lucide/vue'
import type { BreadcrumbItem } from '@/types'

withDefaults(
  defineProps<{
    title: string
    breadcrumbs?: BreadcrumbItem[]
  }>(),
  {
    breadcrumbs: () => [],
  },
)
</script>
