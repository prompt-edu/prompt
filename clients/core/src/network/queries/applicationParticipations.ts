import { axiosInstance } from '@tumaet/prompt-shared-state'
import { ApplicationParticipation } from '../../managementConsole/applicationAdministration/interfaces/applicationParticipation'

export const getApplicationParticipations = async (
  coursePhaseID: string,
): Promise<ApplicationParticipation[]> => {
  try {
    return (await axiosInstance.get(`/api/applications/${coursePhaseID}/participations`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
