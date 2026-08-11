import Papa from 'papaparse'

export interface ParsedCsv {
  headers: string[]
  rows: Record<string, string>[]
}

/**
 * Parses a CSV file in header mode. `dynamicTyping` is disabled so leading zeros in matriculation
 * numbers and university logins are preserved as strings.
 */
export const parseImportCsv = (file: File): Promise<ParsedCsv> =>
  new Promise<ParsedCsv>((resolve, reject) => {
    Papa.parse<Record<string, string>>(file, {
      header: true,
      // 'greedy' drops lines that are empty or contain only delimiters (e.g. a trailing ",,,," that
      // Excel and many university exports append), which would otherwise parse as a blank student row.
      skipEmptyLines: 'greedy',
      dynamicTyping: false,
      transformHeader: (header) => header.trim(),
      complete: (result) => {
        const headers = (result.meta.fields ?? []).filter((h) => h.length > 0)
        if (headers.length === 0) {
          reject(new Error('No columns found in the CSV file.'))
          return
        }
        // A row with more cells than headers is a structural problem (an unescaped comma or an extra
        // column): papaparse assigns cells positionally, so every later column shifts by one and the
        // row would import with the wrong values instead of being rejected. Fail loudly instead.
        const mismatch = result.errors.find(
          (error) => error.type === 'FieldMismatch' || error.type === 'Quotes',
        )
        const extraRow = result.data.findIndex((row) => '__parsed_extra' in row)
        if (mismatch) {
          const line = typeof mismatch.row === 'number' ? mismatch.row + 2 : undefined
          reject(
            new Error(
              `The CSV could not be parsed cleanly${line ? ` (row ${line})` : ''}: ${mismatch.message}. ` +
                'Check for misaligned columns or an unescaped comma.',
            ),
          )
          return
        }
        if (extraRow >= 0) {
          reject(
            new Error(
              `Row ${extraRow + 2} has more values than there are columns. ` +
                'Check for an unescaped comma or an extra column.',
            ),
          )
          return
        }
        resolve({ headers, rows: result.data })
      },
      error: (error) => reject(error),
    })
  })
