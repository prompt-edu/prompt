import { Category, UpdateCategoryRequest } from '../../interfaces/category'
import { assessmentAxiosInstance } from '../assessmentServerConfig'

export const updateCategory = async (
  coursePhaseID: string,
  category: UpdateCategoryRequest,
): Promise<void> => {
  try {
    await assessmentAxiosInstance.put<Category>(
      `assessment/api/course_phase/${coursePhaseID}/category/${category.id}`,
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
