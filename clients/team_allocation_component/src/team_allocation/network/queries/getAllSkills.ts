import type { Skill } from '../../interfaces/skill'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getAllSkills = async (coursePhaseID: string): Promise<Skill[]> => {
  return (
    await teamAllocationAxiosInstance.get(
      `/team-allocation/api/course_phase/${coursePhaseID}/skill`,
    )
  ).data
}
