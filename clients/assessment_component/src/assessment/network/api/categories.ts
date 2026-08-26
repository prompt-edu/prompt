import type { AssessmentType } from '../../interfaces/assessmentType'
import type {
  Category,
  CategoryWithCompetencies,
  CreateCategoryRequest,
  UpdateCategoryRequest,
} from '../../interfaces/category'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/category`

export const categories = {
  listWithCompetencies: (
    coursePhaseID: string,
    assessmentType: AssessmentType,
  ): Promise<CategoryWithCompetencies[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/${assessmentType}/with-competencies`),

  create: (coursePhaseID: string, category: CreateCategoryRequest): Promise<Category> =>
    assessmentRequest.post<Category>(path(coursePhaseID), category),

  update: (coursePhaseID: string, category: UpdateCategoryRequest): Promise<void> =>
    assessmentRequest.put(`${path(coursePhaseID)}/${category.id}`, category),

  remove: (coursePhaseID: string, categoryID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${categoryID}`),
}
