import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CourseTeam } from '../interfaces/TeamMember'

export const courseTeamQueryKey = (courseId: string) => ['courseTeam', courseId] as const

const getCourseTeam = async (courseId: string): Promise<CourseTeam> => {
  return (await axiosInstance.get(`/api/keycloak/${courseId}/group/team`)).data
}

export const useCourseTeam = (courseId: string | undefined) => {
  return useQuery({
    queryKey: courseTeamQueryKey(courseId ?? ''),
    queryFn: () => getCourseTeam(courseId as string),
    enabled: Boolean(courseId),
    staleTime: 30_000,
  })
}
