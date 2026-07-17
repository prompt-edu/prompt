export interface ImportRowResult {
  index: number
  universityLogin: string
  outcome: 'created' | 'updated' | 'failed'
  reason?: string
  courseParticipationId?: string
}

export interface ImportResult {
  created: number
  updated: number
  failed: number
  rows: ImportRowResult[]
}
