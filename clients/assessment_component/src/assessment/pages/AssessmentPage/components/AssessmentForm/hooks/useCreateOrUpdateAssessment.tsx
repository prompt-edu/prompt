import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { useParams } from 'react-router-dom'
import type { CreateOrUpdateAssessmentRequest } from '../../../../../interfaces/assessment'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useCreateOrUpdateAssessment = (
  setError: (error: string | undefined) => void,
  independent = false,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessment: CreateOrUpdateAssessmentRequest) => {
      assessment.coursePhaseID = phaseId ?? ''
      const save = independent
        ? assessmentApi.assessments.saveIndependent
        : assessmentApi.assessments.save
      return save(phaseId ?? '', assessment)
    },
    onSuccess: () => {
      assessmentCache.assessmentWritten(queryClient, phaseId)
      setError(undefined)
    },
    onError: (error: AxiosError<{ error?: string }>) => {
      setError(error.response?.data?.error ?? 'An unexpected error occurred. Please try again.')
    },
  })
}
