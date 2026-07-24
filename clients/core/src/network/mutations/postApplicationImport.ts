import type { ImportApplicationRequest } from '@core/managementConsole/applicationAdministration/interfaces/import/importApplicationRequest'
import type { ImportResult } from '@core/managementConsole/applicationAdministration/interfaces/import/importResult'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const postApplicationImport = async (
  phaseId: string,
  request: ImportApplicationRequest,
): Promise<ImportResult> => {
  try {
    const response = await axiosInstance.post(`/api/applications/${phaseId}/import`, request)
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
