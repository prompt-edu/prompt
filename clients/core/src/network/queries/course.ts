import { Course } from '@tumaet/prompt-shared-state'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getAllCourses = async (): Promise<Course[]> => {
  try {
    return (await axiosInstance.get(`/api/courses/`)).data
  } catch (err: any) {
    console.error(err)

    if (err.response?.status === 401) {
      return [] // case that user has access to no courses
    }
    throw err
  }
}
