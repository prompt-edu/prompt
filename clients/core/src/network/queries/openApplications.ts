import { notAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'
import type { OpenApplicationDetails } from '../../interfaces/application/openApplicationDetails'

export const getAllOpenApplications = async (): Promise<OpenApplicationDetails[]> => {
  try {
    return (await notAuthenticatedAxiosInstance.get(`/api/apply`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
