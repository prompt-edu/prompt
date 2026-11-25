import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const getConfig = async (coursePhaseID: string): Promise<Record<string, boolean>> => {
  return (
    await selfTeamAllocationAxiosInstance.get(
      `/self-team-allocation/api/course_phase/${coursePhaseID}/config`,
    )
  ).data
}
