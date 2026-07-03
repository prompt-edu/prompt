import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getCoursePhaseParticipations = async (
  coursePhaseID: string,
): Promise<AssessmentParticipationWithStudent[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/config/participations`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
