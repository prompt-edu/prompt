import { notAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'
import { PostApplication } from '../../interfaces/application/postApplication'
import { PostApplicationResponse } from '@core/publicPages/application/interfaces/postApplicationConfirmation'

export const postNewApplicationExtern = async (
  phaseId: string,
  application: PostApplication,
): Promise<PostApplicationResponse> => {
  try {
    const response = await notAuthenticatedAxiosInstance.post(
      `/api/apply/${phaseId}`,
      application,
      {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      },
    )
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
