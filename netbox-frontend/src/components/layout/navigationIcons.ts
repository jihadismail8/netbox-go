import type { Component } from 'vue'
import {
  Archive,
  Building,
  Cable,
  Circle,
  Monitor,
  Network,
  Plus,
  Server,
  Settings,
  Shield,
  Upload,
  Users,
  Wifi,
  Wrench,
  Zap,
} from '@lucide/vue'

const navigationIcons: Record<string, Component> = {
  archive: Archive,
  building: Building,
  cable: Cable,
  monitor: Monitor,
  network: Network,
  plus: Plus,
  server: Server,
  settings: Settings,
  shield: Shield,
  upload: Upload,
  users: Users,
  wifi: Wifi,
  wrench: Wrench,
  zap: Zap,
}

export function getNavigationIcon(name: string): Component {
  return navigationIcons[name.toLowerCase()] ?? Circle
}
