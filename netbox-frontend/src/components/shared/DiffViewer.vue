<template>
  <div class="rounded-lg border border-gray-200 dark:border-gray-700">
    <table class="w-full text-sm">
      <tbody>
        <tr
          v-for="diff in diffs"
          :key="diff.key"
          class="border-b border-gray-100 last:border-b-0 dark:border-gray-800"
        >
          <td class="w-1/3 px-3 py-1.5 align-top font-mono text-xs text-gray-500">
            {{ diff.key }}
          </td>
          <td v-if="diff.type === 'added'" class="px-3 py-1.5">
            <span class="text-green-700 dark:text-green-400">
              <Plus :size="11" class="inline" /> {{ formatValue(diff.newValue) }}
            </span>
          </td>
          <td v-else-if="diff.type === 'removed'" class="px-3 py-1.5">
            <span class="text-red-700 line-through dark:text-red-400">
              <Minus :size="11" class="inline" /> {{ formatValue(diff.oldValue) }}
            </span>
          </td>
          <td v-else class="px-3 py-1.5">
            <span class="text-red-700 line-through dark:text-red-400">{{
              formatValue(diff.oldValue)
            }}</span>
            <ArrowRight :size="11" class="mx-1 inline text-gray-400" />
            <span class="text-green-700 dark:text-green-400">{{ formatValue(diff.newValue) }}</span>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Plus, Minus, ArrowRight } from '@lucide/vue'

interface DiffEntry {
  key: string
  type: 'added' | 'removed' | 'changed'
  oldValue?: unknown
  newValue?: unknown
}

const props = defineProps<{
  oldValue?: Record<string, unknown> | null
  newValue?: Record<string, unknown> | null
}>()

const diffs = computed<DiffEntry[]>(() => {
  const oldObj = props.oldValue || {}
  const newObj = props.newValue || {}
  const allKeys = new Set([...Object.keys(oldObj), ...Object.keys(newObj)])
  const result: DiffEntry[] = []
  const skipKeys = ['last_updated', 'created']
  for (const key of allKeys) {
    if (skipKeys.includes(key)) continue
    const inOld = key in oldObj
    const inNew = key in newObj
    if (inNew && !inOld) {
      result.push({ key, type: 'added', newValue: newObj[key] })
    } else if (inOld && !inNew) {
      result.push({ key, type: 'removed', oldValue: oldObj[key] })
    } else if (!isEqual(oldObj[key], newObj[key])) {
      result.push({ key, type: 'changed', oldValue: oldObj[key], newValue: newObj[key] })
    }
  }
  return result
})

function isEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true
  if (typeof a !== typeof b) return false
  if (typeof a === 'object' && a !== null && b !== null) {
    return JSON.stringify(a) === JSON.stringify(b)
  }
  return false
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return '—'
  if (typeof value === 'object') {
    const obj = value as Record<string, unknown>
    if (obj.display) return String(obj.display)
    if (obj.name) return String(obj.name)
    return JSON.stringify(value)
  }
  if (typeof value === 'boolean') return value ? 'True' : 'False'
  return String(value)
}
</script>
