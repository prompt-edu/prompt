import { updateApplicationStatus } from '@core/network/mutations/updateApplicationStatus'
import { type UseMutationResult, useMutation, useQueryClient } from '@tanstack/react-query'
import type { UpdateCoursePhaseParticipationStatus } from '@tumaet/prompt-shared-state'
import { useToast } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const useApplicationStatusUpdate = (): UseMutationResult<
  void,
  Error,
  UpdateCoursePhaseParticipationStatus,
  unknown
> => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const mutation = useMutation({
    mutationFn: (updateApplication: UpdateCoursePhaseParticipationStatus) => {
      return updateApplicationStatus(phaseId ?? 'undefined', updateApplication)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['application_participations', 'students', phaseId],
      })
      toast({
        title: 'Successfully updated the status.',
      })
    },
    onError: () => {
      toast({
        title: 'Error',
        description: 'Failed to update the status.',
        variant: 'destructive',
      })
    },
  })

  return mutation
}
