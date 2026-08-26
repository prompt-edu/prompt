import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { UpdateCompetencyRequest } from '../../../../../../../interfaces/competency'
import { assessmentApi } from '../../../../../../../network/api'
import { assessmentCache } from '../../../../../../../network/cache'

export const useUpdateCompetency = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (competency: UpdateCompetencyRequest) =>
      assessmentApi.competencies.update(phaseId ?? '', competency),
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
