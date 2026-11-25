import type { Team } from '@tumaet/prompt-shared-state'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getAllTeams = async (coursePhaseID: string): Promise<Team[]> => {
  return (
    await teamAllocationAxiosInstance.get(`/team-allocation/api/course_phase/${coursePhaseID}/team`)
  ).data.teams
}
