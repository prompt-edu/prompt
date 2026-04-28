import { AxiosError } from 'axios'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import { CreateOrUpdateCoursePhaseConfigRequest } from '../../../interfaces/coursePhaseConfig'
import { createOrUpdateCoursePhaseConfig } from '../../../network/mutations/createOrUpdateCoursePhaseConfig'

interface CreateOrUpdateCoursePhaseConfigResponseError {
  error?: string
}

interface UseCreateOrUpdateCoursePhaseConfigOptions {
  onSuccess?: () => void
  onError?: (errorMessage: string) => void
}

const getErrorMessage = (error: unknown): string => {
  const responseError = (error as AxiosError<CreateOrUpdateCoursePhaseConfigResponseError>)
    ?.response?.data?.error

  return responseError || 'An unexpected error occurred. Please try again.'
}

export const useCreateOrUpdateCoursePhaseConfig = (
  options?: UseCreateOrUpdateCoursePhaseConfigOptions,
) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: CreateOrUpdateCoursePhaseConfigRequest) =>
      createOrUpdateCoursePhaseConfig(phaseId ?? '', request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['coursePhaseConfig', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['categories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['selfEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['peerEvaluationCategories', phaseId] })
      queryClient.invalidateQueries({ queryKey: ['tutorEvaluationCategories', phaseId] })
      options?.onSuccess?.()
    },
    onError: (error: unknown) => {
      options?.onError?.(getErrorMessage(error))
    },
  })
}
