import type { Team } from '@tumaet/prompt-shared-state'

import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const getAllTeams = async (coursePhaseID: string): Promise<Team[]> => {
  return (
    await selfTeamAllocationAxiosInstance.get(
      `/self-team-allocation/api/course_phase/${coursePhaseID}/team`,
    )
  ).data.teams
}
