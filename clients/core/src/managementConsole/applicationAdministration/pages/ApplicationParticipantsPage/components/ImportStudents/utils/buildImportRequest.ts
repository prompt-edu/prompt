import type {
  ImportApplicationRequest,
  ImportRow,
  ImportStudent,
  NewImportQuestion,
} from '@core/managementConsole/applicationAdministration/interfaces/import/importApplicationRequest'
import { Gender, type PassStatus, StudyDegree } from '@tumaet/prompt-shared-state'
import type { ColumnTarget } from './matchImportColumns'

const MAX_ANSWER_LENGTH = 2000

const matchEnumValue = (values: string[], input: string): string => {
  const found = values.find((value) => value.toLowerCase() === input.trim().toLowerCase())
  return found ?? ''
}

/**
 * Assembles the import request from the parsed CSV rows and the column mapping. Columns mapped to
 * "question" become text application questions; their per-row cell becomes that student's answer.
 */
export const buildImportRequest = (
  headers: string[],
  rows: Record<string, string>[],
  mapping: Record<string, ColumnTarget>,
  passStatus: PassStatus,
): ImportApplicationRequest => {
  const questionHeaders = headers.filter((header) => mapping[header] === 'question')

  // Imported questions get a fixed default length rather than one inferred from this file's longest
  // cell. The server freezes the length permanently and validates every later import against it, so a
  // column that happens to be empty in the first file must not brick re-imports with a length of 1.
  const newQuestions: NewImportQuestion[] = questionHeaders.map((header) => ({
    columnKey: header,
    title: header.trim(),
    allowedLength: MAX_ANSWER_LENGTH,
  }))

  const columnForTarget = (target: ColumnTarget): string | undefined =>
    headers.find((header) => mapping[header] === target)

  const importRows: ImportRow[] = rows.map((row) => {
    const valueFor = (target: ColumnTarget): string => {
      const column = columnForTarget(target)
      return column ? (row[column] ?? '').trim() : ''
    }

    const semesterRaw = valueFor('currentSemester')
    const semester = semesterRaw ? Number.parseInt(semesterRaw, 10) : Number.NaN

    const student: ImportStudent = {
      firstName: valueFor('firstName'),
      lastName: valueFor('lastName'),
      email: valueFor('email'),
      universityLogin: valueFor('universityLogin').toLowerCase(),
      matriculationNumber: valueFor('matriculationNumber'),
      hasUniversityAccount: true,
      gender: matchEnumValue(Object.values(Gender), valueFor('gender')),
      nationality: valueFor('nationality'),
      studyDegree: matchEnumValue(Object.values(StudyDegree), valueFor('studyDegree')),
      studyProgram: valueFor('studyProgram'),
      currentSemester: Number.isNaN(semester) ? null : semester,
    }

    const answers = questionHeaders
      .map((header) => ({ columnKey: header, answer: (row[header] ?? '').trim() }))
      .filter((answer) => answer.answer.length > 0)

    return { student, answers }
  })

  return { passStatus, newQuestions, rows: importRows }
}

export interface UnmatchedEnumValues {
  gender: string[]
  studyDegree: string[]
}

/**
 * Collects the distinct non-empty gender and study-degree values that do not match any enum option.
 * These are silently coerced to the server-side defaults (bachelor / prefer_not_to_say), so the
 * preview step surfaces them: a value the file did state is different from one it omitted.
 */
export const collectUnmatchedEnumValues = (
  headers: string[],
  rows: Record<string, string>[],
  mapping: Record<string, ColumnTarget>,
): UnmatchedEnumValues => {
  const columnForTarget = (target: ColumnTarget): string | undefined =>
    headers.find((header) => mapping[header] === target)

  const collect = (target: ColumnTarget, allowed: string[]): string[] => {
    const column = columnForTarget(target)
    if (!column) {
      return []
    }
    const unmatched = new Set<string>()
    for (const row of rows) {
      const raw = (row[column] ?? '').trim()
      if (raw.length > 0 && matchEnumValue(allowed, raw) === '') {
        unmatched.add(raw)
      }
    }
    return Array.from(unmatched)
  }

  return {
    gender: collect('gender', Object.values(Gender)),
    studyDegree: collect('studyDegree', Object.values(StudyDegree)),
  }
}
