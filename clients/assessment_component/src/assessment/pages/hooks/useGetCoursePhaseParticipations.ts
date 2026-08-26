import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_PARTICIPATIONS: AssessmentParticipationWithStudent[] = []

export const useGetCoursePhaseParticipations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<AssessmentParticipationWithStudent[]>({
    queryKey: assessmentKeys.participants(phaseId),
    queryFn: () => assessmentApi.config.participations(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: data ?? EMPTY_PARTICIPATIONS }
}
