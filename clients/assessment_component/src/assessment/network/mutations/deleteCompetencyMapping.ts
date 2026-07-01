import { assessmentAxiosInstance } from '../assessmentServerConfig'
import type { CompetencyMapping } from './createCompetencyMapping'

export const deleteCompetencyMapping = async (
  coursePhaseID: string,
  mapping: CompetencyMapping,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.delete(
      `assessment/api/course_phase/${coursePhaseID}/competency-mappings`,
      {
        data: mapping,
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )
  } catch (err) {
    console.error(err)
    throw err
  }
}
