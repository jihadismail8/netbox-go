/**
 * NetBox Go Frontend — Application Configuration
 *
 * Colors and theme values derived from the original NetBox v4.4.0
 * which uses Tabler v1.4.0 + Bootstrap v5.3.8.
 */

export const APP_CONFIG = {
  appName: 'NetBox',
  edition: 'Community',
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api',
  pageSizes: [25, 50, 100, 250, 500],
  defaultPageSize: 50,
  searchDebounce: 300,
  toastDuration: 5000,
} as const

export const NETBOX_COLORS = {
  richBlack: '#001423',
  richBlackLight: '#081B2A',
  richBlackLighter: '#0D202E',
  richBlackLightest: '#1A2C39',
  brightTeal: '#00F2D4',
  darkTeal: '#00857D',
  primary: '#00857D',
  secondary: '#6b7280',
  success: '#2fb344',
  info: '#4299e1',
  warning: '#f59f00',
  danger: '#d63939',
  light: '#f9fafb',
  dark: '#1f2937',
  blue: '#066fd1',
  azure: '#4299e1',
  indigo: '#4263eb',
  purple: '#ae3ec9',
  pink: '#d6336c',
  red: '#d63939',
  orange: '#f76707',
  yellow: '#f59f00',
  lime: '#74b816',
  green: '#2fb344',
  teal: '#00857D',
  cyan: '#17a2b8',
} as const

export const STATUS_COLORS: Record<string, string> = {
  active: 'success',
  planned: 'info',
  staging: 'info',
  staged: 'warning',
  provisioning: 'info',
  deprovisioning: 'secondary',
  maintenance: 'warning',
  failed: 'danger',
  offline: 'danger',
  reserved: 'danger',
  deprecated: 'secondary',
  decommissioned: 'secondary',
  decommissioning: 'secondary',
  inventory: 'info',
  available: 'success',
  connected: 'success',
  assigned: 'info',
  unassigned: 'secondary',
}
