import { create } from 'zustand'
import type { CategoryWithCompetencies } from '../interfaces/category'

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
