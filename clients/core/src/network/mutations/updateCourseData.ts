import { axiosInstance } from '@tumaet/prompt-shared-state'
import { serializeUpdateCourse } from '@core/managementConsole/courseOverview/interfaces/postCourse'
import type { UpdateCourseData } from '@tumaet/prompt-shared-state'

export const updateCourseData = async (
  courseID: string,
  courseData: UpdateCourseData,
): Promise<void> => {
  const serializedCourse = serializeUpdateCourse(courseData)
  try {
    await axiosInstance.put(`/api/courses/${courseID}`, serializedCourse, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
