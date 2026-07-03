import type { CreateOrUpdateCoursePhaseConfigRequest } from '../../interfaces/coursePhaseConfig'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createOrUpdateCoursePhaseConfig = async (
  coursePhaseID: string,
  request: CreateOrUpdateCoursePhaseConfigRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put(
      `assessment/api/course_phase/${coursePhaseID}/config`,
      request,
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
