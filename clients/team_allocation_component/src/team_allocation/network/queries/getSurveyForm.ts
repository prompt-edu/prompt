import type { SurveyForm } from '../../interfaces/surveyForm'
import { teamAllocationAxiosInstance } from '../teamAllocationServerConfig'

export const getSurveyForm = async (coursePhaseID: string): Promise<SurveyForm | null> => {
  try {
    return (
      await teamAllocationAxiosInstance.get(
        `/team-allocation/api/course_phase/${coursePhaseID}/survey/form`,
      )
    ).data
  } catch (err: any) {
    console.error(err)
    if (err?.response?.status === 400) {
      // case that survey has not yet started
      return null
    } else {
      throw err
    }
  }
}
