import Papa from 'papaparse'

import type { DecodedCsvFile, ParsedCampusCsv } from '../interfaces/campusGradeExport'
import { resolveCampusColumnIndices } from './campusCsvColumns'

/**
 * Reading and writing the CampusOnline CSV without disturbing anything we did
 * not explicitly change.
 *
 * The file is parsed positionally (`header: false`) rather than into objects.
 * Object mode would collapse duplicate header names into a single key and
 * reshape rows whose field count differs from the header, both of which lose
 * data on the way back out. Addressing the grade column by index keeps every
 * other column byte-for-byte as CampusOnline wrote it.
 *
 * One deliberate deviation from a byte-identical round trip: every field is
 * quoted on the way out, even if it was unquoted on the way in. That matches
 * CampusOnline's own all-quoted style and is valid CSV.
 */

export const MAX_CSV_BYTES = 5 * 1024 * 1024

const DEFAULT_NEWLINE = '\r\n'

const isBlankRow = (row: string[]): boolean => row.length === 1 && row[0] === ''

export const parseCampusCsv = (file: DecodedCsvFile, fileName: string): ParsedCampusCsv => {
  const result = Papa.parse<string[]>(file.text, {
    header: false,
    skipEmptyLines: false,
    dynamicTyping: false,
    delimitersToGuess: [';', ',', '\t'],
  })

  const quoteError = result.errors.find((error) => error.type === 'Quotes')
  if (quoteError) {
    throw new Error(`The CSV file could not be read: ${quoteError.message}`)
  }

  const rows = result.data
  if (rows.length === 0 || rows[0].length === 0) {
    throw new Error('The CSV file is empty.')
  }

  const columnIndices = resolveCampusColumnIndices(rows[0])

  if (rows.slice(1).every(isBlankRow)) {
    throw new Error('The CSV file contains a header row but no students.')
  }

  const newline = result.meta.linebreak || DEFAULT_NEWLINE

  return {
    fileName,
    rows,
    columnIndices,
    delimiter: result.meta.delimiter || ';',
    newline,
    hasTrailingNewline: file.text.endsWith(newline),
    encoding: file.encoding,
  }
}

const escapeCsvField = (value: string): string => `"${value.replace(/"/g, '""')}"`

/**
 * Rebuilds the file from `rows`, restoring the original delimiter, line ending
 * and trailing newline. Blank rows stay blank instead of turning into a single
 * quoted empty field.
 */
export const serializeCampusCsv = (csv: ParsedCampusCsv): string => {
  const lines = csv.rows.map((row) =>
    isBlankRow(row) ? '' : row.map(escapeCsvField).join(csv.delimiter),
  )

  return lines.join(csv.newline) + (csv.hasTrailingNewline ? csv.newline : '')
}

export const buildFilledCsvFileName = (originalFileName: string): string => {
  const baseName = originalFileName.replace(/\.csv$/i, '').trim()
  return baseName.length > 0 ? `${baseName}-with-grades.csv` : 'campus-online-grades.csv'
}
