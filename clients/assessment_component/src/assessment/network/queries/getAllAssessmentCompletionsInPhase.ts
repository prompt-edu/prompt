import type { AssessmentCompletion } from '../../interfaces/assessmentCompletion'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllAssessmentCompletionsInPhase = async (
  coursePhaseID: string,
): Promise<AssessmentCompletion[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/completed`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
