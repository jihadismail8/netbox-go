<template>
  <div class="overflow-hidden rounded-lg border border-gray-200 dark:border-gray-700">
    <table class="w-full text-sm">
      <tbody>
        <tr
          v-for="row in rows"
          :key="row.key"
          class="border-b border-gray-100 last:border-b-0 dark:border-gray-800"
        >
          <td
            class="w-2/5 bg-gray-50 px-4 py-2.5 align-top text-xs font-semibold uppercase tracking-wide text-gray-500 dark:bg-gray-900/40"
          >
            {{ row.label }}
          </td>
          <td class="px-4 py-2.5 align-top text-gray-800 dark:text-gray-200">
            <slot :name="'cell-' + row.key" :value="row.value" :row="item">
              <template v-if="row.presentation === 'status' && isChoice(row.value)">
                <StatusBadge :status="String(row.value.value)" :label="row.value.label" />
              </template>
              <template v-else-if="row.presentation === 'status'">
                <StatusBadge :status="String(row.value ?? '')" />
              </template>
              <template v-else-if="isBoolean(row.value)">
                <span class="inline-flex items-center gap-1.5">
                  <component
                    :is="row.value ? Check : X"
                    :size="16"
                    :class="row.value ? 'text-green-600' : 'text-gray-400'"
                  />
                  {{ row.value ? 'Yes' : 'No' }}
                </span>
              </template>
              <template v-else-if="isLink(row.value)">
                <RouterLink :to="buildLink(row.value)" class="text-primary hover:underline">
                  {{ row.value.display || row.value.id }}
                </RouterLink>
                <CopyToClipboard v-if="row.value.id" :text="String(row.value.id)" class="ml-1" />
              </template>
              <template v-else-if="row.presentation === 'markdown'">
                <SanitizedMarkdown
                  class="prose-sm prose max-w-none dark:prose-invert"
                  :content="typeof row.value === 'string' ? row.value : ''"
                />
              </template>
              <template v-else>{{ formatValue(row.value) }}</template>
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Check, X } from '@lucide/vue'
import type { DetailFieldDef } from '@/types'
import {
  isChoiceDTO,
  isCoreReference,
  resourceDetailField,
  type ChoiceDTO,
  type CoreReference,
  type CoreDetailFieldValue,
  type CoreResourceDTO,
} from '@/features/core/resources'
import StatusBadge from '@/components/shared/StatusBadge.vue'
import CopyToClipboard from '@/components/shared/CopyToClipboard.vue'
import SanitizedMarkdown from '@/components/shared/SanitizedMarkdown.vue'
import { routeForObjectURL } from '@/features/core/links'

const props = defineProps<{
  item: CoreResourceDTO
  fields: DetailFieldDef[]
  skipKeys?: string[]
}>()

interface PropertyRow {
  key: string
  label: string
  value: CoreDetailFieldValue | undefined
  presentation: NonNullable<DetailFieldDef['presentation']>
}

const rows = computed<PropertyRow[]>(() => {
  const skip = new Set(props.skipKeys || [])
  return props.fields
    .filter((field) => !skip.has(field.key))
    .map((field) => ({
      key: field.key,
      label: field.label,
      value: resourceDetailField(props.item, field.key),
      presentation: field.presentation ?? 'default',
    }))
})

function isBoolean(v: unknown): v is boolean {
  return typeof v === 'boolean'
}

function isLink(v: unknown): v is CoreReference {
  return isCoreReference(v)
}

function isChoice(v: unknown): v is ChoiceDTO<string | number> {
  return isChoiceDTO(v)
}

function buildLink(obj: CoreReference): string {
  return routeForObjectURL(obj.url)
}

function formatValue(value: CoreDetailFieldValue | undefined): string {
  if (value === null || value === undefined || value === '') return '—'
  if (isChoiceDTO(value)) return value.label
  if (isCoreReference(value)) return value.display
  if (typeof value === 'boolean') return value ? 'Yes' : 'No'
  return String(value)
}
</script>
