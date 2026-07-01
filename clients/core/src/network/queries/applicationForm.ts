import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { ApplicationForm } from '../../managementConsole/applicationAdministration/interfaces/form/applicationForm'

export const getApplicationForm = async (coursePhaseId: string): Promise<ApplicationForm> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseId}/form`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
