<template>
  <!-- This is the single audited raw-HTML boundary. renderMarkdown escapes all
       input before adding its fixed, allow-listed markup. -->
  <!-- eslint-disable-next-line vue/no-v-html -->
  <div v-html="html" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { renderMarkdown } from '@/utils/markdown'

const props = defineProps<{ content: string }>()
const html = computed(() => renderMarkdown(props.content))
</script>
