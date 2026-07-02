import { CategoryWithCompetencies } from '../interfaces/category'
import { create } from 'zustand'

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
