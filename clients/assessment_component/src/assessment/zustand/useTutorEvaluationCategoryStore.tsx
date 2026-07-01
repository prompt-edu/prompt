import { create } from 'zustand'
import { CategoryWithCompetencies } from '../interfaces/category'
import { Competency } from '../interfaces/competency'

export interface TutorEvaluationCategoryStore {
  tutorEvaluationCategories: CategoryWithCompetencies[]
  allTutorEvaluationCompetencies: Competency[]
  setTutorEvaluationCategories: (categories: CategoryWithCompetencies[]) => void
}

export const useTutorEvaluationCategoryStore = create<TutorEvaluationCategoryStore>((set) => ({
  tutorEvaluationCategories: [],
  allTutorEvaluationCompetencies: [],
  setTutorEvaluationCategories: (categories) =>
    set({
      tutorEvaluationCategories: categories,
      allTutorEvaluationCompetencies: categories.flatMap((c) => c.competencies),
    }),
}))
