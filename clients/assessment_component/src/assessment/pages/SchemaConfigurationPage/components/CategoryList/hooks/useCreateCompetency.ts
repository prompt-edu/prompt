import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CreateCompetencyRequest } from '../../../../../interfaces/competency'
import { assessmentCache } from '../../../../../network/cache'
import { createCompetency } from '../../../../../network/mutations/createCompetency'

export const useCreateCompetency = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (competency: CreateCompetencyRequest) =>
      createCompetency(phaseId ?? '', competency),
    onSuccess: () => {
      assessmentCache.schemaChanged(queryClient, phaseId)
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
