import type { StudentEvaluationResults } from '../../interfaces/evaluationResults'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getMyEvaluationResults = async (
  coursePhaseID: string,
): Promise<StudentEvaluationResults | null> => {
  try {
    const response = await assessmentAxiosInstance.get(
      `assessment/api/course_phase/${coursePhaseID}/evaluation/my-results`,
    )
    // 204 when the phase is not evaluation-only or results are not released yet.
    // React Query rejects an undefined result, so this must be null.
    return response.status === 204 ? null : response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
