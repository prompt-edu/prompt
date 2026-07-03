import type { StudentAssessmentResults } from '../../interfaces/assessmentResults'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getMyAssessmentResults = async (
  coursePhaseID: string,
): Promise<StudentAssessmentResults> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/my-results`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
