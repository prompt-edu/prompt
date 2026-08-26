import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { StudentEvaluationResults } from '../../../interfaces/evaluationResults'
import { assessmentKeys } from '../../../network/cache'
import { getMyEvaluationResults } from '../../../network/queries/getMyEvaluationResults'

export const useGetMyEvaluationResults = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()
  const enabled = (options?.enabled ?? true) && !!phaseId

  const { data, ...queryInfo } = useQuery<StudentEvaluationResults | null>({
    queryKey: assessmentKeys.results.myEvaluation(phaseId),
    queryFn: () => getMyEvaluationResults(phaseId ?? ''),
    enabled,
  })

  return { ...queryInfo, data: enabled ? (data ?? null) : null }
}
