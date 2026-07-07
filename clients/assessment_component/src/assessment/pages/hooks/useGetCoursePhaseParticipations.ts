import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import { getCoursePhaseParticipations } from '../../network/queries/getCoursePhaseParticipations'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

const EMPTY_PARTICIPATIONS: AssessmentParticipationWithStudent[] = []

export const useGetCoursePhaseParticipations = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  const { data, ...queryInfo } = useQuery<AssessmentParticipationWithStudent[]>({
    queryKey: ['participants', phaseId],
    queryFn: () => getCoursePhaseParticipations(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })

  return { ...queryInfo, data: data ?? EMPTY_PARTICIPATIONS }
}
