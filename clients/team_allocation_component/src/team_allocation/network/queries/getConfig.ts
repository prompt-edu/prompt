import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getConfig = async (coursePhaseID: string): Promise<Record<string, boolean>> => {
  return (
    await teamAllocationAxiosInstance.get(
      `/team-allocation/api/course_phase/${coursePhaseID}/config`,
    )
  ).data
}
