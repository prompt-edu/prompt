import { selfTeamAllocationAxiosInstance } from '../selfTeamAllocationServerConfig'

export const updateTimeframe = async (
  coursePhaseID: string,
  startTime: Date,
  endTime: Date,
): Promise<void> => {
  const timeframeUpdate = {
    startTime: startTime,
    endTime: endTime,
  }
  await selfTeamAllocationAxiosInstance.put(
    `/self-team-allocation/api/course_phase/${coursePhaseID}/timeframe`,
    timeframeUpdate,
    {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    },
  )
}
