import { updateCoursePhase } from '@core/network/mutations/updateCoursePhase'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { UpdateCoursePhase } from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

export const useModifyCoursePhase = (onSuccess: () => void, onError: () => void) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (coursePhase: UpdateCoursePhase) => {
      return updateCoursePhase(coursePhase)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['course_phase', phaseId] })
      onSuccess()
    },
    onError: () => {
      onError()
    },
  })
}
