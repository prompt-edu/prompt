import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentCache } from '../../../../../network/cache'

import { unmarkAssessmentAsCompleted } from '../../../../../network/mutations/unmarkAssessmentAsCompleted'

export const useUnmarkAssessmentAsCompleted = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (courseParticipationID: string) => {
      return unmarkAssessmentAsCompleted(phaseId ?? '', courseParticipationID)
    },
    onSuccess: () => {
      assessmentCache.assessmentCompletionChanged(queryClient, phaseId)
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
