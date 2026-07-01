import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'

import { AssessmentType } from '../../interfaces/assessmentType'
import { Evaluation } from '../../interfaces/evaluation'

import { getMyEvaluations } from '../../network/queries/getMyEvaluations'

export const useGetMyEvaluations = (options?: { enabled?: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Evaluation[]>({
    queryKey: ['my-evaluations', phaseId],
    queryFn: () => getMyEvaluations(phaseId ?? ''),
    enabled: options?.enabled,
  })

  const evaluations = useMemo(() => data || [], [data])

  const selfEvaluations = useMemo(
    () => evaluations.filter((evaluation) => evaluation.type === AssessmentType.SELF),
    [evaluations],
  )

  const peerEvaluations = useMemo(
    () => evaluations.filter((evaluation) => evaluation.type === AssessmentType.PEER),
    [evaluations],
  )

  const tutorEvaluations = useMemo(
    () => evaluations.filter((evaluation) => evaluation.type === AssessmentType.TUTOR),
    [evaluations],
  )

  return {
    selfEvaluations,
    peerEvaluations,
    tutorEvaluations,
    evaluations,
    ...queryInfo,
  }
}
