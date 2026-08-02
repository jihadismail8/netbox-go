import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useTableStore = defineStore('table', () => {
  const selectedRows = ref<Record<string, Set<number>>>({})
  const pageSizes = ref<Record<string, number>>({})

  function getSelected(modelKey: string): Set<number> {
    if (!selectedRows.value[modelKey]) {
      selectedRows.value[modelKey] = new Set()
    }
    return selectedRows.value[modelKey]
  }

  function toggleRow(modelKey: string, id: number) {
    const selected = getSelected(modelKey)
    if (selected.has(id)) selected.delete(id)
    else selected.add(id)
  }

  function selectAll(modelKey: string, ids: number[]) {
    const selected = getSelected(modelKey)
    ids.forEach((id) => selected.add(id))
  }

  function clearSelection(modelKey: string) {
    selectedRows.value[modelKey] = new Set()
  }

  function getSelectedArray(modelKey: string): number[] {
    return Array.from(getSelected(modelKey))
  }

  function getPageSize(modelKey: string): number {
    return pageSizes.value[modelKey] || 50
  }

  function setPageSize(modelKey: string, size: number) {
    pageSizes.value[modelKey] = size
  }

  return {
    selectedRows,
    pageSizes,
    getSelected,
    toggleRow,
    selectAll,
    clearSelection,
    getSelectedArray,
    getPageSize,
    setPageSize,
  }
})
