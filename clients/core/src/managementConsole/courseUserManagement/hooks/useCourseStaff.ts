import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CourseStaff } from '../interfaces/StaffMember'

export const courseStaffQueryKey = (courseId: string) => ['courseStaff', courseId] as const

const getCourseStaff = async (courseId: string): Promise<CourseStaff> => {
  return (await axiosInstance.get<CourseStaff>(`/api/keycloak/${courseId}/group/staff`)).data
}

export const useCourseStaff = (courseId: string | undefined) => {
  return useQuery({
    queryKey: courseStaffQueryKey(courseId ?? ''),
    queryFn: () => getCourseStaff(courseId as string),
    enabled: Boolean(courseId),
    staleTime: 30_000,
  })
}
