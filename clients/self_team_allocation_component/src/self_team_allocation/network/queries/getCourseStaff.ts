import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { CourseStaff } from '../../interfaces/courseStaff'

export const getCourseStaff = async (courseID: string): Promise<CourseStaff> => {
  return (await axiosInstance.get(`/api/keycloak/${courseID}/group/staff`)).data
}
