import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentCache } from '../../../network/cache'

import { unreleaseResults } from '../../../network/mutations/unreleaseResults'

export const useUnreleaseResults = () => {
  const queryClient = useQueryClient()
  const { phaseId } = useParams<{ phaseId: string }>()

  const mutation = useMutation({
    mutationFn: () => unreleaseResults(phaseId ?? ''),
    onSuccess: () => {
      assessmentCache.resultsReleaseChanged(queryClient, phaseId)
    },
  })

  return mutation
}
