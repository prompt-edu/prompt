export interface ImportRowResult {
  index: number
  universityLogin: string
  outcome: 'created' | 'updated'
  courseParticipationId?: string
}

export interface ImportResult {
  created: number
  updated: number
  rows: ImportRowResult[]
}
