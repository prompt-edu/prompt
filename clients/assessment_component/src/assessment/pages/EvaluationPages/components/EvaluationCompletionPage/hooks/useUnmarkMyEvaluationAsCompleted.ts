import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { EvaluationCompletionRequest } from '../../../../../interfaces/evaluationCompletion'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useUnmarkMyEvaluationAsCompleted = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluationCompletion: EvaluationCompletionRequest) => {
      return assessmentApi.evaluationCompletions.unmarkMine(phaseId ?? '', evaluationCompletion)
    },
    onSuccess: () => {
      assessmentCache.myEvaluationCompletionChanged(queryClient, phaseId)
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
