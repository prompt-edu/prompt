import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { assessmentApi } from '../../../network/api'
import { assessmentCache } from '../../../network/cache'

export const useMarkAssessmentsAsCompleted = () => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (courseParticipationIDs: string[]) =>
      assessmentApi.completions.markCompleteBatch(phaseId ?? '', courseParticipationIDs),
    onSuccess: () => {
      assessmentCache.assessmentCompletionChanged(queryClient, phaseId)
    },
  })
}
