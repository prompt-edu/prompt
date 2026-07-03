import type { CoursePhaseConfig } from '../../interfaces/coursePhaseConfig'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getCoursePhaseConfig = async (coursePhaseID: string): Promise<CoursePhaseConfig> => {
  try {
    return (
      await assessmentAxiosInstance.get(`assessment/api/course_phase/${coursePhaseID}/config`)
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
