import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { CreateOrUpdateEvaluationRequest } from '../../../../../interfaces/evaluation'
import { createOrUpdateEvaluation } from '../../../../../network/mutations/createOrUpdateEvaluation'

export const useCreateOrUpdateEvaluation = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluation: CreateOrUpdateEvaluationRequest) => {
      return createOrUpdateEvaluation(phaseId ?? '', evaluation)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-evaluations', phaseId] })
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
