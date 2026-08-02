<template>
  <div>
    <label v-if="label" class="block mb-1 text-sm font-medium text-gray-700 dark:text-gray-300">{{
      label
    }}</label>
    <div
      class="flex flex-wrap gap-1 p-2 border border-gray-300 rounded-lg dark:border-gray-600 dark:bg-gray-800"
    >
      <span
        v-for="tag in tags"
        :key="tag"
        class="inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium"
        :style="{ backgroundColor: '#' + tagColor(tag) + '20', color: '#' + tagColor(tag) }"
      >
        {{ tag }}
        <button class="hover:opacity-70" @click.prevent="removeTag(tag)">
          <X :size="10" />
        </button>
      </span>
      <input
        v-model="input"
        type="text"
        :placeholder="tags.length === 0 ? placeholder || 'Add tags...' : ''"
        class="flex-1 p-0 text-sm bg-transparent border-0 outline-none"
        @keydown.enter.prevent="addTag"
        @keydown.delete="removeLast"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { X } from '@lucide/vue'

const props = defineProps<{
  modelValue?: string[]
  label?: string
  placeholder?: string
}>()

const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const input = ref('')
const tags = computed(() => props.modelValue || [])

const colorPalette = [
  '00857D',
  '4299e1',
  '2fb344',
  'f59f00',
  'd63939',
  'ae3ec9',
  'd6336c',
  '4263eb',
]

function tagColor(tag: string): string {
  let hash = 0
  for (let i = 0; i < tag.length; i++) hash = tag.charCodeAt(i) + ((hash << 5) - hash)
  return colorPalette[Math.abs(hash) % colorPalette.length]
}

function addTag() {
  const val = input.value.trim()
  if (val && !tags.value.includes(val)) {
    emit('update:modelValue', [...tags.value, val])
  }
  input.value = ''
}

function removeTag(tag: string) {
  emit(
    'update:modelValue',
    tags.value.filter((t) => t !== tag),
  )
}

function removeLast() {
  if (input.value === '' && tags.value.length > 0) {
    emit('update:modelValue', tags.value.slice(0, -1))
  }
}
</script>
