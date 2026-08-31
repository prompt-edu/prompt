import type {
  InterviewReview,
  UpdateInterviewReviewRequest,
} from '../../interfaces/InterviewReview'
import { interviewAxiosInstance } from '../interviewServerConfig'

export const updateInterviewReview = async (
  coursePhaseID: string,
  courseParticipationID: string,
  review: UpdateInterviewReviewRequest,
): Promise<InterviewReview> => {
  const response = await interviewAxiosInstance.put(
    `interview/api/course_phase/${coursePhaseID}/interview-review/${courseParticipationID}`,
    review,
  )
  return response.data
}
