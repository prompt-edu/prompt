import { describe, expect, it } from 'vitest'

import type { ResourceInstance, ResourceStatus } from '../interfaces/resourceInstance'
import { participantResourceState, resourceState } from './resourceState'

const buildInstance = (
  id: string,
  status: ResourceStatus,
  overrides: Partial<ResourceInstance> = {},
): ResourceInstance => ({
  id,
  resourceConfigId: 'config-1',
  coursePhaseId: 'phase-1',
  status,
  targetName: '',
  resolvedName: '',
  providerType: 'gitlab',
  resourceType: 'group',
  scope: 'per_team',
  nameTemplate: '{{teamName}}',
  members: [],
  createdAt: '',
  updatedAt: '',
  ...overrides,
})

describe('resourceState', () => {
  it('has nothing to report without an instance', () => {
    expect(resourceState(null, null)).toBe('not_provisioned')
  })

  it('reports a queued or running instance as being set up', () => {
    expect(resourceState('pending', null)).toBe('setting_up')
    expect(resourceState('in_progress', true)).toBe('setting_up')
  })

  it('reports a failure regardless of earlier access', () => {
    expect(resourceState('failed', true)).toBe('failed')
  })

  it('tells the person left out of a partial run apart from the ones let in', () => {
    expect(resourceState('partial', true)).toBe('ready')
    expect(resourceState('partial', false)).toBe('not_added')
  })

  it('falls back to the instance status for a run that recorded no members', () => {
    expect(resourceState('created', null)).toBe('ready')
    expect(resourceState('partial', null)).toBe('partly_ready')
  })
})

describe('participantResourceState', () => {
  it('finds a team resource through its members', () => {
    const instances = [
      buildInstance('team-a', 'partial', {
        teamId: 'team-a',
        members: [
          { courseParticipationId: 'alice', granted: true },
          { courseParticipationId: 'bob', granted: false },
        ],
      }),
    ]

    expect(participantResourceState(instances, 'config-1', 'alice')).toBe('ready')
    expect(participantResourceState(instances, 'config-1', 'bob')).toBe('not_added')
    expect(participantResourceState(instances, 'config-1', 'carol')).toBe('not_provisioned')
  })

  it('finds a personal resource through its target', () => {
    const instances = [
      buildInstance('mine', 'created', { scope: 'per_student', courseParticipationId: 'alice' }),
    ]

    expect(participantResourceState(instances, 'config-1', 'alice')).toBe('ready')
  })

  it('ignores the instances of other configs', () => {
    const instances = [
      buildInstance('other', 'failed', {
        resourceConfigId: 'config-2',
        members: [{ courseParticipationId: 'alice', granted: false }],
      }),
    ]

    expect(participantResourceState(instances, 'config-1', 'alice')).toBe('not_provisioned')
  })

  it('shows the least finished resource of someone on two teams', () => {
    const instances = [
      buildInstance('team-a', 'created', {
        members: [{ courseParticipationId: 'alice', granted: true }],
      }),
      buildInstance('team-b', 'failed', {
        members: [{ courseParticipationId: 'alice', granted: false }],
      }),
    ]

    expect(participantResourceState(instances, 'config-1', 'alice')).toBe('failed')
  })
})
