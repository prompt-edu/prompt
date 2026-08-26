import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Evaluation } from '../../interfaces/evaluation'
import { assessmentKeys } from '../../network/cache'
import { getAllEvaluations } from '../../network/queries/getAllEvaluations'

const EMPTY_EVALUATIONS: Evaluation[] = []

export const useGetAllEvaluations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Evaluation[]>({
    queryKey: assessmentKeys.evaluations.inPhase(phaseId),
    queryFn: () => getAllEvaluations(phaseId ?? ''),
    enabled: !!phaseId,
  })

  return { ...queryInfo, data: data ?? EMPTY_EVALUATIONS }
}
