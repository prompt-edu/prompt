import * as XLSX from 'xlsx'
import { UploadedStudent } from '../interfaces/UploadedStudent'
import { useMatchingStore } from '../zustand/useMatchingStore'

/**
 * This hook returns a function that, when called with a File,
 * parses the Excel and updates the Zustand store with the uploaded data.
 */
export const useUploadAndParseXLSX = () => {
  const { setUploadedData } = useMatchingStore()

  /**
   * Parses an Excel file expecting the following headers:
   *  'First name', 'Last name', 'Student number', 'Rank'
   *
   * Returns a Promise of the parsed data. Will throw an error if:
   *  - No sheet is found
   *  - The header row is missing or incorrect
   *  - Required data fields in rows are missing
   */
  const parseFileXLSX = async (file: File): Promise<void> => {
    try {
      // 1. Parse the workbook from the file
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
      const rawData = XLSX.utils.sheet_to_json(worksheet, { header: 1 })
      if (!Array.isArray(rawData) || rawData.length === 0) {
        throw new Error('No data found in the selected sheet.')
      }

      // 4. Separate the header row from the data rows
      const [headerRow, ...dataRows] = rawData
      if (!Array.isArray(headerRow)) {
        throw new Error('Invalid header row. Expected an array of column titles.')
      }

      // 5. Validate the header row matches the expected columns exactly
      const expectedHeaders = ['First name', 'Last name', 'Student number', 'Rank']
      const missingHeaders = expectedHeaders.filter((h) => !headerRow.includes(h))
      if (missingHeaders.length > 0) {
        throw new Error(`Missing headers: ${missingHeaders.join(', ')}`)
      }

      // Map each expected header to its column index
      const headerIndexMap: Record<string, number> = {}
      expectedHeaders.forEach((header) => {
        headerIndexMap[header] = headerRow.indexOf(header)
      })

      // 6. Parse each data row into the UploadedStudent interface
      const parsedData: UploadedStudent[] = dataRows
        .filter((row) => Array.isArray(row) && row.length > 0)
        .map((row, rowIndex) => {
          if (!Array.isArray(row)) {
            throw new Error(`Row ${rowIndex + 2} is not an array of values.`)
          }

          const firstNameCell = row[headerIndexMap['First name']]
          const lastNameCell = row[headerIndexMap['Last name']]
          const studentNumberCell = row[headerIndexMap['Student number']] ?? ''
          const rankCell = row[headerIndexMap['Rank']]

          if (!firstNameCell || !lastNameCell) {
            throw new Error(
              `Row ${rowIndex + 2} is missing required fields (First name, Last name, Student number).`,
            )
          }

          return {
            firstName: String(firstNameCell),
            lastName: String(lastNameCell),
            matriculationNumber: String(studentNumberCell),
            rank: rankCell ? String(rankCell) : undefined,
          }
        })

      // 7. Update Zustand store with parsed data
      setUploadedData(parsedData)
    } catch (error) {
      // Handle any parsing or data-structure errors
      console.error('Error uploading and parsing Excel file:', error)
      throw error // Re-throw to let caller handle it
    }
  }

  return { parseFileXLSX }
}
