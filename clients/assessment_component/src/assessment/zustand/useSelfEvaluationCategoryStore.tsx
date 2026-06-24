import { CategoryWithCompetencies } from '../interfaces/category'
import { create } from 'zustand'

export interface SelfEvaluationCategoryStore {
  selfEvaluationCategories: CategoryWithCompetencies[]
  setSelfEvaluationCategories: (categories: CategoryWithCompetencies[]) => void
}

export const useSelfEvaluationCategoryStore = create<SelfEvaluationCategoryStore>((set) => ({
  selfEvaluationCategories: [],
  setSelfEvaluationCategories: (categories) =>
    set({
      selfEvaluationCategories: categories,
    }),
}))
