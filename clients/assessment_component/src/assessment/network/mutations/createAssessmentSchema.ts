import { CreateAssessmentSchemaRequest } from '../../interfaces/assessmentSchema'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createAssessmentSchema = async (
  coursePhaseID: string,
  request: CreateAssessmentSchemaRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post(
      `assessment/api/course_phase/${coursePhaseID}/assessment-schema`,
      request,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
