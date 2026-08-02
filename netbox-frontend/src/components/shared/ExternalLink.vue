<template>
  <a
    :href="href"
    target="_blank"
    rel="noopener noreferrer"
    class="inline-flex items-center gap-1 text-primary hover:underline"
  >
    <slot>{{ href }}</slot>
    <ExternalLinkIcon :size="size" class="text-primary/60" />
  </a>
</template>

<script setup lang="ts">
import { ExternalLink as ExternalLinkIcon } from '@lucide/vue'

withDefaults(
  defineProps<{
    href: string
    size?: number
  }>(),
  {
    size: 12,
  },
)
</script>
