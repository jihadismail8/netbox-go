import { CORE_PROFILE_MODELS } from '@/router/models/core-profile'

export interface MenuItemButton {
  label: string
  icon: string
  route: string
  color?: string
}

export interface MenuItem {
  label: string
  route: string
  icon?: string
  buttons?: MenuItemButton[]
  permission?: string
}

export interface MenuGroup {
  label: string
  icon: string
  items: MenuItem[]
}

const addButton = (route: string): MenuItemButton => ({
  label: 'Add',
  icon: 'plus',
  route: `${route}add/`,
})

const modelButtons = (route: string): MenuItemButton[] => [addButton(route)]

function moduleItems(module: 'dcim' | 'ipam'): MenuItem[] {
  return CORE_PROFILE_MODELS.filter((model) => model.module === module).map((model) => ({
    label: model.display_name_plural,
    route: model.routePath,
    buttons: modelButtons(model.routePath),
  }))
}

/** Navigation for the published core-workflow-v1 surface only. */
export const NAVIGATION: MenuGroup[] = [
  {
    label: 'Overview',
    icon: 'monitor',
    items: [{ label: 'Dashboard', route: '/' }],
  },
  {
    label: 'DCIM',
    icon: 'server',
    items: moduleItems('dcim'),
  },
  {
    label: 'IPAM',
    icon: 'network',
    items: moduleItems('ipam'),
  },
  {
    label: 'Developer',
    icon: 'wrench',
    items: [{ label: 'API Browser', route: '/api/' }],
  },
]
