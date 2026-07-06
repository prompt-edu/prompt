import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'
import type { InterviewSlot } from '../interfaces/InterviewSlots'

export interface ParticipationStore {
  participations: CoursePhaseParticipationWithStudent[]
  interviewSlots: InterviewSlot[]
  setParticipations: (participations: CoursePhaseParticipationWithStudent[]) => void
  setInterviewSlots: (interviewSlots: InterviewSlot[]) => void
}

export const useParticipationStore = create<ParticipationStore>((set) => ({
  participations: [],
  interviewSlots: [],
  setParticipations: (participations) => set({ participations }),
  setInterviewSlots: (interviewSlots) => set({ interviewSlots }),
}))
