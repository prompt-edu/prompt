import { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'

export interface CoursePhaseStore {
  coursePhase: CoursePhaseWithMetaData | undefined
  setCoursePhase: (coursePhase: CoursePhaseWithMetaData) => void
}

export const useCoursePhaseStore = create<CoursePhaseStore>((set) => ({
  coursePhase: undefined,
  setCoursePhase: (coursePhase) => set({ coursePhase }),
}))
