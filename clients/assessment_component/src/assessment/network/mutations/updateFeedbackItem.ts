import { UpdateFeedbackItemRequest } from '../../interfaces/feedbackItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const updateFeedbackItem = async (
  coursePhaseID: string,
  feedbackItemID: string,
  feedbackItem: UpdateFeedbackItemRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put<UpdateFeedbackItemRequest>(
      `assessment/api/course_phase/${coursePhaseID}/evaluation/feedback-items/${feedbackItemID}`,
      {
        ...feedbackItem,
        coursePhaseID,
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
