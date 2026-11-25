import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const deleteTeam = async (coursePhaseID: string, teamID: string): Promise<void> => {
  await teamAllocationAxiosInstance.delete(
    `/team-allocation/api/course_phase/${coursePhaseID}/team/${teamID}`,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
