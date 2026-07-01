import { EvaluationCompletionRequest } from '../../interfaces/evaluationCompletion'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const markMyEvaluationAsCompleted = async (
  coursePhaseID: string,
  evaluationCompletion: EvaluationCompletionRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post(
      `assessment/api/course_phase/${coursePhaseID}/evaluation/completed/my-completion/mark-complete`,
      evaluationCompletion,
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error('Failed to mark my evaluation as completed:', err)
    throw err
  }
}
