import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const deleteTutor = async (coursePhaseID: string, tutorID: string): Promise<void> => {
  await selfTeamAllocationAxiosInstance.delete(
    `self-team-allocation/api/course_phase/${coursePhaseID}/team/tutor/${tutorID}`,
  )
}
