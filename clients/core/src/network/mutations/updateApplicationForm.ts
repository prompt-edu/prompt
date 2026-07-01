import { axiosInstance } from '@tumaet/prompt-shared-state'
import { UpdateApplicationForm } from '../../managementConsole/applicationAdministration/interfaces/form/updateApplicationForm'

export const updateApplicationForm = async (
  coursePhaseID: string,
  applicationForm: UpdateApplicationForm,
): Promise<void> => {
  try {
    await axiosInstance.put(`/api/applications/${coursePhaseID}/form`, applicationForm, {
      headers: {
        'Content-Type': 'application/json-path+json',
      },
    })
  } catch (err) {
    console.error(err)
    throw err
  }
}
