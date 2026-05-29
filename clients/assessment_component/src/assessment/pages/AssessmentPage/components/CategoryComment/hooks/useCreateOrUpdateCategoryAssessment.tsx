import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { createOrUpdateCategoryAssessment } from '../../../../../network/mutations/createOrUpdateCategoryAssessment'
import { CreateOrUpdateCategoryAssessmentRequest } from '../../../../../interfaces/categoryAssessment'

export const useCreateOrUpdateCategoryAssessment = (
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (req: CreateOrUpdateCategoryAssessmentRequest) => {
      req.coursePhaseID = phaseId ?? ''
      return createOrUpdateCategoryAssessment(phaseId ?? '', req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessments', phaseId] })
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
