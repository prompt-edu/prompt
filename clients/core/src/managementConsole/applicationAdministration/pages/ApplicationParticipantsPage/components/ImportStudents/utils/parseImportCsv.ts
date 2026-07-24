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
      skipEmptyLines: true,
      dynamicTyping: false,
      transformHeader: (header) => header.trim(),
      complete: (result) => {
        const headers = (result.meta.fields ?? []).filter((h) => h.length > 0)
        if (headers.length === 0) {
          reject(new Error('No columns found in the CSV file.'))
          return
        }
        resolve({ headers, rows: result.data })
      },
      error: (error) => reject(error),
    })
  })
