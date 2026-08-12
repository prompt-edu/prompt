import type { CampusColumnIndices } from '../interfaces/campusGradeExport'

/**
 * Header names as CampusOnline writes them. Comparison is case-insensitive
 * because the real header mixes casing styles (`GRADE` next to
 * `Studienplan_Version` and `Examiner`).
 */
export const CAMPUS_COLUMN_REGISTRATION_NUMBER = 'REGISTRATION_NUMBER'
export const CAMPUS_COLUMN_GRADE = 'GRADE'
export const CAMPUS_COLUMN_DATE_OF_ASSESSMENT = 'DATE_OF_ASSESSMENT'
export const CAMPUS_COLUMN_FAMILY_NAME = 'FAMILY_NAME_OF_STUDENT'
export const CAMPUS_COLUMN_FIRST_NAME = 'FIRST_NAME_OF_STUDENT'

/** Without these two the file cannot be matched or filled at all. */
const REQUIRED_CAMPUS_COLUMNS = [CAMPUS_COLUMN_REGISTRATION_NUMBER, CAMPUS_COLUMN_GRADE]

const normalizeHeader = (value: string): string => value.trim().toUpperCase()

const findColumnIndex = (headerRow: string[], columnName: string): number =>
  headerRow.findIndex((header) => normalizeHeader(header) === columnName)

/**
 * Resolves the columns we read from and write to. Throws a lecturer-readable
 * error naming the headers that were actually found when a required column is
 * missing, which is what happens when the wrong file gets uploaded.
 */
export const resolveCampusColumnIndices = (headerRow: string[]): CampusColumnIndices => {
  const missingColumns = REQUIRED_CAMPUS_COLUMNS.filter(
    (columnName) => findColumnIndex(headerRow, columnName) === -1,
  )

  if (missingColumns.length > 0) {
    throw new Error(
      `This does not look like a CampusOnline grade export. Missing ${
        missingColumns.length === 1 ? 'column' : 'columns'
      }: ${missingColumns.join(', ')}. Found: ${headerRow.join(', ')}`,
    )
  }

  return {
    registrationNumber: findColumnIndex(headerRow, CAMPUS_COLUMN_REGISTRATION_NUMBER),
    grade: findColumnIndex(headerRow, CAMPUS_COLUMN_GRADE),
    dateOfAssessment: findColumnIndex(headerRow, CAMPUS_COLUMN_DATE_OF_ASSESSMENT),
    familyName: findColumnIndex(headerRow, CAMPUS_COLUMN_FAMILY_NAME),
    firstName: findColumnIndex(headerRow, CAMPUS_COLUMN_FIRST_NAME),
  }
}
