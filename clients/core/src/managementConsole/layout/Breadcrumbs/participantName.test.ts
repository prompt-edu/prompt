import { describe, expect, it } from 'vitest'

import { participantName } from './participantName'

const participations = [
  {
    courseParticipationID: 'participation-ada',
    student: { id: 'student-ada', firstName: 'Ada', lastName: 'Lovelace' },
  },
  {
    courseParticipationID: 'participation-grace',
    student: { id: 'student-grace', firstName: 'Grace' },
  },
]

describe('participantName', () => {
  it('names the participant of an application details path', () => {
    expect(participantName('participation-ada', participations)).toBe('Ada Lovelace')
  })

  it('names the student of an interview profile path', () => {
    expect(participantName('student-ada', participations)).toBe('Ada Lovelace')
  })

  it('leaves out a missing part of the name', () => {
    expect(participantName('student-grace', participations)).toBe('Grace')
  })

  it('has no name for an id outside the phase', () => {
    expect(participantName('student-alan', participations)).toBeUndefined()
  })
})
