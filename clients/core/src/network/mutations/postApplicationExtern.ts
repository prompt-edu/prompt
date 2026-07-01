import type { PostApplicationResponse } from '@core/publicPages/application/interfaces/postApplicationConfirmation'
import { notAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'
import type { PostApplication } from '../../interfaces/application/postApplication'

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
