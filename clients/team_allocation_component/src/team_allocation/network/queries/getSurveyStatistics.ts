import type { SurveyStatistics } from '../../interfaces/surveyStatistics'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyStatistics = async (coursePhaseID: string): Promise<SurveyStatistics> => {
  return (
    await teamAllocationAxiosInstance.get(
      `/team-allocation/api/course_phase/${coursePhaseID}/survey/statistics`,
    )
  ).data
}
