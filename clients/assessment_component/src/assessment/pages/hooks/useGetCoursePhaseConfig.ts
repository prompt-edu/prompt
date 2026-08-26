import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { assessmentApi } from '../../network/api'
import { assessmentKeys } from '../../network/cache'
import { SHELL_QUERY_STALE_TIME } from './queryConfig'

export const useGetCoursePhaseConfig = () => {
  const { phaseId } = useParams<{ phaseId: string }>()

  return useQuery<CoursePhaseConfig>({
    queryKey: assessmentKeys.coursePhaseConfig(phaseId),
    queryFn: () => assessmentApi.config.get(phaseId ?? ''),
    staleTime: SHELL_QUERY_STALE_TIME,
  })
}
