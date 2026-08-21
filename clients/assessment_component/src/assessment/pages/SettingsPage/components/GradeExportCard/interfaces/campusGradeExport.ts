import type { PassStatus } from '@tumaet/prompt-shared-state'

/** Text encodings we can both read and write without corrupting the file. */
export type CampusCsvEncoding = 'utf-8' | 'utf-8-bom' | 'windows-1252'

export interface DecodedCsvFile {
  text: string
  encoding: CampusCsvEncoding
}

/**
 * Positions of the columns we care about within the header row. `-1` means the
 * column is absent, which is only tolerated for the optional ones.
 */
export interface CampusColumnIndices {
  registrationNumber: number
  grade: number
  dateOfAssessment: number
  familyName: number
  firstName: number
}

/**
 * A CampusOnline CSV kept in the exact shape it arrived in. `rows[0]` is the
 * header row; every other entry is a data row. Blank rows are preserved as
 * `['']` so the download can reproduce them.
 */
export interface ParsedCampusCsv {
  fileName: string
  rows: string[][]
  columnIndices: CampusColumnIndices
  delimiter: string
  newline: string
  hasTrailingNewline: boolean
  encoding: CampusCsvEncoding
}

/** One student of the phase together with their final grade, if they have one. */
export interface StudentGradeEntry {
  courseParticipationID: string
  firstName: string
  lastName: string
  matriculationNumber?: string
  grade: number | null
  completedAt: string | null
  passStatus: PassStatus
}

export type GradeMatchStrategy = 'registrationNumber' | 'name'

export interface FilledGradeRow {
  /** 1-based row number as a spreadsheet shows it: the header row is 1. */
  csvRowNumber: number
  registrationNumber: string
  studentName: string
  grade: string
  dateOfAssessment: string
  matchedBy: GradeMatchStrategy
  /** Set when the cell already held a value that we replaced. */
  overwrittenGrade?: string
}

export type SkippedRowReason = 'noGradeInPrompt' | 'unmatched' | 'ambiguous'

export interface SkippedCsvRow {
  csvRowNumber: number
  registrationNumber: string
  studentName: string
  reason: SkippedRowReason
}

export interface MissingParticipantRow {
  courseParticipationID: string
  studentName: string
  matriculationNumber?: string
  grade: number
}

export interface GradeFillReport {
  totalDataRows: number
  filled: FilledGradeRow[]
  skipped: SkippedCsvRow[]
  /** Graded students whose grade no CSV row consumed. Their grade is not exported. */
  gradedStudentsMissingFromCsv: MissingParticipantRow[]
  overwrittenCount: number
  nameFallbackCount: number
  /** Students marked as failed in PROMPT whose grade is not 5.0. */
  passStatusMismatchCount: number
}

export interface FilledCampusCsv {
  csv: ParsedCampusCsv
  report: GradeFillReport
}
