import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CreateOrUpdateEvaluationRequest } from '../../../../../interfaces/evaluation'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useCreateOrUpdateEvaluation = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluation: CreateOrUpdateEvaluationRequest) => {
      return assessmentApi.evaluations.save(phaseId ?? '', evaluation)
    },
    onSuccess: () => {
      assessmentCache.myEvaluationWritten(queryClient, phaseId)
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
