import { useQuery } from '@tanstack/react-query'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'
import { getCoursePhaseTypes } from '../network/getCoursePhaseTypes'

export function useGetCoursePhaseTypes(forSelf?: boolean) {
  return useQuery<CoursePhaseType[]>({
    queryKey: ['coursePhaseType', forSelf ? 'self' : 'all'],
    queryFn: () => getCoursePhaseTypes(forSelf),
  })
}
