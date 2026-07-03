import type { AssessmentType } from '../../interfaces/assessmentType'
import type { CategoryWithCompetencies } from '../../interfaces/category'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const getAllCategoriesWithCompetencies = async (
  coursePhaseID: string,
  assessmentType: AssessmentType,
): Promise<CategoryWithCompetencies[]> => {
  try {
    return (
      await assessmentAxiosInstance.get(
        `assessment/api/course_phase/${coursePhaseID}/category/${assessmentType}/with-competencies`,
      )
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
