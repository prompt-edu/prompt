import type { Student } from '@tumaet/prompt-shared-state'
import { coreAxiosInstance } from '../coreServerConfig'

export const getStudentsOfCoursePhase = async (coursePhaseID: string): Promise<Student[]> => {
  try {
    const response = await coreAxiosInstance.get(
      `/api/course_phases/${coursePhaseID}/participations/students`,
    )
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
