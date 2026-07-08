import type { FeedbackItem } from '../../interfaces/feedbackItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getFeedbackItemsForStudent = async (
  coursePhaseID: string,
  courseParticipationID: string,
): Promise<FeedbackItem[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/evaluation/feedback-items/course-participation/${courseParticipationID}`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
