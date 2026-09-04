import type { ProviderType } from './providerConfig'

export type Scope = 'per_team' | 'per_student'

export interface ResourceConfig {
  id: string
  coursePhaseId: string
  providerType: ProviderType
  resourceType: string
  scope: Scope
  nameTemplate: string
  permissionMapping: Record<string, string>
  resourceExtraConfig: Record<string, unknown>
  createdAt: string
}

export interface CreateResourceConfigRequest {
  providerType: ProviderType
  resourceType: string
  scope: Scope
  nameTemplate: string
  permissionMapping?: Record<string, string>
  resourceExtraConfig?: Record<string, unknown>
}

export interface UpdateResourceConfigRequest {
  resourceType: string
  scope: Scope
  nameTemplate: string
  permissionMapping?: Record<string, string>
  resourceExtraConfig?: Record<string, unknown>
}
