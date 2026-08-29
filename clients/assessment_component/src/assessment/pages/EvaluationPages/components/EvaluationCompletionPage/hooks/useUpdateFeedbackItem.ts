import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { UpdateFeedbackItemRequest } from '../../../../../interfaces/feedbackItem'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useUpdateFeedbackItem = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (feedbackItem: UpdateFeedbackItemRequest) => {
      return assessmentApi.feedbackItems.update(phaseId ?? '', feedbackItem.id, feedbackItem)
    },
    onSuccess: () => {
      assessmentCache.myFeedbackItemsChanged(queryClient, phaseId)
      setError(undefined)
    },
    onError: (error: any) => {
      if (error?.response?.data?.error) {
        const serverError = error.response.data?.error
        setError(serverError)
      } else {
        setError('An unexpected error occurred. Please try again.')
      }
    },
  })
}
