import { describe, expect, it } from 'vitest'
import { routeForObjectURL } from './links'

describe('profile object links', () => {
  it('maps relative and absolute REST URLs to Vue detail routes', () => {
    expect(routeForObjectURL('/api/dcim/sites/7/')).toBe('/dcim/sites/7/')
    expect(routeForObjectURL('https://netbox.example/api/ipam/vrfs/9/')).toBe('/ipam/vrfs/9/')
  })

  it('rejects undeclared and malformed destinations', () => {
    expect(routeForObjectURL('/api/extras/tags/1/')).toBe('#')
    expect(routeForObjectURL('javascript:alert(1)')).toBe('#')
    expect(routeForObjectURL('/api/dcim/sites/not-an-id/')).toBe('#')
  })
})
