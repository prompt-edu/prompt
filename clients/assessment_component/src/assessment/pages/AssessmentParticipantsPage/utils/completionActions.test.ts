import { describe, expect, it } from 'vitest'
import type { AssessmentCompletion } from '../../../interfaces/assessmentCompletion'
import {
  canMarkAnyAsCompleted,
  canUnmarkAny,
  summarizeMarkResult,
  summarizeUnmarkResult,
} from './completionActions'

const completion = (courseParticipationID: string, completed: boolean): AssessmentCompletion => ({
  courseParticipationID,
  coursePhaseID: 'phase-1',
  completedAt: '2026-01-01T00:00:00Z',
  completed,
  author: 'Author',
  comment: '',
  gradeSuggestion: 2,
})

const COMPLETIONS = [completion('draft', false), completion('final', true)]

describe('canMarkAnyAsCompleted', () => {
  it('is true when a selected student has a saved, non-final assessment', () => {
    expect(canMarkAnyAsCompleted(['final', 'draft'], COMPLETIONS)).toBe(true)
  })

  it('is false when every selected assessment is final or missing', () => {
    expect(canMarkAnyAsCompleted(['final', 'missing'], COMPLETIONS)).toBe(false)
  })
})

describe('canUnmarkAny', () => {
  it('is true when a selected assessment is final', () => {
    expect(canUnmarkAny(['draft', 'final'], COMPLETIONS)).toBe(true)
  })

  it('is false when no selected assessment is final', () => {
    expect(canUnmarkAny(['draft', 'missing'], COMPLETIONS)).toBe(false)
  })
})

describe('summarizeMarkResult', () => {
  it('reports the marked count without a description when nothing was skipped', () => {
    expect(summarizeMarkResult({ marked: ['a'], skipped: [] })).toEqual({
      title: 'Marked 1 assessment as final',
      description: undefined,
      variant: 'default',
    })
  })

  it('groups the skipped students by reason', () => {
    expect(
      summarizeMarkResult({
        marked: ['a', 'b'],
        skipped: [
          { courseParticipationID: 'c', reason: 'remaining_assessments' },
          { courseParticipationID: 'd', reason: 'remaining_assessments' },
          { courseParticipationID: 'e', reason: 'no_completion' },
        ],
      }),
    ).toEqual({
      title: 'Marked 2 assessments as final',
      description: '3 skipped: 2 still have unassessed competencies, 1 has no grade suggestion yet',
      variant: 'default',
    })
  })

  it('flags a run that marked nothing', () => {
    expect(
      summarizeMarkResult({
        marked: [],
        skipped: [{ courseParticipationID: 'a', reason: 'already_completed' }],
      }),
    ).toEqual({
      title: 'No assessments marked as final',
      description: '1 skipped: 1 was already final',
      variant: 'destructive',
    })
  })
})

describe('summarizeUnmarkResult', () => {
  it('reports the reopened count and the skipped students', () => {
    expect(
      summarizeUnmarkResult({
        unmarked: ['a', 'b', 'c'],
        skipped: [
          { courseParticipationID: 'd', reason: 'not_completed' },
          { courseParticipationID: 'e', reason: 'not_completed' },
        ],
      }),
    ).toEqual({
      title: 'Reopened 3 assessments for editing',
      description: '2 skipped: 2 were not final',
      variant: 'default',
    })
  })

  it('flags a run that reopened nothing', () => {
    expect(summarizeUnmarkResult({ unmarked: [], skipped: [] })).toEqual({
      title: 'No assessments reopened',
      description: undefined,
      variant: 'destructive',
    })
  })
})
