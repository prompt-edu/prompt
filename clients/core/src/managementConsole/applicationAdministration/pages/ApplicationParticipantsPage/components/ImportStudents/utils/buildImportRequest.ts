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

  const newQuestions: NewImportQuestion[] = questionHeaders.map((header) => {
    const maxLength = rows.reduce((max, row) => Math.max(max, (row[header] ?? '').length), 1)
    return {
      columnKey: header,
      title: header.trim(),
      allowedLength: Math.min(Math.max(maxLength, 1), MAX_ANSWER_LENGTH),
    }
  })

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
