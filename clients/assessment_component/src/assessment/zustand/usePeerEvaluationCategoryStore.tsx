import { create } from 'zustand'
import type { CategoryWithCompetencies } from '../interfaces/category'
import type { Competency } from '../interfaces/competency'

export interface PeerEvaluationCategoryStore {
  peerEvaluationCategories: CategoryWithCompetencies[]
  setPeerEvaluationCategories: (categories: CategoryWithCompetencies[]) => void
}

export const usePeerEvaluationCategoryStore = create<PeerEvaluationCategoryStore>((set) => ({
  peerEvaluationCategories: [],
  setPeerEvaluationCategories: (categories) =>
    set({
      peerEvaluationCategories: categories,
    }),
}))
