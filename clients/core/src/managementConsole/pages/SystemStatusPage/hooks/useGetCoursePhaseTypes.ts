import { useQuery } from '@tanstack/react-query'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { getCoursePhaseTypes } from '../network/getCoursePhaseTypes'

export function useGetCoursePhaseTypes() {
  return useQuery<CoursePhaseType[]>({
    queryKey: ['coursePhaseType'],
    queryFn: getCoursePhaseTypes,
  })
}
