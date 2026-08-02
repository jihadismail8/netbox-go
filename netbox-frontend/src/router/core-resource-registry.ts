/**
 * Published model registry for core-workflow-v1.
 *
 * The larger legacy catalog remains useful as reference material, but it is not
 * part of the supported runtime and must not be registered here.
 */

import type { ModelConfig } from '@/types'
import { CORE_PROFILE_MODELS } from './models/core-profile'

export const CORE_RESOURCE_REGISTRY: ModelConfig[] = CORE_PROFILE_MODELS

/**
 * Get a model config by route path prefix.
 */
export function getCoreResourceConfig(routePath: string): ModelConfig | undefined {
  return CORE_RESOURCE_REGISTRY.find((model) => routePath.startsWith(model.routePath))
}

/**
 * Get all registered models.
 */
export function getCoreResourceConfigs(): ModelConfig[] {
  return CORE_RESOURCE_REGISTRY
}
