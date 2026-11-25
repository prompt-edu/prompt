import type { Tutor } from '../../interfaces/tutor'
import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const importTutors = async (coursePhaseID: string, tutors: Tutor[]): Promise<void> => {
  await selfTeamAllocationAxiosInstance.post(
    `self-team-allocation/api/course_phase/${coursePhaseID}/team/tutors`,
    tutors,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
