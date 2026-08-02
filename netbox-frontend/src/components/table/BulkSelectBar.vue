<template>
  <Transition name="slide-up">
    <div
      v-if="count > 0"
      class="sticky bottom-0 mx-4 mb-4 flex items-center justify-between rounded-lg bg-gray-900 px-4 py-3 text-white shadow-xl"
    >
      <div class="flex items-center gap-3">
        <CheckSquare :size="18" />
        <span class="text-sm font-medium">{{ count }} selected</span>
      </div>
      <div class="flex items-center gap-2">
        <slot />
        <button
          class="rounded px-3 py-1.5 text-sm text-gray-300 hover:bg-white/10"
          @click="$emit('clear')"
        >
          Clear
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { CheckSquare } from '@lucide/vue'
defineProps<{ count: number }>()
defineEmits<{ clear: [] }>()
</script>
