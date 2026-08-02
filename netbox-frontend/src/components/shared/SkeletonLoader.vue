<template>
  <div class="w-full">
    <!-- Table row skeleton -->
    <div v-if="type === 'table'" class="space-y-2">
      <div v-for="i in rows" :key="i" class="flex gap-4">
        <div
          v-for="(c, idx) in cols"
          :key="idx"
          class="skeleton-pulse h-4 rounded"
          :style="{ width: typeof c === 'string' ? c : c.width }"
        />
      </div>
    </div>

    <!-- Card skeleton -->
    <div
      v-else-if="type === 'card'"
      class="rounded-lg border border-gray-200 p-6 dark:border-gray-700"
    >
      <div class="skeleton-pulse mb-4 h-6 w-1/3 rounded"></div>
      <div class="space-y-3">
        <div
          v-for="i in rows"
          :key="i"
          class="skeleton-pulse h-4 rounded"
          :style="{ width: `${90 - i * 5}%` }"
        ></div>
      </div>
    </div>

    <!-- Line skeleton -->
    <div v-else class="space-y-2">
      <div
        v-for="i in rows"
        :key="i"
        class="skeleton-pulse h-4 rounded"
        :style="{ width: `${100 - i * 8}%` }"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    type?: 'table' | 'card' | 'lines'
    rows?: number
    cols?: (string | { width: string })[]
  }>(),
  {
    type: 'lines',
    rows: 5,
    cols: () => ['60px', '1fr', '120px', '1fr', '80px'],
  },
)
</script>

<style scoped>
.skeleton-pulse {
  background: linear-gradient(
    90deg,
    rgb(229 231 235) 25%,
    rgb(243 244 246) 50%,
    rgb(229 231 235) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.5s infinite;
}
:global(.dark) .skeleton-pulse {
  background: linear-gradient(90deg, rgb(55 65 81) 25%, rgb(75 85 99) 50%, rgb(55 65 81) 75%);
  background-size: 200% 100%;
}
@keyframes skeleton-shimmer {
  0% {
    background-position: 200% 0;
  }
  100% {
    background-position: -200% 0;
  }
}
</style>
