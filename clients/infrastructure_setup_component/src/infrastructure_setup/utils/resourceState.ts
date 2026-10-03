import type { ResourceInstance, ResourceStatus } from '../interfaces/resourceInstance'

// What one resource means for one person, derived from the instance's status and
// whether its latest run let them in.
export type ResourceState =
  | 'not_provisioned'
  | 'setting_up'
  | 'ready'
  | 'partly_ready'
  | 'not_added'
  | 'failed'

export const resourceState = (
  status: ResourceStatus | null,
  granted: boolean | null,
): ResourceState => {
  switch (status) {
    case null:
      return 'not_provisioned'
    case 'pending':
    case 'in_progress':
      return 'setting_up'
    case 'failed':
      return 'failed'
    case 'created':
    case 'partial':
      if (granted === true) return 'ready'
      if (granted === false) return 'not_added'
      // A run from before members were recorded says nothing about this person.
      return status === 'created' ? 'ready' : 'partly_ready'
  }
}

// Least finished first, so a problem on one instance is never hidden behind a success
// on another.
const SEVERITY: ResourceState[] = [
  'failed',
  'not_added',
  'partly_ready',
  'setting_up',
  'ready',
  'not_provisioned',
]

// The state of one config's resource for one participant. The resource is theirs when
// it was provisioned for them personally or its latest run was for them as a member;
// someone on two teams gets the least finished of both.
export const participantResourceState = (
  instances: ResourceInstance[],
  resourceConfigId: string,
  courseParticipationId: string,
): ResourceState => {
  let state: ResourceState = 'not_provisioned'
  for (const instance of instances) {
    if (instance.resourceConfigId !== resourceConfigId) continue

    const member = instance.members.find((m) => m.courseParticipationId === courseParticipationId)
    if (!member && instance.courseParticipationId !== courseParticipationId) continue

    const candidate = resourceState(instance.status, member?.granted ?? null)
    if (SEVERITY.indexOf(candidate) < SEVERITY.indexOf(state)) {
      state = candidate
    }
  }
  return state
}

// How the participants table names a state.
export const RESOURCE_STATE_LABELS: Record<ResourceState, string> = {
  not_provisioned: 'Not provisioned',
  setting_up: 'Being set up',
  ready: 'Ready',
  partly_ready: 'Partly set up',
  not_added: 'Not added yet',
  failed: 'Failed',
}

// How the student page names a state. A failure's cause is the course's to fix, so it
// is not presented as the student's problem.
export const STUDENT_RESOURCE_STATE_LABELS: Record<ResourceState, string> = {
  not_provisioned: 'Not set up yet',
  setting_up: 'Being set up',
  ready: 'Ready',
  partly_ready: 'Ready',
  not_added: 'Not added yet',
  failed: 'Not available yet',
}
