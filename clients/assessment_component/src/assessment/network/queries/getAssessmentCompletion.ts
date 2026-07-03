import type { AssessmentCompletion } from '../../interfaces/assessmentCompletion'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAssessmentCompletion = async (
  coursePhaseID: string,
  courseParticipationID: string,
): Promise<AssessmentCompletion> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/completed/course-participation/${courseParticipationID}`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
