import type { Assessment } from '../../interfaces/assessment'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllAssessmentsInPhase = async (coursePhaseID: string): Promise<Assessment[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/student-assessment`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
