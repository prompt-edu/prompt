import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { EvaluationCompletionRequest } from '../../../../../interfaces/evaluationCompletion'
import { markMyEvaluationAsCompleted } from '../../../../../network/mutations/markMyEvaluationAsCompleted'

export const useMarkMyEvaluationAsCompleted = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluationCompletion: EvaluationCompletionRequest) => {
      return markMyEvaluationAsCompleted(phaseId ?? '', evaluationCompletion)
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
