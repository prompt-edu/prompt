import { create } from 'zustand'
import { CategoryWithCompetencies } from '../interfaces/category'
import { Competency } from '../interfaces/competency'

export interface SelfEvaluationCategoryStore {
  selfEvaluationCategories: CategoryWithCompetencies[]
  allSelfEvaluationCompetencies: Competency[]
  setSelfEvaluationCategories: (categories: CategoryWithCompetencies[]) => void
}

export const useSelfEvaluationCategoryStore = create<SelfEvaluationCategoryStore>((set) => ({
  selfEvaluationCategories: [],
  allSelfEvaluationCompetencies: [],
  setSelfEvaluationCategories: (categories) =>
    set({
      selfEvaluationCategories: categories,
      allSelfEvaluationCompetencies: categories.flatMap((c) => c.competencies),
    }),
}))
