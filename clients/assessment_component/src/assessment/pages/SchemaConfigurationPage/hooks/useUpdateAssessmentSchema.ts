import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { UpdateAssessmentSchemaRequest } from '../../../interfaces/assessmentSchema'
import { assessmentApi } from '../../../network/api'
import { assessmentCache } from '../../../network/cache'

export const useUpdateAssessmentSchema = (
  schemaID: string,
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: UpdateAssessmentSchemaRequest) =>
      assessmentApi.schemas.update(phaseId ?? '', schemaID, request),
    onSuccess: () => {
      assessmentCache.schemaListChanged(queryClient, phaseId)
      setError(undefined)
    },
    onError: (error: any) => {
      if (error?.response?.data?.error) {
        setError(error.response.data.error)
      } else {
        setError('An unexpected error occurred. Please try again.')
      }
    },
  })
}
