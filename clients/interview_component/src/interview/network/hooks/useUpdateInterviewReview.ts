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
    mutationFn: ({ courseParticipationID, review }: UpdateInterviewReviewVariables) =>
      updateInterviewReview(coursePhaseID ?? '', courseParticipationID, review),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['interviewReviews', coursePhaseID] })
    },
  })
}
