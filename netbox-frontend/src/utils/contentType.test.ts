import { describe, it, expect } from 'vitest'
import { buildContentType, contentTypeOf } from './contentType'

describe('buildContentType', () => {
  it('builds content type from a standard route path', () => {
    expect(buildContentType('/dcim/sites/')).toBe('dcim.site')
  })

  it('builds content type for multi-word model', () => {
    expect(buildContentType('/dcim/device-types/')).toBe('dcim.devicetype')
  })

  it('builds content type for IPAM models', () => {
    expect(buildContentType('/ipam/ip-addresses/')).toBe('ipam.ipaddress')
    expect(buildContentType('/ipam/vlan-groups/')).toBe('ipam.vlangroup')
  })

  it('falls back to stripping trailing s for unmapped models', () => {
    expect(buildContentType('/extras/webhooks/')).toBe('extras.webhook')
  })

  it('returns empty string for paths without enough parts', () => {
    expect(buildContentType('/dcim/')).toBe('')
    expect(buildContentType('/')).toBe('')
  })
})

describe('contentTypeOf', () => {
  it('joins module and model with a dot', () => {
    expect(contentTypeOf('dcim', 'site')).toBe('dcim.site')
  })

  it('handles compound model names', () => {
    expect(contentTypeOf('ipam', 'ipaddress')).toBe('ipam.ipaddress')
  })
})
