export type ApplicationMetaData = {
  applicationStartDate?: Date
  applicationEndDate?: Date
  externalStudentsAllowed?: boolean
  universityLoginAvailable?: boolean
  autoAccept?: boolean
  useCustomScores?: boolean
  applicationCsvExportSettings?: Record<string, boolean>
}
