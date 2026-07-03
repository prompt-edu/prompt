import type { Team } from '@tumaet/prompt-shared-state'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllTeams = async (coursePhaseID: string): Promise<Team[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(`assessment/api/course_phase/${coursePhaseID}/config/teams`)
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
