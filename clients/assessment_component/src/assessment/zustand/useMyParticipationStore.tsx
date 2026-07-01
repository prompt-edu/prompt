import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'

export interface MyParticipationStore {
  myParticipation: CoursePhaseParticipationWithStudent | undefined
  setMyParticipation: (myParticipation: CoursePhaseParticipationWithStudent | undefined) => void
}

export const useMyParticipationStore = create<MyParticipationStore>((set) => ({
  myParticipation: undefined,
  setMyParticipation: (myParticipation) => set({ myParticipation }),
}))
