import type { SurveyResponse } from '../../interfaces/surveyResponse'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyOwnResponse = async (coursePhaseID: string): Promise<SurveyResponse> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/survey/answers`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
