<template>
  <Modal
    :model-value="modelValue"
    :title="title"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <p class="text-gray-600 dark:text-gray-300">{{ message }}</p>
    <template #footer>
      <button
        class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
        @click="$emit('update:modelValue', false)"
      >
        {{ cancelLabel }}
      </button>
      <button
        class="rounded-lg px-4 py-2 text-sm font-medium text-white"
        :class="danger ? 'bg-red-600 hover:bg-red-700' : 'bg-primary hover:bg-primary-dark'"
        @click="$emit('confirm')"
      >
        {{ confirmLabel }}
      </button>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import Modal from './Modal.vue'

withDefaults(
  defineProps<{
    modelValue: boolean
    title: string
    message: string
    confirmLabel?: string
    cancelLabel?: string
    danger?: boolean
  }>(),
  {
    confirmLabel: 'Confirm',
    cancelLabel: 'Cancel',
    danger: false,
  },
)

defineEmits<{
  'update:modelValue': [value: boolean]
  confirm: []
}>()
</script>
