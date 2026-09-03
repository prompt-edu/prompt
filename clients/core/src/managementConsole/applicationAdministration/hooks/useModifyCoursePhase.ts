import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

export const useModifyCoursePhase = (onSuccess: () => void, onError: () => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => {
      return coreApi.coursePhases.update(coursePhase)
    },
    onSuccess: () => {
      coreCache.coursePhaseChanged(queryClient, phaseId)
      onSuccess()
    },
    onError: () => {
      onError()
    },
  })
}
