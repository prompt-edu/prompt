import { useMutation, useQueryClient } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { useParams } from 'react-router-dom'
import type { CreateOrUpdateCategoryAssessmentRequest } from '../../../../../interfaces/categoryAssessment'
import { createOrUpdateCategoryAssessment } from '../../../../../network/mutations/createOrUpdateCategoryAssessment'

export const useCreateOrUpdateCategoryAssessment = (
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (req: CreateOrUpdateCategoryAssessmentRequest) =>
      createOrUpdateCategoryAssessment(phaseId ?? '', req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessments', phaseId] })
      setError(undefined)
    },
    onError: (error: unknown) => {
      const serverError = isAxiosError<{ error?: string }>(error)
        ? error.response?.data?.error
        : undefined
      setError(serverError ?? 'An unexpected error occurred. Please try again.')
    },
  })
}
