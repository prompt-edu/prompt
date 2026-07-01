import { create } from 'zustand'
import { ScoreLevelWithParticipation } from '../interfaces/scoreLevelWithParticipation'

export interface ScoreLevelStore {
  scoreLevels: ScoreLevelWithParticipation[]
  setScoreLevels: (scoreLevels: ScoreLevelWithParticipation[]) => void
}

export const useScoreLevelStore = create<ScoreLevelStore>((set) => ({
  scoreLevels: [],
  setScoreLevels: (scoreLevels) => set({ scoreLevels }),
}))
