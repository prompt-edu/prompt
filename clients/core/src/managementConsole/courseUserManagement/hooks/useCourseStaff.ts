import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { CourseStaff } from '../interfaces/StaffMember'

const getCourseStaff = async (courseId: string): Promise<CourseStaff> => {
  return (await axiosInstance.get<CourseStaff>(`/api/keycloak/${courseId}/group/staff`)).data
}

export const useCourseStaff = (courseId: string | undefined) => {
  return useQuery({
    queryKey: coreKeys.courses.staff(courseId ?? ''),
    queryFn: () => getCourseStaff(courseId as string),
    enabled: Boolean(courseId),
    staleTime: 30_000,
  })
}
