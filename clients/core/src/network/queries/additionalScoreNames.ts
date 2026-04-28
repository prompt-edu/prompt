import { AdditionalScore } from '../../managementConsole/applicationAdministration/interfaces/additionalScore/additionalScore'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getAdditionalScoreNames = async (
  coursePhaseId: string,
): Promise<AdditionalScore[]> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseId}/score`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
