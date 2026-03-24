import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { CreateOrUpdateCoursePhaseConfigRequest } from '../../../interfaces/coursePhaseConfig'
import { createOrUpdateCoursePhaseConfig } from '../../../network/mutations/createOrUpdateCoursePhaseConfig'

export const useCreateOrUpdateCoursePhaseConfig = (
  setError: (error: string | undefined) => void,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: CreateOrUpdateCoursePhaseConfigRequest) =>
      createOrUpdateCoursePhaseConfig(phaseId ?? '', request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['coursePhaseConfig', phaseId] })
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
