<template>
  <div>
    <PageHeader title="Dashboard" />

    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
      <ObjectCountWidget
        v-for="resource in resources"
        :key="resource.route"
        :label="resource.label"
        :resource="resource.model"
        :route="resource.route"
        :icon="resource.icon"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { Archive, Building, Cpu, Layers, Network, Server, Tag, Waypoints } from '@lucide/vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import ObjectCountWidget from '@/components/dashboard/ObjectCountWidget.vue'
import { CORE_RESOURCE_REGISTRY } from '@/router/core-resource-registry'
import type { CoreProfileResourceName } from '@/features/core/manifest'
import { usePermissions } from '@/composables/usePermissions'

interface DashboardResource {
  label: string
  model: CoreProfileResourceName
  route: string
  icon: Component
}

const icons: Record<CoreProfileResourceName, Component> = {
  site: Building,
  manufacturer: Building,
  rackrole: Tag,
  racktype: Archive,
  rack: Archive,
  devicerole: Tag,
  devicetype: Cpu,
  interfacetemplate: Waypoints,
  device: Server,
  interface: Waypoints,
  vrf: Layers,
  prefix: Network,
  ipaddress: Network,
}

const { canView } = usePermissions()
const resources = computed<DashboardResource[]>(() =>
  CORE_RESOURCE_REGISTRY.filter((config) => canView(config.module, config.model)).map((config) => ({
    label: config.display_name_plural,
    model: config.model,
    route: config.routePath,
    icon: icons[config.model],
  })),
)
</script>
