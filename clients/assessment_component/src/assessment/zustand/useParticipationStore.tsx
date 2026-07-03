import { create } from 'zustand'

import type { AssessmentParticipationWithStudent } from '../interfaces/assessmentParticipationWithStudent'

export interface ParticipationStore {
  participations: AssessmentParticipationWithStudent[]
  setParticipations: (participations: AssessmentParticipationWithStudent[]) => void
}

export const useParticipationStore = create<ParticipationStore>((set) => ({
  participations: [],
  setParticipations: (participations) => set({ participations }),
}))
