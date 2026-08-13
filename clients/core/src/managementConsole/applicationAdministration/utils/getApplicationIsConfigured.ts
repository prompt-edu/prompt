import type { ApplicationMetaData } from '../interfaces/applicationMetaData'

export function getIsApplicationConfigured(restrictedData: ApplicationMetaData | null): boolean {
  // In import mode the phase is configured as soon as the mode is selected — no
  // application timeframe is required.
  if (restrictedData?.applicationMode === 'import') {
    return true
  }
  return !!(
    restrictedData?.applicationStartDate &&
    restrictedData.applicationEndDate &&
    restrictedData.externalStudentsAllowed !== undefined
  )
}
