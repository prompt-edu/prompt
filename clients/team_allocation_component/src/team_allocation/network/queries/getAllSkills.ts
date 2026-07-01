import type { Skill } from '../../interfaces/skill'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getAllSkills = async (coursePhaseID: string): Promise<Skill[]> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/skill`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
