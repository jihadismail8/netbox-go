import { describe, expect, it } from 'vitest'

import { INTERFACE_TYPE_CHOICES } from './interface-types'

describe('INTERFACE_TYPE_CHOICES', () => {
  it('publishes the complete pinned NetBox baseline without duplicates', () => {
    const values = INTERFACE_TYPE_CHOICES.map(({ value }) => value)

    expect(values).toHaveLength(206)
    expect(new Set(values).size).toBe(values.length)
  })

  it('includes representative choices from every interface family', () => {
    const values = new Set(INTERFACE_TYPE_CHOICES.map(({ value }) => value))

    for (const representative of [
      'virtual',
      '100base-fx',
      '1000base-t',
      '10gbase-x-sfpp',
      '800gbase-x-osfp',
      '100gbase-kr4',
      'ieee802.11be',
      '5g',
      'sonet-oc3840',
      '128gfc-qsfp28',
      'infiniband-xdr',
      'e3',
      'xgs-pon',
      'cisco-stackwise-1t',
      'other',
    ]) {
      expect(values).toContain(representative)
    }
  })
})
