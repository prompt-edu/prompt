import * as XLSX from 'xlsx'
import type { UploadedStudent } from '../interfaces/UploadedStudent'
import { useMatchingStore } from '../zustand/useMatchingStore'

export const useDataDownload = () => {
  const { uploadedFile: file } = useMatchingStore()

  /**
   * Reads the first sheet from the Excel file (using the stored `file` from Zustand).
   * For each row in that sheet, if the `Student number` matches an entry from
   * `matchedStudents`, writes the `Rank` value from `matchedStudents`. Otherwise writes `-`.
   * Finally triggers a download of the modified file.
   */
  const generateAndDownloadFile = async (
    matchedStudents: UploadedStudent[],
    useScoreAsRank: boolean,
  ) => {
    if (!file) {
      throw new Error('No file uploaded.')
    }
    try {
      // 1. Read the workbook from the file
      const arrayBuffer = await file.arrayBuffer()
      const workbook = XLSX.read(arrayBuffer, { type: 'array' })
      const sheetName = workbook.SheetNames[0]
      if (!sheetName) {
        throw new Error('No sheets found in the Excel file.')
      }

      // 2. Access the first worksheet
      const worksheet = workbook.Sheets[sheetName]
      if (!worksheet) {
        throw new Error(`Worksheet "${sheetName}" is empty or could not be read.`)
      }

      // 3. Convert to row-based arrays (header: 1 => each row is an array)
      const rawData: any[][] = XLSX.utils.sheet_to_json(worksheet, { header: 1 })
      if (!Array.isArray(rawData) || rawData.length === 0) {
        throw new Error('No data found in the selected sheet.')
      }

      // 4. Separate the header row from the data rows
      const [headerRow, ...dataRows]: any[][] = rawData
      if (!Array.isArray(headerRow)) {
        throw new Error('Invalid header row. Expected an array of column titles.')
      }

      // 5. Validate the header row (makes sure we have the expected columns)
      const expectedHeaders = ['First name', 'Last name', 'Student number', 'Rank']
      const missingHeaders = expectedHeaders.filter((h) => !headerRow.includes(h))
      if (missingHeaders.length > 0) {
        throw new Error(`Missing headers: ${missingHeaders.join(', ')}`)
      }

      // 6. Get the indices for columns we need
      const studentNumberIndex = headerRow.indexOf('Student number')
      const rankIndex = headerRow.indexOf('Rank')

      // 7. Update each row’s rank if it matches an entry in `matchedStudents`
      dataRows
        .filter((row) => Array.isArray(row) && row.length > 0) // discard empty rows
        .forEach((row) => {
          // Make sure row is array-based
          if (!Array.isArray(row)) return

          // Pull out the current row’s student number
          const matricNum = String(row[studentNumberIndex] || '').trim()

          // Check whether we have a match
          const match = matchedStudents.find((student) => student.matriculationNumber === matricNum)

          // If found, update the rank cell; otherwise write '-'
          // if student has no application score, the score will be 0
          row[rankIndex] = match ? (useScoreAsRank ? match.rank : 1) || '' : '-'
        })

      // 8. Convert the updated rows back into a sheet
      const updatedWorksheet = XLSX.utils.aoa_to_sheet([headerRow, ...dataRows])

      // 9. Replace the old sheet in the workbook with our updated sheet
      workbook.Sheets[sheetName] = updatedWorksheet

      // 10. Download the modified workbook
      XLSX.writeFile(workbook, 'matched_students.xlsx')
    } catch (error) {
      console.error('Error generating and downloading Excel file:', error)
      throw error
    }
  }

  // Return the function so it can be used elsewhere
  return { generateAndDownloadFile }
}
