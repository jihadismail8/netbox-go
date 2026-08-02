<template>
  <div>
    <div v-if="label" class="flex items-center justify-between mb-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ label }}
        <span v-if="required" class="text-red-500">*</span>
      </label>
      <div class="flex items-center gap-2">
        <div class="flex rounded-lg border border-gray-300 dark:border-gray-600">
          <button
            type="button"
            :class="mode === 'write' ? 'bg-primary text-white' : 'text-gray-500'"
            class="px-3 py-1 text-xs font-medium rounded-l-lg"
            @click="mode = 'write'"
          >
            Write
          </button>
          <button
            type="button"
            :class="mode === 'preview' ? 'bg-primary text-white' : 'text-gray-500'"
            class="px-3 py-1 text-xs font-medium rounded-r-lg border-l border-gray-300 dark:border-gray-600"
            @click="mode = 'preview'"
          >
            Preview
          </button>
        </div>
        <button
          type="button"
          class="text-xs text-gray-400 hover:text-primary"
          @click="showCheatSheet = !showCheatSheet"
        >
          <HelpCircle :size="14" />
        </button>
      </div>
    </div>

    <!-- Write mode -->
    <textarea
      v-show="mode === 'write'"
      :value="modelValue || ''"
      :rows="rows"
      :placeholder="placeholder || 'Write markdown...'"
      class="w-full px-3 py-2 text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-800"
      @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />

    <!-- Preview mode -->
    <div
      v-show="mode === 'preview'"
      class="w-full px-3 py-2 overflow-y-auto text-sm border border-gray-300 rounded-lg dark:border-gray-600 dark:bg-gray-800 min-h-[calc(1.5rem*3+1rem)]"
      :style="{ maxHeight: `calc(1.5rem * ${rows} + 1rem)` }"
    >
      <SanitizedMarkdown
        v-if="(modelValue || '').trim()"
        class="prose-sm prose max-w-none dark:prose-invert"
        :content="modelValue || ''"
      />
      <p v-else class="text-gray-400">Nothing to preview.</p>
    </div>

    <!-- Cheat sheet -->
    <div
      v-if="showCheatSheet"
      class="mt-1 rounded-lg border border-gray-200 bg-gray-50 p-3 text-xs text-gray-600 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-400"
    >
      <p class="mb-1 font-semibold">Markdown quick reference:</p>
      <ul class="space-y-0.5 font-mono">
        <li><strong>**bold**</strong> or <em>*italic*</em></li>
        <li># Heading 1 / ## Heading 2</li>
        <li>- Bullet list item</li>
        <li>1. Numbered list item</li>
        <li>`code` or ```code block```</li>
        <li>[link text](https://example.com)</li>
      </ul>
    </div>

    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { HelpCircle } from '@lucide/vue'
import SanitizedMarkdown from '@/components/shared/SanitizedMarkdown.vue'

withDefaults(
  defineProps<{
    modelValue?: string
    label?: string
    required?: boolean
    helpText?: string
    error?: string
    placeholder?: string
    rows?: number
  }>(),
  { required: false, rows: 6 },
)

defineEmits<{ 'update:modelValue': [value: string] }>()

const mode = ref<'write' | 'preview'>('write')
const showCheatSheet = ref(false)
</script>
