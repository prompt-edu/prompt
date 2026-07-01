import { SurveyTimeframe } from '../../interfaces/timeframe'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyTimeframe = async (coursePhaseID: string): Promise<SurveyTimeframe> => {
  try {
    const response = (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/survey/timeframe`,
      )
    ).data

    // Ensure dates are properly parsed from the API response
    return {
      ...response,
      surveyStart: response.surveyStart ? new Date(response.surveyStart) : null,
      surveyDeadline: response.surveyDeadline ? new Date(response.surveyDeadline) : null,
    }
  } catch (err) {
    console.error(err)
    throw err
  }
}
