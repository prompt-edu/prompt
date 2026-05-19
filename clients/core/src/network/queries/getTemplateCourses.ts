import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { Course } from '@tumaet/prompt-shared-state'

export const getTemplateCourses = async (): Promise<Course[]> => {
  try {
    return (await axiosInstance.get(`/api/courses/template`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
