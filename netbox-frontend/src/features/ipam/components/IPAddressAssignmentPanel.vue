<template>
  <section class="rounded-lg border border-gray-200 p-4 dark:border-gray-700">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">Interface assignment</h3>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">
          <template v-if="address.assigned_object">
            Assigned to
            <span class="font-medium">{{ assignmentLabel }}</span>
          </template>
          <template v-else>This IP address is not assigned to an Interface.</template>
        </p>
      </div>
      <div v-if="canChange && !editing" class="flex gap-2">
        <button
          type="button"
          :disabled="saving"
          class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
          @click="beginAssignment"
        >
          {{ address.assigned_object ? 'Change' : 'Assign' }}
        </button>
        <button
          v-if="address.assigned_object"
          type="button"
          :disabled="saving"
          class="rounded-lg border border-red-300 px-3 py-1.5 text-sm font-medium text-red-600 hover:bg-red-50 dark:border-red-700 dark:hover:bg-red-900/20"
          @click="confirmingUnassign = true"
        >
          Unassign
        </button>
      </div>
    </div>
    <p v-if="error" class="mt-3 text-sm text-red-600 dark:text-red-400">{{ error }}</p>

    <div v-if="editing" class="mt-4 space-y-3 border-t border-gray-200 pt-4 dark:border-gray-700">
      <ApiSelectField
        v-model="selectedInterface"
        label="Interface"
        relation-resource="interface"
        :required="true"
        placeholder="Search by Interface name"
      />
      <div class="flex justify-end gap-2">
        <button
          type="button"
          class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm dark:border-gray-600"
          @click="cancelAssignment"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="saving || !selectedInterface"
          class="rounded-lg bg-primary px-3 py-1.5 text-sm font-medium text-white hover:bg-primary-dark disabled:opacity-50"
          @click="saveAssignment"
        >
          {{ saving ? 'Saving...' : 'Save assignment' }}
        </button>
      </div>
    </div>

    <ConfirmModal
      v-model="confirmingUnassign"
      title="Unassign IP address"
      message="Remove this IP address from its Interface? The IP address itself will be kept."
      confirm-label="Unassign"
      :danger="true"
      @confirm="handleUnassign"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ApiSelectField from '@/components/form/ApiSelectField.vue'
import ConfirmModal from '@/components/shared/ConfirmModal.vue'
import { assignIPAddress, unassignIPAddress, type IPAddressDTO } from '@/features/ipam/assignment'
import { getErrorDetail } from '@/api/errors'
import type { CoreRelationSelection } from '@/features/core/resources'

const props = defineProps<{ address: IPAddressDTO; canChange: boolean }>()
const emit = defineEmits<{ updated: [address: IPAddressDTO] }>()

const editing = ref(false)
const saving = ref(false)
const confirmingUnassign = ref(false)
const selectedInterface = ref<CoreRelationSelection>(null)
const error = ref('')

const assignmentLabel = computed(
  () => props.address.assigned_object?.display || `Interface #${props.address.assigned_object?.id}`,
)

function beginAssignment() {
  selectedInterface.value = props.address.assigned_object ?? null
  error.value = ''
  editing.value = true
}

function cancelAssignment() {
  selectedInterface.value = null
  error.value = ''
  editing.value = false
}

async function saveAssignment() {
  saving.value = true
  error.value = ''
  try {
    const updated = await assignIPAddress(props.address.id, selectedInterface.value)
    editing.value = false
    emit('updated', updated)
  } catch (caught: unknown) {
    error.value = getErrorDetail(caught, 'Failed to assign the IP address.')
  } finally {
    saving.value = false
  }
}

async function handleUnassign() {
  confirmingUnassign.value = false
  saving.value = true
  error.value = ''
  try {
    const updated = await unassignIPAddress(props.address.id)
    emit('updated', updated)
  } catch (caught: unknown) {
    error.value = getErrorDetail(caught, 'Failed to unassign the IP address.')
  } finally {
    saving.value = false
  }
}
</script>
