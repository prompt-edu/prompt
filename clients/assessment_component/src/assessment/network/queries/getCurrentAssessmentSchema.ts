import type { AssessmentSchema } from '../../interfaces/assessmentSchema'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getCurrentAssessmentSchema = async (
  coursePhaseID: string,
): Promise<AssessmentSchema> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/assessment-schema/current`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
