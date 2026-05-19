import { axiosInstance } from '@tumaet/prompt-shared-state'
import {
  CopyCourse,
  serializeCopyCourse,
} from '@core/managementConsole/courseOverview/interfaces/copyCourse'

export const copyCourse = async (
  courseID: string,
  courseVariables: CopyCourse,
): Promise<string | undefined> => {
  try {
    const serializedCourse = serializeCopyCourse(courseVariables)
    return (
      await axiosInstance.post(`/api/courses/${courseID}/copy`, serializedCourse, {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data.id // try to get the id of the copied course
  } catch (err) {
    console.error(err)
    throw err
  }
}
