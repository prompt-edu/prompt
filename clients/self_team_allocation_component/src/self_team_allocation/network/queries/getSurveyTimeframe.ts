import type { Timeframe } from '../../interfaces/timeframe'
import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const getTimeframe = async (coursePhaseID: string): Promise<Timeframe> => {
  try {
    const response = (
      await selfTeamAllocationAxiosInstance.get(
        `/self-team-allocation/api/course_phase/${coursePhaseID}/timeframe`,
      )
    ).data

    // Ensure dates are properly parsed from the API response
    return {
      ...response,
      startTime: response.startTime ? new Date(response.startTime) : null,
      endTime: response.endTime ? new Date(response.endTime) : null,
    }
  } catch (err) {
    console.error(err)
    throw err
  }
}
