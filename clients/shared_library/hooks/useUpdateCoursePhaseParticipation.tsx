import { useToast } from '@tumaet/prompt-ui-components'
import { useMutation, UseMutationResult, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import {
  UpdateCoursePhaseParticipation,
  CoursePhaseParticipationWithStudent,
  CoursePhaseParticipationsWithResolution,
} from '@tumaet/prompt-shared-state'
import { updateCoursePhaseParticipation } from '@/network/mutations/updateCoursePhaseParticipationMetaData'

interface MutationContext {
  previousParticipants?:
    | CoursePhaseParticipationsWithResolution
    | CoursePhaseParticipationWithStudent[]
}

export const useUpdateCoursePhaseParticipation = (): UseMutationResult<
  string | undefined,
  Error,
  UpdateCoursePhaseParticipation,
  MutationContext
> => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  // TODO: This requires fixing through a shared library!!!
  const { toast } = useToast()

  const mutation = useMutation<
    string | undefined,
    Error,
    UpdateCoursePhaseParticipation,
    MutationContext
  >({
    mutationFn: (coursePhaseParticipation: UpdateCoursePhaseParticipation) => {
      return updateCoursePhaseParticipation(coursePhaseParticipation)
    },
    // Optimistically update cache before server responds
    onMutate: async (newParticipationData) => {
      // Cancel any outgoing refetches
      await queryClient.cancelQueries({ queryKey: ['participants', phaseId] })

      // Snapshot the previous value
      const previousParticipants = queryClient.getQueryData<
        CoursePhaseParticipationsWithResolution | CoursePhaseParticipationWithStudent[]
      >(['participants', phaseId])

      const applyParticipationUpdate = (
        participants: CoursePhaseParticipationWithStudent[],
      ): CoursePhaseParticipationWithStudent[] =>
        participants.map((participant) => {
          if (participant.courseParticipationID === newParticipationData.courseParticipationID) {
            return {
              ...participant,
              passStatus: newParticipationData.passStatus ?? participant.passStatus,
              restrictedData: {
                ...participant.restrictedData,
                ...newParticipationData.restrictedData,
              },
              studentReadableData: {
                ...participant.studentReadableData,
                ...newParticipationData.studentReadableData,
              },
            }
          }

          return participant
        })

      // Optimistically update the specific participation
      if (previousParticipants) {
        if (Array.isArray(previousParticipants)) {
          queryClient.setQueryData(
            ['participants', phaseId],
            applyParticipationUpdate(previousParticipants),
          )
        } else {
          queryClient.setQueryData(['participants', phaseId], {
            ...previousParticipants,
            participations: applyParticipationUpdate(previousParticipants.participations ?? []),
          })
        }
      }

      return { previousParticipants }
    },
    onSuccess: () => {
      toast({
        title: 'Success',
        description: 'Successfully updated the course participation.',
      })
    },
    onError: (error, _newParticipation, context) => {
      // Rollback on error
      if (context?.previousParticipants) {
        queryClient.setQueryData(['participants', phaseId], context.previousParticipants)
      }
      toast({
        title: 'Error',
        description: 'Failed to update the course participation.',
        variant: 'destructive',
      })
    },
    onSettled: () => {
      // Re-sync with server to ensure cache reflects the true state after success or error
      queryClient.invalidateQueries({ queryKey: ['participants', phaseId] })
    },
  })

  return mutation
}
