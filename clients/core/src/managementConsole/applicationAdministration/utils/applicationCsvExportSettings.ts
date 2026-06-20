export const APPLICATION_CSV_EXPORT_SETTINGS_KEY = 'applicationCsvExportSettings'

export type ApplicationCsvExportSettings = Record<string, boolean>

export const getApplicationCsvExportSettings = (
  restrictedData?: Record<string, unknown> | null,
): ApplicationCsvExportSettings => {
  const value = restrictedData?.[APPLICATION_CSV_EXPORT_SETTINGS_KEY]

  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return {}
  }

  return Object.fromEntries(
    Object.entries(value as Record<string, unknown>).filter(([, setting]) => {
      return typeof setting === 'boolean'
    }),
  ) as ApplicationCsvExportSettings
}

export const shouldExportQuestionToCsv = (
  settings: ApplicationCsvExportSettings,
  questionID: string,
): boolean => settings[questionID] !== false
