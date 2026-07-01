import type { EvaluationCompletionRequest } from '../../interfaces/evaluationCompletion'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const unmarkMyEvaluationAsCompleted = async (
  coursePhaseID: string,
  evaluationCompletion: EvaluationCompletionRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put(
      `assessment/api/course_phase/${coursePhaseID}/evaluation/completed/my-completion/unmark`,
      evaluationCompletion,
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error('Failed to unmark my evaluation as completed:', err)
    throw err
  }
}
