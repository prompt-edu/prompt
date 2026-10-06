import { describe, expect, it } from 'vitest'

import { interviewKeys } from './keys'

const PHASE = 'phase-1'

describe('interviewKeys', () => {
  it('builds the phase keys', () => {
    expect(interviewKeys.slots(PHASE)).toEqual(['interviewSlotsWithAssignments', PHASE])
    expect(interviewKeys.reviews(PHASE)).toEqual(['interviewReviews', PHASE])
    expect(interviewKeys.myAssignment(PHASE)).toEqual(['myInterviewAssignment', PHASE])
  })

  it('keeps a missing id in the key rather than coercing it', () => {
    expect(interviewKeys.slots(undefined)).toEqual(['interviewSlotsWithAssignments', undefined])
  })

  it('reproduces the keys owned by the shared state package', () => {
    expect(interviewKeys.participants(PHASE)).toEqual(['participants', PHASE])
    expect(interviewKeys.coursePhase(PHASE)).toEqual(['course_phase', PHASE])
  })
})
