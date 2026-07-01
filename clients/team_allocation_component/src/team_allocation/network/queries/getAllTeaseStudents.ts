import type { TeaseStudent } from '../../interfaces/tease/student'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getAllTeaseStudents = async (coursePhaseID: string): Promise<TeaseStudent[]> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/tease/course_phase/${coursePhaseID}/students`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
