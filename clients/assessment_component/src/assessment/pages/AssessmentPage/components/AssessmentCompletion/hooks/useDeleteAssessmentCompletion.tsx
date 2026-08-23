import { useMutation, useQueryClient } from '@tanstack/react-query'
import { isAxiosError } from 'axios'
import { useParams } from 'react-router-dom'

import { deleteAssessmentCompletion } from '../../../../../network/mutations/deleteAssessmentCompletion'

export const useDeleteAssessmentCompletion = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (courseParticipationID: string) => {
      return deleteAssessmentCompletion(phaseId ?? '', courseParticipationID)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['assessments', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['scoreLevels', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['assessmentCompletions', phaseId] })
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
