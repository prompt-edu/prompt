import type { PostApplicationResponse } from '@core/publicPages/application/interfaces/postApplicationConfirmation'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { PostApplication } from '../../interfaces/application/postApplication'

export const postNewApplicationAuthenticated = async (
  phaseId: string,
  application: PostApplication,
): Promise<PostApplicationResponse> => {
  try {
    const response = await axiosInstance.post(`/api/apply/authenticated/${phaseId}`, application, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
