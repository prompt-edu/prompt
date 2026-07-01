import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { CreateOrUpdateAssessmentRequest } from '../../../../../interfaces/assessment'
import { createOrUpdateAssessment } from '../../../../../network/mutations/createOrUpdateAssessment'

export const useCreateOrUpdateAssessment = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessment: CreateOrUpdateAssessmentRequest) => {
      assessment.coursePhaseID = phaseId ?? ''
      return createOrUpdateAssessment(phaseId ?? '', assessment)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessments', phaseId] })
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
