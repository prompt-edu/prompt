import type { EvaluationCompletion } from '../../../interfaces/evaluationCompletion'
import type { EvaluationCounts } from '../interfaces/EvaluationCounts'

/**
 * Creates a lookup map from evaluation completions for O(1) access
 * Structure: authorCourseParticipationID -> courseParticipationID -> completed
 */
export const createEvaluationLookup = (
  evaluationCompletions: EvaluationCompletion[] | undefined,
): Map<string, Map<string, boolean>> => {
  const lookup = new Map<string, Map<string, boolean>>()

  if (!evaluationCompletions) {
    return lookup
  }

  evaluationCompletions.forEach((completion) => {
    if (!lookup.has(completion.authorCourseParticipationID)) {
      lookup.set(completion.authorCourseParticipationID, new Map())
    }
    lookup
      .get(completion.authorCourseParticipationID)!
      .set(completion.courseParticipationID, completion.completed)
  })

  return lookup
}

/**
 * Gets evaluation completion counts for a specific participant
 */
export const getEvaluationCounts = (
  courseParticipationID: string,
  targetIds: string[],
  evaluationLookup: Map<string, Map<string, boolean>>,
): EvaluationCounts => {
  const participantEvaluations = evaluationLookup.get(courseParticipationID)

  if (!participantEvaluations) {
    return { completed: 0, total: targetIds.length }
  }

  const completedCount = targetIds.filter(
    (targetId) => participantEvaluations.get(targetId) === true,
  ).length

  return { completed: completedCount, total: targetIds.length }
}
