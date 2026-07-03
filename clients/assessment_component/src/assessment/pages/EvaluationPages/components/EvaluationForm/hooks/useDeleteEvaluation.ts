import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { deleteEvaluation } from '../../../../../network/mutations/deleteEvaluation'

export const useDeleteEvaluation = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (evaluationID: string) => deleteEvaluation(phaseId ?? '', evaluationID),
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
