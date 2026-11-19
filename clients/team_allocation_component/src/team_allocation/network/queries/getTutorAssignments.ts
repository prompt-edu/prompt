import type { TutorAssignment } from '../../interfaces/tutor'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getTutorAssignments = async (coursePhaseID: string): Promise<TutorAssignment[]> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/team/tutors`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
