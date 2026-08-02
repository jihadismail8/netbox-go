<template>
  <div>
    <PageHeader
      :title="isEdit ? 'Edit ' + config.display_name : 'Add ' + config.display_name"
      :breadcrumbs="breadcrumbs"
    />

    <div v-if="loadingItem" class="py-8"><LoadingSpinner /></div>

    <div
      v-else
      class="p-6 bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800"
    >
      <DynamicForm
        v-model="formData"
        :fields="config.fields"
        :errors="errors"
        :loading="saving"
        :editing="isEdit"
        @submit="handleSubmit"
        @cancel="handleCancel"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useCoreResource } from '@/composables/useCoreResource'
import { useNotificationStore } from '@/stores/notifications'
import type { ModelConfig } from '@/types'
import PageHeader from '@/components/layout/PageHeader.vue'
import DynamicForm from '@/components/form/DynamicForm.vue'
import LoadingSpinner from '@/components/shared/LoadingSpinner.vue'
import { getErrorDetail, getErrorRecord, type ValidationErrors } from '@/api/errors'
import { getCoreResourceAdapter } from '@/features/core/adapters'
import type { CoreResourceForm } from '@/features/core/resources'

const props = defineProps<{ config: ModelConfig }>()
const route = useRoute()
const router = useRouter()
const notifications = useNotificationStore()
const resource = props.config.model
const adapter = getCoreResourceAdapter(resource)
const { fetchById, createItem, updateItem } = useCoreResource(resource)

const formData = shallowRef<CoreResourceForm>(adapter.emptyForm())
const errors = ref<ValidationErrors>({})
const saving = ref(false)
const loadingItem = ref(false)

const isEdit = computed(() => !!route.params.id)
const breadcrumbs = computed(() => [
  { label: props.config.display_name_plural, to: props.config.routePath },
  { label: isEdit.value ? 'Edit' : 'Add' },
])

async function handleSubmit() {
  saving.value = true
  errors.value = {}
  try {
    const payload = adapter.mutationFromForm(formData.value, isEdit.value)
    if (isEdit.value) {
      await updateItem(Number(route.params.id), payload)
      notifications.success(props.config.display_name + ' updated successfully')
    } else {
      await createItem(payload)
      notifications.success(props.config.display_name + ' created successfully')
    }
    router.push(props.config.routePath)
  } catch (caught: unknown) {
    errors.value = getErrorRecord(caught)
    notifications.error(getErrorDetail(caught, 'Failed to save'))
  } finally {
    saving.value = false
  }
}

function handleCancel() {
  router.back()
}

onMounted(async () => {
  if (isEdit.value) {
    loadingItem.value = true
    try {
      const item = await fetchById(Number(route.params.id))
      formData.value = adapter.formFromDTO(item)
    } catch {
      notifications.error('Failed to load item')
      router.push(props.config.routePath)
    } finally {
      loadingItem.value = false
    }
  }
})
</script>
