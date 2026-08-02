<template>
  <button
    :type="type"
    :disabled="disabled || confirming"
    class="inline-flex items-center gap-1.5 rounded-lg px-3 py-2 text-sm font-medium transition-colors"
    :class="buttonClass"
    @click="handleClick"
  >
    <component :is="confirming ? AlertTriangle : icon" v-if="icon || confirming" :size="14" />
    {{ confirming ? confirmLabel : label }}
  </button>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Component } from 'vue'
import { AlertTriangle } from '@lucide/vue'

const props = withDefaults(
  defineProps<{
    label: string
    confirmLabel?: string
    type?: 'button' | 'submit'
    disabled?: boolean
    icon?: Component
    danger?: boolean
    variant?: 'primary' | 'danger' | 'ghost'
  }>(),
  {
    type: 'button',
    disabled: false,
    danger: false,
    variant: 'primary',
    confirmLabel: 'Click to confirm',
  },
)

const emit = defineEmits<{ click: [] }>()

const confirming = ref(false)
let resetTimer: ReturnType<typeof setTimeout> | null = null

const buttonClass = computed(() => {
  if (confirming.value) {
    return 'bg-red-600 text-white hover:bg-red-700'
  }
  switch (props.variant) {
    case 'danger':
      return 'bg-red-600 text-white hover:bg-red-700 disabled:opacity-50'
    case 'ghost':
      return 'text-gray-700 border border-gray-300 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700'
    case 'primary':
    default:
      return 'bg-primary text-white hover:bg-primary-dark disabled:opacity-50'
  }
})

function handleClick() {
  if (confirming.value) {
    confirming.value = false
    if (resetTimer) clearTimeout(resetTimer)
    emit('click')
  } else {
    confirming.value = true
    resetTimer = setTimeout(() => {
      confirming.value = false
    }, 3000)
  }
}
</script>
