import Papa from 'papaparse'
import type { UploadedStudent } from '../interfaces/UploadedStudent'
import { useMatchingStore } from '../zustand/useMatchingStore'

/**
 * This hook returns a function that, when called with a CSV File,
 * parses it and updates the Zustand store with the uploaded data.
 */
export const useUploadAndParseCSV = () => {
  const { setUploadedData } = useMatchingStore()

  /**
   * Parses a CSV file expecting the following headers:
   *  "Students first name", "Students last name", "Students matriculation number"
   *
   * Returns a Promise of the parsed data. Will throw an error if:
   *  - The file is empty or cannot be read
   *  - The header row is missing or incorrect
   *  - Required data fields in rows are missing
   */
  const parseFileCSV = async (file: File): Promise<void> => {
    try {
      // 1. Parse the CSV using Papa Parse in 'header' mode
      //    Setting `dynamicTyping: false` ensures all values are treated as strings
      //    so leading zeros are preserved.
      const result = await new Promise<Papa.ParseResult>((resolve, reject) => {
        Papa.parse(file, {
          header: true,
          skipEmptyLines: true,
          dynamicTyping: false, // ensures leading zeros remain as strings
          complete: resolve,
          error: reject,
        })
      })

      // 2. Basic validations
      if (!result?.data || result.data.length === 0) {
        throw new Error('No data found in the CSV file.')
      }

      // 3. Validate the header row
      const expectedHeaders = [
        'Students first name',
        'Students last name',
        'Students matriculation number',
      ]
      const missingHeaders = expectedHeaders.filter((h) => !result.meta.fields?.includes(h))
      if (missingHeaders.length > 0) {
        throw new Error(`Missing headers: ${missingHeaders.join(', ')}`)
      }

      // 4. Parse each row and map it to the UploadedStudent interface
      const parsedData: UploadedStudent[] = result.data.map(
        (row: Record<string, unknown>, rowIndex: number) => {
          const firstName = row['Students first name']
          const lastName = row['Students last name']
          const matriculationNumber = row['Students matriculation number']

          if (!firstName || !lastName || !matriculationNumber) {
            throw new Error(
              `Row ${rowIndex + 2} is missing required fields (first name, last name, or matriculation number).`,
            )
          }

          return {
            firstName: String(firstName),
            lastName: String(lastName),
            matriculationNumber: String(matriculationNumber),
          }
        },
      )

      // 5. Update Zustand store with parsed data
      setUploadedData(parsedData)
    } catch (error) {
      console.error('Error uploading and parsing CSV file:', error)
      throw error // Re-throw to let the caller handle it
    }
  }

  return { parseFileCSV }
}
