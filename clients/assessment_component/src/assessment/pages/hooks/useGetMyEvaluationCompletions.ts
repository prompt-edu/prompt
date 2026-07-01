import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import { EvaluationCompletion } from '../../interfaces/evaluationCompletion'
import { getMyEvaluationCompletions } from '../../network/queries/getMyEvaluationCompletions'

export const useGetMyEvaluationCompletions = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<EvaluationCompletion[]>({
    queryKey: ['my-evaluation-completions', phaseId],
    queryFn: () => getMyEvaluationCompletions(phaseId ?? ''),
    enabled: options?.enabled,
  })

  const evaluationCompletions = useMemo(() => data || [], [data])

  const selfEvaluationCompletion = useMemo(
    () => evaluationCompletions.find((completion) => completion.type === AssessmentType.SELF),
    [evaluationCompletions],
  )

  const peerEvaluationCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === AssessmentType.PEER),
    [evaluationCompletions],
  )

  const tutorEvaluationCompletions = useMemo(
    () => evaluationCompletions.filter((completion) => completion.type === AssessmentType.TUTOR),
    [evaluationCompletions],
  )

  return {
    selfEvaluationCompletion,
    peerEvaluationCompletions,
    tutorEvaluationCompletions,
    evaluationCompletions,
    ...queryInfo,
  }
}
