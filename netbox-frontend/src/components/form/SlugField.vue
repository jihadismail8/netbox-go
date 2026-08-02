<template>
  <div>
    <label class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">
      {{ label || 'Slug' }} <span v-if="required" class="text-red-500">*</span>
    </label>
    <div class="flex gap-1">
      <input
        v-model="localSlug"
        type="text"
        :required="required"
        :placeholder="placeholder"
        class="flex-1 px-3 py-2 font-mono text-sm border border-gray-300 rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-800"
        @input="handleInput"
      />
      <button
        v-if="!locked"
        type="button"
        class="px-3 text-gray-500 border border-gray-300 rounded-lg hover:bg-gray-50 dark:border-gray-600 dark:hover:bg-gray-700"
        title="Lock slug"
        @click="locked = true"
      >
        <Lock :size="14" />
      </button>
      <button
        v-else
        type="button"
        class="px-3 border rounded-lg border-primary text-primary hover:bg-primary/10"
        title="Unlock slug (auto-generate)"
        @click="enableAutomaticSlug"
      >
        <Unlock :size="14" />
      </button>
    </div>
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Lock, Unlock } from '@lucide/vue'

const props = defineProps<{
  modelValue?: string
  sourceValue?: unknown
  label?: string
  required?: boolean
  placeholder?: string
  helpText?: string
  error?: string
}>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

const localSlug = ref(props.modelValue || '')
const locked = ref(false)
const automaticallyGenerated = ref(!props.modelValue)

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
}

watch(
  () => props.modelValue,
  (value) => {
    const next = value || ''
    if (next !== localSlug.value) {
      localSlug.value = next
      automaticallyGenerated.value = !next
    }
  },
)

watch(
  () => props.sourceValue,
  (val) => {
    if (!locked.value && automaticallyGenerated.value && typeof val === 'string' && val) {
      localSlug.value = slugify(val)
      emit('update:modelValue', localSlug.value)
    }
  },
  { immediate: true },
)

function handleInput() {
  automaticallyGenerated.value = false
  emit('update:modelValue', localSlug.value)
}

function enableAutomaticSlug() {
  locked.value = false
  automaticallyGenerated.value = true
  if (typeof props.sourceValue === 'string') {
    localSlug.value = slugify(props.sourceValue)
    emit('update:modelValue', localSlug.value)
  }
}
</script>
