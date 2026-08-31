import type { InterviewReview } from '../../interfaces/InterviewReview'
import { interviewAxiosInstance } from '../interviewServerConfig'

export const getInterviewReviews = async (coursePhaseID: string): Promise<InterviewReview[]> => {
  const response = await interviewAxiosInstance.get(
    `interview/api/course_phase/${coursePhaseID}/interview-review`,
  )
  return response.data ?? []
}
