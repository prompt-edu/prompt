import type { CreateOrUpdateAssessmentSchemaCoursePhaseRequest } from '../../interfaces/assessmentSchema'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createOrUpdateAssessmentSchemaCoursePhase = async (
  request: CreateOrUpdateAssessmentSchemaCoursePhaseRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put(
      `assessment/api/course_phase/${request.coursePhaseID}/assessment-schema`,
      {
        assessmentSchemaID: request.assessmentSchemaID,
      },
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
