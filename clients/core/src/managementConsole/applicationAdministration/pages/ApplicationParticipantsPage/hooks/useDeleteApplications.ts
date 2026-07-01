import { deleteApplications } from '@core/network/mutations/deleteApplications'
import { UseMutationResult, useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const useDeleteApplications = (): UseMutationResult<void, Error, string[], unknown> => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const mutation = useMutation({
    mutationFn: (courseParticipationIDs: string[]) => {
      return deleteApplications(phaseId ?? 'undefined', courseParticipationIDs)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['application_participations', 'students', phaseId],
      })
      toast({
        title: 'Successfully deleted the applications.',
      })
    },
    onError: () => {
      toast({
        title: 'Error',
        description: 'Failed to delete the applications.',
        variant: 'destructive',
      })
    },
  })

  return mutation
}
