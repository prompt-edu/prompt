import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'

import type { StudentEvaluationResults } from '../../../interfaces/evaluationResults'
import { getMyEvaluationResults } from '../../../network/queries/getMyEvaluationResults'

export const useGetMyEvaluationResults = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = (options?.enabled ?? true) && !!phaseId

  const query = useQuery<StudentEvaluationResults | null>({
    queryKey: ['myEvaluationResults', phaseId],
    queryFn: () => getMyEvaluationResults(phaseId ?? ''),
    enabled,
  })

  // A disabled query stays pending forever, which would hang the loading gates above it.
  // refetch() ignores `enabled`, so it has to be neutralized as well.
  if (!enabled) {
    return { ...query, data: null, isPending: false, isError: false, refetch: async () => query }
  }

  return query
}
