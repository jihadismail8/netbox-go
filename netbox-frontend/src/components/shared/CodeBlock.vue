<template>
  <div class="relative">
    <pre
      class="overflow-x-auto rounded-lg bg-gray-900 p-4 text-sm text-gray-100 dark:bg-black/40"
    ><code>{{ code }}</code></pre>
    <CopyToClipboard
      v-if="code"
      :text="code"
      class="absolute right-2 top-2 text-gray-400 hover:text-white"
    />
  </div>
</template>

<script setup lang="ts">
import CopyToClipboard from './CopyToClipboard.vue'

defineProps<{
  code: string
}>()
</script>
