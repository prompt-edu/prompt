import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getOwnCourseIDs = async (): Promise<string[]> => {
  try {
    return (await axiosInstance.get(`/api/courses/self`)).data
  } catch (err: any) {
    console.error(err)
    if (err.response?.status === 401) {
      return [] // case that user has access to no courses
    }
    throw err
  }
}
