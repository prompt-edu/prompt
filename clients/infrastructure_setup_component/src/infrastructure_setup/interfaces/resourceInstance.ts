export type ResourceStatus = 'pending' | 'in_progress' | 'created' | 'failed'

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
  createdAt: string
  updatedAt: string
}
