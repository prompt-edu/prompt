import { create } from 'zustand'
import type { CategoryWithCompetencies } from '../interfaces/category'
import type { Competency } from '../interfaces/competency'

export interface PeerEvaluationCategoryStore {
  peerEvaluationCategories: CategoryWithCompetencies[]
  allPeerEvaluationCompetencies: Competency[]
  setPeerEvaluationCategories: (categories: CategoryWithCompetencies[]) => void
}

export const usePeerEvaluationCategoryStore = create<PeerEvaluationCategoryStore>((set) => ({
  peerEvaluationCategories: [],
  allPeerEvaluationCompetencies: [],
  setPeerEvaluationCategories: (categories) =>
    set({
      peerEvaluationCategories: categories,
      allPeerEvaluationCompetencies: categories.flatMap((c) => c.competencies),
    }),
}))
