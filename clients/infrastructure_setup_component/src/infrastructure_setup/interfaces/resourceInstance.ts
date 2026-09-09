// 'partial' means the resource exists but some members could not be granted access.
export type ResourceStatus = 'pending' | 'in_progress' | 'created' | 'partial' | 'failed'

export interface ResourceInstance {
  id: string
  resourceConfigId: string
  coursePhaseId: string
  teamId?: string
  courseParticipationId?: string
  status: ResourceStatus
  externalId?: string
  externalUrl?: string
  errorMessage?: string
  // Empty until the instance has run at least once.
  targetName: string
  resolvedName: string
  providerType: string
  resourceType: string
  scope: 'per_team' | 'per_student'
  nameTemplate: string
  createdAt: string
  updatedAt: string
}
