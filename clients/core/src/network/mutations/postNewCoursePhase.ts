import { CreateCoursePhase } from '@tumaet/prompt-shared-state'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const postNewCoursePhase = async (
  coursePhase: CreateCoursePhase,
): Promise<string | undefined> => {
  try {
    return (
      await axiosInstance.post(`/api/course_phases/course/${coursePhase.courseID}`, coursePhase, {
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
