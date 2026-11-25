import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const deleteTeam = async (coursePhaseID: string, teamID: string): Promise<void> => {
  await selfTeamAllocationAxiosInstance.delete(
    `/self-team-allocation/api/course_phase/${coursePhaseID}/team/${teamID}`,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
