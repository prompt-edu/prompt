import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { getCoursePhaseConfig } from '../../network/queries/getCoursePhaseConfig'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export const useGetCoursePhaseConfig = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CoursePhaseConfig>({
    queryKey: ['coursePhaseConfig', phaseId],
    queryFn: () => getCoursePhaseConfig(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })
}
