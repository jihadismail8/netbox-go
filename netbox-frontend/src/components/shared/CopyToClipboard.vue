<template>
  <button
    type="button"
    class="inline-flex items-center gap-1 text-gray-400 transition-colors hover:text-primary"
    :title="copied ? 'Copied!' : 'Copy to clipboard'"
    @click="copy"
  >
    <component :is="copied ? Check : Copy" :size="size" />
    <span v-if="showLabel" class="text-xs">{{ copied ? 'Copied' : 'Copy' }}</span>
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Copy, Check } from '@lucide/vue'

const props = withDefaults(
  defineProps<{
    text: string
    size?: number
    showLabel?: boolean
  }>(),
  {
    size: 14,
    showLabel: false,
  },
)

const copied = ref(false)

async function copy() {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(props.text)
    } else {
      // Fallback for non-secure contexts
      const textarea = document.createElement('textarea')
      textarea.value = props.text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 1500)
  } catch {
    // Silently fail
  }
}
</script>
