import type { SurveyResponse } from '../../interfaces/surveyResponse'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const postSurveyResponse = async (
  coursePhaseID: string,
  surveyResponse: SurveyResponse,
): Promise<void> => {
  try {
    await teamAllocationAxiosInstance.post(
      `/team-allocation/api/course_phase/${coursePhaseID}/survey/answers`,
      surveyResponse,
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
