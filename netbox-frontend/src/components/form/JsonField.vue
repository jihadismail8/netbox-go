<template>
  <div>
    <div v-if="label" class="flex items-center justify-between mb-1">
      <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ label }}
        <span v-if="required" class="text-red-500">*</span>
      </label>
      <button
        type="button"
        :disabled="!isValid"
        class="text-xs text-primary hover:underline disabled:opacity-40"
        @click="togglePretty"
      >
        {{ pretty ? 'Minify' : 'Pretty' }}
      </button>
    </div>
    <textarea
      :value="text"
      :rows="rows"
      :placeholder="placeholder || '{}'"
      :class="error ? 'border-red-400' : 'border-gray-300'"
      class="w-full px-3 py-2 font-mono text-sm border rounded-lg focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-gray-600 dark:bg-gray-800"
      @input="onInput"
    />
    <p v-if="error" class="mt-1 text-sm text-red-500">{{ error }}</p>
    <p v-else-if="helpText" class="mt-1 text-xs text-gray-400">{{ helpText }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue?: unknown
    label?: string
    required?: boolean
    helpText?: string
    error?: string
    placeholder?: string
    rows?: number
  }>(),
  { required: false, rows: 6 },
)

const emit = defineEmits<{ 'update:modelValue': [value: unknown] }>()

const text = ref('')
const isValid = ref(true)
const pretty = ref(true)

// Initialize from modelValue
watch(
  () => props.modelValue,
  (val) => {
    if (val === null || val === undefined) {
      text.value = ''
      isValid.value = true
      return
    }
    if (typeof val === 'object') {
      text.value = JSON.stringify(val, null, pretty.value ? 2 : 0)
      isValid.value = true
    } else if (typeof val === 'string') {
      text.value = val
      validate(val)
    }
  },
  { immediate: true },
)

function validate(str: string): boolean {
  if (str.trim() === '') {
    isValid.value = true
    return true
  }
  try {
    JSON.parse(str)
    isValid.value = true
    return true
  } catch {
    isValid.value = false
    return false
  }
}

function onInput(event: Event) {
  const raw = (event.target as HTMLTextAreaElement).value
  text.value = raw
  if (raw.trim() === '') {
    isValid.value = true
    emit('update:modelValue', null)
    return
  }
  if (validate(raw)) {
    emit('update:modelValue', JSON.parse(raw))
  }
}

function togglePretty() {
  if (!isValid.value || text.value.trim() === '') return
  pretty.value = !pretty.value
  const parsed = JSON.parse(text.value)
  text.value = JSON.stringify(parsed, null, pretty.value ? 2 : 0)
}
</script>
