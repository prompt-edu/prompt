import { create } from 'zustand'
import { CategoryWithCompetencies } from '../interfaces/category'

export interface CategoryStore {
  categories: CategoryWithCompetencies[]
  setCategories: (categories: CategoryWithCompetencies[]) => void
}

export const useCategoryStore = create<CategoryStore>((set) => ({
  categories: [],
  setCategories: (categories) => set({ categories }),
}))
