import { describe, it, expect } from 'vitest'
import { hasPermission, permCode } from './usePermissions'

describe('permCode', () => {
  it('builds view permission code', () => {
    expect(permCode('dcim', 'view', 'site')).toBe('dcim.view_site')
  })

  it('builds add permission code', () => {
    expect(permCode('ipam', 'add', 'prefix')).toBe('ipam.add_prefix')
  })

  it('builds change permission code', () => {
    expect(permCode('tenancy', 'change', 'tenant')).toBe('tenancy.change_tenant')
  })

  it('builds delete permission code', () => {
    expect(permCode('dcim', 'delete', 'device')).toBe('dcim.delete_device')
  })

  it('handles multi-word models', () => {
    expect(permCode('dcim', 'view', 'device_type')).toBe('dcim.view_device_type')
  })
})

describe('hasPermission', () => {
  const staffUser = {
    id: 1,
    username: 'operator',
    email: '',
    first_name: '',
    last_name: '',
    is_staff: true,
    is_superuser: false,
  }

  it('does not treat staff status as a model permission', () => {
    expect(hasPermission(staffUser, new Set(), 'view', 'dcim', 'site')).toBe(false)
  })

  it('uses effective permissions and preserves the superuser bypass', () => {
    expect(hasPermission(staffUser, new Set(['dcim.change_site']), 'change', 'dcim', 'site')).toBe(
      true,
    )
    expect(
      hasPermission({ ...staffUser, is_superuser: true }, new Set(), 'delete', 'dcim', 'site'),
    ).toBe(true)
  })
})
