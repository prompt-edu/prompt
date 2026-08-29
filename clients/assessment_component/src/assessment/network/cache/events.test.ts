import { QueryClient } from '@tanstack/react-query'
import { beforeEach, describe, expect, it } from 'vitest'

import { AssessmentType } from '../../interfaces/assessmentType'
import { assessmentCache } from './events'
import { assessmentKeys } from './keys'

const PHASE = 'phase-1'
const OTHER_PHASE = 'phase-2'
const PARTICIPATION = 'participation-1'

let queryClient: QueryClient

const seed = (...keys: readonly (readonly unknown[])[]): void => {
  for (const key of keys) {
    queryClient.setQueryData(key, 'seeded')
  }
}

const isInvalidated = (key: readonly unknown[]): boolean =>
  queryClient.getQueryState(key)?.isInvalidated === true

beforeEach(() => {
  queryClient = new QueryClient()
})

describe('schemaChanged', () => {
  it('invalidates every cache the schema feeds, in this phase and its descendants', () => {
    seed(
      assessmentKeys.categories(PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.SELF, PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.PEER, PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.TUTOR, PHASE),
      assessmentKeys.assessments.all(),
      assessmentKeys.assessments.inPhase(PHASE),
      assessmentKeys.assessments.ofParticipant(PHASE, PARTICIPATION),
    )

    assessmentCache.schemaChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.categories(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.evaluationCategories(AssessmentType.SELF, PHASE))).toBe(
      true,
    )
    expect(isInvalidated(assessmentKeys.evaluationCategories(AssessmentType.PEER, PHASE))).toBe(
      true,
    )
    expect(isInvalidated(assessmentKeys.evaluationCategories(AssessmentType.TUTOR, PHASE))).toBe(
      true,
    )
    expect(isInvalidated(assessmentKeys.assessments.all())).toBe(true)
    expect(isInvalidated(assessmentKeys.assessments.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.assessments.ofParticipant(PHASE, PARTICIPATION))).toBe(true)
  })

  it('leaves another phase categories alone', () => {
    seed(assessmentKeys.categories(OTHER_PHASE))

    assessmentCache.schemaChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.categories(OTHER_PHASE))).toBe(false)
  })

  it('reaches every phase assessments, because the assessment key it uses is unscoped', () => {
    seed(assessmentKeys.assessments.inPhase(OTHER_PHASE))

    assessmentCache.schemaChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.assessments.inPhase(OTHER_PHASE))).toBe(true)
  })
})

describe('assessmentWritten', () => {
  it('invalidates the phase and its participants, not the other phases or the parent', () => {
    seed(
      assessmentKeys.assessments.all(),
      assessmentKeys.assessments.inPhase(PHASE),
      assessmentKeys.assessments.ofParticipant(PHASE, PARTICIPATION),
      assessmentKeys.assessments.inPhase(OTHER_PHASE),
    )

    assessmentCache.assessmentWritten(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.assessments.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.assessments.ofParticipant(PHASE, PARTICIPATION))).toBe(true)
    expect(isInvalidated(assessmentKeys.assessments.inPhase(OTHER_PHASE))).toBe(false)
    expect(isInvalidated(assessmentKeys.assessments.all())).toBe(false)
  })
})

describe('assessmentCompletionChanged', () => {
  it('invalidates the assessments, the score levels and the completions', () => {
    seed(
      assessmentKeys.assessments.inPhase(PHASE),
      assessmentKeys.scoreLevels(PHASE),
      assessmentKeys.assessmentCompletions(PHASE),
    )

    assessmentCache.assessmentCompletionChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.assessments.inPhase(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.scoreLevels(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.assessmentCompletions(PHASE))).toBe(true)
  })
})

describe('coursePhaseConfigChanged', () => {
  it('invalidates the config and the category caches it can switch off', () => {
    seed(
      assessmentKeys.coursePhaseConfig(PHASE),
      assessmentKeys.categories(PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.SELF, PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.PEER, PHASE),
      assessmentKeys.evaluationCategories(AssessmentType.TUTOR, PHASE),
    )

    assessmentCache.coursePhaseConfigChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.coursePhaseConfig(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.categories(PHASE))).toBe(true)
    expect(isInvalidated(assessmentKeys.evaluationCategories(AssessmentType.TUTOR, PHASE))).toBe(
      true,
    )
  })

  it('does not invalidate the assessments', () => {
    seed(assessmentKeys.assessments.all(), assessmentKeys.assessments.inPhase(PHASE))

    assessmentCache.coursePhaseConfigChanged(queryClient, PHASE)

    expect(isInvalidated(assessmentKeys.assessments.all())).toBe(false)
    expect(isInvalidated(assessmentKeys.assessments.inPhase(PHASE))).toBe(false)
  })
})

describe('the remaining events', () => {
  it('invalidates exactly the cache each one names', () => {
    const cases: [() => void, readonly unknown[]][] = [
      [
        () => assessmentCache.schemaListChanged(queryClient, PHASE),
        assessmentKeys.assessmentSchemas.inPhase(PHASE),
      ],
      [
        () => assessmentCache.resultsReleaseChanged(queryClient, PHASE),
        assessmentKeys.coursePhaseConfig(PHASE),
      ],
      [
        () => assessmentCache.actionItemsChanged(queryClient, PHASE),
        assessmentKeys.actionItems.inPhase(PHASE),
      ],
      [
        () => assessmentCache.myEvaluationWritten(queryClient, PHASE),
        assessmentKeys.evaluations.mine(PHASE),
      ],
      [
        () => assessmentCache.myEvaluationCompletionChanged(queryClient, PHASE),
        assessmentKeys.evaluationCompletions.mine(PHASE),
      ],
      [
        () => assessmentCache.myFeedbackItemsChanged(queryClient, PHASE),
        assessmentKeys.feedbackItems.mine(PHASE),
      ],
      [
        () => assessmentCache.coursePhaseMetaDataChanged(queryClient, PHASE),
        assessmentKeys.coursePhase(PHASE),
      ],
    ]

    for (const [fire, key] of cases) {
      queryClient = new QueryClient()
      seed(key)

      fire()

      expect(isInvalidated(key)).toBe(true)
    }
  })
})
