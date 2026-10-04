import { describe, expect, it } from 'vitest'

import { teamAllocationKeys } from './keys'

const PHASE = 'phase-1'

describe('teamAllocationKeys', () => {
  it('builds the phase keys', () => {
    expect(teamAllocationKeys.teams(PHASE)).toEqual(['team_allocation_team', PHASE])
    expect(teamAllocationKeys.skills(PHASE)).toEqual(['team_allocation_skill', PHASE])
    expect(teamAllocationKeys.allocations(PHASE)).toEqual(['team_allocations', PHASE])
    expect(teamAllocationKeys.config(PHASE)).toEqual(['team_allocation_config', PHASE])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(teamAllocationKeys.teams(undefined)).toEqual(['team_allocation_team', undefined])
  })

  it('builds the survey keys', () => {
    expect(teamAllocationKeys.survey.form(PHASE)).toEqual(['team_allocation_survey_form', PHASE])
    expect(teamAllocationKeys.survey.myResponse(PHASE)).toEqual([
      'team_allocation_student_survey_response',
      PHASE,
    ])
    expect(teamAllocationKeys.survey.timeframe(PHASE)).toEqual([
      'team_allocation_survey_timeframe',
      PHASE,
    ])
    expect(teamAllocationKeys.survey.statistics(PHASE)).toEqual([
      'team_allocation_survey_statistics',
      PHASE,
    ])
  })

  it('keeps the TEASE caches apart from the settings caches holding the same data', () => {
    expect(teamAllocationKeys.tease.teams(PHASE)).toEqual(['tease_teams', PHASE])
    expect(teamAllocationKeys.tease.skills(PHASE)).toEqual(['tease_skills', PHASE])
    expect(teamAllocationKeys.tease.students(PHASE)).toEqual(['tease_students', PHASE])
  })

  it('builds the tutor import keys, including the one shared with Self Team Allocation', () => {
    expect(teamAllocationKeys.tutorImport.studentsOfPhase(PHASE)).toEqual(['students', PHASE])
    expect(teamAllocationKeys.tutorImport.studentSearch('ab')).toEqual(['student-search', 'ab'])
  })

  it('reproduces the key owned by the shared state package', () => {
    expect(teamAllocationKeys.participants(PHASE)).toEqual(['participants', PHASE])
  })
})
