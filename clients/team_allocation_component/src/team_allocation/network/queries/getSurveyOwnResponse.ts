import type { SurveyResponse } from '../../interfaces/surveyResponse'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyOwnResponse = async (coursePhaseID: string): Promise<SurveyResponse> => {
  return (
    await teamAllocationAxiosInstance.get(
      `/team-allocation/api/course_phase/${coursePhaseID}/survey/answers`,
    )
  ).data
}
