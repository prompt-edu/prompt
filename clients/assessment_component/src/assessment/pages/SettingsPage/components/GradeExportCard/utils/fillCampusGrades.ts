import { PassStatus } from '@tumaet/prompt-shared-state'

import type {
  FilledCampusCsv,
  FilledGradeRow,
  GradeMatchStrategy,
  MissingParticipantRow,
  ParsedCampusCsv,
  SkippedCsvRow,
  StudentGradeEntry,
} from '../interfaces/campusGradeExport'
import { formatCampusAssessmentDate, formatCampusGrade } from './campusGradeFormat'

const FAILING_GRADE = 5.0

/**
 * CampusOnline pads registration numbers with leading zeros in some exports and
 * not in others, so both sides are compared without them.
 */
export const normalizeRegistrationNumber = (value: string): string =>
  value.trim().replace(/^0+/, '')

/** Diacritic- and case-insensitive so `Müller` matches `MUELLER`'s neighbours. */
export const normalizeStudentName = (value: string): string =>
  value
    .trim()
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLocaleLowerCase()

const buildNameKey = (firstName: string, lastName: string): string =>
  `${normalizeStudentName(lastName)}|${normalizeStudentName(firstName)}`

const groupBy = (
  entries: StudentGradeEntry[],
  keyOf: (entry: StudentGradeEntry) => string | undefined,
): Map<string, StudentGradeEntry[]> => {
  const grouped = new Map<string, StudentGradeEntry[]>()

  for (const entry of entries) {
    const key = keyOf(entry)
    if (!key) continue

    const existing = grouped.get(key)
    if (existing) {
      existing.push(entry)
    } else {
      grouped.set(key, [entry])
    }
  }

  return grouped
}

const readCell = (row: string[], index: number): string =>
  index >= 0 && index < row.length ? row[index] : ''

/** Grows a short row so the target cell exists before we write into it. */
const writeCell = (row: string[], index: number, value: string): void => {
  while (row.length <= index) {
    row.push('')
  }
  row[index] = value
}

const isBlankRow = (row: string[]): boolean => row.length === 1 && row[0] === ''

interface RowMatch {
  entry: StudentGradeEntry
  matchedBy: GradeMatchStrategy
}

/**
 * Writes the PROMPT grades into a copy of the parsed CSV and reports what
 * happened to every row.
 *
 * Matching is by registration number first and by family + first name second.
 * A key that resolves to more than one student is never guessed at: grading the
 * wrong student is the worst failure this feature could produce, so those rows
 * are reported as ambiguous and left untouched.
 */
export const fillCampusGrades = (
  csv: ParsedCampusCsv,
  entries: StudentGradeEntry[],
): FilledCampusCsv => {
  const { registrationNumber, grade, dateOfAssessment, familyName, firstName } = csv.columnIndices
  const canMatchByName = familyName >= 0 && firstName >= 0

  const entriesByRegistrationNumber = groupBy(entries, (entry) =>
    entry.matriculationNumber ? normalizeRegistrationNumber(entry.matriculationNumber) : undefined,
  )
  const entriesByName = groupBy(entries, (entry) => buildNameKey(entry.firstName, entry.lastName))

  const rows = csv.rows.map((row) => [...row])
  const dateSample = csv.rows
    .slice(1)
    .map((row) => readCell(row, dateOfAssessment))
    .find((value) => value.trim() !== '')

  const filled: FilledGradeRow[] = []
  const skipped: SkippedCsvRow[] = []
  const consumedParticipationIDs = new Set<string>()
  let totalDataRows = 0
  let overwrittenCount = 0
  let nameFallbackCount = 0
  let passStatusMismatchCount = 0

  for (let rowIndex = 1; rowIndex < rows.length; rowIndex++) {
    const row = rows[rowIndex]
    if (isBlankRow(row)) continue

    totalDataRows++

    const csvRowNumber = rowIndex + 1
    const rowRegistrationNumber = readCell(row, registrationNumber).trim()
    const rowFamilyName = readCell(row, familyName).trim()
    const rowFirstName = readCell(row, firstName).trim()
    const studentName = [rowFirstName, rowFamilyName].filter(Boolean).join(' ')
    const reportRow = { csvRowNumber, registrationNumber: rowRegistrationNumber, studentName }

    const registrationCandidates = rowRegistrationNumber
      ? (entriesByRegistrationNumber.get(normalizeRegistrationNumber(rowRegistrationNumber)) ?? [])
      : []
    const nameCandidates =
      registrationCandidates.length === 0 && canMatchByName && rowFamilyName && rowFirstName
        ? (entriesByName.get(buildNameKey(rowFirstName, rowFamilyName)) ?? [])
        : []

    const candidates = registrationCandidates.length > 0 ? registrationCandidates : nameCandidates
    if (candidates.length > 1) {
      skipped.push({ ...reportRow, reason: 'ambiguous' })
      continue
    }
    if (candidates.length === 0) {
      skipped.push({ ...reportRow, reason: 'unmatched' })
      continue
    }

    const match: RowMatch = {
      entry: candidates[0],
      matchedBy: registrationCandidates.length > 0 ? 'registrationNumber' : 'name',
    }
    consumedParticipationIDs.add(match.entry.courseParticipationID)

    if (match.entry.grade === null || match.entry.completedAt === null) {
      skipped.push({ ...reportRow, reason: 'noGradeInPrompt' })
      continue
    }

    if (match.matchedBy === 'name') {
      nameFallbackCount++
    }
    if (match.entry.passStatus === PassStatus.FAILED && match.entry.grade !== FAILING_GRADE) {
      passStatusMismatchCount++
    }

    const previousGrade = readCell(row, grade).trim()
    const formattedGrade = formatCampusGrade(match.entry.grade)
    const formattedDate = formatCampusAssessmentDate(match.entry.completedAt, dateSample)

    writeCell(row, grade, formattedGrade)
    if (dateOfAssessment >= 0 && formattedDate) {
      writeCell(row, dateOfAssessment, formattedDate)
    }

    const replacedGrade = previousGrade !== '' && previousGrade !== formattedGrade
    if (replacedGrade) {
      overwrittenCount++
    }

    filled.push({
      ...reportRow,
      grade: formattedGrade,
      dateOfAssessment: formattedDate,
      matchedBy: match.matchedBy,
      ...(replacedGrade ? { overwrittenGrade: previousGrade } : {}),
    })
  }

  const gradedStudentsMissingFromCsv: MissingParticipantRow[] = entries.flatMap((entry) => {
    if (entry.grade === null || consumedParticipationIDs.has(entry.courseParticipationID)) {
      return []
    }

    return [
      {
        courseParticipationID: entry.courseParticipationID,
        studentName: `${entry.firstName} ${entry.lastName}`,
        matriculationNumber: entry.matriculationNumber,
        grade: entry.grade,
      },
    ]
  })

  return {
    csv: { ...csv, rows },
    report: {
      totalDataRows,
      filled,
      skipped,
      gradedStudentsMissingFromCsv,
      overwrittenCount,
      nameFallbackCount,
      passStatusMismatchCount,
    },
  }
}
