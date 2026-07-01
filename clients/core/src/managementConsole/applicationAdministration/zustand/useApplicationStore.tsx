import type { CoursePhaseWithMetaData } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'
import type { AdditionalScore } from '../interfaces/additionalScore/additionalScore'
import type { ApplicationParticipation } from '../interfaces/applicationParticipation'

interface ApplicationStoreState {
  additionalScores: AdditionalScore[]
  participations: ApplicationParticipation[]
  coursePhase: CoursePhaseWithMetaData
}

interface ApplicationStoreAction {
  setAdditionalScores: (additionalScores: AdditionalScore[]) => void
  setParticipations: (participations: ApplicationParticipation[]) => void
  setCoursePhase: (coursePhase: CoursePhaseWithMetaData) => void
}

export const useApplicationStore = create<ApplicationStoreState & ApplicationStoreAction>(
  (set) => ({
    additionalScores: [],
    participations: [],
    coursePhase: {} as CoursePhaseWithMetaData,
    setAdditionalScores: (additionalScores) => set({ additionalScores }),
    setParticipations: (participations) => set({ participations }),
    setCoursePhase: (coursePhase) => set({ coursePhase }),
  }),
)
