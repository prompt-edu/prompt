import { create } from 'zustand'
import type { ScoreLevelWithParticipation } from '../interfaces/scoreLevelWithParticipation'

export interface ScoreLevelStore {
  scoreLevels: ScoreLevelWithParticipation[]
  setScoreLevels: (scoreLevels: ScoreLevelWithParticipation[]) => void
}

export const useScoreLevelStore = create<ScoreLevelStore>((set) => ({
  scoreLevels: [],
  setScoreLevels: (scoreLevels) => set({ scoreLevels }),
}))
