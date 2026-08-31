import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { isAxiosError } from 'axios'
import type { UpdateInterviewReviewRequest } from '../../interfaces/InterviewReview'
import { updateInterviewReview } from '../mutations/updateInterviewReview'

interface UpdateInterviewReviewVariables {
  courseParticipationID: string
  review: UpdateInterviewReviewRequest
}

interface ErrorResponse {
  error?: string
}

export const useUpdateInterviewReview = (coursePhaseID: string | undefined) => {
  const queryClient = useQueryClient()
  const { toast } = useToast()

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
    onError: (error: unknown) => {
      // The card saves on blur, so an unreported failure would look exactly like a successful save.
      queryClient.invalidateQueries({ queryKey: ['interviewReviews', coursePhaseID] })
      toast({
        title: 'Saving the interview review failed',
        description: isAxiosError<ErrorResponse>(error)
          ? (error.response?.data?.error ?? 'Your latest changes were not saved.')
          : 'Your latest changes were not saved.',
        variant: 'destructive',
      })
    },
  })
}
