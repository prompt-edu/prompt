import { assessmentAxiosInstance } from '../assessmentServerConfig'
import type { CompetencyMapping } from '../mutations/createCompetencyMapping'

export const getAllCompetencyMappings = async (
  coursePhaseID: string,
): Promise<CompetencyMapping[]> => {
  try {
    const response = await assessmentAxiosInstance.get(
      `assessment/api/course_phase/${coursePhaseID}/competency-mappings`,
    )
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getCompetencyMappings = async (
  coursePhaseID: string,
  fromCompetencyID: string,
): Promise<CompetencyMapping[]> => {
  try {
    const response = await assessmentAxiosInstance.get(
      `assessment/api/course_phase/${coursePhaseID}/competency-mappings/from/${fromCompetencyID}`,
    )
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
