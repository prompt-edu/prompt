import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { ExportedApplicationAnswersResponse } from '../../managementConsole/applicationAdministration/interfaces/exportedApplicationAnswers'

export const getExportedApplicationAnswers = async (
  coursePhaseId: string,
): Promise<ExportedApplicationAnswersResponse> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseId}/exported-answers`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
