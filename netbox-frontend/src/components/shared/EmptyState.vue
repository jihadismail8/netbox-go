<template>
  <div class="flex flex-col items-center justify-center p-12 text-center">
    <component :is="icon || Inbox" :size="48" class="mb-3 text-gray-300" />
    <p class="text-lg font-medium text-gray-500">{{ message || 'No results found' }}</p>
    <p v-if="hint" class="mt-1 text-sm text-gray-400">{{ hint }}</p>
    <slot name="action" />
  </div>
</template>
<script setup lang="ts">
import type { Component } from 'vue'
import { Inbox } from '@lucide/vue'
defineProps<{ message?: string; hint?: string; icon?: Component }>()
</script>
