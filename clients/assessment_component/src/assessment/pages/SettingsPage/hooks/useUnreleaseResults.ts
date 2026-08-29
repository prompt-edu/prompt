import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../../network/api'
import { assessmentCache } from '../../../network/cache'

export const useUnreleaseResults = () => {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  const mutation = useMutation({
    mutationFn: () => assessmentApi.config.unreleaseResults(phaseId ?? ''),
    onSuccess: () => {
      assessmentCache.resultsReleaseChanged(queryClient, phaseId)
    },
  })

  return mutation
}
