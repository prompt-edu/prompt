import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CourseParticipation } from '@core/managementConsole/shared/interfaces/CourseParticipation'

export const getCourseParticipation = async (courseId: string): Promise<CourseParticipation> => {
  try {
    return (await axiosInstance.get(`/api/courses/${courseId}/participations/self`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
