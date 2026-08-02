<template>
  <Teleport to="body">
    <div v-if="modelValue" class="fixed inset-0 z-50 flex items-center justify-center">
      <div class="absolute inset-0 bg-black/50" @click="close" />
      <div class="relative z-10 w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-gray-800">
        <div
          class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-gray-700"
        >
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ title }}</h3>
          <button class="text-gray-400 hover:text-gray-600" @click="close">
            <X :size="18" />
          </button>
        </div>
        <div class="px-6 py-4">
          <slot />
        </div>
        <div
          v-if="$slots.footer"
          class="flex justify-end gap-2 border-t border-gray-200 px-6 py-4 dark:border-gray-700"
        >
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { X } from '@lucide/vue'

defineProps<{ modelValue: boolean; title: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()

function close() {
  emit('update:modelValue', false)
}
</script>
