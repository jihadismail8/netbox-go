<template>
  <div>
    <PageHeader title="API Browser" :breadcrumbs="[{ label: 'API' }]">
      <template #actions>
        <a
          :href="openApiUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-1.5 rounded-lg border border-gray-300 px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
        >
          <ExternalLink :size="14" /> OpenAPI Schema
        </a>
      </template>
    </PageHeader>
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div
        class="bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800"
      >
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700">
          <input
            v-model="search"
            type="text"
            placeholder="Filter endpoints..."
            class="w-full rounded border border-gray-300 px-2.5 py-1.5 text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div class="max-h-[60vh] overflow-y-auto">
          <p v-if="loading" class="px-4 py-8 text-center text-sm text-gray-400">
            Loading schema...
          </p>
          <p v-else-if="schemaError" class="px-4 py-8 text-center text-sm text-red-500">
            {{ schemaError }}
          </p>
          <div
            v-for="group in filteredGroups"
            :key="group.name"
            class="border-b border-gray-100 dark:border-gray-800"
          >
            <button
              class="flex w-full items-center justify-between px-4 py-2 text-left hover:bg-gray-50 dark:hover:bg-gray-800/50"
              @click="toggleGroup(group.name)"
            >
              <span class="text-sm font-semibold text-gray-900 dark:text-white">{{
                group.name
              }}</span>
              <ChevronDown
                :size="14"
                class="text-gray-400"
                :class="{ 'rotate-180': expandedGroups.has(group.name) }"
              />
            </button>
            <ul v-if="expandedGroups.has(group.name)" class="pb-1">
              <li v-for="ep in group.endpoints" :key="ep.method + ep.path">
                <button
                  class="flex w-full items-center gap-2 px-6 py-1.5 text-left text-xs hover:bg-gray-50 dark:hover:bg-gray-800/50"
                  :class="{ 'text-primary font-medium': selected?.path === ep.path }"
                  @click="selectEndpoint(ep)"
                >
                  <span class="font-mono font-bold w-12" :class="methodClass(ep.method)">{{
                    ep.method
                  }}</span>
                  <span class="truncate text-gray-600 dark:text-gray-400">{{ ep.path }}</span>
                </button>
              </li>
            </ul>
          </div>
        </div>
      </div>
      <div
        class="lg:col-span-2 bg-white border border-gray-200 rounded-lg shadow-sm dark:border-gray-700 dark:bg-gray-800 p-6"
      >
        <div v-if="!selected" class="py-12 text-center text-sm text-gray-400">
          Select an endpoint from the left.
        </div>
        <div v-else>
          <div class="flex items-center gap-2 mb-4">
            <span
              class="font-mono font-bold rounded px-2 py-0.5 text-xs"
              :class="methodClass(selected.method)"
              >{{ selected.method }}</span
            >
            <code class="text-sm text-gray-900 dark:text-white">{{ selected.path }}</code>
          </div>
          <p v-if="selected.description" class="text-sm text-gray-500 mb-4">
            {{ selected.description }}
          </p>
          <div class="flex gap-2">
            <button
              :disabled="trying || !canTrySelected"
              :title="
                canTrySelected ? undefined : 'Only concrete GET operations can be tried here.'
              "
              class="inline-flex items-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-sm font-medium text-white hover:bg-primary-dark disabled:opacity-50"
              @click="tryRequest"
            >
              <Play :size="14" /> {{ trying ? 'Requesting...' : 'Try it out' }}
            </button>
          </div>
          <div v-if="response" class="mt-4">
            <h4 class="text-xs font-semibold text-gray-500 mb-2">
              RESPONSE ({{ response.status }})
            </h4>
            <JsonViewer :data="response.data" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ChevronDown, Play, ExternalLink } from '@lucide/vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import JsonViewer from '@/components/shared/JsonViewer.vue'
import { getErrorRecord } from '@/api/errors'
import { getOpenAPISchema, invokeReadOnlyOperation } from '@/features/core/api'
import { APP_CONFIG } from '@/config/app'

interface Endpoint {
  method: string
  path: string
  description?: string
}
interface EndpointGroup {
  name: string
  endpoints: Endpoint[]
}

const groups = ref<EndpointGroup[]>([])
const search = ref('')
const expandedGroups = ref(new Set<string>(['DCIM']))
const selected = ref<Endpoint | null>(null)
const trying = ref(false)
const response = ref<{ status: number; data: unknown } | null>(null)
const loading = ref(true)
const schemaError = ref('')
const openApiUrl = `${APP_CONFIG.apiBaseUrl.replace(/\/$/, '')}/schema/`
const canTrySelected = computed(
  () => selected.value?.method === 'GET' && !selected.value.path.includes('{'),
)

const filteredGroups = computed(() => {
  if (!search.value) return groups.value
  const q = search.value.toLowerCase()
  return groups.value
    .map((g) => ({ ...g, endpoints: g.endpoints.filter((e) => e.path.includes(q)) }))
    .filter((g) => g.endpoints.length > 0)
})

function methodClass(method: string): string {
  const map: Record<string, string> = {
    GET: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400',
    POST: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-400',
    PUT: 'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-400',
    PATCH: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-400',
    DELETE: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400',
  }
  return map[method] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
}

function toggleGroup(name: string) {
  if (expandedGroups.value.has(name)) expandedGroups.value.delete(name)
  else expandedGroups.value.add(name)
  expandedGroups.value = new Set(expandedGroups.value)
}

function selectEndpoint(ep: Endpoint) {
  selected.value = ep
  response.value = null
}

async function tryRequest() {
  if (!selected.value) return
  trying.value = true
  response.value = null
  try {
    response.value = await invokeReadOnlyOperation(selected.value.path)
  } catch (caught: unknown) {
    response.value = {
      status: 0,
      data: getErrorRecord(caught),
    }
  } finally {
    trying.value = false
  }
}

onMounted(async () => {
  loading.value = true
  schemaError.value = ''
  const moduleGroups: Record<string, EndpointGroup> = {}
  try {
    const schema = await getOpenAPISchema()
    for (const [path, operations] of Object.entries(schema.paths)) {
      for (const [method, operation] of Object.entries(operations)) {
        if (!['get', 'post', 'put', 'patch', 'delete'].includes(method) || !operation) continue
        const tag = operation.tags?.[0] || 'other'
        const name = tag.toUpperCase()
        if (!moduleGroups[tag]) moduleGroups[tag] = { name, endpoints: [] }
        moduleGroups[tag].endpoints.push({
          method: method.toUpperCase(),
          path,
          description: operation.summary || operation.description,
        })
      }
    }
    groups.value = Object.values(moduleGroups)
    for (const group of groups.value) {
      group.endpoints.sort((left, right) =>
        `${left.path}:${left.method}`.localeCompare(`${right.path}:${right.method}`),
      )
    }
  } catch (caught: unknown) {
    schemaError.value = 'Unable to load the OpenAPI schema.'
    response.value = { status: 0, data: getErrorRecord(caught) }
  } finally {
    loading.value = false
  }
})
</script>
