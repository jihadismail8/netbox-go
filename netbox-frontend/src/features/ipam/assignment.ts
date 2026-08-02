import { updateResource } from '@/features/core/api'
import {
  relationID,
  type CoreRelationSelection,
  type IPAddressDTO,
  type IPAddressMutation,
} from '@/features/core/resources'

export type { IPAddressDTO } from '@/features/core/resources'

/** Map the operator-facing Interface selection to the baseline REST pair. */
export function assignmentPatch(
  interfaceValue: CoreRelationSelection | undefined,
): IPAddressMutation {
  if (interfaceValue === null || interfaceValue === undefined) {
    return { assigned_object_type: null, assigned_object_id: null }
  }
  const id = relationID(interfaceValue)
  if (typeof id !== 'number' || !Number.isInteger(id) || id <= 0) {
    throw new Error('Select a valid Interface.')
  }
  return { assigned_object_type: 'dcim.interface', assigned_object_id: id }
}

export async function assignIPAddress(
  id: number,
  interfaceValue: CoreRelationSelection,
): Promise<IPAddressDTO> {
  return updateResource('ipaddress', id, assignmentPatch(interfaceValue))
}

export async function unassignIPAddress(id: number): Promise<IPAddressDTO> {
  return updateResource('ipaddress', id, assignmentPatch(null))
}
