import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CreateOrUpdateAssessmentCompletionRequest } from '../../../../../interfaces/assessmentCompletion'
import { assessmentCache } from '../../../../../network/cache'
import { markAssessmentAsComplete } from '../../../../../network/mutations/markAssessmentAsComplete'

export const useMarkAssessmentAsComplete = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessmentCompletion: CreateOrUpdateAssessmentCompletionRequest) => {
      return markAssessmentAsComplete(phaseId ?? '', assessmentCompletion)
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
