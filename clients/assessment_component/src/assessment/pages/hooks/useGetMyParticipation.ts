import { useQuery } from '@tanstack/react-query'
import {
  type CoursePhaseParticipationWithStudent,
  getOwnCoursePhaseParticipation,
} from '@tumaet/prompt-shared-state'
import { useParams } from 'react-router-dom'

import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export const useGetMyParticipation = (options: { enabled: boolean }) => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CoursePhaseParticipationWithStudent>({
    queryKey: ['course_phase_participation', phaseId],
    queryFn: () => getOwnCoursePhaseParticipation(phaseId ?? ''),
    enabled: options.enabled,
    staleTime: SHELL_QUERY_STALE_TIME,
  })
}
