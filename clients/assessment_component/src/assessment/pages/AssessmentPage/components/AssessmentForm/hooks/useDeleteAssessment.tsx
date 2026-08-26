import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentCache } from '../../../../../network/cache'
import { deleteAssessment } from '../../../../../network/mutations/deleteAssessment'

export const useDeleteAssessment = (setError: (error: string | undefined) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (assessmentID: string) => {
      return deleteAssessment(phaseId ?? '', assessmentID)
    },
    onSuccess: () => {
      assessmentCache.assessmentWritten(queryClient, phaseId)
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
