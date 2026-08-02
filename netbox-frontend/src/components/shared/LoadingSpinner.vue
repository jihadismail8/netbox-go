<template>
  <div class="flex items-center justify-center p-8" :class="{ '!p-0': small }">
    <div
      class="animate-spin rounded-full border-3 border-gray-200 border-t-primary"
      :class="small ? 'h-4 w-4' : 'h-8 w-8'"
    />
  </div>
</template>
<script setup lang="ts">
defineProps<{ small?: boolean }>()
</script>
