import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { CreateAssessmentSchemaRequest } from '../../../interfaces/assessmentSchema'
import { createAssessmentSchema } from '../../../network/mutations/createAssessmentSchema'

export const useCreateAssessmentSchema = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessmentSchema: CreateAssessmentSchemaRequest) =>
      createAssessmentSchema(phaseId ?? '', assessmentSchema),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessmentSchemas', phaseId] })
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
