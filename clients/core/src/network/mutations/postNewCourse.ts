import {
  PostCourse,
  serializePostCourse,
} from '../../managementConsole/courseOverview/interfaces/postCourse'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const postNewCourse = async (course: PostCourse): Promise<string | undefined> => {
  try {
    const serializedCourse = serializePostCourse(course)
    return (
      await axiosInstance.post(`/api/courses/`, serializedCourse, {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data.id // try to get the id of the created course
  } catch (err) {
    console.error(err)
    throw err
  }
}
