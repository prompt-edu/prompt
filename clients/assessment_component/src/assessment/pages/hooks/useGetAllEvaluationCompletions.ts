import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { EvaluationCompletion } from '../../interfaces/evaluationCompletion'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

const EMPTY_EVALUATION_COMPLETIONS: EvaluationCompletion[] = []

export const useGetAllEvaluationCompletions = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<EvaluationCompletion[]>({
    queryKey: assessmentKeys.evaluationCompletions.inPhase(phaseId),
    queryFn: () => assessmentApi.evaluationCompletions.listInPhase(phaseId ?? ''),
  })

  return { ...queryInfo, data: data ?? EMPTY_EVALUATION_COMPLETIONS }
}
