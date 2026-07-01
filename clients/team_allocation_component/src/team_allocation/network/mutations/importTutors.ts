import type { Tutor } from '../../interfaces/tutor'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const importTutors = async (coursePhaseID: string, tutors: Tutor[]): Promise<void> => {
  try {
    await teamAllocationAxiosInstance.post(
      `team-allocation/api/course_phase/${coursePhaseID}/team/tutors`,
      tutors,
      {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
