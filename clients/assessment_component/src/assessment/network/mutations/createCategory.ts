import type { Category, CreateCategoryRequest } from '../../interfaces/category'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const createCategory = async (
  coursePhaseID: string,
  category: CreateCategoryRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.post<Category>(
      `assessment/api/course_phase/${coursePhaseID}/category`,
      category,
      {
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
