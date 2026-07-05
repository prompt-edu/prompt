import { create } from 'zustand'
import type { CoursePhaseConfig } from '../interfaces/coursePhaseConfig'

export interface CoursePhaseConfigStore {
  coursePhaseConfig: CoursePhaseConfig | undefined
  setCoursePhaseConfig: (config: CoursePhaseConfig) => void
}

export const useCoursePhaseConfigStore = create<CoursePhaseConfigStore>((set) => ({
  coursePhaseConfig: undefined,
  setCoursePhaseConfig: (config) => set({ coursePhaseConfig: config }),
}))
