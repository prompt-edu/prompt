import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const createTeams = async (coursePhaseID: string, teamNames: string[]): Promise<void> => {
  const teamsRequest = {
    teamNames: teamNames,
  }
  await teamAllocationAxiosInstance.post(
    `/team-allocation/api/course_phase/${coursePhaseID}/team`,
    teamsRequest,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
