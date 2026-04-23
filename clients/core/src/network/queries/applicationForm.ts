import { ApplicationForm } from '../../managementConsole/applicationAdministration/interfaces/form/applicationForm'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getApplicationForm = async (coursePhaseId: string): Promise<ApplicationForm> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseId}/form`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
