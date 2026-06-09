import { SurveyStatistics } from '../../interfaces/surveyStatistics'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyStatistics = async (coursePhaseID: string): Promise<SurveyStatistics> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/survey/statistics`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
