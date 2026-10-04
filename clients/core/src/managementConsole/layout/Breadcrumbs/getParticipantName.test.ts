import { describe, expect, it } from 'vitest'
import { getParticipantName } from './getParticipantName'

const PARTICIPATIONS = [
  {
    courseParticipationID: 'participation-1',
    student: { id: 'student-1', firstName: 'Carla', lastName: 'Chen' },
  },
  { courseParticipationID: 'participation-2', student: { id: 'student-2', firstName: 'Bruno' } },
]

describe('getParticipantName', () => {
  it('finds a participant by course participation id', () => {
    expect(getParticipantName('participation-1', PARTICIPATIONS)).toBe('Carla Chen')
  })

  it('finds a participant by student id', () => {
    expect(getParticipantName('student-1', PARTICIPATIONS)).toBe('Carla Chen')
  })

  it('uses whichever name parts are present', () => {
    expect(getParticipantName('student-2', PARTICIPATIONS)).toBe('Bruno')
  })

  it('returns undefined for an unknown id', () => {
    expect(getParticipantName('unknown', PARTICIPATIONS)).toBeUndefined()
    expect(getParticipantName('participation-1', [])).toBeUndefined()
  })
})
