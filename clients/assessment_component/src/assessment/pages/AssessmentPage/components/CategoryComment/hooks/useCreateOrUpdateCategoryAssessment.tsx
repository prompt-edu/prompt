import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { AxiosError } from 'axios'

import { createOrUpdateCategoryAssessment } from '../../../../../network/mutations/createOrUpdateCategoryAssessment'
import { CreateOrUpdateCategoryAssessmentRequest } from '../../../../../interfaces/categoryAssessment'

export const useCreateOrUpdateCategoryAssessment = (
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (req: CreateOrUpdateCategoryAssessmentRequest) =>
      createOrUpdateCategoryAssessment(phaseId ?? '', { ...req, coursePhaseID: phaseId ?? '' }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessments', phaseId] })
      setError(undefined)
    },
    onError: (error: unknown) => {
      const serverError =
        error instanceof AxiosError
          ? (error.response?.data as { error?: string } | undefined)?.error
          : undefined
      setError(serverError ?? 'An unexpected error occurred. Please try again.')
    },
  })
}
