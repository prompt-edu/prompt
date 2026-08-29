import type { QueryClient } from '@tanstack/react-query'

import { AssessmentType } from '../../interfaces/assessmentType'
import { assessmentKeys } from './keys'

type Id = string | undefined
type CacheKeys = readonly (readonly unknown[])[]

const EVALUATION_TYPES = [AssessmentType.SELF, AssessmentType.PEER, AssessmentType.TUTOR] as const

const invalidate = (queryClient: QueryClient, keys: CacheKeys): void => {
  for (const queryKey of keys) {
    queryClient.invalidateQueries({ queryKey })
  }
}

const categoryKeys = (phaseId: Id): CacheKeys => [
  assessmentKeys.categories(phaseId),
  ...EVALUATION_TYPES.map((evaluationType) =>
    assessmentKeys.evaluationCategories(evaluationType, phaseId),
  ),
]

export const assessmentCache = {
  schemaChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [...categoryKeys(phaseId), assessmentKeys.assessments.all()]),

  schemaListChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.assessmentSchemas.inPhase(phaseId)]),

  coursePhaseConfigChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.coursePhaseConfig(phaseId), ...categoryKeys(phaseId)]),

  resultsReleaseChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.coursePhaseConfig(phaseId)]),

  assessmentWritten: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.assessments.inPhase(phaseId)]),

  assessmentCompletionChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [
      assessmentKeys.assessments.inPhase(phaseId),
      assessmentKeys.scoreLevels(phaseId),
      assessmentKeys.assessmentCompletions(phaseId),
    ]),

  actionItemsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.actionItems.inPhase(phaseId)]),

  myEvaluationWritten: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.evaluations.mine(phaseId)]),

  myEvaluationCompletionChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.evaluationCompletions.mine(phaseId)]),

  myFeedbackItemsChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.feedbackItems.mine(phaseId)]),

  coursePhaseMetaDataChanged: (queryClient: QueryClient, phaseId: Id): void =>
    invalidate(queryClient, [assessmentKeys.coursePhase(phaseId)]),
}
