import { describe, expect, it } from 'vitest'

import { selfTeamAllocationKeys } from './keys'

const PHASE = 'phase-1'

describe('selfTeamAllocationKeys', () => {
  it('builds the phase keys', () => {
    expect(selfTeamAllocationKeys.teams(PHASE)).toEqual(['self_team_allocations', PHASE])
    expect(selfTeamAllocationKeys.timeframe(PHASE)).toEqual(['timeframe', PHASE])
    expect(selfTeamAllocationKeys.config(PHASE)).toEqual(['team_allocation_config', PHASE])
    expect(selfTeamAllocationKeys.myParticipation(PHASE)).toEqual([
      'course_phase_participation',
      PHASE,
    ])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(selfTeamAllocationKeys.teams(undefined)).toEqual(['self_team_allocations', undefined])
  })

  it('builds the tutor import key shared with Team Allocation', () => {
    expect(selfTeamAllocationKeys.tutorImport.studentsOfPhase(PHASE)).toEqual(['students', PHASE])
  })
})
