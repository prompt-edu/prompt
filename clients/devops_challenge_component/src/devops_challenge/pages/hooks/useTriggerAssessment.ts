import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { triggerAssessment } from '../../network/mutations/triggerAssessment'

export const useTriggerAssessment = (setError: (error: string | null) => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => triggerAssessment(phaseId ?? ''),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devOpsDeveloperProfile', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['devOpsPassedStudentsCount', phaseId] })
      setError(null)
    },
    onError: (error: any) => {
      queryClient.invalidateQueries({ queryKey: ['devOpsDeveloperProfile', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['devOpsPassedStudentsCount', phaseId] })
      if (error?.response?.data?.error) {
        const serverError = error.response.data?.error
        setError(serverError)
      } else {
        setError('An unexpected error occurred. Please try again.')
      }
    },
  })
}
