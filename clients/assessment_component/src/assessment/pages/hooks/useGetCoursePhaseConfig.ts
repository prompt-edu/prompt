import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export interface CoursePhaseConfigQueryOptions {
  // The sidebar renders outside the phase routes, so it passes the phase ID explicitly.
  coursePhaseID?: string
}

export const useGetCoursePhaseConfig = (options: CoursePhaseConfigQueryOptions = {}) => {
  const { phaseId: routePhaseId } = useParams<{ phaseId: string }>()
  const phaseId = options.coursePhaseID ?? routePhaseId

  return useQuery<CoursePhaseConfig>({
    queryKey: assessmentKeys.coursePhaseConfig(phaseId),
    queryFn: () => assessmentApi.config.get(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })
}
