<template>
  <form class="space-y-4" @submit.prevent="$emit('submit')">
    <div
      v-if="formErrorMessages.length > 0"
      role="alert"
      class="rounded-lg border border-red-300 bg-red-50 p-3 text-sm text-red-700 dark:border-red-700 dark:bg-red-900/20 dark:text-red-300"
    >
      <p v-for="message in formErrorMessages" :key="message">{{ message }}</p>
    </div>
    <template v-for="(groupFields, groupName) in groupedFields" :key="groupName">
      <div
        v-if="groupName !== 'undefined'"
        class="pb-2 text-sm font-semibold text-gray-500 border-b border-gray-200 dark:border-gray-700"
      >
        {{ groupName }}
      </div>
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <DynamicField
          v-for="field in groupFields"
          :key="field.key"
          :field="field"
          :model-value="formField(formData, field.key)"
          :form-data="formData"
          :disabled="isFieldDisabled(field)"
          :required="isFieldRequired(field)"
          :error="fieldError(errors?.[field.key])"
          @update:model-value="updateField(field.key, $event)"
        />
      </div>
    </template>

    <div class="flex justify-end gap-2 pt-4 border-t border-gray-200 dark:border-gray-700">
      <button
        type="button"
        class="px-4 py-2 text-sm font-medium text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
        @click="$emit('cancel')"
      >
        Cancel
      </button>
      <button
        type="submit"
        :disabled="loading"
        class="px-4 py-2 text-sm font-medium text-white rounded-lg bg-primary hover:bg-primary-dark disabled:opacity-50"
      >
        {{ loading ? 'Saving...' : 'Save' }}
      </button>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { FormFieldDef } from '@/types'
import type { ValidationErrors } from '@/api/errors'
import { formField, withFormField, type CoreResourceForm } from '@/features/core/resources'
import DynamicField from './DynamicField.vue'

const props = defineProps<{
  fields: FormFieldDef[]
  modelValue?: CoreResourceForm
  errors?: ValidationErrors
  loading?: boolean
  editing?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: CoreResourceForm]
  submit: []
  cancel: []
}>()

const formData = ref<CoreResourceForm>({ ...(props.modelValue ?? {}) })

watch(
  () => props.modelValue,
  (value) => {
    formData.value = { ...(value ?? {}) }
  },
  { deep: true },
)

const groupedFields = computed(() => {
  const groups: Record<string, FormFieldDef[]> = {}
  for (const field of props.fields) {
    const g = field.group || 'General'
    if (!groups[g]) groups[g] = []
    groups[g].push(field)
  }
  return groups
})

const formErrorMessages = computed(() => {
  const value = props.errors?.non_field_errors ?? props.errors?.detail
  if (Array.isArray(value)) return value.map(String)
  return typeof value === 'string' && value ? [value] : []
})

function updateField(key: string, value: unknown) {
  formData.value = withFormField(formData.value, key, value)
  const changed = new Set([key])
  let foundDependency = true
  while (foundDependency) {
    foundDependency = false
    for (const field of props.fields) {
      if (
        changed.has(field.key) ||
        !field.clearWhenFieldChanges?.some((dependency) => changed.has(dependency)) ||
        formField(formData.value, field.key) === null ||
        formField(formData.value, field.key) === undefined
      ) {
        continue
      }
      formData.value = withFormField(formData.value, field.key, null)
      changed.add(field.key)
      foundDependency = true
    }
  }
  emit('update:modelValue', formData.value)
}

function isFieldDisabled(field: FormFieldDef): boolean {
  return Boolean(
    (props.editing && field.immutableOnEdit) ||
    (field.disabledWhenFieldTruthy && formField(formData.value, field.disabledWhenFieldTruthy)) ||
    (field.disabledUnlessFieldTruthy &&
      !formField(formData.value, field.disabledUnlessFieldTruthy)),
  )
}

function isFieldRequired(field: FormFieldDef): boolean {
  return Boolean(
    field.required ||
    (field.requiredWhenFieldTruthy && formField(formData.value, field.requiredWhenFieldTruthy)),
  )
}

function fieldError(value: unknown): string | undefined {
  if (Array.isArray(value)) return value.map(String).join(' ')
  return typeof value === 'string' && value ? value : undefined
}
</script>
