import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const createTeam = async (coursePhaseID: string, teamNames: string[]): Promise<void> => {
  const teamRequest = {
    teamNames: teamNames,
  }
  await selfTeamAllocationAxiosInstance.post(
    `/self-team-allocation/api/course_phase/${coursePhaseID}/team`,
    teamRequest,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
