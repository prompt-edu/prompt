import type { FeedbackItem } from '../../interfaces/feedbackItem'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllFeedbackItems = async (coursePhaseID: string): Promise<FeedbackItem[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/evaluation/feedback-items`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
