import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../../../../network/api'
import { assessmentCache } from '../../../../../network/cache'

export const useDeleteAssessment = (
  setError: (error: string | undefined) => void,
  independent = false,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessmentID: string) => {
      const remove = independent
        ? assessmentApi.assessments.removeIndependent
        : assessmentApi.assessments.remove
      return remove(phaseId ?? '', assessmentID)
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
