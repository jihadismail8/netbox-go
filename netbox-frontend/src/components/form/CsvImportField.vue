<template>
  <div class="space-y-4">
    <div class="flex gap-2">
      <button
        type="button"
        :class="
          inputMode === 'paste' ? 'border-primary text-primary' : 'border-gray-300 text-gray-500'
        "
        class="border-b-2 px-3 py-1.5 text-sm font-medium"
        @click="inputMode = 'paste'"
      >
        Paste CSV
      </button>
      <button
        type="button"
        :class="
          inputMode === 'file' ? 'border-primary text-primary' : 'border-gray-300 text-gray-500'
        "
        class="border-b-2 px-3 py-1.5 text-sm font-medium"
        @click="inputMode = 'file'"
      >
        Upload File
      </button>
    </div>

    <textarea
      v-if="inputMode === 'paste'"
      :value="csvText"
      rows="10"
      placeholder="name,description&#10;Site A,Main DC"
      class="w-full px-3 py-2 font-mono text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-800"
      @input="onTextInput"
    />

    <div
      v-else
      :class="dragOver ? 'border-primary bg-primary/5' : 'border-gray-300'"
      class="flex flex-col items-center justify-center py-8 border-2 border-dashed rounded-lg"
      @dragover.prevent="dragOver = true"
      @dragleave.prevent="dragOver = false"
      @drop.prevent="onDrop"
    >
      <Upload :size="28" class="text-gray-400" />
      <p class="mt-1 text-sm text-gray-500">Drag &amp; drop a CSV file, or</p>
      <label class="mt-1 cursor-pointer text-sm text-primary hover:underline"
        >Browse<input type="file" accept=".csv,text/csv" class="hidden" @change="onFileSelect"
      /></label>
      <p v-if="fileName" class="mt-1 text-xs text-primary">{{ fileName }}</p>
    </div>

    <div
      v-if="headers.length > 0"
      class="rounded-lg border border-gray-200 p-4 dark:border-gray-700"
    >
      <h4 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-300">Column Mapping</h4>
      <div class="grid grid-cols-1 gap-2 md:grid-cols-2">
        <div v-for="header in headers" :key="header" class="flex items-center gap-2">
          <span class="w-32 truncate text-xs font-mono text-gray-500">{{ header }}</span>
          <select
            v-model="mapping[header]"
            class="flex-1 rounded border border-gray-300 px-2 py-1 text-xs dark:border-gray-600 dark:bg-gray-800"
          >
            <option value="">— Skip —</option>
            <option v-for="f in fields" :key="f.key" :value="f.key">
              {{ f.label || f.key }}{{ f.required ? ' *' : '' }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <div
      v-if="mappedRows.length > 0"
      class="rounded-lg border border-gray-200 p-4 dark:border-gray-700"
    >
      <h4 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-300">
        Preview ({{ mappedRows.length }} rows, showing first {{ previewRows.length }})
      </h4>
      <div class="overflow-x-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="px-2 py-1 text-left text-gray-500">#</th>
              <th
                v-for="key in mappedKeys"
                :key="key"
                class="px-2 py-1 text-left text-gray-700 dark:text-gray-300"
              >
                {{ key }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(row, idx) in previewRows"
              :key="idx"
              class="border-b border-gray-100 dark:border-gray-800"
            >
              <td class="px-2 py-1 text-gray-400">{{ idx + 1 }}</td>
              <td
                v-for="key in mappedKeys"
                :key="key"
                class="px-2 py-1 text-gray-700 dark:text-gray-300"
              >
                {{ row[key] ?? '—' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <p v-if="error" class="text-sm text-red-500">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Upload } from '@lucide/vue'
import { parseCsv } from '@/utils/csv'

export interface FormFieldLike {
  key: string
  label?: string
  required?: boolean
}

const props = defineProps<{ modelValue?: string; fields: FormFieldLike[]; error?: string }>()
const emit = defineEmits<{
  'update:modelValue': [value: string]
  parsed: [rows: Record<string, string>[]]
}>()

const inputMode = ref<'paste' | 'file'>('paste')
const csvText = ref(props.modelValue || '')
const fileName = ref('')
const dragOver = ref(false)
const mapping = ref<Record<string, string>>({})

const parsedRaw = computed(() => (csvText.value.trim() ? parseCsv(csvText.value) : []))
const headers = computed(() => parsedRaw.value[0] || [])
const mappedKeys = computed(() => Object.values(mapping.value).filter((v) => v) as string[])

const mappedRows = computed<Record<string, string>[]>(() => {
  if (parsedRaw.value.length < 2) return []
  return parsedRaw.value.slice(1).map((row) => {
    const obj: Record<string, string> = {}
    for (let i = 0; i < headers.value.length; i++) {
      const target = mapping.value[headers.value[i]]
      if (target) obj[target] = row[i] || ''
    }
    return obj
  })
})
const previewRows = computed(() => mappedRows.value.slice(0, 5))

watch(
  headers,
  (hs) => {
    for (const h of hs) {
      if (!mapping.value[h]) {
        const match = props.fields.find(
          (f) => f.key === h || f.key === h.replace(/\s+/g, '_').toLowerCase(),
        )
        mapping.value[h] = match?.key || h
      }
    }
  },
  { immediate: true },
)

watch(mappedRows, (rows) => emit('parsed', rows))

function onTextInput(event: Event) {
  csvText.value = (event.target as HTMLTextAreaElement).value
  emit('update:modelValue', csvText.value)
}

function onFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files?.length) return
  const file = input.files[0]
  fileName.value = file.name
  file.text().then((text) => {
    csvText.value = text
    emit('update:modelValue', text)
  })
}

function onDrop(event: DragEvent) {
  dragOver.value = false
  const files = event.dataTransfer?.files
  if (!files?.length) return
  const file = files[0]
  fileName.value = file.name
  file.text().then((text) => {
    csvText.value = text
    emit('update:modelValue', text)
  })
}
</script>
