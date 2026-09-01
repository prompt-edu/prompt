import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

export const useCourseStaff = (courseId: string | undefined) => {
  return useQuery({
    queryKey: coreKeys.courses.staff(courseId ?? ''),
    queryFn: () => coreApi.keycloak.courseStaff(courseId as string),
    enabled: Boolean(courseId),
    staleTime: 30_000,
  })
}
