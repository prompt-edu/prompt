import { CreateOrUpdateCategoryAssessmentRequest } from '../../interfaces/categoryAssessment'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createOrUpdateCategoryAssessment = async (
  coursePhaseID: string,
  categoryAssessment: CreateOrUpdateCategoryAssessmentRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post<void>(
      `assessment/api/course_phase/${coursePhaseID}/category-assessment`,
      categoryAssessment,
      {
        headers: {
          'Content-Type': 'application/json',
        },
        timeout: 10000,
      },
    )
  } catch (err) {
    console.error('Failed to create or update category assessment:', err)
    throw err
  }
}
