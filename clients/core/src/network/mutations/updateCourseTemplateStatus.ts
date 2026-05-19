import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CourseTemplateStatus } from '@core/interfaces/courseTemplateStatus'

export const updateCourseTemplateStatus = async (
  courseID: string,
  updateRequest: CourseTemplateStatus,
): Promise<void> => {
  try {
    return await axiosInstance.put(`/api/courses/${courseID}/template`, updateRequest, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
