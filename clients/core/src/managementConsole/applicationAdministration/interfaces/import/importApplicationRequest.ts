import type { PassStatus } from '@tumaet/prompt-shared-state'

export interface ImportStudent {
  firstName: string
  lastName: string
  email: string
  matriculationNumber: string
  universityLogin: string
  hasUniversityAccount: boolean
  gender: string
  nationality: string
  studyDegree: string
  studyProgram: string
  currentSemester: number | null
}

export interface NewImportQuestion {
  columnKey: string
  title: string
  allowedLength: number
}

export interface ImportAnswer {
  columnKey: string
  answer: string
}

export interface ImportRow {
  student: ImportStudent
  answers: ImportAnswer[]
}

export interface ImportApplicationRequest {
  passStatus: PassStatus
  newQuestions: NewImportQuestion[]
  rows: ImportRow[]
}
