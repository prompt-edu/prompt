import { coreApi } from '@core/network/api'
import { coreCache } from '@core/network/cache'
import { type UseMutationResult, useMutation, useQueryClient } from '@tanstack/react-query'
import { useToast } from '@tumaet/prompt-ui-components'
import { useParams } from 'react-router-dom'

export const useDeleteApplications = (): UseMutationResult<void, Error, string[], unknown> => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

  const mutation = useMutation({
    mutationFn: (courseParticipationIDs: string[]) => {
      return coreApi.applications.remove(phaseId ?? 'undefined', courseParticipationIDs)
    },
    onSuccess: () => {
      coreCache.applicationParticipantsChanged(queryClient, phaseId)
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
