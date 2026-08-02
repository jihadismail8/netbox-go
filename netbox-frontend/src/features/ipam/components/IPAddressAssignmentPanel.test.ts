import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import IPAddressAssignmentPanel from './IPAddressAssignmentPanel.vue'
import type { IPAddressDTO } from '@/features/core/resources'

const mocks = vi.hoisted(() => ({
  assignIPAddress: vi.fn(),
  unassignIPAddress: vi.fn(),
}))

vi.mock('@/features/ipam/assignment', async () => {
  const actual = await vi.importActual('@/features/ipam/assignment')
  return {
    ...actual,
    assignIPAddress: mocks.assignIPAddress,
    unassignIPAddress: mocks.unassignIPAddress,
  }
})

const ApiSelectStub = {
  template: '<button class="choose" @click="$emit(\'update:modelValue\', 77)">Choose</button>',
  emits: ['update:modelValue'],
}

const ConfirmStub = {
  props: ['modelValue'],
  template:
    '<button v-if="modelValue" class="confirm" @click="$emit(\'confirm\')">Confirm</button>',
  emits: ['confirm', 'update:modelValue'],
}

function ipAddress(overrides: Partial<IPAddressDTO> = {}): IPAddressDTO {
  return {
    id: 9,
    url: '/api/ipam/ip-addresses/9/',
    display: '192.0.2.9/24',
    created: '2026-07-18T10:00:00Z',
    last_updated: '2026-07-18T10:00:00Z',
    address: '192.0.2.9/24',
    vrf: null,
    status: { value: 'active', label: 'Active' },
    role: null,
    dns_name: '',
    description: '',
    comments: '',
    assigned_object_type: null,
    assigned_object_id: null,
    family: { value: 4, label: 'IPv4' },
    assigned_object: null,
    ...overrides,
  }
}

describe('IPAddressAssignmentPanel', () => {
  beforeEach(() => {
    mocks.assignIPAddress.mockReset()
    mocks.unassignIPAddress.mockReset()
  })

  it('assigns an unassigned address through the feature API', async () => {
    const updated = ipAddress({
      assigned_object_type: 'dcim.interface',
      assigned_object_id: 77,
      assigned_object: {
        id: 77,
        url: '/api/dcim/interfaces/77/',
        display: 'edge-1: xe-0/0/0',
      },
    })
    mocks.assignIPAddress.mockResolvedValue(updated)
    const wrapper = mount(IPAddressAssignmentPanel, {
      props: { address: ipAddress(), canChange: true },
      global: { stubs: { ApiSelectField: ApiSelectStub, ConfirmModal: ConfirmStub } },
    })

    await wrapper.get('button').trigger('click')
    await wrapper.get('.choose').trigger('click')
    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Save assignment')!
      .trigger('click')
    await flushPromises()

    expect(mocks.assignIPAddress).toHaveBeenCalledWith(9, 77)
    expect(wrapper.emitted('updated')).toEqual([[updated]])
  })

  it('requires confirmation and preserves the address while unassigning', async () => {
    const updated = ipAddress()
    mocks.unassignIPAddress.mockResolvedValue(updated)
    const wrapper = mount(IPAddressAssignmentPanel, {
      props: {
        address: ipAddress({
          assigned_object_type: 'dcim.interface',
          assigned_object_id: 77,
          assigned_object: {
            id: 77,
            url: '/api/dcim/interfaces/77/',
            display: 'edge-1: xe-0/0/0',
          },
        }),
        canChange: true,
      },
      global: { stubs: { ApiSelectField: ApiSelectStub, ConfirmModal: ConfirmStub } },
    })

    await wrapper
      .findAll('button')
      .find((button) => button.text() === 'Unassign')!
      .trigger('click')
    expect(mocks.unassignIPAddress).not.toHaveBeenCalled()
    await wrapper.get('.confirm').trigger('click')
    await flushPromises()

    expect(mocks.unassignIPAddress).toHaveBeenCalledWith(9)
    expect(wrapper.emitted('updated')).toEqual([[updated]])
  })
})
