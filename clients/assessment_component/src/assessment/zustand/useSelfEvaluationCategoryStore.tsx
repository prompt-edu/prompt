import { create } from 'zustand'
import type { CategoryWithCompetencies } from '../interfaces/category'

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
