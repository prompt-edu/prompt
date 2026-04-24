import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'
import type { TeamAllocationProfile } from '../../interfaces/companyImport'

export const setAllocationProfile = async (
  coursePhaseID: string,
  profile: TeamAllocationProfile,
): Promise<void> => {
  await teamAllocationAxiosInstance.put(
    `/team-allocation/api/course_phase/${coursePhaseID}/survey/profile`,
    { profile },
  )
}
