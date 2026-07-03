import type { FeedbackItem } from '../../interfaces/feedbackItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getFeedbackItemsForTutorInPhase = async (
  coursePhaseID: string,
  tutorParticipationID: string,
): Promise<FeedbackItem[]> => {
  const response = await assessmentAxiosInstance.get<FeedbackItem[]>(
    `assessment/api/course_phase/${coursePhaseID}/evaluation/feedback-items/tutor/${tutorParticipationID}`,
  )
  return response.data
}
