import type { CreateOrUpdateAssessmentCompletionRequest } from '../../interfaces/assessmentCompletion'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const updateAssessmentCompletion = async (
  coursePhaseID: string,
  assessmentCompletion: CreateOrUpdateAssessmentCompletionRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put(
      `assessment/api/course_phase/${coursePhaseID}/student-assessment/completed`,
      assessmentCompletion,
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
