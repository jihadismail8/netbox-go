<template>
  <div class="flex flex-col items-center justify-center py-16 text-center">
    <div class="mb-4 rounded-full bg-red-100 p-4 dark:bg-red-900/30">
      <AlertCircle :size="32" class="text-red-600 dark:text-red-400" />
    </div>
    <h3 class="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
      {{ title }}
    </h3>
    <p class="mb-4 max-w-md text-sm text-gray-500 dark:text-gray-400">
      {{ message }}
    </p>
    <div v-if="$slots.default || retry" class="flex items-center gap-2">
      <button
        v-if="retry"
        class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-dark"
        @click="$emit('retry')"
      >
        <RotateCcw :size="16" /> Try Again
      </button>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { AlertCircle, RotateCcw } from '@lucide/vue'

withDefaults(
  defineProps<{
    title?: string
    message?: string
    retry?: boolean
  }>(),
  {
    title: 'Something went wrong',
    message: 'An error occurred while loading this data.',
    retry: false,
  },
)

defineEmits<{ retry: [] }>()
</script>
