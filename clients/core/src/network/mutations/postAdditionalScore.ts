import { axiosInstance } from '@tumaet/prompt-shared-state'
import { AdditionalScoreUpload } from '../../managementConsole/applicationAdministration/interfaces/additionalScore/additionalScoreUpload'

export const postAdditionalScore = async (
  phaseId: string,
  additionalScore: AdditionalScoreUpload,
): Promise<void> => {
  try {
    return await axiosInstance.post(`/api/applications/${phaseId}/score`, additionalScore, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
