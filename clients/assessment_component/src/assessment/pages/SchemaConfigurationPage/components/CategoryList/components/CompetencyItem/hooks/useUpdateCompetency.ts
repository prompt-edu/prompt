import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { UpdateCompetencyRequest } from '../../../../../../../interfaces/competency'
import { assessmentCache } from '../../../../../../../network/cache'
import { updateCompetency } from '../../../../../../../network/mutations/updateCompetency'

export const useUpdateCompetency = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (competency: UpdateCompetencyRequest) =>
      updateCompetency(phaseId ?? '', competency),
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
