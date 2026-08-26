import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { AssessmentType } from '../../interfaces/assessmentType'
import type { Evaluation } from '../../interfaces/evaluation'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'

import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export const useGetMyEvaluations = (options: { enabled: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<Evaluation[]>({
    queryKey: assessmentKeys.evaluations.mine(phaseId),
    queryFn: () => assessmentApi.evaluations.listMine(phaseId ?? ''),
    enabled: options.enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
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
