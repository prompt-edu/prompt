import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CourseTemplateStatus } from '@core/interfaces/courseTemplateStatus'

export const checkCourseTemplateStatus = async (
  courseID: string,
): Promise<CourseTemplateStatus> => {
  try {
    const response = await axiosInstance.get(`/api/courses/${courseID}/template`, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
