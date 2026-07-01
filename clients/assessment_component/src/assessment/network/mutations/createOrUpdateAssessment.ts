import { CreateOrUpdateAssessmentRequest } from '../../interfaces/assessment'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createOrUpdateAssessment = async (
  coursePhaseID: string,
  assessment: CreateOrUpdateAssessmentRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post<void>(
      `assessment/api/course_phase/${coursePhaseID}/student-assessment`,
      assessment,
      {
        headers: {
          'Content-Type': 'application/json',
        },
        timeout: 10000, // 10 second timeout
      },
    )
  } catch (err) {
    console.error('Failed to create or update assessment:', err)
    throw err
  }
}
