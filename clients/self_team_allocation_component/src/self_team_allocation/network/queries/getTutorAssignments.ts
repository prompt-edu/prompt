import type { TutorAssignment } from '../../interfaces/tutor'
import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const getTutorAssignments = async (coursePhaseID: string): Promise<TutorAssignment[]> => {
  return (
    await selfTeamAllocationAxiosInstance.get(
      `/self-team-allocation/api/course_phase/${coursePhaseID}/team/tutors`,
    )
  ).data
}
