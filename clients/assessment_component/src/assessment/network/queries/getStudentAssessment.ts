import type { StudentAssessment } from '../../interfaces/studentAssessment'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getStudentAssessment = async (
  coursePhaseID: string,
  courseParticipationID: string,
): Promise<StudentAssessment> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment/${courseParticipationID}`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
