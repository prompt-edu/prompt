import { axiosInstance } from '@tumaet/prompt-shared-state'
import { PostApplication } from '../../interfaces/application/postApplication'
import { PostApplicationResponse } from '@core/publicPages/application/interfaces/postApplicationConfirmation'

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
