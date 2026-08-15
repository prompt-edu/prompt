import type { UpdateAssessmentSchemaRequest } from '../../interfaces/assessmentSchema'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const updateAssessmentSchema = async (
  coursePhaseID: string,
  schemaID: string,
  request: UpdateAssessmentSchemaRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put(
      `assessment/api/course_phase/${coursePhaseID}/assessment-schema/${schemaID}`,
      request,
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
