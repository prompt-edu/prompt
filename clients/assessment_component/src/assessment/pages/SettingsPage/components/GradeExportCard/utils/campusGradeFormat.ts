import { format, isValid, parseISO } from 'date-fns'

/**
 * How grades and dates are written into the CampusOnline CSV.
 *
 * Nothing in PROMPT can tell us what CampusOnline's import parser accepts, so
 * both formats are single constants: if a real import rejects `1.7` in favour
 * of `1,7`, changing `CAMPUS_GRADE_DECIMAL_SEPARATOR` is the entire fix.
 */

export const CAMPUS_GRADE_DECIMAL_SEPARATOR = '.'

/** German date format, the default CampusOnline uses. */
export const CAMPUS_DATE_FORMAT = 'dd.MM.yyyy'

const ISO_DATE_FORMAT = 'yyyy-MM-dd'
const ISO_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}/

export const formatCampusGrade = (grade: number): string =>
  grade.toFixed(1).replace('.', CAMPUS_GRADE_DECIMAL_SEPARATOR)

/**
 * Formats the completion timestamp for `DATE_OF_ASSESSMENT`. When the uploaded
 * file already carries dates, their shape wins over our default so we do not
 * mix two formats in one column.
 */
export const formatCampusAssessmentDate = (completedAt: string, sample?: string): string => {
  const completionDate = parseISO(completedAt)
  if (!isValid(completionDate)) {
    return ''
  }

  const dateFormat =
    sample && ISO_DATE_PATTERN.test(sample.trim()) ? ISO_DATE_FORMAT : CAMPUS_DATE_FORMAT

  return format(completionDate, dateFormat)
}
