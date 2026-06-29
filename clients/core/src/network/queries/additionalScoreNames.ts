import { AdditionalScore } from '../../managementConsole/applicationAdministration/interfaces/additionalScore/additionalScore'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getAdditionalScoreNames = async (
  coursePhaseId: string,
): Promise<AdditionalScore[]> => {
  try {
    const { data } = await axiosInstance.get<AdditionalScore[] | null>(
      `/api/applications/${coursePhaseId}/score`,
    )
    return data ?? []
  } catch (err) {
    console.error(err)
    throw err
  }
}
