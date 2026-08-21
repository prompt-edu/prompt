import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateInterviewReviewRequest } from '../../interfaces/InterviewReview'
import { updateInterviewReview } from '../mutations/updateInterviewReview'

interface UpdateInterviewReviewVariables {
  courseParticipationID: string
  review: UpdateInterviewReviewRequest
}

export const useUpdateInterviewReview = (coursePhaseID: string | undefined) => {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ courseParticipationID, review }: UpdateInterviewReviewVariables) => {
      if (!coursePhaseID) {
        throw new Error('Cannot update interview review without a course phase ID')
      }
      return updateInterviewReview(coursePhaseID, courseParticipationID, review)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewReviews', coursePhaseID] })
    },
  })
}
