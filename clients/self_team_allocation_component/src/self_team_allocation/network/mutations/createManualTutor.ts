import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export interface CreateManualTutorRequest {
  firstName: string
  lastName: string
}

export const createManualTutor = async (
  coursePhaseID: string,
  teamID: string,
  tutor: CreateManualTutorRequest,
): Promise<void> => {
  await selfTeamAllocationAxiosInstance.post(
    `self-team-allocation/api/course_phase/${coursePhaseID}/team/${teamID}/tutor`,
    tutor,
    {
      headers: {
        'Content-Type': 'application/json',
      },
    },
  )
}
