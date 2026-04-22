import { create } from 'zustand'
import { CoursePhaseParticipationWithStudent } from '@tumaet/prompt-shared-state'
import { DeveloperProfile } from '../interfaces/DeveloperProfile'

interface DevOpsChallengeStoreState {
  coursePhaseParticipation?: CoursePhaseParticipationWithStudent
  developerProfile?: DeveloperProfile
}

interface DevOpsChallengeStoreActions {
  setDeveloperProfile: (developerProfile?: DeveloperProfile) => void
  setCoursePhaseParticipation: (
    coursePhaseParticipation: CoursePhaseParticipationWithStudent,
  ) => void
}

export const useDevOpsChallengeStore = create<
  DevOpsChallengeStoreState & DevOpsChallengeStoreActions
>((set) => ({
  developerProfile: undefined,
  setDeveloperProfile: (developerProfile) => set({ developerProfile }),
  setCoursePhaseParticipation: (coursePhaseParticipation) => set({ coursePhaseParticipation }),
}))
