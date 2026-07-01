import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { EvaluationCompletionRequest } from '../../../../../interfaces/evaluationCompletion'
import { unmarkMyEvaluationAsCompleted } from '../../../../../network/mutations/unmarkMyEvaluationAsCompleted'

export const useUnmarkMyEvaluationAsCompleted = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluationCompletion: EvaluationCompletionRequest) => {
      return unmarkMyEvaluationAsCompleted(phaseId ?? '', evaluationCompletion)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-evaluation-completions', phaseId] })
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
