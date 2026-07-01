import { Team } from '@tumaet/prompt-shared-state'
import { create } from 'zustand'

export interface TeamStore {
  teams: Team[]
  setTeams: (teams: Team[]) => void
}

export const useTeamStore = create<TeamStore>((set) => ({
  teams: [],
  setTeams: (teams) => set({ teams }),
}))
