import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'

export function useGetCoursePhaseTypes(forSelf?: boolean) {
  return useQuery<CoursePhaseType[]>({
    queryKey: coreKeys.coursePhases.typesForScope(forSelf ? 'self' : 'all'),
    queryFn: () => coreApi.coursePhases.listTypesForScope(forSelf),
  })
}
