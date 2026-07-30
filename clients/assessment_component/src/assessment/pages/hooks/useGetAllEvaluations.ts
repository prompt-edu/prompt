import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { Evaluation } from '../../interfaces/evaluation'
import { getAllEvaluations } from '../../network/queries/getAllEvaluations'

const EMPTY_EVALUATIONS: Evaluation[] = []

export const useGetAllEvaluations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Evaluation[]>({
    queryKey: ['evaluations', phaseId],
    queryFn: () => getAllEvaluations(phaseId ?? ''),
    enabled: !!phaseId,
  })

  return { ...queryInfo, data: data ?? EMPTY_EVALUATIONS }
}
