import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { UpdateAssessmentSchemaRequest } from '../../../interfaces/assessmentSchema'
import { updateAssessmentSchema } from '../../../network/mutations/updateAssessmentSchema'

export const useUpdateAssessmentSchema = (
  schemaID: string,
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: UpdateAssessmentSchemaRequest) =>
      updateAssessmentSchema(phaseId ?? '', schemaID, request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessmentSchemas', phaseId] })
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
