<template>
  <span :title="absoluteTime" class="text-gray-500">{{ relative }}</span>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    datetime: string | Date
    interval?: number
  }>(),
  {
    interval: 30000,
  },
)

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

const timestamp = computed(() => {
  const d = typeof props.datetime === 'string' ? new Date(props.datetime) : props.datetime
  return isNaN(d.getTime()) ? 0 : d.getTime()
})

const relative = computed(() => {
  const diff = now.value - timestamp.value
  if (diff < 0) return 'in the future'
  const seconds = Math.floor(diff / 1000)
  if (seconds < 60) return 'just now'
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes} minute${minutes === 1 ? '' : 's'} ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} hour${hours === 1 ? '' : 's'} ago`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days} day${days === 1 ? '' : 's'} ago`
  const months = Math.floor(days / 30)
  if (months < 12) return `${months} month${months === 1 ? '' : 's'} ago`
  const years = Math.floor(months / 12)
  return `${years} year${years === 1 ? '' : 's'} ago`
})

const absoluteTime = computed(() => {
  if (!timestamp.value) return ''
  return new Date(timestamp.value).toLocaleString()
})

onMounted(() => {
  if (props.interval > 0) {
    timer = setInterval(() => {
      now.value = Date.now()
    }, props.interval)
  }
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>
