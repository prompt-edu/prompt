import { describe, expect, it } from 'vitest'

import { isTutorEvaluationDeadlinePassed } from './isTutorEvaluationDeadlinePassed'

const NOW = new Date('2026-06-15T12:00:00Z')

describe('isTutorEvaluationDeadlinePassed', () => {
  it('is true once the deadline lies in the past', () => {
    expect(isTutorEvaluationDeadlinePassed(new Date('2026-06-14T12:00:00Z'), NOW)).toBe(true)
  })

  it('is false while the deadline lies ahead', () => {
    expect(isTutorEvaluationDeadlinePassed(new Date('2026-06-16T12:00:00Z'), NOW)).toBe(false)
  })

  it('reads the ISO string the server sends', () => {
    expect(isTutorEvaluationDeadlinePassed('2026-06-14T12:00:00Z', NOW)).toBe(true)
  })

  it('treats a missing deadline as open', () => {
    expect(isTutorEvaluationDeadlinePassed(undefined, NOW)).toBe(false)
  })

  it("treats Go's zero time, sent for an unset deadline, as open", () => {
    expect(isTutorEvaluationDeadlinePassed('0001-01-01T00:00:00Z', NOW)).toBe(false)
  })
})
