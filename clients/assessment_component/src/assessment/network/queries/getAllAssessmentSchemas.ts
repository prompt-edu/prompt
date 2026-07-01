import type { AssessmentSchema } from '../../interfaces/assessmentSchema'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllAssessmentSchemas = async (
  coursePhaseID: string,
): Promise<AssessmentSchema[]> => {
  const { data } = await assessmentAxiosInstance.get<AssessmentSchema[]>(
    `assessment/api/course_phase/${coursePhaseID}/assessment-schema`,
  )
  return data
}
