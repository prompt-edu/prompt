import type { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'
import type { UploadedStudent } from '../interfaces/UploadedStudent'

export interface MatchingStore {
  participations: CoursePhaseParticipationWithStudent[]
  uploadedData: UploadedStudent[]
  uploadedFile: File | null
  setParticipations: (participations: CoursePhaseParticipationWithStudent[]) => void
  setUploadedData: (uploadedData: UploadedStudent[]) => void
  setUploadedFile: (file: File | null) => void
}

export const useMatchingStore = create<MatchingStore>((set) => ({
  participations: [],
  uploadedData: [],
  uploadedFile: null,
  setParticipations: (participations) => set({ participations }),
  setUploadedData: (uploadedData: UploadedStudent[]) => set({ uploadedData }),
  setUploadedFile: (uploadedFile: File | null) => set({ uploadedFile }),
}))
