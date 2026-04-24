import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const createTeams = async (
  coursePhaseID: string,
  teamNames: string[],
  teamType?: 'standard' | 'field_bucket' | 'company_project',
  replaceExisting?: boolean,
  teamSizeConstraints?: Record<string, { lowerBound: number; upperBound: number }>,
): Promise<void> => {
  try {
    const teamsRequest = {
      teamNames: teamNames,
      ...(teamType ? { teamType } : {}),
      ...(replaceExisting ? { replaceExisting } : {}),
      ...(teamSizeConstraints ? { teamSizeConstraints } : {}),
    }
    await teamAllocationAxiosInstance.post(
      `/team-allocation/api/course_phase/${coursePhaseID}/team`,
      teamsRequest,
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
