import type { IndividualScore } from '../../../../../interfaces/additionalScore/individualScore'
import type { ApplicationParticipation } from '../../../../../interfaces/applicationParticipation'

export const matchStudents = (
  csvData: string[][],
  matchBy: string,
  matchColumn: string,
  scoreColumn: string,
  threshold: number | null,
  onError: (message: string) => void,
  applications: ApplicationParticipation[],
  setState: React.Dispatch<
    React.SetStateAction<{
      page: number
      additionalScores: IndividualScore[]
      unmatchedApplications: ApplicationParticipation[]
      rowsWithError: string[][]
      numberOfBelowThreshold: number | null
      open: boolean
    }>
  >,
) => {
  const headerRow = csvData[0]
  const matchColumnIndex = headerRow.indexOf(matchColumn)
  const scoreColumnIndex = headerRow.indexOf(scoreColumn)

  if (matchColumnIndex === -1 || scoreColumnIndex === -1) {
    onError('The CSV file does not contain the required columns.')
  }

  const matchedApplications: IndividualScore[] = []
  let belowThreshold: number = 0
  const unmatched: ApplicationParticipation[] = []
  const errorRows: string[][] = []

  errorRows.push(csvData[0]) // Add the header row to the error rows

  applications.forEach((app) => {
    const matchValue = app.student[matchBy as keyof typeof app.student]
    const matchedRow = csvData.find((row) => row[matchColumnIndex] === matchValue)

    if (matchedRow) {
      const commaSeparatedScores = matchedRow[scoreColumnIndex].replace(',', '.')
      const score = parseFloat(commaSeparatedScores)
      if (!isNaN(score)) {
        matchedApplications.push({
          courseParticipationID: app.courseParticipationID,
          score,
        })
        if (threshold !== null && score < threshold) {
          belowThreshold += 1
        }
      } else {
        errorRows.push(matchedRow)
        unmatched.push(app)
      }
    } else {
      unmatched.push(app)
    }
  })

  setState((prev) => ({
    ...prev,
    additionalScores: matchedApplications,
    unmatchedApplications: unmatched,
    rowsWithError: errorRows,
    numberOfBelowThreshold: threshold !== null ? belowThreshold : null,
  }))
}
