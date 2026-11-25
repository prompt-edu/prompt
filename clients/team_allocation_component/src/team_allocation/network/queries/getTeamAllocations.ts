import type { Allocation } from '../../interfaces/allocation'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getTeamAllocations = async (coursePhaseID: string): Promise<Allocation[]> => {
  return (
    await teamAllocationAxiosInstance.get(
      `/team-allocation/api/tease/course_phase/${coursePhaseID}/allocations`,
    )
  ).data
}
