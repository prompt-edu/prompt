import type { PassStatus } from '@tumaet/prompt-shared-state'

export interface ApplicationAssessment {
  Score?: number | null
  restrictedData?: { [key: string]: any }
  passStatus?: PassStatus
}
