import { describe, expect, it } from 'vitest'
import { getApplicationParticipantPath } from './getApplicationParticipantPath'

const COURSES = [
  {
    id: 'course-1',
    coursePhases: [
      { id: 'application-phase', coursePhaseType: 'Application' },
      { id: 'interview-phase', coursePhaseType: 'Interview' },
    ],
  },
  { id: 'course-2', coursePhases: [{ id: 'interview-phase-2', coursePhaseType: 'Interview' }] },
]

describe('getApplicationParticipantPath', () => {
  it('links to the participant in the course application phase', () => {
    expect(getApplicationParticipantPath(COURSES, 'course-1', 'participation-1')).toBe(
      '/management/course/course-1/application-phase/participants/participation-1',
    )
  })

  it('returns undefined when the course has no application phase', () => {
    expect(getApplicationParticipantPath(COURSES, 'course-2', 'participation-1')).toBeUndefined()
  })

  it('returns undefined when the course or participation is unknown', () => {
    expect(getApplicationParticipantPath(COURSES, 'unknown', 'participation-1')).toBeUndefined()
    expect(getApplicationParticipantPath(COURSES, undefined, 'participation-1')).toBeUndefined()
    expect(getApplicationParticipantPath(COURSES, 'course-1', undefined)).toBeUndefined()
  })
})
